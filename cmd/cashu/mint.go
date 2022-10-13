package main

import (
	"github.com/gohumble/cashu-feni/cashu"
	"github.com/gohumble/cashu-feni/ledger"
	"net/http"
)

// todo -- this responses are currently not used.
type Mint struct {
	HttpServer *http.Server
	Ledger     *ledger.Ledger
}

type MintResponse cashu.BlindedMessages
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
	Fst string
	Snd string
}
type GetMintResponse struct {
	Pr   string
	Hash string
}

type MeltRequest struct {
	Proofs  cashu.Proofs
	Amount  int64
	Invoice string
}
type CheckRequest struct {
	Proofs cashu.Proofs
}

type CheckFeesResponse struct {
	Fee int64 `json:"fee"`
}
type CheckFeesRequest struct {
	Pr string `json:"pr"`
}
type SplitRequest struct {
	Proofs  cashu.Proofs `json:"proofs"`
	Amount  int64        `json:"amount"`
	Outputs struct {
		BlindedMessages cashu.BlindedMessages `json:"blinded_messages"`
	} `json:"outputs"`
	// todo -- remove output data in future version. This is only used for backward compatibility
	// check https://github.com/callebtc/cashu/pull/20
	OutputData struct {
		BlindedMessages cashu.BlindedMessages `json:"blinded_messages"`
	} `json:"output_data" swaggerignore:"true"`
}
