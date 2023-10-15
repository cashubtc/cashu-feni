package wallet

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/caarlos0/env/v6"
	"math/rand"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/crypto"
	"github.com/cashubtc/cashu-feni/db"
	"github.com/cashubtc/cashu-feni/mint"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/google/uuid"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
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

type Wallet struct {
	//	keys    map[uint64]*secp256k1.PublicKey // current public keys from mint server
	keySets       []crypto.KeySet // current keySet id from mint server.
	proofs        []cashu.Proof
	currentKeySet *crypto.KeySet
	Client        *Client
	Storage       db.MintStorage
	Config        Config
}
type Option func(w *Wallet)

func WithName(name string) Option {
	return func(w *Wallet) {
		w.Config.Wallet = name
	}
}
func New(opts ...Option) *Wallet {
	wallet := &Wallet{
		proofs: make([]cashu.Proof, 0),
	}
	wallet.startClientConfiguration()
	for _, opt := range opts {
		opt(wallet)
	}
	wallet.initializeDatabase(wallet.Config.Wallet)

	wallet.Client = &Client{Url: fmt.Sprintf("%s:%s", wallet.Config.MintServerHost, wallet.Config.MintServerPort)}
	wallet.LoadDefaultMint()

	proofs, err := wallet.Storage.GetUsedProofs()
	if err != nil {
		return nil
	}
	wallet.proofs = proofs
	return wallet
}

// constructOutputs takes in a slice of amounts and a slice of secrets, and
// constructs a MintRequest with blinded messages and a slice of private keys
// corresponding to the given amounts and secrets.
func constructOutputs(amounts []uint64, secrets []string) (cashu.MintRequest, []*secp256k1.PrivateKey) {
	// Create a new empty MintRequest with a slice of blinded messages.
	payloads := cashu.MintRequest{Outputs: make(cashu.BlindedMessages, 0)}
	// Create an empty slice of private keys.
	privateKeys := make([]*secp256k1.PrivateKey, 0)
	// For each pair of amount and secret in the input slices,
	for _, pair := range Zip[string, uint64](secrets, amounts) {
		// Generate a private key.
		r, err := secp256k1.GeneratePrivateKey()
		if err != nil {
			// If there is an error generating the private key, panic.
			panic(err)
		}
		// Compute the first step of the two-step blind signature protocol using the given secret and private key.
		pub, r := crypto.FirstStepAlice(pair.First, r)
		// Append the private key to the slice of private keys.
		privateKeys = append(privateKeys, r)
		// Append a new blinded message to the MintRequest using the given amount and the computed public key.
		payloads.Outputs = append(payloads.Outputs,
			cashu.BlindedMessage{Amount: pair.Second, B_: fmt.Sprintf("%x", pub.SerializeCompressed())})
	}
	// Return the MintRequest and the slice of private keys.
	return payloads, privateKeys
}

func (w *Wallet) checkUsedSecrets(amounts []uint64, secrets []string) error {
	proofs := w.Storage.ProofsUsed(secrets)
	if len(proofs) > 0 {
		return fmt.Errorf("proofs already used")
	}
	return nil
}

func (w *Wallet) AvailableBalance() uint64 {
	return SumProofs(w.proofs)
}

