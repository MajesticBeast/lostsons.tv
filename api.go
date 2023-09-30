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
	r.Get("/healthDb", s.handleHealthDb)
	r.Get("/healthHttp", s.handleHealthHTTP)

	// Mount subrouters router
	r.Mount("/admin", s.adminRouter())
	r.Mount("/clips", s.clipsRouter())

	// Start server
	log.Fatal(http.ListenAndServe(":3000", r))
}

//
// Route Handlers
//

// --> healthDb
func (s *APIServer) handleHealthDb(w http.ResponseWriter, r *http.Request) {
	err := s.store.db.Ping(r.Context())
	if err != nil {
		s.responseWithError(w, http.StatusInternalServerError, "dead")
	}

	s.responseWithJSON(w, http.StatusOK, map[string]string{"db": "alive"})
}

// --> healthHTTP
func (s *APIServer) handleHealthHTTP(w http.ResponseWriter, r *http.Request) {
	s.responseWithJSON(w, http.StatusOK, map[string]string{"http": "alive"})
}

// --> index
func (s *APIServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	s.responseWithJSON(w, http.StatusOK, map[string]string{"message": "hello world"})
}

// json responses
func (s *APIServer) responseWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		s.responseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
}

func (s *APIServer) responseWithError(w http.ResponseWriter, code int, message string) {
	s.responseWithJSON(w, code, map[string]string{"error": message})
}
