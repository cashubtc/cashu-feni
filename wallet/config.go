package wallet

import (
	"github.com/caarlos0/env/v6"
	"github.com/cashubtc/cashu-feni/lightning"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"path"
	"time"
)

type Config struct {
	Debug          bool   `env:"DEBUG"`
	Lightning      bool   `env:"LIGHTNING"`
	MintServerHost string `env:"MINT_HOST"`
	MintServerPort string `env:"MINT_PORT"`
	Wallet         string `env:"WALLET"`
}

func (w *Wallet) defaultConfig() {
	log.Infof("Loading default configuration")
	w.Config = Config{
		Debug:          true,
		Lightning:      true,
		MintServerHost: "https://8333.space",
		MintServerPort: "3339",
		Wallet:         "wallet",
	}

}
func (w *Wallet) startClientConfiguration() {
	loaded := false
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	p := path.Join(dirname, ".cashu", ".env")
	err = godotenv.Load(p)
	if err != nil {
		w.defaultConfig()
		loaded = true
	}
	if !loaded {
		err = env.Parse(&w.Config)
		if err != nil {
			w.defaultConfig()
		}
	}

	// initialize the default wallet (no other option selected using -w)
	lightning.Config.Lightning.Enabled = w.Config.Lightning

	rand.Seed(time.Now().UnixNano())

}
