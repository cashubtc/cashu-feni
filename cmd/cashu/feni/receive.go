package feni

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/wallet"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	RootCmd.Command().AddCommand(receiveCommand)
	receiveCommand.PersistentFlags().StringVarP(&lockFlag, "lock", "l", "", "Lock tokens (P2SH)")

}

var receiveCommand = &cobra.Command{
	Use:    "receive",
	Short:  "Receive tokens",
	Long:   `Receive cashu tokens from another user`,
	PreRun: RunCommandWithWallet(RootCmd, preRun),
	Run:    RunCommandWithWallet(RootCmd, receive),
}

type Tokens struct {
	Token []Token `json:"token"`
	Memo  string  `json:"memo"`
}
type Proofs struct {
	ID     string `json:"id"`
	Amount int    `json:"amount"`
	Secret string `json:"secret"`
	C      string `json:"C"`
}
type Token struct {
	Mint   string        `json:"mint"`
	Proofs []cashu.Proof `json:"proofs"`
}

func (t Tokens) String() string {
	tokenBytes, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}

	encodedToken := base64.URLEncoding.EncodeToString(tokenBytes)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("cashuA%s", encodedToken)
}

func NewTokens(t string) *Tokens {
	token := &Tokens{}
	decodedCoin, err := base64.URLEncoding.DecodeString(t[len("cashuA"):])
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(decodedCoin, &token)
	if err != nil {
		log.Fatal(err)
	}
	return token
}

func receive(wallet *wallet.Wallet, params cobraParameter) {
	var script, signature string
	coin := params.args[0]
	if lockFlag != "" {
		if !flagIsPay2ScriptHash() {
			log.Fatal("lock has wrong format. Expected P2SH:<address>")
		}
		addressSplit := strings.Split(lockFlag, "P2SH:")[1]
		p2shScripts, err := getUnusedLocks(wallet, addressSplit)
		if err != nil {
			log.Fatal(err)
		}
		if len(p2shScripts) != 1 {
			log.Fatal("lock not found.")
		}
		script = p2shScripts[0].Script
		signature = p2shScripts[0].Signature
	}
	tokens := NewTokens(coin)

	/*trust := verifyMints(*cmd, tokens)
	if !trust {
		log.Fatal("Aborted!")
	}*/
	for _, token := range tokens.Token {
		defaultUrl := wallet.Client.Url
		defer func() {
			wallet.Client.Url = defaultUrl
		}()
		wallet.Client.Url = token.Mint
		_, _, err := redeem(wallet, token.Proofs, script, signature)
		if err != nil {
			log.Fatal(err)
		}
	}
}
func redeem(w *wallet.Wallet, proofs []cashu.Proof, scndScript, scndSignature string) (keep []cashu.Proof, send []cashu.Proof, err error) {
	if scndScript != "" && scndSignature != "" {
		log.Infof("Unlock script: %s", scndScript)
		for i := range proofs {
			proofs[i].Script = &cashu.P2SHScript{
				Script:    scndScript,
				Signature: scndSignature}
		}
	}
	return w.Split(proofs, wallet.SumProofs(proofs), "")
}
func getUnusedLocks(wallet *wallet.Wallet, addressSplit string) ([]cashu.P2SHScript, error) {
	return wallet.Storage.GetScripts(addressSplit)
}

/*
func verifyMints(cmd cobra.Command, token Tokens) (trust bool) {
	trust = true
	for _, m := range token.Token {
		_, exists := lo.Find[crypto.KeySet](Wallet.keySets, func(s crypto.KeySet) bool {
			for _, k := range m.Ks {
				if k == s.Id {
					return true
				}
			}
			return false
		})
		if exists {
			continue
		}
		trust = false
		u := Wallet.Client.Url
		Wallet.Client.Url = m.Mint
		// fetch unknown keysets from mint
		wks, err := Wallet.Client.KeySets()
		if err != nil {
			panic(err)
		}
		// check if keyset id from token matches mint response
		kid, mintKeySetExists := lo.Find[string](wks.KeySets, func(s string) bool {
			for _, k := range m.Ks {
				if k == s {
					return true
				}
			}
			return false
		})
		if !mintKeySetExists {
			panic("mint does not have this keyset.")
		}
		// ask user to verify trust
		cmd.Printf("Warning: Tokens are from a mint you don't know yet.\n")
		cmd.Printf("Mint URL: %s\n", Wallet.Client.Url)
		cmd.Printf("Mint keyset: %s\n", kid)
		cmd.Printf("Do you trust this mint and want to receive the tokens? (y/n)\n")
		trust = ask(&cmd)
		if trust {
			_, err = Wallet.persistCurrentKeysSet()
			if err != nil {
				panic(err)
			}
		}
		Wallet.Client.Url = u // reset the url
	}
	return
}
*/
