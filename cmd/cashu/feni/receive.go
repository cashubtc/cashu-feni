package feni

import (
	"encoding/base64"
	"encoding/json"
	"github.com/gohumble/cashu-feni/cashu"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	RootCmd.AddCommand(receiveCommand)
	receiveCommand.PersistentFlags().StringVarP(&lockFlag, "lock", "l", "", "Lock coins (P2SH)")

}

var receiveCommand = &cobra.Command{
	Use:    "receive",
	Short:  "Receive coins",
	Long:   `Receive cashu coins from another user`,
	PreRun: PreRunFeni,
	Run:    receive,
}

func receive(cmd *cobra.Command, args []string) {
	var script, signature string
	coin := args[0]
	if lockFlag != "" {
		if !flagIsPay2ScriptHash() {
			log.Fatal("lock has wrong format. Expected P2SH:<address>")
		}
		addressSplit := strings.Split(lockFlag, "P2SH:")[1]
		p2shScripts, err := getUnusedLocks(addressSplit)
		if err != nil {
			log.Fatal(err)
		}
		if len(p2shScripts) != 1 {
			log.Fatal("lock not found.")
		}
		script = p2shScripts[0].Script
		signature = p2shScripts[0].Signature
	}
	proofs := make([]cashu.Proof, 0)
	decodedCoin, err := base64.URLEncoding.DecodeString(coin)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(decodedCoin, &proofs)
	if err != nil {
		log.Fatal(err)
	}
	_, _, err = Wallet.redeem(proofs, script, signature)
	if err != nil {
		log.Fatal(err)
	}
}
