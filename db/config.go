package db

// Will handle the yaml configuration for the proxy.
import (
	"encoding/json"
	"errors"
	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
	"os"
)

type Configuration struct {
	Database struct {
		MySql *struct {
			Host string `json:"host" yaml:"host"`
			Port string `json:"port" yaml:"port"`
		}
		Sqlite *struct {
			Path string `json:"path" yaml:"path"`
		} `json:"sqlite" yaml:"sqlite"`
	} `json:"database" yaml:"database"`
}

var Config Configuration

const name = "config.yaml"

func init() {
	if _, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
		Config.Database.Sqlite.Path = "data"
		cfg, err := json.Marshal(Config)
		if err != nil {
			panic(err)
		}
		log.Warnf("could not load configuration. using default mint configuration instead")
		log.Warnf(string(cfg))
	} else {
		c := configor.New(&configor.Config{Silent: true})
		err = c.Load(&Config, name)
		if err != nil {
			panic(err)
		}
	}
}