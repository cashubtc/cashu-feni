package feni

import (
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
	WalletClient = NewClient()

	// initialize the default wallet (no other option selected using -w)
	lightning.Config.Lightning.Enabled = Config.Lightning
	InitializeDatabase(Config.Wallet)

	rand.Seed(time.Now().UnixNano())

	Wallet = MintWallet{proofs: make([]cashu.Proof, 0), keys: make(map[uint64]*secp256k1.PublicKey)}
	mintServerPublicKeys, err := WalletClient.Keys()
	if err != nil {
		panic(err)
	}
	Wallet.keys = mintServerPublicKeys
	keySet, err := WalletClient.KeySets()
	if err != nil {
		panic(err)
	}
	Wallet.keySet = keySet.KeySets[len(keySet.KeySets)-1]
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
