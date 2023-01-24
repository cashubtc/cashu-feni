package feni

import (
	"encoding/base64"
	"encoding/json"
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/crypto"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	RootCmd.AddCommand(receiveCommand)
	receiveCommand.PersistentFlags().StringVarP(&lockFlag, "lock", "l", "", "Lock tokens (P2SH)")

}

var receiveCommand = &cobra.Command{
	Use:    "receive",
	Short:  "Receive tokens",
	Long:   `Receive cashu tokens from another user`,
	PreRun: PreRunFeni,
	Run:    receive,
}

type Token struct {
	Proofs []cashu.Proof `json:"proofs"`
	Mints  Mints         `json:"mints"`
}
type Mint struct {
	URL string   `json:"url"`
	Ks  []string `json:"ks"`
}
type Mints map[string]Mint

func receive(cmd *cobra.Command, args []string) {
	var script, signature string
	coin := args[0]
	if lockFlag != "" {
		if !flagIsPay2ScriptHash() {
			log.Fatal("lock has wrong format. Expected P2SH:<address>")
		}
		addressSplit := strings.Split(lockFlag, "P2SH:")[1]
		p2shScripts, err := getUnusedLocks(addressSplit)
		if err != nil {
			log.Fatal(err)
		}
		if len(p2shScripts) != 1 {
			log.Fatal("lock not found.")
		}
		script = p2shScripts[0].Script
		signature = p2shScripts[0].Signature
	}
	token := Token{}
	decodedCoin, err := base64.URLEncoding.DecodeString(coin)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(decodedCoin, &token)
	if err != nil {
		log.Fatal(err)
	}
	if len(token.Mints) == 0 {
		_, _, err = Wallet.redeem(token.Proofs, script, signature)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	trust := verifyMints(*cmd, token)
	if !trust {
		log.Fatal("Aborted!")
	}
	defaultUrl := Wallet.client.Url
	defer func() {
		Wallet.client.Url = defaultUrl
	}()
	for _, mint := range token.Mints {
		Wallet.client.Url = mint.URL
		_, _, err = Wallet.redeem(token.Proofs, script, signature)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func verifyMints(cmd cobra.Command, token Token) (trust bool) {
	trust = true
	for _, m := range token.Mints {
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
		u := Wallet.client.Url
		Wallet.client.Url = m.URL
		// fetch unknown keysets from mint
		wks, err := Wallet.client.KeySets()
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
		cmd.Printf("Mint URL: %s\n", Wallet.client.Url)
		cmd.Printf("Mint keyset: %s\n", kid)
		cmd.Printf("Do you trust this mint and want to receive the tokens? (y/n)\n")
		trust = ask(&cmd)
		if trust {
			_, err = Wallet.persistCurrentKeysSet()
			if err != nil {
				panic(err)
			}
		}
		Wallet.client.Url = u // reset the url
	}
	return
}
