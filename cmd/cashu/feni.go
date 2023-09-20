package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/cashubtc/cashu-feni/cmd/cashu/feni"
)

var sendRegex = regexp.MustCompile("send [0-9]")
var advancedPrompt = &feni.CobraPrompt{
	RootCmd:                  feni.RootCmd,
	PersistFlagValues:        true,
	ShowHelpCommandAndFlags:  false,
	DisableCompletionCommand: true,
	AddDefaultExitCommand:    false,

	GoPromptOptions: []prompt.Option{
		prompt.OptionTitle("cashu-feni"),
		prompt.OptionPrefix(">(cashu-feni)> "),
		prompt.OptionMaxSuggestion(10),
	},
	DynamicSuggestionsFunc: func(annotationValue string, document *prompt.Document) []prompt.Suggest {
		if document.Text == "-w " || document.Text == "--wallet " {
			fmt.Println(document.Text)
			if suggestions := feni.GetWalletsDynamic(annotationValue); suggestions != nil {
				return suggestions
			}
		} else if document.Text == "locks " || document.Text == "-l " {
			if suggestions := feni.GetLocksDynamic(annotationValue); suggestions != nil {
				return suggestions
			}
		} else if sendRegex.MatchString(document.Text) {
			document.Text = fmt.Sprintf("%s %s", document.Text, "-m ")
			if suggestions := feni.GetMintsDynamic(annotationValue); suggestions != nil {
				return suggestions
			}
		}

		return nil
	},
	OnErrorFunc: func(err error) {
		if strings.Contains(err.Error(), "unknown command") {
			feni.RootCmd.PrintErrln(err)
			return
		}

		feni.RootCmd.PrintErr(err)
		os.Exit(1)
	},
}

func main() {
	feni.StartClientConfiguration()
	advancedPrompt.RootCmd.PersistentFlags().StringVarP(&feni.WalletUsed, "wallet", "w", "wallet", "Name of your wallet")
	advancedPrompt.RootCmd.PersistentFlags().StringVarP(&feni.Host, "host", "H", "", "Mint host address")
	advancedPrompt.Run()
}
