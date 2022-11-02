package db

import (
	"github.com/gohumble/cashu-feni/cashu"
	"github.com/gohumble/cashu-feni/lightning"
	"github.com/gohumble/cashu-feni/lightning/lnbits"
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
	StoreLightningInvoice(i lightning.Invoice) error
	GetLightningInvoice(hash string) (lightning.Invoice, error)
	GetLightningInvoices(paid bool) ([]lnbits.Invoice, error) // todo -- the return type of this interface function must be of type lightning.Invoice
	UpdateLightningInvoice(hash string, issued bool, paid bool) error

	Migrate(interface{}) error
}
