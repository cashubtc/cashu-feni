package cashu

import (
	"encoding/hex"
	"github.com/cashubtc/cashu-feni/lightning"
	"github.com/cashubtc/cashu-feni/lightning/lnbits"
	cashuLog "github.com/cashubtc/cashu-feni/log"
	"github.com/google/uuid"
	"strings"
	"time"
)

type ProofsUsed struct {
	Amount   uint64 `json:"amount"`
	Secret   string `json:"secret" gorm:"primaryKey"`
	C        string `json:"C"`
	TimeUsed time.Time
}
type ProofStatus int

const (
	ProofStatusSpent ProofStatus = iota
	ProofStatusPending
	ProofStatusReserved
)

type Proof struct {
	Id           string      `json:"id"`
	Amount       uint64      `json:"amount"`
	Secret       string      `json:"secret" gorm:"primaryKey"`
	C            string      `json:"C"`
	Status       ProofStatus `json:"-"`
	Reserved     bool        `json:"-,omitempty"`
	Script       *P2SHScript `gorm:"-" json:"script,omitempty" structs:"Script,omitempty"`
	SendId       uuid.UUID   `json:"-,omitempty" structs:"SendId,omitempty"`
	TimeCreated  time.Time   `json:"-,omitempty" structs:"TimeCreated,omitempty"`
	TimeReserved time.Time   `json:"-,omitempty" structs:"TimeReserved,omitempty"`
}

func IsPay2ScriptHash(s string) bool {
	return len(strings.Split(s, "P2SH:")) == 2
}
func (p Proof) Log() map[string]interface{} {
	return cashuLog.ToMap(p)
}

type P2SHScript struct {
	Script    string `json:"script"`
	Signature string `json:"signature"`
	Address   string `json:"address"`
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
	Id     string `json:"id"`
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
func CreateInvoice() lightning.Invoicer {
	if lightning.Config.Lightning.Enabled {
		return lnbits.NewInvoice()
	}
	return nil
}

type Mint struct {
	Url     string   `json:"url"`
	KeySets []string `json:"ks"`
}
type MintResponse struct {
	Promises []BlindedSignature `json:"promises"`
}

type MintRequest struct {
	Outputs BlindedMessages `json:"outputs"`
}
type MeltResponse struct {
	Paid     bool               `json:"paid"`
	Preimage string             `json:"preimage"`
	Change   []BlindedSignature `json:"change,omitempty"`
}
type GetKeysResponse map[int]string
type SplitResponse struct {
	Fst []BlindedSignature `json:"fst"`
	Snd []BlindedSignature `json:"snd"`
}
type GetKeySetsResponse struct {
	KeySets []string `json:"keysets"`
}
type GetMintResponse struct {
	Pr   string `json:"pr"`
	Hash string `json:"hash"`
}

type MeltRequest struct {
	Proofs  Proofs           `json:"proofs"`
	Pr      string           `json:"pr"`
	Outputs []BlindedMessage `json:"outputs,omitempty"`
}
type CheckSpendableRequest struct {
	Proofs Proofs `json:"proofs"`
}
type CheckSpendableResponse struct {
	Spendable []bool `json:"spendable"`
}

type CheckFeesResponse struct {
	Fee uint64 `json:"fee"`
}
type CheckFeesRequest struct {
	Pr string `json:"pr"`
}

type SplitRequest struct {
	Proofs  Proofs           `json:"proofs"`
	Amount  uint64           `json:"amount"`
	Outputs []BlindedMessage `json:"outputs"`
}
