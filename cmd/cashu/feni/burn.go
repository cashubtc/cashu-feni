package feni

import (
	"encoding/base64"
	"encoding/json"
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/wallet"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var burnCommand = &cobra.Command{
	Use:    "burn",
	Short:  "Burn spent tokens",
	Long:   ``,
	PreRun: RunCommandWithWallet(RootCmd, preRun),
	Run:    RunCommandWithWallet(RootCmd, burnCmd),
}
var all bool
var force bool

func init() {
	burnCommand.PersistentFlags().BoolVarP(&all, "all", "a", false, "burn all spent tokens.")
	burnCommand.PersistentFlags().BoolVarP(&force, "force", "f", false, "force check on all tokens.")
	RootCmd.Command().AddCommand(burnCommand)
}
func burnCmd(wallet *wallet.Wallet, params cobraParameter) {
	var token string
	if len(params.args) == 1 {
		token = params.args[0]
	}
	if !(all || force || token != "") || (token != "" && all) {
		params.cmd.Println("Error: enter a token or use --all to burn all pending tokens or --force to check all tokens.")
		return
	}
	proofs := make([]cashu.Proof, 0)
	var err error
	if all {
		proofs, err = wallet.Storage.GetReservedProofs()
		if err != nil {
			log.Fatal(err)
		}
	} else if force {
		// invalidate all wallet proofs
	} else {
		p, err := base64.URLEncoding.DecodeString(token)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(p, &proofs)
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(proofs) == 0 {
		return
	}

	err = wallet.Invalidate(proofs)
	if err != nil {
		log.Fatal(err)
	}
}
