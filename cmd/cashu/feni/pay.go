package feni

import (
	"fmt"
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/wallet"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"math"
	"strings"
)

func init() {
	RootCmd.Command().AddCommand(payCommand)

}

var payCommand = &cobra.Command{
	Use:   "pay <invoice>",
	Short: "Pay lightning invoice",
	Long:  `Pay a lightning invoice using cashu tokens.`,
	Run:   RunCommandWithWallet(RootCmd, pay),
}

func ask(cmd *cobra.Command) bool {
	reader := cmd.InOrStdin()
	in := [2]byte{}
	for i := 0; i <= 1; i++ {
		c := make([]byte, 1)
		_, err := reader.Read(c)
		if err != nil {
			return false
		}
		// ATTENTION: newline is somehow converted to \r. (probably go-prompt)
		// 13 == \r
		if c[0] == 13 {
			c[0] = 10 // \n
		}
		cmd.Printf("%s", c)
		in[i] = c[0]
		s := strings.ToLower(fmt.Sprintf("%s", in))
		if strings.Compare(s, "n\n") == 0 {
			break
		} else if strings.Compare(s, "y\n") == 0 {
			return true
		} else {
			continue
		}
	}
	return false
}
func pay(wallet *wallet.Wallet, params cobraParameter) {
	if len(params.args) != 1 {
		params.cmd.Help()
		return
	}
	invoice := params.args[0]
	fee, err := wallet.Client.CheckFee(cashu.CheckFeesRequest{Pr: invoice})
	if err != nil {
		log.Fatal(err)
	}
	bold, err := decodepay.Decodepay(invoice)
	if err != nil {
		params.cmd.Println("invalid invoice")
		return
	}
	amount := math.Ceil(float64((uint64(bold.MSatoshi) + fee.Fee*1000) / 1000))
	if amount < 0 {
		log.Fatal("amount is not positive")
	}
	if wallet.AvailableBalance() < uint64(amount) {
		log.Fatal("Error: Balance to low.")
	}
	params.cmd.Printf("Pay %d sat (%f sat incl. fees)?\n", uint64(amount)-fee.Fee, amount)
	params.cmd.Println("continue? [Y/n]")
	if !ask(params.cmd) {
		params.cmd.Println("canceled...")
		return
	}
	_, sendProofs, err := wallet.SplitToSend(uint64(amount), "", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Paying Lightning invoice ...")
	changeProofs, err := wallet.PayLightning(sendProofs, invoice)
	if changeProofs != nil {
		err = wallet.StoreProofs(changeProofs)
		if err != nil {
			log.Fatal(err)
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}
