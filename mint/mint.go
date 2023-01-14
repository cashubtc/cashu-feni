package mint

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/cashubtc/cashu-feni/bitcoin"
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/crypto"
	"github.com/cashubtc/cashu-feni/db"
	"github.com/cashubtc/cashu-feni/lightning"
	"github.com/cashubtc/cashu-feni/lightning/lnbits"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// Mint implements all functions for a cashu ledger.
type Mint struct {
	// proofsUsed list of all proofs ever used
	proofsUsed []string
	// masterKey used to derive mints private key
	masterKey string
	keySets   map[string]*crypto.KeySet
	KeySetId  string
	database  db.MintStorage
	client    lightning.Client
}

// New creates a new ledger and derives keys
func New(masterKey string, opt ...Options) *Mint {
	l := &Mint{
		masterKey:  masterKey,
		proofsUsed: make([]string, 0),
		keySets:    make(map[string]*crypto.KeySet, 0),
	}
	// apply ledger options
	for _, o := range opt {
		o(l)
	}
	if l.database != nil {
		p, err := l.database.GetUsedProofs()
		if err != nil {
			log.Warnf("could not load used proofs")
			return l
		}
		lo.ForEach[cashu.Proof](p, func(proof cashu.Proof, i int) {
			l.proofsUsed = append(l.proofsUsed, proof.Secret)
		})
	}

	return l
}
func (m Mint) setProofsPending(proofs []cashu.Proof) error {
	for _, proof := range proofs {
		if proof.Status == cashu.ProofStatusPending {
			return fmt.Errorf("proofs already pending.")
		}
		proof.Status = cashu.ProofStatusPending
		err := m.database.StoreProof(proof)
		if err != nil {
			return err
		}
	}
	return nil
}
func (m Mint) unsetProofsPending(proofs []cashu.Proof) {
	m.database.GetUsedProofs()
}
func (m Mint) LoadKeySet(id string) *crypto.KeySet {
	return m.keySets[id]
}

var couldNotCreateClient = fmt.Errorf("could not create lightning client. Please check your configuration")

// NewLightningClient will create a new lightning client implementation based on the ln config
func NewLightningClient() (lightning.Client, error) {
	cfg := lightning.Config.Lightning
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.Lnbits != nil {
		return lnbits.NewClient(cfg.Lnbits.AdminKey, cfg.Lnbits.Url), nil
	}
	return nil, couldNotCreateClient
}

type Options func(l *Mint)

func WithInitialKeySet(derivationPath string) Options {
	return func(l *Mint) {
		k := crypto.NewKeySet(l.masterKey, derivationPath)
		l.keySets[k.Id] = k
		l.KeySetId = k.Id
	}
}

func WithClient(client lightning.Client) Options {
	return func(l *Mint) {
		l.client = client
	}
}

func WithStorage(database db.MintStorage) Options {
	return func(l *Mint) {
		l.database = database
	}
}
func (m Mint) GetKeySetIds() []string {
	return lo.Keys(m.keySets)
}
func (m Mint) GetKeySet() []string {
	return lo.Keys(m.keySets)
}

// requestMint will create and return the lightning invoice for a mint
func (m *Mint) RequestMint(amount uint64) (lightning.Invoicer, error) {
	// signed amount is int64 (arm intel compatibility)
	signedAmount := int64(amount)
	if m.client == nil {
		invoice := lnbits.NewInvoice()
		invoice.SetAmount(signedAmount)
		invoice.SetHash("invalid")
		return invoice, nil
	}
	invoice, err := m.client.CreateInvoice(signedAmount, "requested feni mint")
	if err != nil {
		return invoice, err
	}
	err = m.database.StoreLightningInvoice(invoice)
	if err != nil {
		return invoice, err
	}
	return invoice, nil
}
func (m *Mint) CheckFees(pr string) (uint64, error) {
	decodedInvoice, err := decodepay.Decodepay(pr)
	if err != nil {
		return 0, err
	}
	amount := uint64(math.Ceil(float64(decodedInvoice.MSatoshi / 1000)))
	// hack: check if it's internal, if it exists, it will return paid = False,
	// if id does not exist (not internal), it returns paid = None
	invoice, err := m.client.InvoiceStatus(decodedInvoice.PaymentHash)
	if err != nil {
		// invoice was not found. pay fees
		return lightning.FeeReserve(amount*1000, false), nil
	}
	internal := invoice.IsPaid() == false
	return lightning.FeeReserve(amount*1000, internal), nil
}

