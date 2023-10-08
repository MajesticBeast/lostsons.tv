package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	"github.com/gtuk/discordwebhook"
	"github.com/majesticbeast/lostsons.tv/logger"
	"github.com/majesticbeast/lostsons.tv/mux"
)

var tokenAuth *jwtauth.JWTAuth

func init() {
	tokenAuth = jwtauth.New("HS256", []byte(os.Getenv("JWT_SECRET")), nil)
}

type APIServer struct {
	store *PostgresStore
	log   logger.Logger
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func NewAPIServer(store *PostgresStore, log logger.Logger) *APIServer {
	return &APIServer{
		store: store,
		log:   log,
	}
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
	r.Get("/", makeHTTPHandleFunc(s.handleIndex))
	r.Get("/healthDB", makeHTTPHandleFunc(s.handleHealthDB))
	r.Get("/healthHTTP", makeHTTPHandleFunc(s.handleHealthHTTP))

	// Mux webhook route
	r.Post("/mux-webhook", makeHTTPHandleFunc(s.handleMuxWebhook))

	/********************
	 * Mount subrouters *
	 ********************/

	// Protected subroutes
	r.Group(func(r chi.Router) {
		// Seek, verify, and validate JWT tokens
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator)
		r.Mount("/admin", s.adminRouter())
	})

	// Unprotected subroutes (they may have their own protected routes)
	r.Mount("/clips", s.clipsRouter())
	r.Mount("/users", s.usersRouter())
	r.Mount("/games", s.gamesRouter())
	r.Mount("/auth", s.authRouter())

	// Start server
	s.log.Info("Starting server on port 3000")
	s.log.Error(http.ListenAndServe(":3000", r).Error())
}

// Routes

func (s *APIServer) handleIndex(w http.ResponseWriter, r *http.Request) error {
	t, err := template.ParseFiles("./templates/index.html")
	if err != nil {
		log.Fatal(err)
	}

	if err := t.Execute(w, nil); err != nil {
		log.Fatal(err)
	}

	return responseWithJSON(w, http.StatusOK, map[string]string{"message": "hello world"})
}

func (s *APIServer) handleHealthDB(w http.ResponseWriter, r *http.Request) error {
	err := s.store.db.Ping(r.Context())
	if err != nil {
		return responseWithError(w, http.StatusInternalServerError, "dead")
	}
	return responseWithJSON(w, http.StatusOK, map[string]string{"db": "alive"})

}

func (s *APIServer) handleHealthHTTP(w http.ResponseWriter, r *http.Request) error {
	return responseWithJSON(w, http.StatusOK, map[string]string{"http": "alive"})
}

func (s *APIServer) handleMuxWebhook(w http.ResponseWriter, r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("error reading mux webhook response body: %w", err)
		return responseWithError(w, http.StatusBadRequest, err.Error())
	}

	err = IsValidMuxSignature(r, body)
	if err != nil {
		err = fmt.Errorf("error validating mux signature: %w", err)
		return responseWithError(w, http.StatusBadRequest, err.Error())
	}

	assetResponse := mux.WebhookResponse{}
	if err := json.Unmarshal(body, &assetResponse); err != nil {
		err = fmt.Errorf("error unmarshalling mux webhook response body: %w", err)
		return responseWithError(w, http.StatusBadRequest, err.Error())
	}

	err = PostToDiscordWebhook(assetResponse)
	if err != nil {
		return responseWithError(w, http.StatusInternalServerError, err.Error())
	}

	return nil
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

// Mux webhook signature validation
func generateHmacSignature(webhookSecret, payload string) string {
	h := hmac.New(sha256.New, []byte(webhookSecret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// IsValidMuxSignature validates the mux webhook signature
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

	return nil
}

// JSON responses
func responseWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(payload)
}

func responseWithError(w http.ResponseWriter, code int, payload string) error {
	return responseWithJSON(w, code, map[string]string{"error": payload})
}
