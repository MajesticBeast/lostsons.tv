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
	r.Post("/delete", makeHTTPHandleFunc(s.handleDeleteUser))

	return r
}

// Route for creating a new user
func (s *APIServer) handleCreateUser(w http.ResponseWriter, r *http.Request) error {
	user, err := parseUserForm(r)
	if err != nil {
		return err
	}

	// Check if user or email already exists
	if _, err := s.store.GetUserByUsername(user.Username); err == nil {
		return fmt.Errorf("user already exists")
	}

	if _, err := s.store.GetUserByEmail(user.Email); err == nil {
		return fmt.Errorf("email already exists")
	}

	// Create user
	if err := s.store.CreateUser(user); err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	return responseWithJSON(w, http.StatusOK, "success")
}

// Route for deleting a user
func (s *APIServer) handleDeleteUser(w http.ResponseWriter, r *http.Request) error {
	user := User{
		ID:       r.PostFormValue("id"),
		Username: r.PostFormValue("username"),
		Email:    r.PostFormValue("email"),
	}

	// Check if user exists
	if _, err := s.store.GetUserByUsername(user.Username); err != nil {
		return fmt.Errorf("user does not exist")
	}

	// Need to delete foreign key references first
	if err := s.store.UpdateClipsUserIDToDeleted(user.ID); err != nil {
		return fmt.Errorf("error updating clips.user_id to 0000: %w", err)
	}

	if err := s.store.UpdateClipsUsersUserIDToDeleted(user.ID); err != nil {
		return fmt.Errorf("error updating clips_users.user_id to 0000: %w", err)
	}

	// Delete user
	if err := s.store.DeleteUser(user); err != nil {
		return fmt.Errorf("error deleting user: %w", err)
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

func validateUserForm(userForm NewUserForm) error {
	if userForm.Username != "majesticbeast" && userForm.Username != "devient" && userForm.Username != "ivorygun" {
		return fmt.Errorf("username is invalid")
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
