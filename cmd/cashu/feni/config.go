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
		Lightning:      false,
		MintServerHost: "https://8333.space",
		MintServerPort: "3339",
		Wallet:         "wallet",
	}

}
func init() {
	WalletClient = &Client{
		Url: "http://0.0.0.0:3338",
	}
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	p := path.Join(dirname, ".cashu", ".env")
	err = godotenv.Load(p)
	if err != nil {
		defaultConfig()
		return
	}
	err = env.Parse(&Config)
	if err != nil {
		defaultConfig()
		return
	}
	// initialize the default wallet (no other option selected using -w)
	lightning.Config.Lightning.Enabled = Config.Lightning
	InitializeDatabase(Config.Wallet)

	rand.Seed(time.Now().UnixNano())

	Wallet = MintWallet{proofs: make([]cashu.Proof, 0)}
	WalletClient = &Client{Url: fmt.Sprintf("%s:%s", Config.MintServerHost, Config.MintServerPort)}

	loadMint()

}
func loadMint() {
	activeKeys, err := WalletClient.Keys()
	if err != nil {
		panic(err)
	}
	keyset, _ := persistKeysSet(activeKeys)
	Wallet.keySets = append(Wallet.keySets, keyset)

	k, err := WalletClient.KeySets()
	if err != nil {
		panic(err)
	}
	for _, set := range k.KeySets {
		if set == keyset.Id {
			continue
		}
		err = checkAndPersistKeySet(set)
		if err != nil {
			panic(err)
		}

	}
}
func persistKeysSet(keys map[uint64]*secp256k1.PublicKey) (crypto.KeySet, error) {
	keySet := crypto.KeySet{MintUrl: WalletClient.Url, FirstSeen: time.Now(), PublicKeys: crypto.PublicKeyList{}}
	keySet.SetPublicKeyList(keys)
	keySet.DeriveKeySetId()
	err := storage.StoreKeySet(keySet)
	if err != nil {
		return keySet, err
	}
	return keySet, nil
}
func checkAndPersistKeySet(id string) error {
	var ks crypto.KeySet
	var err error
	if ks, err = storage.GetKeySet(id); err != nil {
		keys, err := WalletClient.KeysForKeySet(id)
		if err != nil {
			return err
		}
		ks, err = persistKeysSet(keys)
		if err != nil {
			return err
		}
	}
	Wallet.keySets = append(Wallet.keySets, ks)
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
