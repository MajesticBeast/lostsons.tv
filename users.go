package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *APIServer) usersRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/new", makeHTTPHandleFunc(s.handleCreateUser))

	return r
}

// Route for creating a new user
func (s *APIServer) handleCreateUser(w http.ResponseWriter, r *http.Request) error {
	user := parseUserForm(r)

	// Check if user already exists
	if _, err := s.store.GetUserByUsername(user.Username); err == nil {
		return fmt.Errorf("user already exists")
	}

	// Create user
	if err := s.store.CreateUser(user); err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	return responseWithJSON(w, http.StatusOK, "success")
}

func parseUserForm(r *http.Request) User {
	userForm := new(NewUserForm)

	userForm.Username = r.FormValue("username")
	userForm.Email = r.FormValue("email")

	// Create a new user object
	user := User{
		Username: userForm.Username,
		Email:    userForm.Email,
	}

	return user
}
