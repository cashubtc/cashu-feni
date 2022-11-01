package feni

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gohumble/cashu-feni/cashu"
	"github.com/spf13/cobra"
	"strconv"
)

func init() {
	RootCmd.AddCommand(sendCommand)
	sendCommand.PersistentFlags().StringVarP(&lockFlag, "lock", "l", "", "Lock coins (P2SH)")
}

var lockFlag string

var sendCommand = &cobra.Command{
	Use:    "send",
	Short:  "Send coins",
	Long:   `Send cashu coins to another user`,
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
	coin, err := serialize_proofs(sendProofs, hide)
	if err != nil {
		panic(err)
	}
	fmt.Println(coin)
}

func serialize_proofs(proofs []cashu.Proof, hideSecrets bool) (string, error) {
	if hideSecrets {
		for i := range proofs {
			proofs[i].Secret = ""
		}
	}
	jsonProofs, err := json.Marshal(proofs)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(jsonProofs), nil
}
