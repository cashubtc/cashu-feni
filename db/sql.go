package db

import (
	"errors"
	"github.com/gohumble/cashu-feni/cashu"
	"github.com/gohumble/cashu-feni/lightning"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path"
)

type SqlDatabase struct {
	db *gorm.DB
}

func NewSqlDatabase() MintStorage {
	if Config.Database.Sqlite != nil {
		return createSqliteDatabase()
	}
	return nil
}

func createSqliteDatabase() MintStorage {
	filePath := Config.Database.Sqlite.Path
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	return SqlDatabase{db: open(sqlite.Open(path.Join(filePath, "database.db")))}
}
func open(dialector gorm.Dialector) *gorm.DB {
	orm, err := gorm.Open(dialector,
		&gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, FullSaveAssociations: true})
	if err != nil {
		panic(err)
	}
	migrationInvoice := cashu.CreateInvoice()
	// do not migrate invoice, if lightning is not enabled
	if migrationInvoice != nil {
		err = orm.AutoMigrate(migrationInvoice)
		if err != nil {
			panic(err)
		}
	}
	err = orm.AutoMigrate(&cashu.Proof{})
	if err != nil {
		panic(err)
	}
	err = orm.AutoMigrate(&cashu.Promise{})
	if err != nil {
		panic(err)
	}
	return orm
}

// getUsedProofs reads all proofs from db
func (s SqlDatabase) GetUsedProofs() []cashu.Proof {
	proofs := make([]cashu.Proof, 0)
	s.db.Find(&proofs)
	return proofs
}

// invalidateProof will write proof to db
func (s SqlDatabase) InvalidateProof(p cashu.Proof) error {
	return s.db.Create(&p).Error
}

// storePromise will write promise to db
func (s SqlDatabase) StorePromise(p cashu.Promise) error {
	return s.db.Create(&p).Error
}

// storeLightningInvoice will store lightning invoice in db
func (s SqlDatabase) StoreLightningInvoice(i lightning.Invoice) error {
	return s.db.Create(i).Error
}

// getLightningInvoice reads lighting invoice from db
func (s SqlDatabase) GetLightningInvoice(hash string) (lightning.Invoice, error) {
	invoice := cashu.CreateInvoice()
	invoice.SetHash(hash)
	tx := s.db.Find(invoice)
	return invoice, tx.Error
}

// updateLightningInvoice updates lightning invoice in db
func (s SqlDatabase) UpdateLightningInvoice(hash string, issued bool) error {
	i, err := s.GetLightningInvoice(hash)
	if err != nil {
		return err
	}
	i.SetIssued(issued)
	return s.db.Save(i).Error
}
