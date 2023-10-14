package feni

import (
	"fmt"
	"github.com/cashubtc/cashu-feni/wallet"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var pendingCommand = &cobra.Command{
	Use:    "pending",
	Short:  "Show pending tokens",
	Long:   ``,
	PreRun: RunCommandWithWallet(RootCmd, preRun),
	Run:    RunCommandWithWallet(RootCmd, pendingCmd),
}

func init() {
	RootCmd.Command().AddCommand(pendingCommand)
}
func pendingCmd(wallet *wallet.Wallet, params cobraParameter) {
	reserved, err := storage.GetReservedProofs()
	if err != nil {
		log.Fatal(err)
	}
	if len(reserved) > 0 {
		fmt.Println("--------------------------")
	}
	for i, proof := range reserved {
		fmt.Printf("#%d Amount: %d sat Time: %s, ID: %s\n", i, proof.Amount, proof.TimeReserved, proof.SendId.String())
	}
}
