package lightning

import (
	"fmt"
	cashuLog "github.com/cashubtc/cashu-feni/log"
	"time"
)

// If these interfaces are not implemented, library users of cashu MUST provide the lightning capabilities them self.
// Please be sure of not enabling lightning in the configuration, if you provide lightning services yourself.
// When using cashu as stand alone, this lightning interface could be provided by multiple lighting services like
// LND, CLN, or any abstraction layer like LNBits.

// Invoice should create a lightning invoice somewhere.
type Invoice interface {
	cashuLog.Loggable
	fmt.Stringer      // toJson
	SetHash(h string) // set the payment hash
	GetHash() string  // get the payment hash

	SetPaid(i bool) // SetPaid to true, if lightning invoice was paid

	SetIssued(i bool) // SetIssued to true, if lightning invoice was paid
	IsIssued() bool   // IsIssued returns true, if lightning invoice is paid

	SetAmount(a int64) // SetAmount of the lightning invoice
	GetAmount() int64  // GetAmount of the lightning invoice

	GetPaymentRequest() string // GetPaymentRequest should return the payment request (probably bech encoded)
	SetPaymentRequest(string)  // SetPaymentRequest

	SetTimeCreated(t time.Time)
	SetTimePaid(t time.Time)
}

// Payment should give information about the payment status
type Payment interface {
	IsPaid() bool        // IsPaid must return true, if payment is fulfilled
	GetPreimage() string // GetPreimage must return the preimage of the payment
}

// Client should be able to perform lightning services
type Client interface {
	InvoiceStatus(paymentHash string) (Payment, error)        // InvoiceStatus should return Payment information for a payment hash
	Pay(paymentRequest string) (Invoice, error)               // Pay should pay the payment request.
	CreateInvoice(amount int64, memo string) (Invoice, error) // CreateInvoice should create an invoice for given amount and memo
}