// checkLightningInvoice will check the lightning invoice amount matches the outputs amount.
func (m *Mint) checkLightningInvoice(amounts []uint64, paymentHash string) (bool, error) {
	invoice, err := m.database.GetLightningInvoice(paymentHash)
	if err != nil {
		return false, err
	}
	if invoice.IsIssued() {
		return false, fmt.Errorf("tokens already issued for this invoice.")
	}
	payment, err := m.client.InvoiceStatus(paymentHash)
	if err != nil {
		return false, err
	}
	// sum all amounts
	total := lo.SumBy[uint64](amounts, func(amount uint64) uint64 {
		return amount
	})
	// validate total and invoice amount
	if total > uint64(invoice.GetAmount()) {
		return false, fmt.Errorf("requested amount too high: %d. Invoice amount: %d", total, invoice.GetAmount())
	}
	if err != nil {
		return false, err
	}
	if payment.IsPaid() {
		err = m.database.UpdateLightningInvoice(paymentHash, db.UpdateInvoicePaid(true), db.UpdateInvoiceWithIssued(true))
		if err != nil {
			// todo -- check if we rly want to return false here!
			return false, err
		}
	}
	return payment.IsPaid(), err
}

// payLightningInvoice will pay pr using master wallet
func (m *Mint) payLightningInvoice(pr string, feeLimitMSat uint64) (lightning.Payment, error) {
	invoice, err := m.client.Pay(pr)
	if err != nil {
		return lnbits.LNbitsPayment{}, err
	}
	return m.client.InvoiceStatus(invoice.GetHash())
}

func (m Mint) mint(messages cashu.BlindedMessages, pr string, keySet *crypto.KeySet) ([]cashu.BlindedSignature, error) {
	publicKeys := make([]*secp256k1.PublicKey, 0)
	var amounts []uint64
	for _, msg := range messages {
		amounts = append(amounts, msg.Amount)
		hkey := make([]byte, 0)
		hkey, err := hex.DecodeString(msg.B_)
		publicKey, err := secp256k1.ParsePubKey(hkey)
		if err != nil {
			return nil, err
		}
		publicKeys = append(publicKeys, publicKey)
	}
	// if the client is not nil, ledger is running on lightning
	if m.client != nil {
		paid, err := m.checkLightningInvoice(amounts, pr)
		if err != nil {
			return nil, err
		}
		if !paid {
			return nil, fmt.Errorf("Lightning invoice not paid yet.")
		}
	}
	promises := make([]cashu.BlindedSignature, 0)
	for i, key := range publicKeys {
		sig, err := m.generatePromise(amounts[i], keySet, key)
		if err != nil {
			return nil, err
		}
		promises = append(promises, sig)
	}
	return promises, nil
}

func (m Mint) Mint(messages cashu.BlindedMessages, pr string, keySet *crypto.KeySet) ([]cashu.BlindedSignature, error) {
	// mint generates promises for keys. checks lightning invoice before creating promise.
	return m.mint(messages, pr, keySet)
}
func (m Mint) MintWithoutKeySet(messages cashu.BlindedMessages, pr string) ([]cashu.BlindedSignature, error) {
	// mint generates promises for keys. checks lightning invoice before creating promise.
	return m.mint(messages, pr, m.LoadKeySet(m.KeySetId))
}

// generatePromise will generate promise and signature for given amount using public key
func (m *Mint) generatePromise(amount uint64, keySet *crypto.KeySet, B_ *secp256k1.PublicKey) (cashu.BlindedSignature, error) {
	C_ := crypto.SecondStepBob(*B_, *m.keySets[keySet.Id].PrivateKeys.GetKeyByAmount(uint64(amount)).Key)
	err := m.database.StorePromise(cashu.Promise{Amount: amount, B_b: hex.EncodeToString(B_.SerializeCompressed()), C_c: hex.EncodeToString(C_.SerializeCompressed())})
	if err != nil {
		return cashu.BlindedSignature{}, err
	}
	return cashu.BlindedSignature{C_: hex.EncodeToString(C_.SerializeCompressed()), Amount: amount}, nil
}

