package feni

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(balanceCommand)
}

var balanceCommand = &cobra.Command{
	Use:    "balance",
	Short:  "Check your balance blyad",
	Long:   ``,
	PreRun: PreRunFeni,
	Run:    balance,
}

func balance(cmd *cobra.Command, args []string) {
	fmt.Println(Wallet.balancePerKeySet())
}
