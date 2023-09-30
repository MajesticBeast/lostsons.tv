package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
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

// --> healthDb
func (s *APIServer) handleHealthDB(w http.ResponseWriter, r *http.Request) {
	err := s.store.db.Ping(r.Context())
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "dead")
	}

	responseWithJSON(w, http.StatusOK, map[string]string{"db": "alive"})
}

// --> healthHTTP
func (s *APIServer) handleHealthHTTP(w http.ResponseWriter, r *http.Request) {
	responseWithJSON(w, http.StatusOK, map[string]string{"http": "alive"})
}

// --> index
func (s *APIServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	responseWithJSON(w, http.StatusOK, map[string]string{"message": "hello world"})
}

// --> mux-webhook
func (s *APIServer) handleMuxWebhook(w http.ResponseWriter, r *http.Request) {

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
