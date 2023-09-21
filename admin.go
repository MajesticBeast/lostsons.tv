package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *APIServer) adminRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", s.handleAdminIndex)

	return r
}

// Admin Handlers
//
// --> /admin/index
func (s *APIServer) handleAdminIndex(w http.ResponseWriter, r *http.Request) {
	clips, err := s.store.GetAllClips()
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.ParseFiles("./templates/admin/index.html")
	if err != nil {
		log.Fatal(err)
	}

	if err := t.Execute(w, clips); err != nil {
		log.Fatal(err)
	}
}
