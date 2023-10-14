package feni

import (
	"fmt"
	"github.com/cashubtc/cashu-feni/db"
	"github.com/cashubtc/cashu-feni/lightning"
	"github.com/cashubtc/cashu-feni/wallet"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
	"time"
)

var invoiceCommand = &cobra.Command{
	Use:    "invoice",
	Short:  "Creates a new invoice, if lightning is enabled",
	Long:   ``,
	PreRun: RunCommandWithWallet(RootCmd, preRun),
	Run:    RunCommandWithWallet(RootCmd, mintCmd),
}
var hash string

func init() {
	invoiceCommand.PersistentFlags().StringVarP(&hash, "hash", "", "", "the hash of the mint you want to claim")
	RootCmd.Command().AddCommand(invoiceCommand)
}
func mintCmd(wallet *wallet.Wallet, params cobraParameter) {
	amount, err := strconv.Atoi(params.args[0])
	if err != nil {
		panic(err)
	}
	if amount > 0 {
		if !Config.Lightning {
			if _, err := wallet.Mint(uint64(amount), hash); err != nil {
				log.Error(err)
			}
			return
		}
		if hash == "" {
			var invoice lightning.Invoicer
			invoice, err = wallet.Client.GetMint(int64(amount))
			if err != nil {
				panic(err)
			}
			err = storage.StoreLightningInvoice(invoice)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Pay invoice to mint %d sat:\n", amount)
			fmt.Printf("Invoice: %s\n", invoice.GetPaymentRequest())
			fmt.Printf("Execute this command if you abort the check:\nfeni invoice {amount} --hash %s\n", invoice.GetHash())
			fmt.Printf("Checking invoice ...")
			for {
				time.Sleep(time.Second * 3)
				proofs, err := wallet.Mint(uint64(amount), invoice.GetHash())
				if err != nil {
					fmt.Print(".")
					continue
				}
				if len(proofs) == 0 {
					fmt.Print(".")
					continue
				}
				fmt.Println("Invoice paid.")
				err = wallet.Storage.UpdateLightningInvoice(invoice.GetHash(), db.UpdateInvoicePaid(true))
				if err != nil {
					log.Fatal(err)
				}
				return
			}
		} else {
			wallet.Mint(uint64(amount), hash)
		}
	}
}
