package feni

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/gohumble/cashu-feni/api"
	"github.com/gohumble/cashu-feni/cashu"
	"github.com/gohumble/cashu-feni/crypto"
	"github.com/gohumble/cashu-feni/mint"
	"github.com/google/uuid"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"strings"
	"time"
)

type Pair[T, U any] struct {
	First  T
	Second U
}

func Zip[T, U any](ts []T, us []U) []Pair[T, U] {
	if len(ts) != len(us) {
		panic("slices have different length")
	}
	pairs := make([]Pair[T, U], len(ts))
	for i := 0; i < len(ts); i++ {
		pairs[i] = Pair[T, U]{ts[i], us[i]}
	}
	return pairs
}

type MintWallet struct {
	keys   map[int64]*secp256k1.PublicKey // current public keys from mint server
	keySet string                         // current keySet id from mint server.
	proofs []cashu.Proof
}

var Wallet MintWallet

func init() {

	rand.Seed(time.Now().UnixNano())

	Wallet = MintWallet{proofs: make([]cashu.Proof, 0), keys: make(map[int64]*secp256k1.PublicKey)}
	mintServerPublickeys, err := WalletClient.Keys()
	if err != nil {
		panic(err)
	}
	Wallet.keys = mintServerPublickeys
	keySet, err := WalletClient.KeySets()
	if err != nil {
		panic(err)
	}
	Wallet.keySet = keySet.KeySets[len(keySet.KeySets)-1]
}

func constructOutputs(amounts []int64, secrets []string) (api.MintRequest, []*secp256k1.PrivateKey) {
	payloads := api.MintRequest{BlindedMessages: make(cashu.BlindedMessages, 0)}
	privateKeys := make([]*secp256k1.PrivateKey, 0)
	for _, pair := range Zip[string, int64](secrets, amounts) {
		r, err := secp256k1.GeneratePrivateKey()
		if err != nil {
			panic(err)
		}
		pub, r := crypto.FirstStepAlice(pair.First, r)
		privateKeys = append(privateKeys, r)
		payloads.BlindedMessages = append(payloads.BlindedMessages,
			cashu.BlindedMessage{Amount: pair.Second, B_: fmt.Sprintf("%x", pub.SerializeCompressed())})
	}
	return payloads, privateKeys
}

func (w MintWallet) checkUsedSecrets(amounts []int64, secrets []string) error {
	proofs := storage.ProofsUsed(secrets)
	if len(proofs) > 0 {
		return fmt.Errorf("proofs already used")
	}
	return nil
}

func (w MintWallet) mint(amounts []int64, paymentHash string) []cashu.Proof {
	secrets := make([]string, 0)
	for range amounts {
		secrets = append(secrets, generateSecret())
	}
	err := w.checkUsedSecrets(amounts, secrets)
	if err != nil {
		panic(err)
	}
	req, privateKeys := constructOutputs(amounts, secrets)
	blindedSignatures, err := WalletClient.Mint(req, paymentHash)
	if err != nil {
		panic(err)
	}
	return w.constructProofs(*blindedSignatures, secrets, privateKeys)
}

func (w MintWallet) constructProofs(promises []cashu.BlindedSignature, secrets []string, privateKeys []*secp256k1.PrivateKey) []cashu.Proof {
	proofs := make([]cashu.Proof, 0)
	for i, promise := range promises {
		h, err := hex.DecodeString(promise.C_)
		if err != nil {
			return nil
		}
		C_, err := secp256k1.ParsePubKey(h)
		if err != nil {
			return nil
		}
		C := crypto.ThirdStepAlice(*C_, *privateKeys[i], *w.keys[promise.Amount])
		proofs = append(proofs, cashu.Proof{
			Id:     w.keySet,
			Amount: promise.Amount,
			C:      fmt.Sprintf("%x", C.SerializeCompressed()),
			Secret: secrets[i],
		})
	}
	return proofs
}

type KeySetBalance struct {
	Balance   int64
	Available int64
}
type Balance map[string]KeySetBalance

