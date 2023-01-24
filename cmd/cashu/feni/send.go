package feni

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/crypto"
	"github.com/cashubtc/cashu-feni/db"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"strconv"
)

func init() {
	RootCmd.AddCommand(sendCommand)
	sendCommand.PersistentFlags().StringVarP(&lockFlag, "lock", "l", "", "Lock tokens (P2SH)")
}

var lockFlag string

var sendCommand = &cobra.Command{
	Use:    "send <amount> <mint_id>",
	Short:  "Send tokens",
	Long:   `Send cashu tokens to another user`,
	PreRun: PreRunFeni,
	Annotations: map[string]string{
		DynamicSuggestionsAnnotation: getLocksAnnotationValue, // get suggestion for p2sh
	},
	Run: send,
}
var filteredKeySets []crypto.KeySet
var GetMintsDynamic = func(annotationValue string) []prompt.Suggest {
	keysets, err := storage.GetKeySet()
	if err != nil {
		return nil
	}
	suggestions := make([]prompt.Suggest, 0)
	setBalance := make(map[string]uint64)
	setBalanceAvailable := make(map[string]uint64)
	balances, err := Wallet.balancePerKeySet()
	if err != nil {
		panic(err)
	}
	filteredKeySets = lo.UniqBy[crypto.KeySet, string](keysets, func(k crypto.KeySet) string {
		setBalance[k.MintUrl] = balances[k.Id].Balance
		setBalanceAvailable[k.MintUrl] = balances[k.Id].Available
		return k.MintUrl
	})

	for i, set := range filteredKeySets {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        fmt.Sprintf("%d", i),
			Description: fmt.Sprintf("Balance: %d sat (available: %d) URL: %s\n", setBalance[set.MintUrl], setBalanceAvailable[set.MintUrl], set.MintUrl)})
	}
	return suggestions
}

func askMintSelection(cmd *cobra.Command) error {
	keysets, err := storage.GetKeySet()
	if err != nil {
		return nil
	}
	setBalance := make(map[string]uint64)
	setBalanceAvailable := make(map[string]uint64)
	balances, err := Wallet.balancePerKeySet()
	if err != nil {
		panic(err)
	}
	filteredKeySets = lo.UniqBy[crypto.KeySet, string](keysets, func(k crypto.KeySet) string {
		setBalance[k.MintUrl] = balances[k.Id].Balance
		setBalanceAvailable[k.MintUrl] = balances[k.Id].Available
		return k.MintUrl
	})

	for i, set := range filteredKeySets {
		cmd.Printf("Mint: %d Balance: %d sat (available: %d) URL: %s\n", i+1, setBalance[set.MintUrl], setBalanceAvailable[set.MintUrl], set.MintUrl)
	}
	cmd.Printf("Select mint [1-%d, press enter default 1]\n\n", len(filteredKeySets))
	Wallet.client.Url = filteredKeySets[askInt(cmd)-1].MintUrl
	Wallet.loadDefaultMint()
	return nil
}
func askInt(cmd *cobra.Command) int {
	reader := cmd.InOrStdin()
	in := []byte{}
	for i := 0; i <= 8; i++ {
		c := make([]byte, 1)
		_, err := reader.Read(c)
		if err != nil {
			return 0
		}
		if c[0] == 13 {
			in = bytes.Trim(in, "\x00")
			break
		}
		cmd.Printf(string(in))
		in = append(in, c[0])

	}
	s, err := strconv.Atoi(string(in))
	fmt.Printf("%d, %v", s, err)
	return s
	return 0
}

func send(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		cmd.Help()
		return
	}
	if lockFlag != "" && len(lockFlag) < 22 {
		fmt.Println("Error: lock has to be at least 22 characters long.")
		return
	}
	var p2sh bool
	if lockFlag != "" && flagIsPay2ScriptHash() {
		p2sh = true
	}
	mint, _ := strconv.Atoi(args[1])
	Wallet.client.Url = filteredKeySets[mint].MintUrl
	Wallet.loadDefaultMint()
	amount, err := strconv.Atoi(args[0])
	if err != nil {
		panic(err)
	}
	_, sendProofs, err := Wallet.SplitToSend(uint64(amount), lockFlag, true)
	if err != nil {
		panic(err)
	}
	var hide bool
	if lockFlag != "" && !p2sh {
		hide = true
	}
	token, err := serializeToken(sendProofs, hide)
	if err != nil {
		panic(err)
	}
	fmt.Println(token)
}

// serializeToken function serializes a slice of cashu.Proof structures into a Token structure and returns the result as a string.
// If the hideSecrets flag is set to true, the Secret field of each proof will be set to an empty string before serialization.
// The serialized data is returned as a base64-encoded string.
// If an error occurs, the empty string is returned as the result and an error is returned as the second return value.
func serializeToken(proofs []cashu.Proof, hideSecrets bool) (string, error) {
	// Create a new Token structure with the given proofs and an empty Mints map.
	token := Token{Proofs: proofs, Mints: Mints{}}

	// Iterate over each proof in the `proofs` slice.
	for i := range proofs {
		// If `hideSecrets` is true, set the `Secret` field of the current proof to an empty string.
		if hideSecrets {
			proofs[i].Secret = ""
		}

		// Try to find a `crypto.KeySet` structure in the `Wallet.keySets` slice that has the same `Id` as the current proof.
		keyset, ok := lo.Find[crypto.KeySet](Wallet.keySets, func(k crypto.KeySet) bool {
			return k.Id == proofs[i].Id
		})
		// If a matching `crypto.KeySet` was not found, return an error.
		if !ok {
			return "", fmt.Errorf("error finding keyset")
		}
		if _, ok := token.Mints[keyset.Id]; !ok {
			ks, err := storage.GetKeySet(db.KeySetWithId(proofs[i].Id))
			if err != nil {
				return "", err
			}
			token.Mints[keyset.MintUrl] = Mint{URL: ks[0].MintUrl, Ks: []string{keyset.Id}}
		}
	}
	// Marshal the `token` structure as a JSON string.
	jsonProofs, err := json.Marshal(token)
	// If an error occurred while marshaling the JSON, return the error.
	if err != nil {
		return "", err
	}
	// Return the base64-encoded version of the JSON string.
	return base64.URLEncoding.EncodeToString(jsonProofs), nil

}
