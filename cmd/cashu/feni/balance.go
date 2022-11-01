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
	balances := Wallet.balancePerKeySet()
	fmt.Printf("You have balances in %d keysets:\n", len(balances))
	for s, setBalance := range balances {
		fmt.Printf("Keysets: %s Balance: %d sat (available: %d)\n", s, setBalance.Balance, setBalance.Available)
	}
}
