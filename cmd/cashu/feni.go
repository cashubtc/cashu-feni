package main

import (
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/cashubtc/cashu-feni/cmd/cashu/feni"
)

var advancedPrompt = &feni.CobraPrompt{
	RootCmd:                  feni.RootCmd,
	PersistFlagValues:        true,
	ShowHelpCommandAndFlags:  false,
	DisableCompletionCommand: true,
	AddDefaultExitCommand:    false,
	DynamicSuggestionsFunc:   feni.DynamicSuggestion(feni.RootCmd),
	GoPromptOptions: []prompt.Option{
		prompt.OptionTitle("cashu-feni"),
		prompt.OptionPrefix(">(cashu-feni)> "),
		prompt.OptionMaxSuggestion(10),
	},

	OnErrorFunc: func(err error) {
		if strings.Contains(err.Error(), "unknown command") {
			feni.RootCmd.Command().PrintErrln(err)
			return
		}

		feni.RootCmd.Command().PrintErr(err)
		os.Exit(1)
	},
}

func main() {

	advancedPrompt.RootCmd.Command().PersistentFlags().StringVarP(&feni.WalletName, "wallet", "w", "wallet", "Name of your wallet")
	advancedPrompt.RootCmd.Command().PersistentFlags().StringVarP(&advancedPrompt.RootCmd.Wallet().Config.MintServerHost, "host", "H", "", "Mint host address")
	advancedPrompt.Run()
}
