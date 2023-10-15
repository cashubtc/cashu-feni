package feni

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/wallet"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.Command().AddCommand(locksCommand)

}

const getLocksAnnotationValue = "GetLocks"

var GetLocksDynamic = func(wallet *wallet.Wallet, annotationValue string) []prompt.Suggest {
	scripts, err := wallet.Storage.GetScripts("")
	if err != nil {
		return nil
	}
	suggestions := make([]prompt.Suggest, 0)
	for _, script := range scripts {
		suggestions = append(suggestions, prompt.Suggest{Text: fmt.Sprintf("P2SH:%s", script.Address), Description: fmt.Sprintf("Your P2SH lock for receiving and sending cashu tokens")})

	}
	return suggestions
}
var locksCommand = &cobra.Command{
	Use:   "locks",
	Short: "Show unused receiving locks",
	Long:  `Generates a receiving lock for cashu tokens.`,
	Annotations: map[string]string{
		DynamicSuggestionsAnnotation: getLocksAnnotationValue,
	},
	PreRun: RunCommandWithWallet(RootCmd, preRun),
	Run:    RunCommandWithWallet(RootCmd, locks),
}

func locks(wallet *wallet.Wallet, params cobraParameter) {
	scriptLocks := getP2SHLocks(wallet)
	for _, l := range scriptLocks {
		fmt.Printf("P2SH:%s\n", l.Address)
	}
}

func getP2SHLocks(wallet *wallet.Wallet) []cashu.P2SHScript {
	scripts, err := wallet.Storage.GetScripts("")
	if err != nil {
		return nil
	}
	return scripts
}
