package lightning

// Will handle the yaml configuration for the proxy.
import (
	"github.com/jinzhu/configor"
	"math"
)

// Configuration for lnbits
type Configuration struct {
	Lnbits struct {
		Enabled                bool    `json:"enabled" yaml:"enabled"`
		LightningFeePercent    float64 `json:"lightning_fee_percent" yaml:"lightning_fee_percent"`
		LightningReserveFeeMin float64 `json:"lightning_reserve_fee_min" yaml:"lightning_reserve_fee_min"`
		AdminKey               string  `yaml:"admin_key"`
		Url                    string  `yaml:"url"`
	} `json:"lnbits" yaml:"lnbits"`
}

var Config Configuration

func init() {
	err := configor.Load(&Config, "config.yaml")
	if err != nil {
		panic(err)
	}
}

func FeeReserve(amountMsat int64, internal bool) int64 {
	if internal {
		return 0
	}
	return int64(math.Max(Config.Lnbits.LightningReserveFeeMin, float64(amountMsat)*Config.Lnbits.LightningFeePercent/1000))
}
