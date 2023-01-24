package feni

import (
	"encoding/hex"
	"fmt"
	"github.com/cashubtc/cashu-feni/api"
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/lightning"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/imroc/req"
	"strings"
	"time"
)

type Client struct {
	Url string
}

//var WalletClient *Client

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

func parseKeys(resp *req.Resp, err error) (map[uint64]*secp256k1.PublicKey, error) {
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
func (c Client) Keys() (map[uint64]*secp256k1.PublicKey, error) {
	return parseKeys(req.Get(fmt.Sprintf("%s/keys", c.Url)))
}
func (c Client) KeysForKeySet(kid string) (map[uint64]*secp256k1.PublicKey, error) {
	kid = strings.ReplaceAll(strings.ReplaceAll(kid, "/", "_"), "+", "-")
	return parseKeys(req.Get(fmt.Sprintf("%s/keys/%s", c.Url, kid)))
}

func (c Client) KeySets() (*api.GetKeySetsResponse, error) {
	resp, err := req.Get(fmt.Sprintf("%s/keysets", c.Url))
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
	resp, err := req.Post(fmt.Sprintf("%s/check", c.Url), req.BodyJSON(data))
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
	resp, err := req.Post(fmt.Sprintf("%s/split", c.Url), req.BodyJSON(data))
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
	resp, err := req.Post(fmt.Sprintf("%s/melt", c.Url), req.BodyJSON(data))
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
	requestUrl := fmt.Sprintf("%s/mint", c.Url)
	if paymentHash != "" {
		requestUrl += fmt.Sprintf("?payment_hash=%s", paymentHash)
	}
	resp, err := req.Post(requestUrl, req.BodyJSON(data))
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
	resp, err := req.Get(fmt.Sprintf("%s/mint?amount=%d", c.Url, amount))
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
	resp, err := req.Post(fmt.Sprintf("%s/checkfees", c.Url), req.BodyJSON(CheckFeesRequest))
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
