package db

import (
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/lightning"
	"github.com/cashubtc/cashu-feni/lightning/invoice"
	"time"
)

type MintStorage interface {
	GetUsedProofs() ([]cashu.Proof, error)
	GetReservedProofs() ([]cashu.Proof, error)
	ProofsUsed([]string) []cashu.Proof
	StoreProof(proof cashu.Proof) error
	DeleteProof(proof cashu.Proof) error
	StoreUsedProofs(proof cashu.ProofsUsed) error
	StorePromise(p cashu.Promise) error
	StoreScript(p cashu.P2SHScript) error
	GetScripts(address string) ([]cashu.P2SHScript, error)
	StoreLightningInvoice(i lightning.Invoicer) error
	GetLightningInvoice(hash string) (lightning.Invoicer, error)
	GetLightningInvoices(paid bool) ([]invoice.Invoice, error) // todo -- the return type of this interface function must be of type lightning.Invoicer
	UpdateLightningInvoice(hash string, options ...UpdateInvoiceOptions) error

	Migrate(interface{}) error
}

func UpdateInvoiceWithIssued(issued bool) UpdateInvoiceOptions {
	return func(invoice lightning.Invoicer) {
		invoice.SetIssued(issued)
	}
}
func UpdateInvoicePaid(paid bool) UpdateInvoiceOptions {
	return func(invoice lightning.Invoicer) {
		invoice.SetPaid(paid)
	}
}
func UpdateInvoiceTimePaid(t time.Time) UpdateInvoiceOptions {
	return func(invoice lightning.Invoicer) {
		invoice.SetTimePaid(t)
	}
}

type UpdateInvoiceOptions func(invoice lightning.Invoicer)
