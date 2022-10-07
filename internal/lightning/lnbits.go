package lightning

import (
	"fmt"
	"time"

	"github.com/imroc/req"
)

var LnbitsClient *Client

// NewClient returns a new lnbits api client. Pass your API key and url here.
func NewClient(key, url string) *Client {
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

// Invoice creates an invoice associated with this wallet.
func (c *Client) CreateInvoice(params InvoiceParams) (lntx Invoice, err error) {
	resp, err := req.Post(c.url+"/api/v1/payments", c.header, req.BodyJSON(&params))
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

	err = resp.ToJSON(&lntx)
	if err == nil {
		lntx.Amount = params.Amount
	}
	return
}

// Pay pays a given invoice with funds from the wallet.
func Pay(params PaymentParams, c *Client) (wtx Invoice, err error) {
	r := req.New()
	r.SetTimeout(time.Hour * 24)
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
func (c Client) GetPaymentStatus(payment_hash string) (payment LNbitsPayment, err error) {
	return c.GetInvoiceStatus(payment_hash)
}

// Payment state of a payment
func (c Client) GetInvoiceStatus(payment_hash string) (payment LNbitsPayment, err error) {
	resp, err := req.Get(c.url+fmt.Sprintf("/api/v1/payments/%s", payment_hash), c.header, nil)
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

	err = resp.ToJSON(&payment)
	return
}
