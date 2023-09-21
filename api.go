package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type APIServer struct {
	store *PostgresStore
}

func NewAPIServer(store *PostgresStore) *APIServer {
	return &APIServer{store: store}
}

func (s *APIServer) Run() {

	// Initialize main router and routes
	r := chi.NewRouter()
	r.Get("/", s.handleIndex)
	r.Get("/health", s.handleHealth)

	// Mount subrouters router
	r.Mount("/admin", s.adminRouter())

	// Start server
	log.Fatal(http.ListenAndServe(":3000", r))
}

//
// Route Handlers
//

// --> /health
func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.responseWithJSON(w, http.StatusOK, map[string]string{"message": "alive"})
}

// --> /index
func (s *APIServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	s.responseWithJSON(w, http.StatusOK, map[string]string{"message": "hello world"})
}

func (s *APIServer) responseWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		s.responseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (s *APIServer) responseWithError(w http.ResponseWriter, code int, message string) {
	s.responseWithJSON(w, code, map[string]string{"error": message})
}