func (w *Wallet) StoreProofs(proofs []cashu.Proof) error {
	for _, proof := range proofs {
		w.proofs = append(w.proofs, proof)
		err := w.Storage.StoreProof(proof)
		if err != nil {
			return err
		}
	}
	return nil
}
func (w *Wallet) Mint(amount uint64, paymentHash string) ([]cashu.Proof, error) {
	split := mint.AmountSplit(amount)
	proofs := w.mint(split, paymentHash)
	if len(proofs) == 0 {
		return nil, fmt.Errorf("received no proofs.")
	}
	err := w.StoreProofs(proofs)
	if err != nil {
		return nil, err
	}
	if paymentHash != "" {
		err = w.Storage.UpdateLightningInvoice(
			paymentHash,
			db.UpdateInvoicePaid(true),
			db.UpdateInvoiceTimePaid(time.Now()),
		)
		if err != nil {
			return nil, err
		}
	}
	w.proofs = append(w.proofs, proofs...)
	return proofs, nil
}
func (w *Wallet) mint(amounts []uint64, paymentHash string) []cashu.Proof {
	secrets := make([]string, 0)
	for range amounts {
		secrets = append(secrets, generateSecret())
	}
	err := w.checkUsedSecrets(amounts, secrets)
	if err != nil {
		panic(err)
	}
	req, privateKeys := constructOutputs(amounts, secrets)
	blindedSignatures, err := w.Client.Mint(req, paymentHash)
	if err != nil {
		panic(err)
	}
	return w.constructProofs(blindedSignatures.Promises, secrets, privateKeys)
}

func (w *Wallet) constructProofs(promises []cashu.BlindedSignature, secrets []string, privateKeys []*secp256k1.PrivateKey) []cashu.Proof {
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
		C := crypto.ThirdStepAlice(*C_, *privateKeys[i], *w.currentKeySet.PublicKeys.GetKeyByAmount(promise.Amount).Key)
		proofs = append(proofs, cashu.Proof{
			Id:     w.currentKeySet.Id,
			Amount: promise.Amount,
			C:      fmt.Sprintf("%x", C.SerializeCompressed()),
			Secret: secrets[i],
		})
	}
	return proofs
}

type Balance struct {
	Balance   uint64
	Available uint64
	Mint      Mint
}
type Mint struct {
	URL string   `json:"url"`
	Ks  []string `json:"ks"`
}
type Mints map[string]Mint

type Balances []*Balance

func (b Balances) ById(id string) *Balance {
	for _, ba := range b {
		if found := slices.Contains[string](ba.Mint.Ks, id); found {
			return ba
		}
	}
	return nil
}
func (w *Wallet) getProofsPerMintUrl() cashu.Proofs {
	return w.proofs
}
func (w *Wallet) Balances() (Balances, error) {
	balances := make(Balances, 0)
	for _, proof := range w.proofs {
		proofBalance, foundBalance := lo.Find[*Balance](balances, func(b *Balance) bool {
			return slices.Contains[string](b.Mint.Ks, proof.Id)
		})
		if foundBalance {
			proofBalance.Balance += proof.Amount
		} else {
			proofBalance = &Balance{
				Balance: proof.Amount,
			}
		}
		if !proof.Reserved {
			proofBalance.Available += proof.Amount
		}
		keySet, found := lo.Find[crypto.KeySet](w.keySets, func(k crypto.KeySet) bool {
			return k.Id == proof.Id
		})
		if found {
			u, err := url.Parse(keySet.MintUrl)
			if err != nil {
				return nil, err
			}
			proofBalance.Mint.URL = u.String()
			proofBalance.Mint.Ks = []string{keySet.Id}
		}
		if !foundBalance {
			balances = append(balances, proofBalance)
		}
	}
	return balances, nil
}

