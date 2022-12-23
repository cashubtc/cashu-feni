package api

import (
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/mint"
	"github.com/nbd-wtf/go-nostr"
	"net/http"
)

// todo -- this responses are currently not used.
type Api struct {
	HttpServer *http.Server
	Mint       *mint.Mint
	Nostr      *nostr.RelayPool
}

type MintResponse []cashu.BlindedSignature

type MintRequest struct {
	BlindedMessages cashu.BlindedMessages `json:"blinded_messages"`
}
type MeltResponse struct {
	Paid     bool   `json:"paid"`
	Preimage string `json:"preimage"`
}
type GetKeysResponse map[int]string
type CheckResponse map[string]bool
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
	Proofs  cashu.Proofs `json:"proofs"`
	Invoice string       `json:"invoice"`
}
type CheckRequest struct {
	Proofs cashu.Proofs `json:"proofs"`
}

type CheckFeesResponse struct {
	Fee uint64 `json:"fee"`
}
type CheckFeesRequest struct {
	Pr string `json:"pr"`
}
type SplitRequest struct {
	Proofs  cashu.Proofs `json:"proofs"`
	Amount  uint64       `json:"amount"`
	Outputs struct {
		BlindedMessages cashu.BlindedMessages `json:"blinded_messages"`
	} `json:"outputs,omitempty"`
	// todo -- remove output data in future version. This is only used for backward compatibility
	// check https://github.com/callebtc/cashu/pull/20
	OutputData *struct {
		BlindedMessages cashu.BlindedMessages `json:"blinded_messages"`
	} `json:"output_data,omitempty" swaggerignore:"true"`
}
