package mint

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/gohumble/cashu-feni/internal/cashu"
	"github.com/gohumble/cashu-feni/internal/core"
	"github.com/gohumble/cashu-feni/internal/lightning"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"strconv"
	"time"
)

const (
	ResourceSwaggerPathPrefix = "/swagger/"
)

func New() *Mint {
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", Config.Mint.Host, Config.Mint.Port),
		WriteTimeout: 90 * time.Second,
		ReadTimeout:  90 * time.Second,
	}
	m := &Mint{
		HttpServer: srv,
		ledger:     NewLedger(Config.Mint.PrivateKey),
	}
	lightning.LnbitsClient = lightning.NewClient(lightning.Config.Lnbits.AdminKey, lightning.Config.Lnbits.Url)
	m.HttpServer.Handler = m.newRouter()
	log.Trace("created mint server")
	return m
}
func responseError(w http.ResponseWriter, err cashu.ErrorResponse) {
	log.WithFields(log.Fields{"error.message": err.Error(), "code": err.Code}).Error(err)
	response := err.String()
	_, writeError := fmt.Fprintf(w, response)
	if writeError != nil {
		log.WithFields(log.Fields{"error.message": writeError.Error()}).Error(writeError)
	}

}
func Use(h http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}
	return h
}

// LoggingMiddleware will log all incoming requests
func LoggingMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{"resource": r.URL.String()}).Infof("incoming request")
		h.ServeHTTP(w, r)
	}
}

func (m Mint) StartServer() {
	if Config.Mint.Tls.Enabled {
		log.Println(m.HttpServer.ListenAndServeTLS(Config.Mint.Tls.CertFile, Config.Mint.Tls.KeyFile))
	} else {
		log.Println(m.HttpServer.ListenAndServe())
	}
}
func (m Mint) newRouter() *mux.Router {
	router := mux.NewRouter()
	// route to receive mint public keys
	router.HandleFunc("/keys", Use(m.getKeys, LoggingMiddleware)).Methods(http.MethodGet)
	// route to get mint (create tokens)
	router.HandleFunc("/mint", Use(m.getMint, LoggingMiddleware)).Methods(http.MethodGet)
	// route to real mint (with LIGHTNING enabled)
	router.HandleFunc("/mint", Use(m.mint, LoggingMiddleware)).Methods(http.MethodPost)
	// route to burn / melt a tx
	router.HandleFunc("/melt", Use(m.melt, LoggingMiddleware)).Methods(http.MethodPost)
	// route to check spendable proofs
	router.HandleFunc("/check", Use(m.check, LoggingMiddleware)).Methods(http.MethodPost)
	// route to check routing fees
	router.HandleFunc("/checkfees", Use(m.checkFee, LoggingMiddleware)).Methods(http.MethodPost)
	// route to split proofs (send money)
	router.HandleFunc("/split", Use(m.split, LoggingMiddleware)).Methods(http.MethodGet, http.MethodPost)
	appendSwaggoHandler(router)
	return router
}

// appendSwaggoHandler will append routes for the documentation to the router
func appendSwaggoHandler(router *mux.Router) {
	router.PathPrefix(ResourceSwaggerPathPrefix).Handler(httpSwagger.Handler(
		httpSwagger.URL(Config.DocReference), //The url pointing to API definition"
	))
}

// checkFee checks fee for payment
// @Summary Check Fee
// @Description Check fees for lightning payment.
// @Produce  json
// @Success 200 {object} CheckFeesResponse
// @Failure 500 {object} ErrorResponse
// @Router /checkfees [post]
// @Param CheckFeesRequest body CheckFeesRequest true "Model containing lightning invoice"
// @Tags POST
func (m Mint) checkFee(w http.ResponseWriter, r *http.Request) {
	feesRequest := CheckFeesRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&feesRequest)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	fee, err := m.ledger.checkFees(feesRequest.Pr)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	response := CheckFeesResponse{Fee: fee / 1000}
	res, err := json.Marshal(response)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	_, writeError := fmt.Fprintf(w, string(res))
	if writeError != nil {
		log.WithFields(log.Fields{"error.message": writeError.Error()}).Error(writeError)
	}
}

// getMint is the http handler function for GET /mint
// @Summary Request Mint
// @Description Requests the minting of tokens belonging to a paid payment request.
// @Produce  json
// @Success 200 {object} GetMintResponse
// @Failure 500 {object} ErrorResponse
// @Router /mint [get]
// @Param        amount    query     string  false  "amount of the mint"
// @Tags GET
func (m Mint) getMint(w http.ResponseWriter, r *http.Request) {
	amount := r.URL.Query().Get("amount")
	ai, err := strconv.Atoi(amount)
	if err != nil {
		log.Errorf("error checking amount")
	}
	invoice, err := requestMint(lightning.LnbitsClient, int64(ai))
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	log.WithField("invoice", invoice).Infof("created lightning invoice")
	_, err = fmt.Fprintf(w, `{"pr": "%s", "hash": "%s"}`, invoice.Pr, invoice.Hash)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
}

