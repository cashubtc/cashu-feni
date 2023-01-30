package api

import (
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/mint"
	"net/http"
)

// todo -- this responses are currently not used.
type Api struct {
	HttpServer *http.Server
	Mint       *mint.Mint
}
type Mint struct {
	Url     string   `json:"url"`
	KeySets []string `json:"ks"`
}
type MintResponse struct {
	Promises []cashu.BlindedSignature `json:"promises"`
}

type MintRequest struct {
	Outputs cashu.BlindedMessages `json:"outputs"`
}
type MeltResponse struct {
	Paid     bool   `json:"paid"`
	Preimage string `json:"preimage"`
}
type GetKeysResponse map[int]string
type SplitResponse struct {
	Fst []cashu.BlindedSignature `json:"fst"`
	Snd []cashu.BlindedSignature `json:"snd"`
}
type GetKeySetsResponse struct {
	KeySets []string `json:"keysets"`
}
type GetMintResponse struct {
	Pr   string
	Hash string
}

type MeltRequest struct {
	Proofs cashu.Proofs `json:"proofs"`
	Pr     string       `json:"pr"`
}
type CheckSpendableRequest struct {
	Proofs cashu.Proofs `json:"proofs"`
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
	Proofs  cashu.Proofs           `json:"proofs"`
	Amount  uint64                 `json:"amount"`
	Outputs []cashu.BlindedMessage `json:"outputs"`
}