func (w MintWallet) balancePerKeySet() Balance {
	b := Balance{}
	for _, proof := range w.proofs {
		proofBalance, ok := b[proof.Id]
		if ok {
			proofBalance.Balance += proof.Amount
		} else {
			proofBalance = KeySetBalance{
				Balance: proof.Amount,
			}
		}
		if !proof.Reserved {
			proofBalance.Available += proof.Amount
		}
		b[proof.Id] = proofBalance
	}
	return b
}
func generateSecrets(secret string, n int) []string {
	secrets := make([]string, 0)
	var generator func(i int)
	if len(strings.Split(secret, "P2SH:")) == 2 {
		generator = func(i int) {
			secrets = append(secrets, fmt.Sprintf("%s:%s", secret, generateSecret()))
		}
	} else {
		generator = func(i int) {
			secrets = append(secrets, fmt.Sprintf("%d:%s", i, secret))
		}
	}
	for i := 0; i < n; i++ {
		generator(i)
	}
	return secrets

}
func generateSecret() string {
	return base64.RawURLEncoding.EncodeToString([]byte(RandStringRunes(16)))
}
func getUnusedLocks(addressSplit string) ([]cashu.P2SHScript, error) {
	return storage.GetScripts(addressSplit)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (w MintWallet) GetSpendableProofs() []cashu.Proof {
	spendable := make([]cashu.Proof, 0)
	for _, proof := range w.proofs {
		if proof.Reserved {
			continue
		}
		if proof.Id != w.keySet {
			continue
		}
		spendable = append(spendable, proof)
	}
	return spendable
}

func (w MintWallet) SplitToSend(amount int64, scndSecret string, setReserved bool) (keep []cashu.Proof, send []cashu.Proof, err error) {
	spendableProofs := w.GetSpendableProofs()
	if SumProofs(spendableProofs) < amount {
		return nil, nil, fmt.Errorf("balance to low.")
	}
	keepProofs, SendProofs, err := w.Split(spendableProofs, amount, scndSecret)
	if err != nil {
		return nil, nil, err
	}
	if setReserved {
		err = w.setReserved(SendProofs, true)
		if err != nil {
			return nil, nil, err
		}
	}
	return keepProofs, SendProofs, err
}
func (w MintWallet) setReserved(p []cashu.Proof, reserved bool) error {
	for _, proof := range p {
		proof.Reserved = reserved
		proof.SendId = uuid.New()
		err := storage.StoreProof(proof)
		if err != nil {
			return err
		}
	}
	return nil
}
func (w MintWallet) redeem(proofs []cashu.Proof, scndScript, scndSignature string) (keep []cashu.Proof, send []cashu.Proof, err error) {
	if scndScript != "" && scndSignature != "" {
		log.Infof("Unlock script: %s", scndScript)
		for i := range proofs {
			proofs[i].Script = &cashu.P2SHScript{
				Script:    scndScript,
				Signature: scndSignature}
		}
	}
	return w.Split(proofs, SumProofs(proofs), "")
}
func (w *MintWallet) Split(proofs []cashu.Proof, amount int64, scndSecret string) (keep []cashu.Proof, send []cashu.Proof, err error) {
	if len(proofs) < 0 {
		return nil, nil, fmt.Errorf("no proofs provided.")
	}
	frstProofs, scndProofs, err := w.split(proofs, amount, scndSecret)
	if err != nil {
		return nil, nil, err
	}
	if len(frstProofs) == 0 && len(scndProofs) == 0 {
		return nil, nil, fmt.Errorf("received no splits.")
	}
	usedSecrets := make([]string, 0)
	for _, proof := range proofs {
		usedSecrets = append(usedSecrets, proof.Secret)
	}
	w.proofs = lo.Filter[cashu.Proof](w.proofs, func(p cashu.Proof, i int) bool {
		_, found := lo.Find[string](usedSecrets, func(secret string) bool {
			return secret == p.Secret
		})
		return !found
	})
	w.proofs = append(w.proofs, frstProofs...)
	w.proofs = append(w.proofs, scndProofs...)
	err = storeProofs(append(frstProofs, scndProofs...))
	if err != nil {
		return nil, nil, err
	}
	for _, proof := range proofs {
		err = invalidate(proof)
		if err != nil {
			return nil, nil, err
		}
	}
	return frstProofs, scndProofs, nil
}
func (w MintWallet) split(proofs []cashu.Proof, amount int64, scndSecret string) (keep []cashu.Proof, send []cashu.Proof, err error) {

	total := SumProofs(proofs)
	frstAmt := total - amount
	scndAmt := amount
	frstOutputs := mint.AmountSplit(frstAmt)
	scndOutputs := mint.AmountSplit(scndAmt)
	amounts := append(frstOutputs, scndOutputs...)
	secrets := make([]string, 0)
	if scndSecret == "" {
		for range amounts {
			secrets = append(secrets, generateSecret())
		}
	} else {
		scndSecrets := generateSecrets(scndSecret, len(scndOutputs))
		if len(scndSecrets) != len(scndOutputs) {
			return nil, nil, fmt.Errorf("number of scnd_secrets does not match number of outputs.")
		}
		for range frstOutputs {
			secrets = append(secrets, generateSecret())
		}
		secrets = append(secrets, scndSecrets...)
	}
	if len(secrets) != len(amounts) {
		return nil, nil, fmt.Errorf("number of secrets does not match number of outputs")
	}
	// TODO -- check used secrets(secrtes)
	payloads, rs := constructOutputs(amounts, secrets)
	response, err := WalletClient.Split(api.SplitRequest{Amount: amount, Proofs: proofs, Outputs: payloads})
	if err != nil {
		return nil, nil, err
	}

	return w.constructProofs(response.Fst, secrets[:len(response.Fst)], rs[:len(response.Fst)]),
		w.constructProofs(response.Snd, secrets[len(response.Fst):], rs[len(response.Fst):]), nil
}

func SumProofs(p []cashu.Proof) int64 {
	var sum int64
	for _, proof := range p {
		sum += proof.Amount
	}
	return sum
}
