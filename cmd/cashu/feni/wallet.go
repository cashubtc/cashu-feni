package feni

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/gohumble/cashu-feni/api"
	"github.com/gohumble/cashu-feni/cashu"
	"github.com/gohumble/cashu-feni/crypto"
	"math/rand"
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
	keys   map[uint64]*secp256k1.PublicKey // current public keys from mint server
	keySet string                          // current keySet id from mint server.
	proofs []cashu.Proof
}

var Wallet MintWallet

func init() {

	rand.Seed(time.Now().UnixNano())

	Wallet = MintWallet{proofs: make([]cashu.Proof, 0), keys: make(map[uint64]*secp256k1.PublicKey)}
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

func constructOutputs(amounts []uint64, secrets []string) (api.MintRequest, []*secp256k1.PrivateKey) {
	payloads := api.MintRequest{BlindedMessages: make(cashu.BlindedMessages, 0)}
	privateKeys := make([]*secp256k1.PrivateKey, 0)
	for _, pair := range Zip[string, uint64](secrets, amounts) {
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

func (w MintWallet) checkUsedSecrets(amounts []uint64, secrets []string) error {
	proofs := storage.ProofsUsed(secrets)
	if len(proofs) > 0 {
		return fmt.Errorf("proofs already used")
	}
	return nil
}

func (w MintWallet) mint(amounts []uint64, paymentHash string) []cashu.Proof {
	secrets := make([]string, 0)
	for range amounts {
		secrets = append(secrets, generateSecrets())
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
	Balance   uint64
	Available uint64
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
func generateSecrets() string {
	return base64.RawURLEncoding.EncodeToString([]byte(RandStringRunes(16)))
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
