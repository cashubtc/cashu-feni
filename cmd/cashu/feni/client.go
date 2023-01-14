package feni

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/cashubtc/cashu-feni/api"
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/lightning"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	httpNostr "github.com/gohumble/go-nostr-http"
	"github.com/imroc/req"
	"github.com/nbd-wtf/go-nostr"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Client struct {
	privateKey string
	client     *http.Client
	url        string
	header     req.Header
}

var WalletClient *Client

func NewClient() *Client {
	relay, err := nostr.RelayConnect(context.Background(), "ws://91.237.88.218:2700")
	if err != nil {
		panic(err)
	}
	WalletClient = &Client{
		url:    fmt.Sprintf("%s:%s", Config.MintServerHost, Config.MintServerPort),
		header: req.Header{"NOSTR-TO-PUBLIC-KEY": "6af64dc47572d2c55804fac39a4f3a120a0e97c634fe8a37ea58437bebf804fa"},
	}
	privateKet, err := api.GetPrivateKey()
	if err != nil {
		panic(err)
	}
	publicKey, err := nostr.GetPublicKey(privateKet)
	if err != nil {
		return nil
	}
	log.WithField("publicKey", publicKey).Infof("Nostr public key")
	httpNostr.Configuration.PrivateKey = privateKet
	WalletClient.client = httpNostr.NewClient(relay, publicKey)
	req.SetClient(WalletClient.client)

	return WalletClient
}

func checkError(resp *req.Resp) error {
	if resp.Response().StatusCode >= 300 {
		var reqErr cashu.ErrorResponse
		err := resp.ToJSON(&reqErr)
		if err != nil {
			return nil
		}
		err = reqErr
		return err
	}
	return nil
}
func (c Client) Keys() (map[uint64]*secp256k1.PublicKey, error) {
	resp, err := req.Get(fmt.Sprintf("%s/keys", c.url), c.header)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	response := make(map[uint64]string)
	keys := make(map[uint64]*secp256k1.PublicKey)
	err = resp.ToJSON(&response)
	for u, s := range response {
		h, err := hex.DecodeString(s)
		if err != nil {
			panic(err)
		}
		key, err := secp256k1.ParsePubKey(h)
		if err != nil {
			return nil, err
		}
		keys[u] = key
	}
	return keys, nil
}

func (c Client) KeySets() (*api.GetKeySetsResponse, error) {
	resp, err := req.Get(fmt.Sprintf("%s/keysets", c.url), c.header)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	keySets := api.GetKeySetsResponse{}
	err = resp.ToJSON(&keySets)
	return &keySets, nil
}

func (c Client) Check(data api.CheckRequest) (api.CheckResponse, error) {
	resp, err := req.Post(fmt.Sprintf("%s/check", c.url), req.BodyJSON(data), c.header)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	check := api.CheckResponse{}
	err = resp.ToJSON(&check)
	return check, nil
}

func (c Client) Split(data api.SplitRequest) (*api.SplitResponse, error) {
	resp, err := req.Post(fmt.Sprintf("%s/split", c.url), req.BodyJSON(data), c.header)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	split := api.SplitResponse{}
	err = resp.ToJSON(&split)
	return &split, nil
}
func (c Client) Melt(data api.MeltRequest) (*api.MeltResponse, error) {
	resp, err := req.Post(fmt.Sprintf("%s/melt", c.url), req.BodyJSON(data), c.header)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	melt := api.MeltResponse{}
	err = resp.ToJSON(&melt)
	return &melt, nil
}

func (c Client) Mint(data api.MintRequest, paymentHash string) (*api.MintResponse, error) {
	requestUrl := fmt.Sprintf("%s/mint", c.url)
	if paymentHash != "" {
		requestUrl += fmt.Sprintf("?payment_hash=%s", paymentHash)
	}
	resp, err := req.Post(requestUrl, req.BodyJSON(data), c.header)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	mint := api.MintResponse{}
	err = resp.ToJSON(&mint)
	return &mint, nil
}
func (c Client) GetMint(amount int64) (lightning.Invoicer, error) {
	resp, err := req.Get(fmt.Sprintf("%s/mint?amount=%d", c.url, amount), c.header)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	mint := api.GetMintResponse{}
	err = resp.ToJSON(&mint)
	invoice := cashu.CreateInvoice()
	invoice.SetAmount(amount)
	invoice.SetHash(mint.Hash)
	invoice.SetPaymentRequest(mint.Pr)
	invoice.SetTimeCreated(time.Now())
	return invoice, nil
}

func (c Client) CheckFee(CheckFeesRequest api.CheckFeesRequest) (*api.CheckFeesResponse, error) {
	resp, err := req.Post(fmt.Sprintf("%s/checkfees", c.url), req.BodyJSON(CheckFeesRequest), c.header)
	if err != nil {
		return nil, err
	}
	if err = checkError(resp); err != nil {
		return nil, err
	}
	fees := api.CheckFeesResponse{}
	err = resp.ToJSON(&fees)
	return &fees, nil
}
