package db

import (
	"github.com/gohumble/cashu-feni/core"
	"github.com/gohumble/cashu-feni/lightning"
)

type MintStorage interface {
	GetUsedProofs() []core.Proof
	InvalidateProof(proof core.Proof) error
	StorePromise(p core.Promise) error
	StoreLightningInvoice(i lightning.Invoice) error
	GetLightningInvoice(hash string) (*lightning.Invoice, error)
	UpdateLightningInvoice(hash string, issued bool) error
}
