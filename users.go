package main

import (
	"fmt"
	"net/http"

	ev "github.com/AfterShip/email-verifier"
	"github.com/go-chi/chi/v5"
)

func (s *APIServer) usersRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/new", makeHTTPHandleFunc(s.handleCreateUser))

	return r
}

// Route for creating a new user
func (s *APIServer) handleCreateUser(w http.ResponseWriter, r *http.Request) error {
	user, err := parseUserForm(r)
	if err != nil {
		return err
	}

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

func parseUserForm(r *http.Request) (User, error) {
	userForm := new(NewUserForm)

	userForm.Username = r.PostFormValue("username")
	userForm.Email = r.PostFormValue("email")

	err := validateUserForm(*userForm)
	if err != nil {
		return User{}, err
	}

	user := User{
		Username: userForm.Username,
		Email:    userForm.Email,
	}

	return user, nil
}

// Create a function to validate the NewUserForm
func validateUserForm(userForm NewUserForm) error {
	if userForm.Username == "" {
		return fmt.Errorf("username is required")
	}

	if userForm.Email == "" {
		return fmt.Errorf("email is required")
	}

	ev := ev.NewVerifier()
	ret, err := ev.Verify(userForm.Email)
	if err != nil {
		return fmt.Errorf("error verifying email: %w", err)
	}

	if !ret.Syntax.Valid {
		return fmt.Errorf("email is invalid")
	}

	return nil
}
