package db

import (
	"errors"
	"github.com/gohumble/cashu-feni/cashu"
	"github.com/gohumble/cashu-feni/lightning"
	cashuLog "github.com/gohumble/cashu-feni/log"
	log "github.com/sirupsen/logrus"
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

	db := SqlDatabase{db: open(sqlite.Open(path.Join(filePath, Config.Database.Sqlite.FileName)))}

	return db
}

func open(dialector gorm.Dialector) *gorm.DB {
	orm, err := gorm.Open(dialector,
		&gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, FullSaveAssociations: true})
	if err != nil {
		panic(err)
	}

	return orm
}

func (s SqlDatabase) Migrate(object interface{}) error {
	// do not migrate invoice, if lightning is not enabled
	if object != nil {
		err := s.db.AutoMigrate(object)
		if err != nil {
			panic(err)
		}
	}
	return nil
}
func (s SqlDatabase) ProofsUsed(in []string) []cashu.Proof {
	proofs := make([]cashu.Proof, 0)
	s.db.Where(in).Find(&proofs)
	return proofs
}

// GetUsedProofs reads all proofs from db
func (s SqlDatabase) GetUsedProofs() []cashu.Proof {
	proofs := make([]cashu.Proof, 0)
	s.db.Find(&proofs)
	return proofs
}

// InvalidateProof will write proof to db
func (s SqlDatabase) StoreProof(p cashu.Proof) error {
	log.WithFields(p.Log()).Info("invalidating proof")
	return s.db.Create(&p).Error
}

func (s SqlDatabase) StoreScript(p cashu.P2SHScript) error {
	log.Info("storing script")
	return s.db.Create(&p).Error
}

// StorePromise will write promise to db
func (s SqlDatabase) StorePromise(p cashu.Promise) error {
	log.WithFields(p.Log()).Info("storing promise")
	return s.db.Create(&p).Error
}

// StoreLightningInvoice will store lightning invoice in db
func (s SqlDatabase) StoreLightningInvoice(i lightning.Invoice) error {
	log.WithFields(i.Log()).Info("storing lightning invoice")
	return s.db.Create(i).Error
}

// GetLightningInvoice reads lighting invoice from db
func (s SqlDatabase) GetLightningInvoice(hash string) (lightning.Invoice, error) {
	invoice := cashu.CreateInvoice()
	invoice.SetHash(hash)
	tx := s.db.Find(invoice)
	log.WithFields(cashuLog.WithLoggable(invoice, tx.Error)).Info("storing lightning invoice")
	return invoice, tx.Error
}

// UpdateLightningInvoice updates lightning invoice in db
func (s SqlDatabase) UpdateLightningInvoice(hash string, issued bool) error {
	i, err := s.GetLightningInvoice(hash)
	if err != nil {
		return err
	}
	i.SetIssued(issued)
	return s.db.Save(i).Error
}
