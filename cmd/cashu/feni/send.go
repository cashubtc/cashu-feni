package feni

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/crypto"
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
	Use:    "send",
	Short:  "Send tokens",
	Long:   `Send cashu tokens to another user`,
	PreRun: PreRunFeni,
	Annotations: map[string]string{
		DynamicSuggestionsAnnotation: getLocksAnnotationValue,
	},
	Run: send,
}

func send(cmd *cobra.Command, args []string) {
	if lockFlag != "" && len(lockFlag) < 22 {
		fmt.Println("Error: lock has to be at least 22 characters long.")
		return
	}
	var p2sh bool
	if lockFlag != "" && flagIsPay2ScriptHash() {
		p2sh = true
	}
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
	coin, err := serializeToken(sendProofs, hide)
	if err != nil {
		panic(err)
	}
	fmt.Println(coin)
}

// serializeToken function serializes a slice of cashu.Proof structures into a Token structure and returns the result as a string.
// If the hideSecrets flag is set to true, the Secret field of each proof will be set to an empty string before serialization.
// The serialized data is returned as a base64-encoded string.
// If an error occurs, the empty string is returned as the result and an error is returned as the second return value.
func serializeToken(proofs []cashu.Proof, hideSecrets bool) (string, error) {
	// Create a new Token structure with the given proofs and an empty Mints map.
	token := Token{Proofs: proofs, Mints: map[string]map[string]interface{}{}}

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
			ks, err := storage.GetKeySet(proofs[i].Id)
			if err != nil {
				return "", err
			}
			token.Mints[keyset.MintUrl] = make(map[string]interface{}, 0)
			token.Mints[keyset.MintUrl]["url"] = ks.MintUrl
			token.Mints[keyset.MintUrl]["ks"] = make([]string, 0)
		}
		token.Mints[keyset.MintUrl]["ks"] = append(token.Mints[keyset.MintUrl]["ks"].([]string), keyset.Id)
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
