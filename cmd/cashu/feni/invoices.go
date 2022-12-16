package feni

import (
	"fmt"
	"github.com/cashubtc/cashu-feni/lightning/invoice"
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
	invoices := make([]invoice.Invoice, 0)
	invoices, err := storage.GetLightningInvoices(false)
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