// generatePromises will generate multiple promises and signatures
func (m *Mint) generatePromises(amounts []uint64, keySet *crypto.KeySet, keys []*secp256k1.PublicKey) ([]cashu.BlindedSignature, error) {
	promises := make([]cashu.BlindedSignature, 0)
	for i, key := range keys {
		p, err := m.generatePromise(amounts[i], keySet, key)
		if err != nil {
			return nil, err
		}
		promises = append(promises, p)
	}
	return promises, nil
}

// verifyProof will verify proof
func (m *Mint) verifyProof(proof cashu.Proof) error {
	if !m.checkSpendable(proof) {
		return fmt.Errorf("tokens already spent. Secret: %s", proof.Secret)
	}
	secretKey := m.keySets[m.KeySetId].PrivateKeys.GetKeyByAmount(uint64(proof.Amount)).Key
	pubKey, err := hex.DecodeString(proof.C)
	if err != nil {
		return err
	}
	C, err := secp256k1.ParsePubKey(pubKey)
	if err != nil {
		return err
	}
	if crypto.Verify(*secretKey, *C, proof.Secret, crypto.HashToCurve) ||
		crypto.Verify(*secretKey, *C, proof.Secret, crypto.LegacyHashToCurve) {
		return nil
	}
	return fmt.Errorf("could not verify proofs.")
}

func verifyScript(proof cashu.Proof) (addr *btcutil.AddressScriptHash, err error) {
	if proof.Script == nil || proof.Script.Script == "" || proof.Script.Signature == "" {
		if cashu.IsPay2ScriptHash(proof.Secret) {
			return nil, fmt.Errorf("secret indicates a script but no script is present")
		} else {
			// secret indicates no script, so treat script as valid
			return nil, nil
		}
	}

	// decode payloads
	pubScriptKey, err := base64.URLEncoding.DecodeString(proof.Script.Script)
	if err != nil {
		return
	}
	sig, err := base64.URLEncoding.DecodeString(proof.Script.Signature)
	if err != nil {
		return
	}
	return bitcoin.VerifyScript(pubScriptKey, sig)
}

// verifyOutputs verify output data
func verifyOutputs(total, amount uint64, outputs []cashu.BlindedMessage) (bool, error) {
	fstAmt, sndAmt := total-amount, amount
	fstOutputs := AmountSplit(fstAmt)
	sndOutputs := AmountSplit(sndAmt)
	expected := append(fstOutputs, sndOutputs...)
	given := make([]uint64, 0)
	for _, o := range outputs {
		given = append(given, o.Amount)
	}
	return reflect.DeepEqual(given, expected), nil
}
func verifyNoDuplicateProofs(proofs []cashu.Proof) bool {
	secrets := make([]string, 0)
	for _, proof := range proofs {
		secrets = append(secrets, proof.Secret)
	}
	secretUniqueMap := make(map[string]struct{}, 0)
	for _, secret := range secrets {
		secretUniqueMap[secret] = struct{}{}
	}
	if len(secrets) != len(secretUniqueMap) {
		return false
	}
	return true
}

// verifyNoDuplicates checks if there are any duplicates
func verifyNoDuplicateOutputs(outputs []cashu.BlindedMessage) bool {
	B_s := make([]string, 0)
	for _, datum := range outputs {
		B_s = append(B_s, datum.B_)
	}
	B_sm := make(map[string]struct{})
	for _, b_ := range B_s {
		B_sm[b_] = struct{}{}
	}
	if len(B_s) != len(B_sm) {
		return false
	}
	return true
}

// checkSpendables checks multiple proofs
func (m *Mint) CheckSpendables(proofs []cashu.Proof) map[int]bool {
	result := make(map[int]bool, 0)
	for i, proof := range proofs {
		result[i] = m.checkSpendable(proof)
	}
	return result
}

