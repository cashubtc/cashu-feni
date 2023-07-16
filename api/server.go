package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/crypto"
	"github.com/cashubtc/cashu-feni/db"
	"github.com/cashubtc/cashu-feni/mint"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	ResourceSwaggerPathPrefix = "/swagger/"
)

// todo -- this responses are currently not used.
type Api struct {
	HttpServer *http.Server
	Mint       *mint.Mint
}

func New() *Api {
	err := Config.Load()
	if err != nil {
		panic(err)
	}
	// currently using sql storage only.
	// this should be extensible for future versions.
	sqlStorage := db.NewSqlDatabase()
	err = sqlStorage.Migrate(cashu.Proof{})
	if err != nil {
		panic(err)
	}
	err = sqlStorage.Migrate(cashu.Promise{})
	if err != nil {
		panic(err)
	}
	err = sqlStorage.Migrate(crypto.KeySet{})
	if err != nil {
		panic(err)
	}
	err = sqlStorage.Migrate(cashu.CreateInvoice())
	if err != nil {
		panic(err)
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", Config.Mint.Host, Config.Mint.Port),
		WriteTimeout: 90 * time.Second,
		ReadTimeout:  90 * time.Second,
	}

	lnBitsClient, err := mint.NewLightningClient()
	if err != nil {
		panic(err)
	}
	m := &Api{
		HttpServer: srv,
		Mint: mint.New(Config.Mint.PrivateKey,
			mint.WithClient(lnBitsClient),
			mint.WithStorage(sqlStorage),
			mint.WithInitialKeySet(Config.Mint.DerivationPath),
		),
	}

	m.HttpServer.Handler = newRouter(m)
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
		log.WithFields(log.Fields{"resource": r.URL.String(), "ip": r.RemoteAddr}).Infof("incoming request")
		h.ServeHTTP(w, r)
	}
}

func (api Api) StartServer() {
	if Config.Mint.Tls.Enabled {
		log.Println(api.HttpServer.ListenAndServeTLS(Config.Mint.Tls.CertFile, Config.Mint.Tls.KeyFile))
	} else {
		log.Println(api.HttpServer.ListenAndServe())
	}
}
func newRouter(a *Api) *mux.Router {
	router := mux.NewRouter()
	// route to receive mint public keys
	router.HandleFunc("/keys", Use(a.getKeys, LoggingMiddleware)).Methods(http.MethodGet)
	router.HandleFunc("/keys/{id}", Use(a.getKeysByKeySet, LoggingMiddleware)).Methods(http.MethodGet)
	router.HandleFunc("/keysets", Use(a.getKeySets, LoggingMiddleware)).Methods(http.MethodGet)
	// route to get mint (create tokens)
	router.HandleFunc("/mint", Use(a.getMint, LoggingMiddleware)).Methods(http.MethodGet)
	// route to real mint (with LIGHTNING enabled)
	router.HandleFunc("/mint", Use(a.mint, LoggingMiddleware)).Methods(http.MethodPost)
	// route to burn / melt a tx
	router.HandleFunc("/melt", Use(a.melt, LoggingMiddleware)).Methods(http.MethodPost)
	// route to check spendable proofs
	router.HandleFunc("/check", Use(a.check, LoggingMiddleware)).Methods(http.MethodPost)
	// route to check routing fees
	router.HandleFunc("/checkfees", Use(a.checkFee, LoggingMiddleware)).Methods(http.MethodPost)
	// route to split proofs (send money)
	router.HandleFunc("/split", Use(a.split, LoggingMiddleware)).Methods(http.MethodGet, http.MethodPost)
	appendSwaggoHandler(router)
	return router
}

