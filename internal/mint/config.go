package mint

// Will handle the yaml configuration for the proxy.
import (
	"github.com/jinzhu/configor"
)

type ServerConfiguration struct {
	PrivateKey string `json:"mint_private_key" yaml:"mint_private_key"`
	Host       string `json:"host" yaml:"host"`
	Port       string `json:"port" yaml:"port"`
}
type Configuration struct {
	DocReference string `yaml:"doc_ref" json:"doc_ref"`
	LogLevel     string `yaml:"log_level" json:"log_level"`
	Mint         struct {
		PrivateKey string `json:"mint_private_key" yaml:"mint_private_key"`
		Host       string `json:"host" yaml:"host"`
		Port       string `json:"port" yaml:"port"`
		Tls        struct {
			Enabled  bool   `json:"enabled" yaml:"enabled"`
			KeyFile  string `json:"key_path" yaml:"key_path"`
			CertFile string `json:"cert_path" yaml:"cert_path"`
		} `json:"tls" yaml:"tls"`
	} `json:"mint" yaml:"mint"`
}

var Config Configuration

func init() {
	err := configor.Load(&Config, "config.yaml")
	if err != nil {
		panic(err)
	}
}
