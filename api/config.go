package api

// Will handle the yaml configuration for the proxy.
import (
	"encoding/json"
	"errors"
	"flag"
	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
	"os"
)

type ServerConfiguration struct {
	PrivateKey string `json:"mint_private_key" yaml:"mint_private_key"`
	Host       string `json:"host" yaml:"host"`
	Port       string `json:"port" yaml:"port"`
}
type Configuration struct {
	DocReference string `yaml:"doc_ref" json:"doc_ref"`
	LogLevel     string `yaml:"log_level" json:"log_level"`

	Mint struct {
		PrivateKey     string `json:"private_key" yaml:"private_key"`
		DerivationPath string `json:"derivation_path" yaml:"derivation_path"`
		Host           string `json:"host" yaml:"host"`
		Port           string `json:"port" yaml:"port"`
		Tls            struct {
			Enabled  bool   `json:"enabled" yaml:"enabled"`
			KeyFile  string `json:"key_path" yaml:"key_path"`
			CertFile string `json:"cert_path" yaml:"cert_path"`
		} `json:"tls" yaml:"tls"`
	} `json:"mint" yaml:"mint"`
}

var Config Configuration

const name = "config.yaml"

func (c Configuration) Load() error {
	if _, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
		var host = flag.String("host", "", "the default mint host name")
		var port = flag.String("port", "", "the default mint port")

		Config.Mint.Tls.Enabled = false
		flag.Parse()
		if *port == "" {
			*port = "3338"
		}
		if *host == "" {
			*host = "0.0.0.0"
		}
		Config.Mint.Host = *host
		Config.Mint.Port = *port

		Config.Mint.PrivateKey = "supersecretprivatekey"
		Config.Mint.DerivationPath = "0/0/0/0"
		Config.LogLevel = "trace"
		Config.DocReference = "http://0.0.0.0:3338/swagger/doc.json"
		cfg, err := json.Marshal(Config)
		if err != nil {
			return err
		}
		log.Warnf("could not load configuration. using default mint configuration instead")
		log.Warnf(string(cfg))
	} else {
		c := configor.New(&configor.Config{Silent: true})
		err = c.Load(&Config, name)
		if err != nil {
			return err
		}
	}
	return nil
}
func init() {

}
