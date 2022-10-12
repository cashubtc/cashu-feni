package mint

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/gohumble/cashu-feni/bitcoin"
	"github.com/gohumble/cashu-feni/cashu"
	"github.com/gohumble/cashu-feni/crypto"
	"github.com/gohumble/cashu-feni/db"
	"github.com/gohumble/cashu-feni/lightning"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/samber/lo"
	"math"
	"reflect"
	"strconv"
	"strings"
)

const MaxOrder = 64

// Ledger implements all functions for a cashu ledger.
type Ledger struct {
	// proofsUsed list of all proofs ever used
	proofsUsed []string
	// masterKey used to derive mints private key
	masterKey string
	// privateKeys map of amount:privateKey
	privateKeys map[int64]*secp256k1.PrivateKey
	// publicKeys map of amount:publicKey
	publicKeys map[int64]*secp256k1.PublicKey
	database   db.MintStorage
}

// NewLedger creates a new ledger and derives keys
func NewLedger(masterKey string, db db.MintStorage) *Ledger {

	l := &Ledger{
		masterKey:   masterKey,
		proofsUsed:  make([]string, 0),
		privateKeys: make(map[int64]*secp256k1.PrivateKey),
		publicKeys:  make(map[int64]*secp256k1.PublicKey),
		database:    db,
	}
	l.deriveKeys()
	l.derivePublicKeys()
	lo.ForEach[cashu.Proof](l.database.GetUsedProofs(), func(proof cashu.Proof, i int) {
		l.proofsUsed = append(l.proofsUsed, proof.Secret)
	})
	return l
}

// deriveKeys will generate private keys for the mint server
func (l *Ledger) deriveKeys() {
	for i := 0; i < MaxOrder; i++ {
		hasher := sha256.New()
		hasher.Write([]byte(l.masterKey + strconv.Itoa(i)))
		l.privateKeys[int64(math.Pow(2, float64(i)))] = secp256k1.PrivKeyFromBytes(hasher.Sum(nil)[:32])
	}
}

// derivePublicKeys will generate public keys for the mint server
func (l *Ledger) derivePublicKeys() {
	for amt, key := range l.privateKeys {
		l.publicKeys[amt] = key.PubKey()
	}
}

// requestMint will create and return the lightning invoice for a mint
func (l *Ledger) RequestMint(c *lightning.Client, amount int64) (lightning.Invoice, error) {
	invoice, err := c.CreateInvoice(lightning.InvoiceParams{Amount: amount})
	if err != nil {
		return invoice, err
	}
	err = l.database.StoreLightningInvoice(invoice)
	if err != nil {
		return invoice, err
	}
	return invoice, nil
}
func (l *Ledger) CheckFees(pr string) (int64, error) {
	decodedInvoice, err := decodepay.Decodepay(pr)
	if err != nil {
		return 0, err
	}
	amount := int64(math.Ceil(float64(decodedInvoice.MSatoshi / 1000)))
	// hack: check if it's internal, if it exists, it will return paid = False,
	// if id does not exist (not internal), it returns paid = None
	invoice, err := lightning.LnbitsClient.GetInvoiceStatus(decodedInvoice.PaymentHash)
	if err != nil {
		// invoice was not found. pay fees
		return lightning.FeeReserve(amount*1000, false), nil
	}
	internal := invoice.Paid == false
	return lightning.FeeReserve(amount*1000, internal), nil
}

// checkLightningInvoice will check the lightning invoice amount matches the outputs amount.
func (l *Ledger) checkLightningInvoice(c *lightning.Client, amounts []int64, paymentHash string) (bool, error) {
	invoice, err := l.database.GetLightningInvoice(paymentHash)
	if err != nil {
		return false, err
	}
	if invoice.Issued {
		return false, fmt.Errorf("tokens already issued for this invoice.")
	}
	payment, err := c.GetInvoiceStatus(paymentHash)
	// sum all amounts
	total := lo.SumBy[int64](amounts, func(amount int64) int64 {
		return amount
	})
	// validate total and invoice amount
	if total > invoice.Amount {
		return false, fmt.Errorf("requested amount too high: %d. Invoice amount: %d", total, invoice.Amount)
	}
	if err != nil {
		return false, err
	}
	if payment.Paid {
		err = l.database.UpdateLightningInvoice(paymentHash, true)
		if err != nil {
			return false, err
		}
	}
	return payment.Paid, err
}