func generateSecrets(secret string, n int) []string {
	secrets := make([]string, 0)
	var generator func(i int)
	if cashu.IsPay2ScriptHash(secret) {
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

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (w *Wallet) PayLightning(proofs []cashu.Proof, invoice string) ([]cashu.Proof, error) {
	secrets := make([]string, 0)
	amounts := []uint64{0, 0, 0, 0}
	for i := 0; i < 4; i++ {
		secrets = append(secrets, generateSecret())
	}
	payloads, rs := constructOutputs(amounts, secrets)
	res, err := w.Client.Melt(cashu.MeltRequest{Proofs: proofs, Pr: invoice, Outputs: payloads.Outputs})
	if err != nil {
		return nil, err
	}
	if res.Paid {
		changeProofs := w.constructProofs(res.Change, secrets, rs)
		err = w.Invalidate(proofs)
		if err != nil {
			return changeProofs, err
		}
		return changeProofs, nil
	}
	return nil, fmt.Errorf("could not pay invoice")
}
func (w *Wallet) getKeySet(id string) (crypto.KeySet, error) {
	k, found := lo.Find[crypto.KeySet](w.keySets, func(k crypto.KeySet) bool {
		return k.Id == id
	})
	if !found {
		return k, fmt.Errorf("keyset does not exist")
	}
	return k, nil
}

func (w *Wallet) GetSpendableProofs() ([]cashu.Proof, error) {
	spendable := make([]cashu.Proof, 0)
	for _, proof := range lo.Filter[cashu.Proof](w.proofs, func(p cashu.Proof, i int) bool {
		return p.Id == w.currentKeySet.Id
	}) {
		if proof.Reserved {
			continue
		}
		keySet, err := w.getKeySet(proof.Id)
		if err != nil {
			return nil, err
		}
		if proof.Id != keySet.Id {
			continue
		}
		spendable = append(spendable, proof)
	}
	return spendable, nil
}

func (w *Wallet) SplitToSend(amount uint64, scndSecret string, setReserved bool) (keep []cashu.Proof, send []cashu.Proof, err error) {
	spendableProofs, err := w.GetSpendableProofs()
	if err != nil {
		return nil, nil, err
	}
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
func (w *Wallet) setReserved(p []cashu.Proof, reserved bool) error {
	for _, proof := range p {
		proof.Reserved = reserved
		proof.SendId = uuid.New()
		proof.TimeReserved = time.Now()
		err := w.Storage.StoreProof(proof)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Wallet) Split(proofs []cashu.Proof, amount uint64, scndSecret string) (keep []cashu.Proof, send []cashu.Proof, err error) {
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
	err = w.StoreProofs(append(frstProofs, scndProofs...))
	if err != nil {
		return nil, nil, err
	}
	for _, proof := range proofs {
		err = w.invalidateProof(proof)
		if err != nil {
			return nil, nil, err
		}
	}
	return frstProofs, scndProofs, nil
}
func (w *Wallet) split(proofs []cashu.Proof, amount uint64, scndSecret string) (keep []cashu.Proof, send []cashu.Proof, err error) {

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
	response, err := w.Client.Split(cashu.SplitRequest{Amount: amount, Proofs: proofs, Outputs: payloads.Outputs})
	if err != nil {
		return nil, nil, err
	}

	return w.constructProofs(response.Fst, secrets[:len(response.Fst)], rs[:len(response.Fst)]),
		w.constructProofs(response.Snd, secrets[len(response.Fst):], rs[len(response.Fst):]), nil
}

func SumProofs(p []cashu.Proof) uint64 {
	var sum uint64
	for _, proof := range p {
		sum += proof.Amount
	}
	return sum
}

func (w *Wallet) Invalidate(proofs []cashu.Proof) error {
	if len(proofs) == 0 {
		proofs = w.proofs
	}
	resp, err := w.Client.Check(cashu.CheckSpendableRequest{Proofs: proofs})
	if err != nil {
		return err
	}
	invalidatedProofs := make([]cashu.Proof, 0)
	for i, spendable := range resp.Spendable {
		if !spendable {
			invalidatedProofs = append(invalidatedProofs, proofs[i])
			err = w.invalidateProof(proofs[i])
			if err != nil {
				return err
			}
		}
	}
	invalidatedSecrets := make([]string, 0)
	for _, proof := range invalidatedProofs {
		invalidatedSecrets = append(invalidatedSecrets, proof.Secret)
	}
	w.proofs = lo.Filter[cashu.Proof](w.proofs, func(p cashu.Proof, i int) bool {
		_, found := lo.Find[string](invalidatedSecrets, func(secret string) bool {
			return secret == p.Secret
		})
		return !found
	})
	return nil
}
func (w *Wallet) invalidateProof(proof cashu.Proof) error {
	err := w.Storage.DeleteProof(proof)
	if err != nil {
		return err
	}
	return w.Storage.StoreUsedProofs(
		cashu.ProofsUsed{
			Secret:   proof.Secret,
			Amount:   proof.Amount,
			C:        proof.C,
			TimeUsed: time.Now(),
		},
	)
}

func (w *Wallet) loadMint(keySetId string) {
	/*keySet, err := Storage.GetKeySet(db.KeySetWithId(keySetId))
	if err != nil {
		panic(err)
	}
	*/
	for _, set := range w.keySets {
		if set.Id == keySetId {
			w.currentKeySet = &set
		}
	}
	w.Client.Url = w.currentKeySet.MintUrl
	w.LoadDefaultMint()
}
func (w *Wallet) setCurrentKeySet(keySet crypto.KeySet) {
	for _, set := range w.keySets {
		if set.Id == keySet.Id {
			w.currentKeySet = &keySet
		}
	}
}
func (w *Wallet) loadPersistedKeySets() {
	persistedKeySets, err := w.Storage.GetKeySet()
	if err != nil {
		panic(err)
	}
	w.keySets = persistedKeySets
}
func (w *Wallet) LoadDefaultMint() {
	keySet, _ := w.persistCurrentKeysSet()
	w.loadPersistedKeySets()
	w.setCurrentKeySet(keySet)
	k, err := w.Client.KeySets()
	if err != nil {
		panic(err)
	}
	for _, set := range k.KeySets {
		if _, found := lo.Find[crypto.KeySet](w.keySets, func(k crypto.KeySet) bool {
			return set == k.Id
		}); !found {
			err = w.checkAndPersistKeySet(set)
			if err != nil {
				panic(err)
			}
		}
	}

}

func (w *Wallet) persistCurrentKeysSet() (crypto.KeySet, error) {
	activeKeys, err := w.Client.Keys()
	if err != nil {
		panic(err)
	}
	return w.persistKeysSet(activeKeys)
}

func (w *Wallet) persistKeysSet(keys map[uint64]*secp256k1.PublicKey) (crypto.KeySet, error) {
	keySet := crypto.KeySet{MintUrl: w.Client.Url, FirstSeen: time.Now(), PublicKeys: crypto.PublicKeyList{}}
	keySet.SetPublicKeyList(keys)
	keySet.DeriveKeySetId()
	err := w.Storage.StoreKeySet(keySet)
	if err != nil {
		return keySet, err
	}
	return keySet, nil
}

func (w *Wallet) checkAndPersistKeySet(id string) error {
	var ks []crypto.KeySet
	var err error
	if ks, err = w.Storage.GetKeySet(db.KeySetWithId(id)); err != nil || len(ks) == 0 {
		keys, err := w.Client.KeysForKeySet(id)
		if err != nil {
			return err
		}
		k, err := w.persistKeysSet(keys)
		ks = append(ks, k)
		if err != nil {
			return err
		}
	}
	w.keySets = append(w.keySets, ks...)
	return nil
}

func (w *Wallet) initializeDatabase(wallet string) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	walletPath := path.Join(dirname, ".cashu", wallet)
	db.Config.Database.Sqlite = &db.SqliteConfig{Path: walletPath, FileName: "wallet.sqlite3"}
	err = env.Parse(&w.Config)
	if err != nil {
		panic(err)
	}
	w.Storage = db.NewSqlDatabase()
	err = w.Storage.Migrate(cashu.Proof{})
	if err != nil {
		panic(err)
	}
	err = w.Storage.Migrate(cashu.ProofsUsed{})
	if err != nil {
		panic(err)
	}
	err = w.Storage.Migrate(crypto.KeySet{})
	if err != nil {
		panic(err)
	}
	err = w.Storage.Migrate(cashu.P2SHScript{})
	if err != nil {
		panic(err)
	}
	err = w.Storage.Migrate(cashu.CreateInvoice())
	if err != nil {
		panic(err)
	}
}
