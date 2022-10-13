package lightning

import (
	"fmt"
)

type Invoice interface {
	fmt.Stringer
	SetHash(h string)
	GetHash() string

	SetIssued(i bool)
	IsIssued() bool

	SetAmount(a int64)
	GetAmount() int64

	GetPaymentRequest() string
}

type Payment interface {
	IsPaid() bool
	GetPreimage() string
}

type Client interface {
	InvoiceStatus(paymentHash string) (Payment, error)
	Pay(paymentRequest string) (Invoice, error)
	CreateInvoice(amount int64, memo string) (Invoice, error)
}
