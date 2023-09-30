package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (s *APIServer) clipsRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", s.handleAdminIndex)

	return r
}

// Admin Handlers
//
// --> /admin/index
func (s *APIServer) handleCreateClip(w http.ResponseWriter, r *http.Request) error {

	newForm := new(NewClipForm)
	if err := r.ParseMultipartForm(45 << 20); err != nil {
		err = fmt.Errorf("error parsing multipart form: %w", err)
		return err
	}
	newForm.Description = r.FormValue("description")
	newForm.Game = r.FormValue("game")
	newForm.Username = r.FormValue("username")
	newForm.Tags = strings.Split(r.FormValue("tags"), " ")

	return nil
}
