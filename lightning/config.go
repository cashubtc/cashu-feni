package lightning

// Will handle the yaml configuration for the proxy.
import (
	"encoding/json"
	"errors"
	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
	"math"
	"os"
)

// Configuration for lnbits
type Configuration struct {
	Lightning struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
		Lnbits  *struct {
			LightningFeePercent    float64 `json:"lightning_fee_percent" yaml:"lightning_fee_percent"`
			LightningReserveFeeMin float64 `json:"lightning_reserve_fee_min" yaml:"lightning_reserve_fee_min"`
			AdminKey               string  `yaml:"admin_key"`
			Url                    string  `yaml:"url"`
		} `json:"lnbits" yaml:"lnbits"`
	} `json:"lightning" json:"lightning"`
}

var Config Configuration

const name = "config.yaml"

func init() {
	if _, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
		// no lightning client configuration found. starting mint without lightning support.
		Config.Lightning.Enabled = false
		cfg, err := json.Marshal(Config)
		if err != nil {
			panic(err)
		}
		log.Warnf("could not load configuration. using default mint configuration instead")
		log.Warnf(string(cfg))
	} else {
		c := configor.New(&configor.Config{Silent: true})
		err = c.Load(&Config, "config.yaml")
		if err != nil {
			panic(err)
		}
	}

}

func FeeReserve(amountMsat int64, internal bool) int64 {
	if internal {
		return 0
	}
	return int64(math.Max(Config.Lightning.Lnbits.LightningReserveFeeMin, float64(amountMsat)*Config.Lightning.Lnbits.LightningFeePercent/1000))
}
