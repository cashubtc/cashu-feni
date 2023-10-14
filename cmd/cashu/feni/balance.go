package feni

import (
	"fmt"
	"github.com/cashubtc/cashu-feni/wallet"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.Command().AddCommand(balanceCommand)
}

var balanceCommand = &cobra.Command{
	Use:    "balance",
	Short:  "Check your balance",
	Long:   ``,
	PreRun: RunCommandWithWallet(RootCmd, preRun),
	Run:    RunCommandWithWallet(RootCmd, balance),
}

func balance(wallet *wallet.Wallet, params cobraParameter) {
	balances, err := wallet.Balances()
	if err != nil {
		panic(err)
	}
	fmt.Printf("You have balances in %d keysets:\n", len(balances))
	for _, setBalance := range balances {
		fmt.Printf("Keysets: %v Balance: %d sat URL: %s\n", setBalance.Mint.Ks, setBalance.Available, setBalance.Mint.URL)
	}
}
