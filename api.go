package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gtuk/discordwebhook"
	"github.com/majesticbeast/lostsons.tv/mux"
)

type APIServer struct {
	store *PostgresStore
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func NewAPIServer(store *PostgresStore) *APIServer {
	return &APIServer{store: store}
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			responseWithJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func (s *APIServer) Run() {

	// Initialize main router and routes
	r := chi.NewRouter()
	r.Get("/", s.handleIndex)
	r.Get("/healthDB", s.handleHealthDB)
	r.Get("/healthHTTP", s.handleHealthHTTP)

	// Mux webhook route
	r.Post("/mux-webhook", s.handleMuxWebhook)

	// Mount subrouters router
	r.Mount("/admin", s.adminRouter())
	r.Mount("/clips", s.clipsRouter())

	// Start server
	fmt.Println("Starting server on port 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}

//
// Route Handlers
//

func (s *APIServer) handleHealthDB(w http.ResponseWriter, r *http.Request) {
	err := s.store.db.Ping(r.Context())
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "dead")
	}

	responseWithJSON(w, http.StatusOK, map[string]string{"db": "alive"})
}

func (s *APIServer) handleHealthHTTP(w http.ResponseWriter, r *http.Request) {
	responseWithJSON(w, http.StatusOK, map[string]string{"http": "alive"})
}

func (s *APIServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	responseWithJSON(w, http.StatusOK, map[string]string{"message": "hello world"})
}

func (s *APIServer) handleMuxWebhook(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("error reading mux webhook response body: %w", err)
		responseWithError(w, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	err = IsValidMuxSignature(r, body)
	if err != nil {
		err = fmt.Errorf("error validating mux signature: %w", err)
		responseWithError(w, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	assetResponse := mux.WebhookResponse{}
	if err := json.Unmarshal(body, &assetResponse); err != nil {
		err = fmt.Errorf("error unmarshalling mux webhook response body: %w", err)
		responseWithError(w, http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	err = PostToDiscordWebhook(assetResponse)
	if err != nil {
		err = fmt.Errorf("error posting to discord webhook: %w", err)
		responseWithError(w, http.StatusInternalServerError, err.Error())
		log.Println(err)
		return
	}

}

func PostToDiscordWebhook(assetResponse mux.WebhookResponse) error {
	username := "lostsons.tv"
	content := fmt.Sprintf("New clip { %s }\nPlaybackID: %s", assetResponse.Type, assetResponse.Data.PlaybackIds[0].ID)
	url := os.Getenv("DISCORD_WEBHOOK_URL")
	message := discordwebhook.Message{
		Username: &username,
		Content:  &content,
	}

	err := discordwebhook.SendMessage(url, message)
	if err != nil {
		err = fmt.Errorf("error sending message to discord webhook: %w", err)
		return err
	}

	return nil
}

func generateHmacSignature(webhookSecret, payload string) string {
	h := hmac.New(sha256.New, []byte(webhookSecret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

func IsValidMuxSignature(r *http.Request, body []byte) error {
	muxSignature := r.Header.Get("Mux-Signature")

	if muxSignature == "" {
		return errors.New("no Mux-Signature in request header")
	}

	muxSignatureArr := strings.Split(muxSignature, ",")

	if len(muxSignatureArr) != 2 {
		return fmt.Errorf("Mux-Signature in request header should be 2 values long: %s", muxSignatureArr)
	}

	timestampArr := strings.Split(muxSignatureArr[0], "=")
	v1SignatureArr := strings.Split(muxSignatureArr[1], "=")

	if len(timestampArr) != 2 || len(v1SignatureArr) != 2 {
		return fmt.Errorf("missing timestamp: %s or missing v1Signature: %s", timestampArr, v1SignatureArr)
	}

	timestamp := timestampArr[1]
	v1Signature := v1SignatureArr[1]

	webhookSecret := os.Getenv("DISCORD_WEBHOOK_SECRET")
	payload := fmt.Sprintf("%s.%s", timestamp, string(body))
	sha := generateHmacSignature(webhookSecret, payload)

	if sha != v1Signature {
		return errors.New("not a valid mux webhook signature")
	}

	fmt.Println("timestamp sha:", sha)
	fmt.Println("v1Signature:", v1Signature)
	return nil
}

// json responses
func responseWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
}

func responseWithError(w http.ResponseWriter, code int, message string) {
	responseWithJSON(w, code, map[string]string{"error": message})
}
