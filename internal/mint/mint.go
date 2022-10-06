package mint

import (
	"github.com/gohumble/cashu-feni/internal/core"
	"net/http"
)

type Mint struct {
	HttpServer *http.Server
	ledger     *Ledger
}

type MintResponse core.BlindedMessages
type MintRequest struct {
	BlindedMessages core.BlindedMessages `json:"blinded_messages"`
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
	Proofs  core.Proofs
	Amount  int64
	Invoice string
}
type CheckRequest struct {
	Proofs core.Proofs
}

type CheckFeesResponse struct {
	Fee int64 `json:"fee"`
}
type CheckFeesRequest struct {
	Pr string `json:"pr"`
}
type SplitRequest struct {
	Proofs  core.Proofs `json:"proofs"`
	Amount  int64       `json:"amount"`
	Outputs struct {
		BlindedMessages core.BlindedMessages `json:"blinded_messages"`
	} `json:"outputs"`
	// todo -- remove output data in future version. This is only used for backward compatibility
	// check https://github.com/callebtc/cashu/pull/20
	OutputData struct {
		BlindedMessages core.BlindedMessages `json:"blinded_messages"`
	} `json:"output_data" swaggerignore:"true"`
}
