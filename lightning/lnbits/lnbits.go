package lnbits

import (
	"fmt"
	"github.com/gohumble/cashu-feni/lightning"
	"time"

	"github.com/imroc/req"
)

// NewClient returns a new lnbits api client. Pass your API key and url here.
func NewClient(key, url string) lightning.Client {
	return &Client{
		url: url,
		// info: this header holds the ADMIN key for the entire API
		// it can be used to create wallets for example
		// if you want to check the balance of a user, use w.Inkey
		// if you want to make a payment, use w.Adminkey
		header: req.Header{
			"Content-Type": "application/json",
			"Accept":       "application/json",
			"X-Api-Key":    key,
		},
	}
}

// Info returns wallet information
func (c Client) Status() (wtx Wallet, err error) {
	resp, err := req.Get(c.url+"/api/v1/wallet", c.header, nil)
	if err != nil {
		return
	}

	if resp.Response().StatusCode >= 300 {
		var reqErr Error
		err = resp.ToJSON(&reqErr)
		if err != nil {
			return
		}
		err = reqErr
		return
	}

	err = resp.ToJSON(&wtx)
	return
}

func NewInvoice() lightning.Invoice {
	return &Invoice{}
}

// Invoice creates an invoice associated with this wallet.
func (c *Client) CreateInvoice(amount int64, memo string) (lightning.Invoice, error) {
	params := InvoiceParams{Amount: amount, Memo: memo}
	resp, err := req.Post(c.url+"/api/v1/payments", c.header, req.BodyJSON(&params))
	if err != nil {
		return nil, err
	}

	if resp.Response().StatusCode >= 300 {
		var reqErr Error
		err = resp.ToJSON(&reqErr)
		if err != nil {
			return nil, err
		}
		err = reqErr
		return nil, err
	}
	invoice := &Invoice{}
	err = resp.ToJSON(invoice)
	if err == nil {
		invoice.SetAmount(params.Amount)
		return invoice, nil
	}
	return nil, err
}

// Pay pays a given invoice with funds from the wallet.
func (c *Client) Pay(paymentRequest string) (wtx lightning.Invoice, err error) {
	r := req.New()
	r.SetTimeout(time.Hour * 24)
	params := PaymentParams{Out: true, Bolt11: paymentRequest}
	resp, err := r.Post(c.url+"/api/v1/payments", c.header, req.BodyJSON(&params))
	if err != nil {
		return
	}

	if resp.Response().StatusCode >= 300 {
		var reqErr Error
		err = resp.ToJSON(&reqErr)
		if err != nil {
			return
		}
		err = reqErr
		return
	}

	err = resp.ToJSON(&wtx)
	return
}

// Payment state of a payment
func (c Client) GetPaymentStatus(payment_hash string) (payment lightning.Payment, err error) {
	return c.InvoiceStatus(payment_hash)
}

// Payment state of a payment
func (c Client) InvoiceStatus(paymentHash string) (lightning.Payment, error) {
	resp, err := req.Get(c.url+fmt.Sprintf("/api/v1/payments/%s", paymentHash), c.header, nil)
	if err != nil {
		return nil, err
	}

	if resp.Response().StatusCode >= 300 {
		var reqErr Error
		err = resp.ToJSON(&reqErr)
		if err != nil {
			return nil, err
		}
		err = reqErr
		return nil, err
	}
	payment := LNbitsPayment{}
	return &payment, resp.ToJSON(&payment)
}
