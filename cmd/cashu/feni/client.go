package feni

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/cashubtc/cashu-feni/api"
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/lightning"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/imroc/req"
	"github.com/nbd-wtf/go-nostr"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

type Client struct {
	mintPublicKey   string
	nostr           *nostr.RelayPool
	url             string
	clientPublicKey string
}

var WalletClient *Client

func NewClient() *Client {
	WalletClient = &Client{
		nostr:         nostr.NewRelayPool(),
		mintPublicKey: "2a4c4cce1523ffab2b727d6f0a0ce983dbe6868dde850ad65874961d62bca69d",
	}
	err := api.ConnectNostr(WalletClient.nostr, []string{"ws://91.237.88.218:2700/"})
	if err != nil {
		panic(err)
	}
	publicKey, err := nostr.GetPublicKey(*WalletClient.nostr.SecretKey)
	if err != nil {
		return nil
	}
	WalletClient.clientPublicKey = publicKey
	req.SetClient(&http.Client{
		Transport: WalletClient,
	})
	api.SubscribeNostrEvents(WalletClient.nostr, api.GetSubscriptionFilter(WalletClient.clientPublicKey), func(message nostr.Event) {
		if message.Tags.ContainsAny("p", []string{WalletClient.clientPublicKey}) {
			rs, err := api.ComputeSharedSecret(*WalletClient.nostr.SecretKey, message.PubKey)
			if err != nil {
				return
			}
			resp, err := api.Decrypt(message.Content, rs)
			if err != nil {
				return
			}
			response := http.Response{
				Body:          io.NopCloser(bytes.NewBufferString(resp)),
				Status:        "200 OK",
				StatusCode:    200,
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				ContentLength: int64(len(message.Content)),
				Request:       r,
				Header:        make(http.Header, 0),
			}
			log.WithField("content", message.Content).Infof("received mint response")
		}
	})
	return WalletClient
}

func (s *Client) RoundTrip(r *http.Request) (*http.Response, error) {
	request, err := httputil.DumpRequestOut(r, true)
	if err != nil {
		return nil, err
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	var response http.Response

	api.PublishNostrEvents(string(request), s.mintPublicKey, s.nostr)
	wg.Wait()
	return &response, nil
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
	resp, err := req.Get(fmt.Sprintf("%s/keys", c.url))
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
	resp, err := req.Get(fmt.Sprintf("%s/keysets", c.url))
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
	resp, err := req.Post(fmt.Sprintf("%s/check", c.url), req.BodyJSON(data))
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
	resp, err := req.Post(fmt.Sprintf("%s/split", c.url), req.BodyJSON(data))
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
	resp, err := req.Post(fmt.Sprintf("%s/melt", c.url), req.BodyJSON(data))
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
	resp, err := req.Get(fmt.Sprintf("%s/mint?amount=%d", c.url, amount))
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
	resp, err := req.Post(fmt.Sprintf("%s/checkfees", c.url), req.BodyJSON(CheckFeesRequest))
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
