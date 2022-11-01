package feni

import (
	"github.com/caarlos0/env/v6"
	"github.com/gohumble/cashu-feni/cashu"
	"github.com/gohumble/cashu-feni/crypto"
	"github.com/gohumble/cashu-feni/db"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
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
		MintServerHost: "localhost",
		MintServerPort: "3338",
		Wallet:         "wallet",
	}

}
func init() {
	WalletClient = &Client{
		url: "http://0.0.0.0:3338",
	}
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	err = godotenv.Load(path.Join(dirname, ".cashu", ".env"))
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
	InitializeDatabase(Config.Wallet)
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
