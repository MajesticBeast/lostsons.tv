package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *APIServer) gamesRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/new", makeHTTPHandleFunc(s.handleCreateGame))

	return r
}

// Route for creating a new game
func (s *APIServer) handleCreateGame(w http.ResponseWriter, r *http.Request) error {
	game, err := parseGameForm(r)
	if err != nil {
		return err
	}

	// Check if game already exists
	if _, err := s.store.GetGameByName(game.Name); err == nil {
		return fmt.Errorf("game already exists")
	}

	// Create game
	if err := s.store.CreateGame(game); err != nil {
		return fmt.Errorf("error creating game: %w", err)
	}

	return responseWithJSON(w, http.StatusOK, "success")
}

func parseGameForm(r *http.Request) (Game, error) {
	gameForm := new(NewGameForm)

	gameForm.Name = r.PostFormValue("name")

	err := validateGameForm(*gameForm)
	if err != nil {
		return Game{}, err
	}

	game := Game{
		Name: gameForm.Name,
	}

	return game, nil
}

func validateGameForm(gameForm NewGameForm) error {
	if gameForm.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	return nil
}
