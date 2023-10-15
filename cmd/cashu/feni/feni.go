package feni

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/cashubtc/cashu-feni/wallet"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path"
)

const getWalletsAnnotationValue = "GetWallets"

var GetWalletsDynamic = func(annotationValue string) []prompt.Suggest {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	err = godotenv.Load()
	files, err := os.ReadDir(path.Join(dirname, ".cashu"))
	if err != nil {
		log.Fatal(err)
	}
	suggestions := make([]prompt.Suggest, 0)
	for _, file := range files {
		if file.IsDir() {
			suggestions = append(suggestions, prompt.Suggest{Text: file.Name(), Description: fmt.Sprintf("Wallet %s in your home folder", file.Name())})
		}

	}
	return suggestions
}

var RootCmd = &RootCommand{
	wallet: &wallet.Wallet{},
	cmd: &cobra.Command{
		Use:   "feni",
		Short: "Cashu Feni is a cashu wallet application",
		Long:  ``,
		Annotations: map[string]string{
			DynamicSuggestionsAnnotation: getWalletsAnnotationValue,
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	},
}
