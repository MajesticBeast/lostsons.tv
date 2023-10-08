package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

// init function to load env vars
func init() {
	godotenv.Load()
}

func (s *APIServer) authRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/discord", makeHTTPHandleFunc(s.handleDiscordLogin))
	r.Get("/discord/callback", makeHTTPHandleFunc(s.handleDiscordCallback))

	return r
}

// Route for discord login
func (s *APIServer) handleDiscordLogin(w http.ResponseWriter, r *http.Request) error {
	oauth2Config := &oauth2.Config{
		ClientID:    os.Getenv("DISCORD_OAUTH_ID"),
		RedirectURL: os.Getenv("DISCORD_OAUTH_REDIRECT"),
		Endpoint:    oauth2.Endpoint{AuthURL: os.Getenv("DISCORD_OAUTH_ENDPOINT")},
		Scopes:      []string{"identify", "email"},
	}

	url := oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)

	return nil
}

// Route for discord callback
func (s *APIServer) handleDiscordCallback(w http.ResponseWriter, r *http.Request) error {
	oauth2Config := &oauth2.Config{
		ClientID:     os.Getenv("DISCORD_OAUTH_ID"),
		ClientSecret: os.Getenv("DISCORD_OAUTH_SECRET"),
		RedirectURL:  os.Getenv("DISCORD_OAUTH_REDIRECT"),
		Endpoint:     oauth2.Endpoint{TokenURL: os.Getenv("DISCORD_OAUTH_ENDPOINT")},
		Scopes:       []string{"identify", "email"},
	}

	code := r.URL.Query().Get("code")
	token, err := oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		return err
	}

	if token.Valid() {
		// Create a JWT with the access token as the claim
		jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"access_token": token.AccessToken,
		})

		// Sign the JWT with the secret
		jwtString, err := jwtToken.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
		if err != nil {
			return err
		}

		// Create and set the JWT as a secure cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "jwt",
			Value:    jwtString,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
	} else {
		return err
	}

	http.Redirect(w, r, "/", http.StatusFound)

	return nil

}
