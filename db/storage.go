package db

import (
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/crypto"
	"github.com/cashubtc/cashu-feni/lightning"
	"github.com/cashubtc/cashu-feni/lightning/invoice"
	"gorm.io/gorm"
	"time"
)

type MintStorage interface {
	GetUsedProofs(secrets ...string) ([]cashu.Proof, error)
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
	GetKeySet(options ...GetKeySetOptions) ([]crypto.KeySet, error)
	StoreKeySet(k crypto.KeySet) error
	Migrate(interface{}) error
}

func KeySetWithId(id string) GetKeySetOptions {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	}
}

func KeySetWithMintUrl(mintUrl string) GetKeySetOptions {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("mint_url = ?", mintUrl)
	}
}

type GetKeySetOptions func(db *gorm.DB) *gorm.DB

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
