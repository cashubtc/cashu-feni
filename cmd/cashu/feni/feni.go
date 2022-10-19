package feni

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/gohumble/cashu-feni/db"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path"
)

var WalletUsed string
var Host string

var storage db.MintStorage

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

func PreRunFeni(cmd *cobra.Command, args []string) {
	InitializeDatabase(WalletUsed)
	Wallet.proofs = storage.GetUsedProofs()
}

var RootCmd = &cobra.Command{
	Use:   "feni",
	Short: "Cashu Feni is a cashu wallet application",
	Long:  ``,
	Annotations: map[string]string{
		DynamicSuggestionsAnnotation: getWalletsAnnotationValue,
	},
	Run: func(cmd *cobra.Command, args []string) {
	},
}
