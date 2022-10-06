package lightning

import (
	"github.com/gohumble/cashu-feni/internal/core"
	"github.com/imroc/req"
)

type Client struct {
	header     req.Header
	url        string
	AdminKey   string
	InvoiceKey string
}

type InvoiceParams struct {
	Out                 bool   `json:"out"`                            // must be True if invoice is payed, False if invoice is received
	Amount              int64  `json:"amount"`                         // amount in Satoshi
	Memo                string `json:"memo,omitempty"`                 // the invoice memo.
	Webhook             string `json:"webhook,omitempty"`              // the webhook to fire back to when payment is received.
	DescriptionHash     string `json:"description_hash,omitempty"`     // the invoice description hash.
	UnhashedDescription string `json:"unhashed_description,omitempty"` // the unhashed invoice description.
}

type PaymentParams struct {
	Out          bool   `json:"out"`
	Bolt11       string `json:"bolt11"`
	FeeLimitMSat int64  `json:"feeLimitMSat"`
}
type PayParams struct {
	// the BOLT11 payment request you want to pay.
	PaymentRequest string `json:"payment_request"`

	// custom data you may want to associate with this invoice. optional.
	PassThru map[string]interface{} `json:"passThru"`
}

type TransferParams struct {
	Memo         string `json:"memo"`           // the transfer description.
	NumSatoshis  int64  `json:"num_satoshis"`   // the transfer amount.
	DestWalletId string `json:"dest_wallet_id"` // the key or id of the destination
}

type Error struct {
	Detail string `json:"detail"`
}

func (err Error) Error() string {
	return err.Detail
}

type Wallet struct {
	ID       string `json:"id" gorm:"id"`
	Adminkey string `json:"adminkey"`
	Inkey    string `json:"inkey"`
	Balance  int64  `json:"balance"`
	Name     string `json:"name"`
	User     string `json:"user"`
}

type Payment struct {
	CheckingID    string      `json:"checking_id"`
	Pending       bool        `json:"pending"`
	Amount        int64       `json:"amount"`
	Fee           int64       `json:"fee"`
	Memo          string      `json:"memo"`
	Time          int         `json:"time"`
	Bolt11        string      `json:"bolt11"`
	Preimage      string      `json:"preimage"`
	PaymentHash   string      `json:"payment_hash"`
	Extra         struct{}    `json:"extra"`
	WalletID      string      `json:"wallet_id"`
	Webhook       interface{} `json:"webhook"`
	WebhookStatus interface{} `json:"webhook_status"`
}
type Invoice struct {
	Amount int64  `json:"amount"`
	Pr     string `json:"payment_request"`
	Hash   string `json:"payment_hash" gorm:"primaryKey"`
	Issued bool   `json:"issued"`
}

func (i Invoice) String() string {
	return core.ToJson(i)
}

type LNbitsPayment struct {
	Paid     bool    `json:"paid"`
	Preimage string  `json:"preimage"`
	Details  Payment `json:"details,omitempty"`
}

type Payments []Payment