// checkSpendable returns true if proof was not used before
func (m *Mint) checkSpendable(proof cashu.Proof) bool {
	_, found := lo.Find[string](m.proofsUsed, func(p string) bool {
		return p == proof.Secret
	})
	return !found
}

// AmountSplit will convert amount into binary and return array with decimal binary values
func AmountSplit(amount uint64) []uint64 {
	bin := reverse(strconv.FormatUint(amount, 2))
	rv := make([]uint64, 0)
	for i, b := range []byte(bin) {
		if b == 49 {
			rv = append(rv, uint64(math.Pow(2, float64(i))))
		}
	}
	return rv
}

// reverse string. used to reverse binary representation of amount
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// verifySplitAmount will verify amount
func verifySplitAmount(amount uint64) (uint64, error) {
	return verifyAmount(amount)
}

// verifyAmount make sure that amount is bigger than zero and smaller than 2^MaxOrder
func verifyAmount(amount uint64) (uint64, error) {
	if amount < 0 || amount > uint64(math.Pow(2, crypto.MaxOrder)) {
		return 0, fmt.Errorf("invalid split amount: %d", amount)
	}
	return amount, nil
}

// verifyEquationBalanced verify that equation is balanced.
func verifyEquationBalanced(proofs []cashu.Proof, outs []cashu.BlindedSignature) (bool, error) {
	var sumInputs uint64
	var sumOutputs uint64
	// sum proof amounts
	for _, proof := range proofs {
		in, err := verifyAmount(proof.Amount)
		if err != nil {
			return false, err
		}
		sumInputs += in
	}
	// sum output amounts
	for _, out := range outs {
		in, err := verifyAmount(out.Amount)
		if err != nil {
			return false, err
		}
		sumOutputs += in
	}
	// sum of outputs minus sum of inputs must be zero
	return sumOutputs-sumInputs == 0, nil
}