// payLightningInvoice will pay pr using master wallet
func (l *Ledger) payLightningInvoice(c *lightning.Client, pr string, feesMsat int64) (lightning.LNbitsPayment, error) {
	invoice, err := lightning.Pay(lightning.PaymentParams{Out: true, Bolt11: pr, FeeLimitMSat: feesMsat}, c)
	if err != nil {
		return lightning.LNbitsPayment{}, err
	}
	return c.GetPaymentStatus(invoice.Hash)
}

// mint generates promises for keys. checks lightning invoice before creating promise.
func (l *Ledger) Mint(c *lightning.Client, keys []*secp256k1.PublicKey, amounts []int64, pr string) ([]cashu.BlindedSignature, error) {
	if lightning.Config.Lnbits.Enabled {
		paid, err := l.checkLightningInvoice(c, amounts, pr)
		if err != nil {
			return nil, err
		}
		if !paid {
			return nil, fmt.Errorf("Lightning invoice not paid yet.")
		}
	}
	promises := make([]cashu.BlindedSignature, 0)
	for i, key := range keys {
		sig, err := l.generatePromise(amounts[i], key)
		if err != nil {
			return nil, err
		}
		promises = append(promises, sig)
	}
	return promises, nil
}

// generatePromise will generate promise and signature for given amount using public key
func (l *Ledger) generatePromise(amount int64, B_ *secp256k1.PublicKey) (cashu.BlindedSignature, error) {
	C_ := crypto.SecondStepBob(*B_, *l.privateKeys[amount])
	err := l.database.StorePromise(cashu.Promise{Amount: amount, B_b: hex.EncodeToString(B_.SerializeCompressed()), C_c: hex.EncodeToString(C_.SerializeCompressed())})
	if err != nil {
		return cashu.BlindedSignature{}, err
	}
	return cashu.BlindedSignature{C_: hex.EncodeToString(C_.SerializeCompressed()), Amount: amount}, nil
}

// generatePromises will generate multiple promises and signatures
func (l *Ledger) generatePromises(amounts []int64, keys []*secp256k1.PublicKey) ([]cashu.BlindedSignature, error) {
	promises := make([]cashu.BlindedSignature, 0)
	for i, key := range keys {
		p, err := l.generatePromise(amounts[i], key)
		if err != nil {
			return nil, err
		}
		promises = append(promises, p)
	}
	return promises, nil
}

