package feni

import (
	"encoding/base64"
	"fmt"
	"github.com/gohumble/cashu-feni/bitcoin"
	"github.com/gohumble/cashu-feni/cashu"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(lockCommand)

}

var lockCommand = &cobra.Command{
	Use:    "lock",
	Short:  "Generate receiving lock",
	Long:   `Generates a receiving lock for cashu coins.`,
	PreRun: PreRunFeni,
	Run:    lock,
}

func lock(cmd *cobra.Command, args []string) {
	fmt.Println(createP2SHLock())
}

func createP2SHLock() *cashu.P2SHScript {
	key := bitcoin.Step0CarolPrivateKey()
	txInRedeemScript := bitcoin.Step0CarolCheckSigRedeemScript(*key.PubKey())
	txInP2SHAdress, err := bitcoin.Step1CarolCreateP2SHAddress(txInRedeemScript)
	if err != nil {
		return nil
	}
	txInSignature, err := bitcoin.Step2CarolSignTx(txInRedeemScript, key)
	if err != nil {
		return nil
	}
	txInRedeemScriptB64 := base64.RawURLEncoding.EncodeToString(txInRedeemScript)
	txInSignatureB64 := base64.RawURLEncoding.EncodeToString(txInSignature.SignatureScript)
	p2SHScript := cashu.P2SHScript{Script: txInRedeemScriptB64, Signature: txInSignatureB64, Address: txInP2SHAdress.EncodeAddress()}
	err = storage.StoreScript(p2SHScript)
	if err != nil {
		return nil
	}
	return &p2SHScript
}
