package feni

import (
	"bytes"
	"fmt"
	"github.com/cashubtc/cashu-feni/wallet"
	"strconv"

	"github.com/c-bata/go-prompt"
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/crypto"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.Command().AddCommand(sendCommand)
	sendCommand.PersistentFlags().StringVarP(&lockFlag, "lock", "l", "", "Lock tokens (P2SH)")
}

var lockFlag string

var sendCommand = &cobra.Command{
	Use:   "send <amount> <mint_id>",
	Short: "Send tokens",
	Long:  `Send cashu tokens to another user`,
	Annotations: map[string]string{
		DynamicSuggestionsAnnotation: getLocksAnnotationValue, // get suggestion for p2sh
	},
	PreRun: RunCommandWithWallet(RootCmd, preRun),
	Run:    RunCommandWithWallet(RootCmd, send),
}
var filteredKeySets []crypto.KeySet
var GetMintsDynamic = func(wallet *wallet.Wallet, annotationValue string) []prompt.Suggest {
	keysets, err := wallet.Storage.GetKeySet()
	if err != nil {
		return nil
	}
	suggestions := make([]prompt.Suggest, 0)
	setBalanceAvailable := make(map[string]uint64)
	balances, err := wallet.Balances()
	if err != nil {
		panic(err)
	}
	filteredKeySets = lo.UniqBy[crypto.KeySet, string](keysets, func(k crypto.KeySet) string {
		if b := balances.ById(k.Id); b != nil {
			setBalanceAvailable[k.MintUrl] = b.Available
		}

		return k.MintUrl
	})

	for i, set := range filteredKeySets {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        fmt.Sprintf("%d", i),
			Description: fmt.Sprintf("Balance: %d sat URL: %s\n", setBalanceAvailable[set.MintUrl], set.MintUrl)})
	}
	return suggestions
}

func askMintSelection(wallet *wallet.Wallet, cmd *cobra.Command) error {
	keysets, err := wallet.Storage.GetKeySet()
	if err != nil {
		return nil
	}
	setBalance := make(map[string]uint64)
	setBalanceAvailable := make(map[string]uint64)
	balances, err := wallet.Balances()
	if err != nil {
		panic(err)
	}
	filteredKeySets = lo.UniqBy[crypto.KeySet, string](keysets, func(k crypto.KeySet) string {
		setBalance[k.MintUrl] = balances.ById(k.Id).Balance
		setBalanceAvailable[k.MintUrl] = balances.ById(k.Id).Available
		return k.MintUrl
	})

	for i, set := range filteredKeySets {
		cmd.Printf("Mint: %d Balance: %d sat (available: %d) URL: %s\n", i+1, setBalance[set.MintUrl], setBalanceAvailable[set.MintUrl], set.MintUrl)
	}
	cmd.Printf("Select mint [1-%d, press enter default 1]\n\n", len(filteredKeySets))
	wallet.Client.Url = filteredKeySets[askInt(cmd)-1].MintUrl
	wallet.LoadDefaultMint()
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
}

func send(wallet *wallet.Wallet, params cobraParameter) {
	if len(params.args) < 2 {
		params.cmd.Help()
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

	mint, _ := strconv.Atoi(params.args[1])
	wallet.Client.Url = filteredKeySets[mint].MintUrl
	wallet.LoadDefaultMint()

	amount, err := strconv.ParseUint(params.args[0], 10, 64)
	if err != nil {
		fmt.Println("invalid amount")
		return
	}
	_, sendProofs, err := wallet.SplitToSend(amount, lockFlag, true)
	if err != nil {
		fmt.Println(err)
		return
	}
	var hide bool
	if lockFlag != "" && !p2sh {
		hide = true
	}
	token, err := serializeToken(wallet.Client, sendProofs, hide)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(token)
}

// serializeToken function serializes a slice of cashu.Proof structures into a Token structure and returns the result as a string.
// If the hideSecrets flag is set to true, the Secret field of each proof will be set to an empty string before serialization.
// The serialized data is returned as a base64-encoded string.
// If an error occurs, the empty string is returned as the result and an error is returned as the second return value.
func serializeToken(client *wallet.Client, proofs []cashu.Proof, hideSecrets bool) (string, error) {
	// Create a new Token structure with the given proofs and an empty Mints map.
	token := Tokens{Token: make([]Token, 0)}
	token.Token = append(token.Token, Token{Proofs: proofs, Mint: client.Url})
	// Iterate over each proof in the `proofs` slice.
	for i := range proofs {
		// If `hideSecrets` is true, set the `Secret` field of the current proof to an empty string.
		if hideSecrets {
			proofs[i].Secret = ""
		}
	}
	// Return the base64-encoded version of the JSON string.
	return token.String(), nil

}