// verifyProof will verify proof
func (l *Ledger) verifyProof(proof cashu.Proof) error {
	if !l.checkSpendable(proof) {
		return fmt.Errorf("tokens already spent. Secret: %s", proof.Secret)
	}
	secretKey := l.privateKeys[proof.Amount]
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
		if len(strings.Split(proof.Secret, "P2SH:")) == 2 {
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
func verifyOutputs(total, amount int64, outputs []cashu.BlindedMessage) (bool, error) {
	fstAmt, sndAmt := total-amount, amount
	fstOutputs := amountSplit(fstAmt)
	sndOutputs := amountSplit(sndAmt)
	expected := append(fstOutputs, sndOutputs...)
	given := make([]int64, 0)
	for _, o := range outputs {
		given = append(given, o.Amount)
	}
	return reflect.DeepEqual(given, expected), nil
}

// verifyNoDuplicates checks if there are any duplicates
func verifyNoDuplicates(proofs []cashu.Proof, outputs []cashu.BlindedMessage) bool {
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
func (l *Ledger) CheckSpendables(proofs []cashu.Proof) map[int]bool {
	result := make(map[int]bool, 0)
	for i, proof := range proofs {
		result[i] = l.checkSpendable(proof)
	}
	return result
}

// checkSpendable returns true if proof was not used before
func (l *Ledger) checkSpendable(proof cashu.Proof) bool {
	_, found := lo.Find[string](l.proofsUsed, func(p string) bool {
		return p == proof.Secret
	})
	return !found
}

// amountSplit will convert amount into binary and return array with decimal binary values
func amountSplit(amount int64) []int64 {
	bin := reverse(strconv.FormatInt(int64(amount), 2))
	rv := make([]int64, 0)
	for i, b := range []byte(bin) {
		if b == 49 {
			rv = append(rv, int64(math.Pow(2, float64(i))))
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
func verifySplitAmount(amount int64) (int64, error) {
	return verifyAmount(amount)
}

// verifyAmount make sure that amount is bigger than zero and smaller than 2^MaxOrder
func verifyAmount(amount int64) (int64, error) {
	if amount < 0 || amount > 2^MaxOrder {
		return 0, fmt.Errorf("invalid split amount: %d", amount)
	}
	return amount, nil
}

// verifyEquationBalanced verify that equation is balanced.
func verifyEquationBalanced(proofs []cashu.Proof, outs []cashu.BlindedSignature) (bool, error) {
	var sumInputs int64
	var sumOutputs int64
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
func (l *Ledger) invalidateProofs(proofs []cashu.Proof) error {
	proofMsgs := make(map[string]struct{})
	// get unique proofs
	for _, proof := range proofs {
		proofMsgs[proof.Secret] = struct{}{}
	}
	// append to proofs used
	for pm := range proofMsgs {
		l.proofsUsed = append(l.proofsUsed, pm)
	}
	// make proofs used unique again...
	l.proofsUsed = lo.Uniq[string](l.proofsUsed)
	// invalidate all proofs
	for _, proof := range proofs {
		err := l.database.InvalidateProof(proof)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetPublicKeys will return current public keys for all amounts
func (l *Ledger) GetPublicKeys() map[int64]string {
	ret := make(map[int64]string, 0)
	for i, key := range l.publicKeys {
		ret[i] = hex.EncodeToString(key.SerializeCompressed())
	}
	return ret
}

/*
if not all( [self._verify_proof(p) for p in proofs]):
raise Exception ("could not verify proofs.")
*/
// melt will meld proofs
func (l *Ledger) Melt(proofs []cashu.Proof, amount int64, invoice string) (status bool, preimage string, err error) {
	var total int64
	for _, proof := range proofs {
		// verify every proof and sum total amount
		err = l.verifyProof(proof)
		if err != nil {
			return false, "", err
		}
		total += proof.Amount
	}
	// decode invoice and use this amount instead of melt amount
	bolt, err := decodepay.Decodepay(invoice)
	amount = int64(math.Ceil(float64(bolt.MSatoshi / 1000)))
	fee, err := l.CheckFees(invoice)
	if err != nil {
		return false, "", err
	}
	if !(total >= amount+(fee/1000)) {
		return false, "", fmt.Errorf("provided proofs not enough for Lightning payment")
	}
	payment, err := l.payLightningInvoice(lightning.LnbitsClient, invoice, fee)
	if err != nil {
		return false, "", err
	}
	if payment.Paid == true {
		err = l.invalidateProofs(proofs)
		if err != nil {
			return false, "", err
		}
	}
	return payment.Paid, payment.Preimage, nil
}

// split will split proofs. creates BlindedSignatures from BlindedMessages.
func (l *Ledger) Split(proofs []cashu.Proof, amount int64, outputs []cashu.BlindedMessage) ([]cashu.BlindedSignature, []cashu.BlindedSignature, error) {
	total := lo.SumBy[cashu.Proof](proofs, func(p cashu.Proof) int64 {
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
	if !verifyNoDuplicates(proofs, outputs) {
		return nil, nil, fmt.Errorf("duplicate proofs or promises.")
	}

	// verify proofs
	for _, proof := range proofs {
		err := l.verifyProof(proof)
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
	err = l.invalidateProofs(proofs)
	if err != nil {
		return nil, nil, err
	}
	// create first outputs and second outputs
	outsFts := amountSplit(total - amount)
	outsSnd := amountSplit(amount)
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
	fstPromise, err := l.generatePromises(outsFts, B_fst)
	if err != nil {
		return nil, nil, err
	}
	sendPromise, err := l.generatePromises(outsSnd, B_snd)
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

func verifySecretCriteria(proofs []cashu.Proof) error {
	for _, proof := range proofs {
		if proof.Secret == "" {
			return fmt.Errorf("secrets do not match criteria.")
		}
	}
	return nil
}