// appendSwaggoHandler will append routes for the documentation to the router
func appendSwaggoHandler(router *mux.Router) {
	router.PathPrefix(ResourceSwaggerPathPrefix).Handler(httpSwagger.Handler(
		httpSwagger.URL(Config.DocReference), //The url pointing to API definition
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
func (api Api) checkFee(w http.ResponseWriter, r *http.Request) {
	feesRequest := cashu.CheckFeesRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&feesRequest)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	fee, err := api.Mint.CheckFees(feesRequest.Pr)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	response := cashu.CheckFeesResponse{Fee: fee / 1000}
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
// @Summary Request Api
// @Description Requests the minting of tokens belonging to a paid payment request.
// @Produce  json
// @Success 200 {object} GetMintResponse
// @Failure 500 {object} ErrorResponse
// @Router /mint [get]
// @Param        amount    query     string  false  "amount of the mint"
// @Tags GET
func (api Api) getMint(w http.ResponseWriter, r *http.Request) {
	amount := r.URL.Query().Get("amount")
	ai, err := strconv.Atoi(amount)
	if err != nil {
		log.Errorf("error checking amount")
	}
	invoice, err := api.Mint.RequestMint(uint64(ai))
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	log.WithField("invoice", invoice).Infof("created lightning invoice")
	hash, err := crypto.EncryptAESGCM([]byte(api.Mint.MasterSha526), []byte(invoice.GetHash()))
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	_, err = fmt.Fprintf(w, `{"pr": "%s", "hash": "%x"}`, invoice.GetPaymentRequest(), hash)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
}

// mint is the http handler function for POST /mint
// @Summary Api
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
func (api Api) mint(w http.ResponseWriter, r *http.Request) {
	var pr []byte
	var err error
	hash := r.URL.Query().Get("hash")
	if len(hash) == 0 {
		hash = r.URL.Query().Get("payment_hash")
		if len(hash) == 0 {
			responseError(w, cashu.NewErrorResponse(fmt.Errorf("invalid hash parameter")))
			return
		}
	}

	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	pr, err = crypto.DecryptAESGCM([]byte(api.Mint.MasterSha526), decodedHash)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}

	mintRequest := cashu.MintRequest{Outputs: make(cashu.BlindedMessages, 0)}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&mintRequest)
	if err != nil {
		panic(err)
	}

	promises, err := api.Mint.MintWithoutKeySet(mintRequest.Outputs, string(pr))
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	resp := cashu.MintResponse{Promises: promises}
	data, err := json.Marshal(resp)
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
func (api Api) melt(w http.ResponseWriter, r *http.Request) {

	payload := cashu.MeltRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		panic(err)
	}

	payment, err := api.Mint.Melt(payload.Proofs, payload.Pr)
	if err != nil {
		log.WithFields(log.Fields{"error.message": err.Error()}).Errorf("error in melt")
		return
	}
	// TODO -- add Change
	response := cashu.MeltResponse{Paid: payment.IsPaid(), Preimage: payment.GetPreimage()}
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
func (api Api) getKeys(w http.ResponseWriter, r *http.Request) {
	key, err := json.Marshal(api.Mint.GetPublicKeys())
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

// getKeys is the http handler function for GET /keys
// @Summary Keys
// @Description Get the public keys of the mint
// @Produce  json
// @Success 200 {object} GetKeysResponse
// @Failure 500 {object} ErrorResponse
// @Router /keys/{keyset_id} [get]
// @Tags GET
func (api Api) getKeysByKeySet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keysetId := vars["id"]
	keysetId = strings.ReplaceAll(strings.ReplaceAll(keysetId, "_", "/"), "-", "+")
	keyset, err := api.Mint.LoadKeySet(keysetId)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	key, err := json.Marshal(crypto.GetKeySetPublicKeys(keyset))
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

func (api Api) getKeySets(w http.ResponseWriter, r *http.Request) {
	response := cashu.GetKeySetsResponse{KeySets: api.Mint.GetKeySetIds()}
	res, err := json.Marshal(response)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	w.Write(res)
}

// check is the http handler function for POST /check
// @Summary Check spendable
// @Description Get currently available public keys
// @Produce  json
// @Success 200 {object} CheckSpendableResponse
// @Failure 500 {object} ErrorResponse
// @Router /check [post]
// @Param CheckSpendableRequest body CheckSpendableRequest true "Model containing proofs to check"
// @Tags POST
func (api Api) check(w http.ResponseWriter, r *http.Request) {
	payload := cashu.CheckSpendableRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
	}
	spendable := api.Mint.CheckSpendables(payload.Proofs)
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
// @Description Requests a set of tokens with amount "total" to be split into two newly minted sets with amount "split" and "total-split".
// @Produce  json
// @Success 200 {object} SplitResponse
// @Failure 500 {object} ErrorResponse
// @Router /split [post]
// @Param SplitRequest body SplitRequest true "Model containing proofs to split"
// @Tags POST
func (api Api) split(w http.ResponseWriter, r *http.Request) {

	payload := cashu.SplitRequest{}
	buf, _ := io.ReadAll(r.Body)
	body := io.NopCloser(bytes.NewBuffer(buf))
	bodyInvalidAmount := io.NopCloser(bytes.NewBuffer(buf))
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&payload)
	if err != nil {
		switch err.(type) {
		case *json.UnmarshalTypeError:
			amt := struct {
				Amount int64
			}{}
			decoder := json.NewDecoder(bodyInvalidAmount)
			err := decoder.Decode(&amt)
			if err == nil {
				err = fmt.Errorf("invalid split amount: %d", amt.Amount)
				responseError(w, cashu.NewErrorResponse(err))
				return
			}
		}
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	proofs := payload.Proofs
	outputs := payload.Outputs
	keyset, err := api.Mint.LoadKeySet(api.Mint.KeySetId)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	promises, err := api.Mint.Split(proofs, outputs, keyset)
	if err != nil {
		responseError(w, cashu.NewErrorResponse(err))
		return
	}
	response := cashu.SplitResponse{Promises: promises}
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
