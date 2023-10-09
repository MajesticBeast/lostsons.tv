package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
)

func (s *APIServer) adminRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/", s.handleAdminIndex)
	r.Get("/games", s.handleAdminGames)
	r.Get("/users", s.handleAdminUsers)
	r.Get("/clips", s.handleAdminClips)

	return r
}

// Admin Handlers
func (s *APIServer) handleAdminIndex(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())

	// Create user struct
	user := User{
		Username: claims["username"].(string),
		Email:    claims["email"].(string),
	}

	t, err := template.ParseFiles("./templates/admin/index.html")
	if err != nil {
		log.Fatal(err)
	}

	if err := t.Execute(w, user); err != nil {
		log.Fatal(err)
	}
}

// List of games
func (s *APIServer) handleAdminGames(w http.ResponseWriter, r *http.Request) {
	games, err := s.store.GetAllGames()
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.ParseFiles("./templates/admin/games.html")
	if err != nil {
		log.Fatal(err)
	}

	if err := t.Execute(w, games); err != nil {
		log.Fatal(err)
	}
}

// List of users
func (s *APIServer) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.GetAllUsers()
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.ParseFiles("./templates/admin/users.html")
	if err != nil {
		log.Fatal(err)
	}

	if err := t.Execute(w, users); err != nil {
		log.Fatal(err)
	}
}

// List of clips
func (s *APIServer) handleAdminClips(w http.ResponseWriter, r *http.Request) {
	clips, err := s.store.GetAllClips()
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.ParseFiles("./templates/admin/clips.html")
	if err != nil {
		log.Fatal(err)
	}

	if err := t.Execute(w, clips); err != nil {
		log.Fatal(err)
	}
}
