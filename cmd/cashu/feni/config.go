package feni

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/crypto"
	"github.com/cashubtc/cashu-feni/db"
	"github.com/cashubtc/cashu-feni/lightning"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/joho/godotenv"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"path"
	"time"
)

var Config WalletConfig

type WalletConfig struct {
	Debug          bool   `env:"DEBUG"`
	Lightning      bool   `env:"LIGHTNING"`
	MintServerHost string `env:"MINT_HOST"`
	MintServerPort string `env:"MINT_PORT"`
	Wallet         string `env:"WALLET"`
}

func defaultConfig() {
	log.Infof("Loading default configuration")
	Config = WalletConfig{
		Debug:          true,
		Lightning:      true,
		MintServerHost: "https://8333.space",
		MintServerPort: "3339",
		Wallet:         "wallet",
	}

}
func init() {
	loaded := false
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	p := path.Join(dirname, ".cashu", ".env")
	err = godotenv.Load(p)
	if err != nil {
		defaultConfig()
		loaded = true
	}
	if !loaded {
		err = env.Parse(&Config)
		if err != nil {
			defaultConfig()
		}
	}

	// initialize the default wallet (no other option selected using -w)
	lightning.Config.Lightning.Enabled = Config.Lightning
	InitializeDatabase(Config.Wallet)

	rand.Seed(time.Now().UnixNano())

	Wallet = MintWallet{
		proofs: make([]cashu.Proof, 0),
		client: &Client{Url: fmt.Sprintf("%s:%s", Config.MintServerHost, Config.MintServerPort)},
	}

	Wallet.loadDefaultMint()

}
func (w *MintWallet) loadMint(keySetId string) {
	/*keySet, err := storage.GetKeySet(db.KeySetWithId(keySetId))
	if err != nil {
		panic(err)
	}
	*/
	for _, set := range w.keySets {
		if set.Id == keySetId {
			w.currentKeySet = &set
		}
	}
	w.client.Url = w.currentKeySet.MintUrl
	w.loadDefaultMint()
}
func (w *MintWallet) setCurrentKeySet(keySet crypto.KeySet) {
	for _, set := range w.keySets {
		if set.Id == keySet.Id {
			w.currentKeySet = &keySet
		}
	}
}
func (w *MintWallet) loadPersistedKeySets() {
	persistedKeySets, err := storage.GetKeySet()
	if err != nil {
		panic(err)
	}
	w.keySets = persistedKeySets
}
func (w *MintWallet) loadDefaultMint() {
	keySet, _ := w.persistCurrentKeysSet()
	w.loadPersistedKeySets()
	w.setCurrentKeySet(keySet)
	k, err := w.client.KeySets()
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
func (w *MintWallet) persistCurrentKeysSet() (crypto.KeySet, error) {
	activeKeys, err := w.client.Keys()
	if err != nil {
		panic(err)
	}
	return w.persistKeysSet(activeKeys)
}
func (w *MintWallet) persistKeysSet(keys map[uint64]*secp256k1.PublicKey) (crypto.KeySet, error) {
	keySet := crypto.KeySet{MintUrl: w.client.Url, FirstSeen: time.Now(), PublicKeys: crypto.PublicKeyList{}}
	keySet.SetPublicKeyList(keys)
	keySet.DeriveKeySetId()
	err := storage.StoreKeySet(keySet)
	if err != nil {
		return keySet, err
	}
	return keySet, nil
}
func (w *MintWallet) checkAndPersistKeySet(id string) error {
	var ks []crypto.KeySet
	var err error
	if ks, err = storage.GetKeySet(db.KeySetWithId(id)); err != nil || len(ks) == 0 {
		keys, err := w.client.KeysForKeySet(id)
		if err != nil {
			return err
		}
		k, err := w.persistKeysSet(keys)
		ks = append(ks, k)
		if err != nil {
			return err
		}
	}
	Wallet.keySets = append(Wallet.keySets, ks...)
	return nil
}
func InitializeDatabase(wallet string) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	walletPath := path.Join(dirname, ".cashu", wallet)
	db.Config.Database.Sqlite = &db.SqliteConfig{Path: walletPath, FileName: "wallet.sqlite3"}
	err = env.Parse(&Config)
	if err != nil {
		panic(err)
	}
	storage = db.NewSqlDatabase()
	err = storage.Migrate(cashu.Proof{})
	if err != nil {
		panic(err)
	}
	err = storage.Migrate(cashu.ProofsUsed{})
	if err != nil {
		panic(err)
	}
	err = storage.Migrate(crypto.KeySet{})
	if err != nil {
		panic(err)
	}
	err = storage.Migrate(cashu.P2SHScript{})
	if err != nil {
		panic(err)
	}
	err = storage.Migrate(cashu.CreateInvoice())
	if err != nil {
		panic(err)
	}
}