// invalidateProofs will invalidate multiple proofs at once by persisting them into proof table
func (m *Mint) invalidateProofs(proofs []cashu.Proof) error {
	proofMsgs := make(map[string]struct{})
	// get unique proofs
	for _, proof := range proofs {
		proofMsgs[proof.Secret] = struct{}{}
	}
	// append to proofs used
	for pm := range proofMsgs {
		m.proofsUsed = append(m.proofsUsed, pm)
	}
	// make proofs used unique again...
	m.proofsUsed = lo.Uniq[string](m.proofsUsed)
	// invalidate all proofs
	for _, proof := range proofs {
		err := m.database.StoreProof(proof)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetPublicKeys will return current public keys for all amounts
func (m *Mint) GetPublicKeys() map[uint64]string {
	return crypto.GetKeySetPublicKeys(m.keySets[m.KeySetId])
}

/*
if not all( [self._verify_proof(p) for p in proofs]):
raise Exception ("could not verify proofs.")
*/
// melt will meld proofs
func (m *Mint) Melt(proofs []cashu.Proof, invoice string) (payment lightning.Payment, err error) {
	var total uint64
	for _, proof := range proofs {
		// verify every proof and sum total amount
		err = m.verifyProof(proof)
		if err != nil {
			return nil, err
		}
		total += proof.Amount
	}
	// decode invoice and use this amount instead of melt amount
	bolt, err := decodepay.Decodepay(invoice)
	amount := uint64(math.Ceil(float64(bolt.MSatoshi / 1000)))
	fee, err := m.CheckFees(invoice)
	if err != nil {
		return nil, err
	}
	if !(total >= amount+(fee/1000)) {
		return nil, fmt.Errorf("provided proofs not enough for Lightning payment")
	}
	payment, err = m.payLightningInvoice(invoice, fee)
	if err != nil {
		return nil, err
	}
	if payment.IsPaid() == true {
		err = m.invalidateProofs(proofs)
		if err != nil {
			return nil, err
		}
	}
	return payment, nil
}

// split will split proofs. creates BlindedSignatures from BlindedMessages.
func (m *Mint) Split(proofs []cashu.Proof, amount uint64, outputs []cashu.BlindedMessage, keySet *crypto.KeySet) ([]cashu.BlindedSignature, []cashu.BlindedSignature, error) {
	total := lo.SumBy[cashu.Proof](proofs, func(p cashu.Proof) uint64 {
		return p.Amount
	})
	if amount > total {
		return nil, nil, fmt.Errorf("split amount is higher than the total sum.")
	}
	// verifySplitAmount
	amount, err := verifySplitAmount(amount)

	if err != nil {
		return nil, nil, err
	}
	// verify script
	for _, proof := range proofs {
		addr, err := verifyScript(proof)
		if err != nil {
			// Python test adoption
			// this should be removed in future versions
			switch err.Error() {
			case "pay to script hash is not push only":
				return nil, nil, fmt.Errorf("('%v', EvalScriptError('EvalScript: OP_RETURN called'))", fmt.Errorf("Script evaluation failed:"))
			case "false stack entry at end of script execution":
				return nil, nil, fmt.Errorf("('%v', VerifyScriptError('scriptPubKey returned false'))", fmt.Errorf("Script verification failed:"))
			}
			return nil, nil, err
		}
		if addr != nil {
			ss := strings.Split(proof.Secret, ":")
			if len(ss) != 3 {
				return nil, nil, fmt.Errorf("script verification failed.")
			}
			addrs := addr.String()
			if ss[1] != addrs {
				return nil, nil, fmt.Errorf("script verification failed.")
			}
		}
	}
	// _verify_secret_criteria
	if err = verifySecretCriteria(proofs); err != nil {
		return nil, nil, fmt.Errorf("no secret in proof.")
	}
	// check for duplicates
	if !verifyNoDuplicateProofs(proofs) {
		return nil, nil, fmt.Errorf("duplicate proofs.")
	}
	if !verifyNoDuplicateOutputs(outputs) {
		return nil, nil, fmt.Errorf("duplicate outputs.")
	}
	// verify proofs
	for _, proof := range proofs {
		err := m.verifyProof(proof)
		if err != nil {
			return nil, nil, err
		}
	}

	// check outputs
	_, err = verifyOutputs(total, amount, outputs)
	if err != nil {
		return nil, nil, err
	}
	// invalidate proofs
	err = m.invalidateProofs(proofs)
	if err != nil {
		return nil, nil, err
	}
	// create first outputs and second outputs
	outsFts := AmountSplit(total - amount)
	outsSnd := AmountSplit(amount)
	B_fst := make([]*secp256k1.PublicKey, 0)
	B_snd := make([]*secp256k1.PublicKey, 0)
	for _, data := range outputs[:len(outsFts)] {
		b, err := hex.DecodeString(data.B_)
		if err != nil {
			return nil, nil, err
		}
		key, err := secp256k1.ParsePubKey(b)
		if err != nil {
			return nil, nil, err
		}
		B_fst = append(B_fst, key)
	}

	for _, data := range outputs[len(outsFts):] {
		b, err := hex.DecodeString(data.B_)
		if err != nil {
			return nil, nil, err
		}
		key, err := secp256k1.ParsePubKey(b)
		if err != nil {
			return nil, nil, err
		}
		B_snd = append(B_snd, key)
	}
	// create promises for outputs
	fstPromise, err := m.generatePromises(outsFts, keySet, B_fst)
	if err != nil {
		return nil, nil, err
	}
	sendPromise, err := m.generatePromises(outsSnd, keySet, B_snd)
	if err != nil {
		return nil, nil, err
	}
	outs := append(fstPromise, sendPromise...)
	// check eq is balanced
	_, err = verifyEquationBalanced(proofs, outs)
	if err != nil {
		return nil, nil, err
	}
	return fstPromise, sendPromise, nil
}

// verifySecretCriteria verifies that a secret is present and is not too long (DOS prevention).
func verifySecretCriteria(proofs []cashu.Proof) error {
	for _, proof := range proofs {
		if proof.Secret == "" {
			return fmt.Errorf("secrets do not match criteria.")
		}
		if len(proof.Secret) > 64 {
			return fmt.Errorf("secret too long.")
		}
	}
	return nil
}
