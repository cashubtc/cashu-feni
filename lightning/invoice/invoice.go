package invoice

import (
	"encoding/json"
	cashuLog "github.com/cashubtc/cashu-feni/log"
	"time"
)

type Invoice struct {
	Amount   int64     `json:"amount"`
	Pr       string    `json:"payment_request"`
	Hash     string    `json:"payment_hash" gorm:"primaryKey"`
	Issued   bool      `json:"issued"`
	Preimage string    `json:"preimage"`
	Paid     bool      `json:"paid"`
	Create   time.Time `json:"time_created"`
	TimePaid time.Time `json:"time_paid"`
}

func (i Invoice) Log() map[string]interface{} {
	return cashuLog.ToMap(i)
}

func (i Invoice) String() string {
	b, err := json.Marshal(i)
	if err != nil {
		return err.Error()
	}
	return string(b)

}
func (i *Invoice) SetHash(h string) {
	i.Hash = h
}

func (i *Invoice) GetHash() string {
	return i.Hash
}
func (i *Invoice) SetTimeCreated(t time.Time) {
	i.Create = t
}
func (i *Invoice) SetTimePaid(t time.Time) {
	i.TimePaid = t
}

func (i *Invoice) SetPaymentRequest(pr string) {
	i.Pr = pr
}
func (i *Invoice) GetPaymentRequest() string {
	return i.Pr
}
func (i *Invoice) SetPaid(paid bool) {
	i.Paid = paid
}
func (i *Invoice) SetIssued(issued bool) {
	i.Issued = issued
}

func (i *Invoice) SetAmount(amount int64) {
	i.Amount = amount
}
func (i *Invoice) GetAmount() int64 {
	return i.Amount
}
func (i *Invoice) IsIssued() bool {
	return i.Issued
}
