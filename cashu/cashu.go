package cashu

import (
	"encoding/hex"
	"encoding/json"
	"github.com/gohumble/cashu-feni/lightning"
	"github.com/gohumble/cashu-feni/lightning/lnbits"
	cashuLog "github.com/gohumble/cashu-feni/log"
	"github.com/google/uuid"
	"time"
)

type ProofsUsed struct {
	Amount   uint64 `json:"amount"`
	Secret   string `json:"secret" gorm:"primaryKey"`
	C        string `json:"C"`
	TimeUsed time.Time
}
type Proof struct {
	Id           string `json:"id"`
	Amount       uint64 `json:"amount"`
	Secret       string `json:"secret" gorm:"primaryKey"`
	C            string `json:"C"`
	Reserved     bool
	Script       *P2SHScript `gorm:"-" json:"script" structs:"Script,omitempty"`
	SendId       uuid.UUID   `json:"send_id,omitempty" structs:"SendId,omitempty"`
	TimeCreated  time.Time   `json:"time_created,omitempty" structs:"TimeCreated,omitempty"`
	TimeReserved time.Time   `json:"time_reserved,omitempty" structs:"TimeReserved,omitempty"`
}


func (p Proof) Log() map[string]interface{} {
	return cashuLog.ToMap(p)
}

type P2SHScript struct {
	Script    string
	Signature string
	Address   string
}

func (p Proof) Decode() ([]byte, error) {
	return hex.DecodeString(p.C)
}

type Proofs []Proof

type Promise struct {
	B_b    string `json:"C_b" gorm:"primaryKey"`
	C_c    string `json:"C_c"`
	Amount uint64 `json:"amount"`
}

func (p Promise) Log() map[string]interface{} {
	return cashuLog.ToMap(p)
}

type BlindedMessages []BlindedMessage

type BlindedMessage struct {
	Amount uint64 `json:"amount"`
	B_     string `json:"B_"`
}
type BlindedSignature struct {
	Amount uint64 `json:"amount"`
	C_     string `json:"C_"`
}

type ErrorResponse struct {
	Err  string `json:"error"`
	Code int    `json:"code"`
}
type ErrorOptions func(err *ErrorResponse)

func WithCode(code int) ErrorOptions {
	return func(err *ErrorResponse) {
		err.Code = code
	}
}
func NewErrorResponse(err error, options ...ErrorOptions) ErrorResponse {
	e := ErrorResponse{
		Err: err.Error(),
	}
	for _, o := range options {
		o(&e)
	}
	return e
}

func (e ErrorResponse) String() string {
	return cashuLog.ToJson(e)
}

func (e ErrorResponse) Error() string {
	return e.Err
}

// CreateInvoice will generate a blank invoice
func CreateInvoice() lightning.Invoice {
	if !lightning.Config.Lightning.Enabled {
		return nil
	}
	if lightning.Config.Lightning.Lnbits != nil {
		return lnbits.NewInvoice()
	}
	return nil
}
