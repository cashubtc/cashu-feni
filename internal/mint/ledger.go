package mint

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/gohumble/cashu-feni/internal/core"
	"github.com/gohumble/cashu-feni/internal/lightning"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/samber/lo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"math"
	"os"
	"path"
	"reflect"
	"strconv"
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
}

// NewLedger creates a new ledger and derives keys
func NewLedger(masterKey string) *Ledger {
	if _, err := os.Stat(Config.DbPath); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(Config.DbPath, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	orm, err := gorm.Open(sqlite.Open(path.Join(Config.DbPath, "database.db")),
		&gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, FullSaveAssociations: true})
	if err != nil {
		panic(err)
	}

	err = orm.AutoMigrate(&lightning.Invoice{})
	if err != nil {
		panic(err)
	}
	err = orm.AutoMigrate(&core.Proof{})
	if err != nil {
		panic(err)
	}
	err = orm.AutoMigrate(&core.Promise{})
	if err != nil {
		panic(err)
	}
	Database = orm

	l := &Ledger{
		masterKey:   masterKey,
		proofsUsed:  make([]string, 0),
		privateKeys: make(map[int64]*secp256k1.PrivateKey),
		publicKeys:  make(map[int64]*secp256k1.PublicKey),
	}
	l.deriveKeys()
	l.derivePublicKeys()
	lo.ForEach[core.Proof](getUsedProofs(orm), func(proof core.Proof, i int) {
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
func requestMint(c *lightning.Client, amount int64) (lightning.Invoice, error) {
	invoice, err := c.CreateInvoice(lightning.InvoiceParams{Amount: amount})
	if err != nil {
		return invoice, err
	}
	err = storeLightningInvoice(invoice)
	if err != nil {
		return invoice, err
	}
	return invoice, nil
}
func (l *Ledger) checkFees(pr string) (int64, error) {
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

// checkLightningInvoice will check payment status for pr
func (l *Ledger) checkLightningInvoice(c *lightning.Client, paymentHash string) (bool, error) {
	invoice, err := getLightningInvoice(paymentHash)
	if err != nil {
		return false, err
	}
	if invoice.Issued {
		return false, fmt.Errorf("tokens already issued for this invoice.")
	}
	payment, err := c.GetInvoiceStatus(paymentHash)
	if err != nil {
		return false, err
	}
	if payment.Paid {
		err = updateLightningInvoice(paymentHash, true)
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
func (l *Ledger) mint(c *lightning.Client, keys []*secp256k1.PublicKey, amounts []int64, pr string) ([]core.BlindedSignature, error) {
	if lightning.Config.Lnbits.Enabled {
		payed, err := l.checkLightningInvoice(c, pr)
		if err != nil {
			return nil, err
		}
		if !payed {
			return nil, fmt.Errorf("Lightning invoice not paid yet.")
		}
	}
	promises := make([]core.BlindedSignature, 0)
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
func (l *Ledger) generatePromise(amount int64, B_ *secp256k1.PublicKey) (core.BlindedSignature, error) {
	C_ := core.SecondStepBob(*B_, *l.privateKeys[amount])
	err := storePromise(core.Promise{Amount: amount, B_b: hex.EncodeToString(B_.SerializeCompressed()), C_c: hex.EncodeToString(C_.SerializeCompressed())})
	if err != nil {
		return core.BlindedSignature{}, err
	}
	return core.BlindedSignature{C_: hex.EncodeToString(C_.SerializeCompressed()), Amount: amount}, nil
}

// generatePromises will generate multiple promises and signatures
func (l *Ledger) generatePromises(amounts []int64, keys []*secp256k1.PublicKey) ([]core.BlindedSignature, error) {
	promises := make([]core.BlindedSignature, 0)
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
func (l *Ledger) verifyProof(proof core.Proof) (bool, error) {
	if !l.checkSpendable(proof) {
		return false, fmt.Errorf("tokens already spent. Secret: %s", proof.Secret)
	}
	secretKey := l.privateKeys[proof.Amount]
	pubKey, err := hex.DecodeString(proof.C)
	if err != nil {
		return false, err
	}
	C, err := secp256k1.ParsePubKey(pubKey)
	if err != nil {
		return false, err
	}
	return core.Verify(*secretKey, *C, proof.Secret), nil
}

// verifyOutputs verify output data
func verifyOutputs(total, amount int64, outputs []core.BlindedMessage) (bool, error) {
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
func verifyNoDuplicates(proofs []core.Proof, outputs []core.BlindedMessage) bool {
	secrets := make([]string, 0)
	for _, proof := range proofs {
		secrets = append(secrets, proof.Secret)
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
func (l *Ledger) checkSpendables(proofs []core.Proof) map[int]bool {
	result := make(map[int]bool, 0)
	for i, proof := range proofs {
		result[i] = l.checkSpendable(proof)
	}
	return result
}

// checkSpendable returns true if proof was not used before
func (l *Ledger) checkSpendable(proof core.Proof) bool {
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
		return 0, fmt.Errorf("invalid amount: %d", amount)
	}
	return amount, nil
}

// verifyEquationBalanced verify that equation is balanced.
func verifyEquationBalanced(proofs []core.Proof, outs []core.BlindedSignature) (bool, error) {
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
func (l *Ledger) invalidateProofs(proofs []core.Proof) error {
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
		err := invalidateProof(proof)
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
func (l *Ledger) melt(proofs []core.Proof, amount int64, invoice string) (status bool, preimage string, err error) {
	var total int64
	for _, proof := range proofs {
		// verify every proof and sum total amount
		ok, err := l.verifyProof(proof)
		if err != nil {
			return false, "", err
		}
		if !ok {
			return false, "", fmt.Errorf("could not verify proof")
		}
		total += proof.Amount
	}
	// decode invoice and use this amount instead of melt amount
	bolt, err := decodepay.Decodepay(invoice)
	amount = int64(math.Ceil(float64(bolt.MSatoshi / 1000)))
	fee, err := l.checkFees(invoice)
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
func (l *Ledger) split(proofs []core.Proof, amount int64, outputs []core.BlindedMessage) ([]core.BlindedSignature, []core.BlindedSignature, error) {
	// verifySplitAmount
	amount, err := verifySplitAmount(amount)
	if err != nil {
		return nil, nil, err
	}
	var total int64
	// verify proofs
	for _, proof := range proofs {
		vp, err := l.verifyProof(proof)
		if err != nil {
			return nil, nil, err
		}
		if !vp {
			return nil, nil, fmt.Errorf("invalid proof")
		}
		total += proof.Amount
	}
	// check for duplicates
	if !verifyNoDuplicates(proofs, outputs) {
		return nil, nil, fmt.Errorf("duplicates")
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
