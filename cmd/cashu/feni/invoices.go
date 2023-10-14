package feni

import (
	"fmt"
	"github.com/cashubtc/cashu-feni/lightning/invoice"
	"github.com/cashubtc/cashu-feni/wallet"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var invoicesCommand = &cobra.Command{
	Use:   "invoices",
	Short: "List all pending invoices",
	Long:  ``,
	Run:   RunCommandWithWallet(RootCmd, invoicesCmd),
}

func init() {
	RootCmd.Command().AddCommand(invoicesCommand)
}

func invoicesCmd(wallet *wallet.Wallet, params cobraParameter) {
	invoices := make([]invoice.Invoice, 0)
	invoices, err := wallet.Storage.GetLightningInvoices(false)
	if err != nil {
		log.Fatal(err)
	}
	for _, iv := range invoices {
		fmt.Println("--------------------------")
		fmt.Printf("Paid: %t\n", iv.IsIssued())
		fmt.Printf("Incoming: %t\n", iv.GetAmount() > 0)
		fmt.Printf("Amount: %d\n", iv.GetAmount())
		fmt.Printf("Hash: %s\n", iv.GetHash())
		fmt.Printf("PR: %s\n", iv.GetPaymentRequest())
	}
}
