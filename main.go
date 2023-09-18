package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	r.Get("/", handleIndex)
	r.Get("/health", handleHealth)
	http.ListenAndServe(":3000", r)
}

//
// Route Handlers
//

// --> /health
func handleHealth(w http.ResponseWriter, r *http.Request) {
	responseWithJSON(w, http.StatusOK, map[string]string{"message": "alive"})
}

// --> /
func handleIndex(w http.ResponseWriter, r *http.Request) {
	responseWithJSON(w, http.StatusOK, map[string]string{"message": "hello world"})
}

func responseWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func responseWithError(w http.ResponseWriter, code int, message string) {
	responseWithJSON(w, code, map[string]string{"error": message})
}
