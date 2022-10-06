package mint

import (
	"github.com/gohumble/cashu-feni/internal/core"
	"github.com/gohumble/cashu-feni/internal/lightning"
	"gorm.io/gorm"
)

var Database *gorm.DB

// getUsedProofs reads all proofs from db
func getUsedProofs(db *gorm.DB) []core.Proof {
	proofs := make([]core.Proof, 0)
	db.Find(&proofs)
	return proofs
}

// invalidateProof will write proof to db
func invalidateProof(p core.Proof) error {
	return Database.Create(&p).Error
}

// storePromise will write promise to db
func storePromise(p core.Promise) error {
	return Database.Create(&p).Error
}

// storeLightningInvoice will store lightning invoice in db
func storeLightningInvoice(i lightning.Invoice) error {
	return Database.Create(&i).Error
}

// getLightningInvoice reads lighting invoice from db
func getLightningInvoice(hash string) (*lightning.Invoice, error) {
	invoice := &lightning.Invoice{Hash: hash}
	tx := Database.Find(invoice)
	return invoice, tx.Error
}

// updateLightningInvoice updates lightning invoice in db
func updateLightningInvoice(hash string, issued bool) error {
	i, err := getLightningInvoice(hash)
	if err != nil {
		return err
	}
	i.Issued = issued
	return Database.Save(i).Error
}