// mint is the http handler function for POST /mint
// @Summary Mint
// @Description Requests the minting of tokens belonging to a paid payment request.
// @Description
// @Description Parameters: pr: payment_request of the Lightning paid invoice.
// @Description
// @Description Body (JSON): payloads: contains a list of blinded messages waiting to be signed.
// @Description
// @Description NOTE:
// @Description
// @Description * This needs to be replaced by the preimage otherwise someone knowing the payment_request can request the tokens instead of the rightful owner.
// @Description * The blinded message should ideally be provided to the server before payment in the GET /mint endpoint so that the server knows to sign only these tokens when the invoice is paid.
// @Produce  json
// @Success 200 {object} MintResponse
// @Failure 500 {object} ErrorResponse
// @Router /mint [post]
// @Param core.BlindedMessages body core.BlindedMessages true "Model containing proofs to mint"
// @Param        payment_hash    query     string  false  "payment hash for the mint"
// @Tags POST
func (m Mint) mint(w http.ResponseWriter, r *http.Request) {
	pr := r.URL.Query().Get("payment_hash")
	amounts := make([]int64, 0)
	B_s := make([]*secp256k1.PublicKey, 0)
	blindedMessages := MintRequest{BlindedMessages: make(core.BlindedMessages, 0)}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&blindedMessages)
	if err != nil {
		panic(err)
	}
	for _, msg := range blindedMessages.BlindedMessages {
		amounts = append(amounts, msg.Amount)
		hkey := make([]byte, 0)
		hkey, err = hex.DecodeString(msg.B_)
		publicKey, err := secp256k1.ParsePubKey(hkey)
		if err != nil {
			responseError(w, cashu.NewErrorResponse(err))
			return
		}
		B_s = append(B_s, publicKey)
	}
	promises, err := m.ledger.mint(lightning.LnbitsClient, B_s, amounts, pr)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	data, err := json.Marshal(promises)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	_, err = fmt.Fprintf(w, string(data))
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
}

// melt is the http handler function for POST /melt
// @Summary Melt
// @Description Requests tokens to be destroyed and sent out via Lightning.
// @Produce  json
// @Success 200 {object} MeltResponse
// @Failure 500 {object} ErrorResponse
// @Router /melt [post]
// @Param MeltRequest body MeltRequest true "Model containing proofs to melt"
// @Tags POST
func (m Mint) melt(w http.ResponseWriter, r *http.Request) {
	payload := MeltRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		panic(err)
	}
	ok, preimage, err := m.ledger.melt(payload.Proofs, payload.Amount, payload.Invoice)
	if err != nil {
		log.WithFields(log.Fields{"error.message": err.Error()}).Errorf("error in melt")
	}
	response := MeltResponse{Paid: ok, Preimage: preimage}
	res, err := json.Marshal(response)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	_, err = fmt.Fprintf(w, string(res))
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
}

// getKeys is the http handler function for GET /keys
// @Summary Keys
// @Description Get the public keys of the mint
// @Produce  json
// @Success 200 {object} GetKeysResponse
// @Failure 500 {object} ErrorResponse
// @Router /keys [get]
// @Tags GET
func (m Mint) getKeys(w http.ResponseWriter, r *http.Request) {
	key, err := json.Marshal(m.ledger.GetPublicKeys())
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	w.WriteHeader(200)
	_, err = fmt.Fprintf(w, string(key))
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
}

// check is the http handler function for POST /check
// @Summary Check spendable
// @Description Get currently available public keys
// @Produce  json
// @Success 200 {object} CheckResponse
// @Failure 500 {object} ErrorResponse
// @Router /check [post]
// @Param CheckRequest body CheckRequest true "Model containing proofs to check"
// @Tags POST
func (m Mint) check(w http.ResponseWriter, r *http.Request) {
	payload := CheckRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
	}
	spendable := m.ledger.checkSpendables(payload.Proofs)
	res, err := json.Marshal(spendable)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	_, err = fmt.Fprintf(w, string(res))
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
}

// split is the http handler function for POST /split
// @Summary Split your proofs
// @Description Requetst a set of tokens with amount "total" to be split into two newly minted sets with amount "split" and "total-split".
// @Produce  json
// @Success 200 {object} SplitResponse
// @Failure 500 {object} ErrorResponse
// @Router /split [post]
// @Param SplitRequest body SplitRequest true "Model containing proofs to split"
// @Tags POST
func (m Mint) split(w http.ResponseWriter, r *http.Request) {
	payload := SplitRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		panic(err)
	}
	proofs := payload.Proofs
	amount := payload.Amount
	// todo -- remove this mapping from output_data to outputs.
	// https://github.com/callebtc/cashu/pull/20
	if payload.Outputs.BlindedMessages == nil {
		payload.Outputs.BlindedMessages = payload.OutputData.BlindedMessages
	}
	outputs := payload.Outputs
	fstPromise, sendPromise, err := m.ledger.split(proofs, amount, outputs.BlindedMessages)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	fstb, err := json.Marshal(fstPromise)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	sstb, err := json.Marshal(sendPromise)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	_, err = fmt.Fprintf(w, fmt.Sprintf(`{"fst": %s, "snd": %s}`, string(fstb), string(sstb)))
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
}
