package feni

import (
	"fmt"
	"github.com/gohumble/cashu-feni/lightning/lnbits"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var invoicesCommand = &cobra.Command{
	Use:    "invoices",
	Short:  "List all pending invoices",
	Long:   ``,
	PreRun: PreRunFeni,
	Run:    invoicesCmd,
}

func init() {
	RootCmd.AddCommand(invoicesCommand)
}

func invoicesCmd(cmd *cobra.Command, args []string) {
	invoices := make([]lnbits.Invoice, 0)
	invoices, err := storage.GetLightningInvoices(false)
	if err != nil {
		log.Fatal(err)
	}
	for _, invoice := range invoices {
		fmt.Println("--------------------------")
		fmt.Printf("Paid: %t\n", invoice.IsIssued())
		fmt.Printf("Incoming: %t\n", invoice.GetAmount() > 0)
		fmt.Printf("Amount: %d\n", invoice.GetAmount())
		fmt.Printf("Hash: %s\n", invoice.GetHash())
		fmt.Printf("PR: %s\n", invoice.GetPaymentRequest())
	}
}
