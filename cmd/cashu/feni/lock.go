package feni

import (
	"encoding/base64"
	"fmt"
	"github.com/cashubtc/cashu-feni/bitcoin"
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/wallet"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.Command().AddCommand(lockCommand)

}

var lockCommand = &cobra.Command{
	Use:    "lock",
	Short:  "Generate receiving lock",
	Long:   `Generates a receiving lock for cashu tokens.`,
	PreRun: RunCommandWithWallet(RootCmd, preRun),
	Run:    RunCommandWithWallet(RootCmd, lock),
}

func flagIsPay2ScriptHash() bool {
	return cashu.IsPay2ScriptHash(lockFlag)
}

func lock(wallet *wallet.Wallet, params cobraParameter) {
	fmt.Println(createP2SHLock())
}

func createP2SHLock() *cashu.P2SHScript {
	key := bitcoin.Step0CarolPrivateKey()
	txInRedeemScript := bitcoin.Step0CarolCheckSigRedeemScript(*key.PubKey())
	fmt.Println(txInRedeemScript)
	txInP2SHAdress, err := bitcoin.Step1CarolCreateP2SHAddress(txInRedeemScript)
	if err != nil {
		return nil
	}
	txInSignature, err := bitcoin.Step2CarolSignTx(txInRedeemScript, key)
	if err != nil {
		return nil
	}
	txInRedeemScriptB64 := base64.URLEncoding.EncodeToString(txInRedeemScript)
	txInSignatureB64 := base64.URLEncoding.EncodeToString(txInSignature.SignatureScript)
	p2SHScript := cashu.P2SHScript{Script: txInRedeemScriptB64, Signature: txInSignatureB64, Address: txInP2SHAdress.EncodeAddress()}
	err = storage.StoreScript(p2SHScript)
	if err != nil {
		return nil
	}
	return &p2SHScript
}
