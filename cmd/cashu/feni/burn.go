package feni

import (
	"encoding/base64"
	"encoding/json"
	"github.com/cashubtc/cashu-feni/cashu"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var burnCommand = &cobra.Command{
	Use:    "burn",
	Short:  "Burn spent tokens",
	Long:   ``,
	PreRun: PreRunFeni,
	Run:    burnCmd,
}
var all bool
var force bool

func init() {
	burnCommand.PersistentFlags().BoolVarP(&all, "all", "a", false, "burn all spent tokens.")
	burnCommand.PersistentFlags().BoolVarP(&force, "force", "f", false, "force check on all tokens.")
	RootCmd.AddCommand(burnCommand)
}
func burnCmd(cmd *cobra.Command, args []string) {
	var token string
	if len(args) == 1 {
		token = args[0]
	}
	if !(all || force || token != "") || (token != "" && all) {
		cmd.Println("Error: enter a token or use --all to burn all pending tokens or --force to check all tokens.")
		return
	}
	proofs := make([]cashu.Proof, 0)
	var err error
	if all {
		proofs, err = storage.GetReservedProofs()
		if err != nil {
			log.Fatal(err)
		}
	} else if force {
		proofs = Wallet.proofs
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

	err = invalidate(proofs)
	if err != nil {
		log.Fatal(err)
	}
}
