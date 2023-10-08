package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

var discordEndpoint = oauth2.Endpoint{
	AuthURL:  "https://discord.com/api/oauth2/authorize?client_id=1083877630912249927&redirect_uri=https%3A%2F%2Flostsons.tv%2Fauth%2Fdiscord%2Fcallback&response_type=code&scope=identify%20email",
	TokenURL: "https://discord.com/api/oauth2/token",
}

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

func (s *APIServer) handleDiscordLogin(w http.ResponseWriter, r *http.Request) error {
	oauth2Config := &oauth2.Config{
		ClientID:    os.Getenv("DISCORD_OAUTH_ID"),
		RedirectURL: os.Getenv("DISCORD_OAUTH_REDIRECT"),
		Endpoint:    discordEndpoint,
		Scopes:      []string{"identify", "email"},
	}

	url := oauth2Config.AuthCodeURL(os.Getenv("STATE"), oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)

	return nil
}

func (s *APIServer) handleDiscordCallback(w http.ResponseWriter, r *http.Request) error {
	// Check if state is valid #CSRF
	state := r.URL.Query().Get("state")
	if state != os.Getenv("STATE") {
		return fmt.Errorf("invalid state")
	}

	oauth2Config := &oauth2.Config{
		ClientID:     os.Getenv("DISCORD_OAUTH_ID"),
		ClientSecret: os.Getenv("DISCORD_OAUTH_SECRET"),
		RedirectURL:  os.Getenv("DISCORD_OAUTH_REDIRECT"),
		Endpoint:     discordEndpoint,
		Scopes:       []string{"identify"},
	}

	code := r.URL.Query().Get("code")
	token, err := oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		err = fmt.Errorf("error exchanging code (%s) for token: %w", code, err)
		return err
	}

	// Get users Discord info
	req, err := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		err = fmt.Errorf("error creating request to Discord API: %w", err)
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("error getting response from Discord API: %w", err)
		return err
	}
	defer resp.Body.Close()

	// Create user struct for later use
	user := User{}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		err = fmt.Errorf("error decoding response body: %w", err)
		return err
	}

	// If user doesn't exist, enter into DB, then continue as normal
	if _, err := s.store.GetUserByUsername(user.Username); err != nil {
		if err := s.store.CreateUser(user); err != nil {
			err = fmt.Errorf("error creating user: %w", err)
			return err
		}
	}

	// Create a JWT with the access token as the claim
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"access_token": token.AccessToken,
		"username":     user.Username,
		"email":        user.Email,
	})

	// Sign the JWT with the secret
	jwtString, err := jwtToken.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		err = fmt.Errorf("error signing JWT: %w", err)
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

	http.Redirect(w, r, "/", http.StatusFound)

	return nil

}

// // function to authenticate jwt and cookie
// func (s *APIServer) authenticateJWT(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Get JWT from cookie
// 		cookie, err := r.Cookie("jwt")
// 		if err != nil {
// 			err = fmt.Errorf("error getting jwt cookie: %w", err)
// 			http.Error(w, err.Error(), http.StatusUnauthorized)
// 			return
// 		}

// 		// Parse JWT
// 		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
// 			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { // Check if signing method is HMAC
// 				err = fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 				http.Error(w, err.Error(), http.StatusUnauthorized)
// 				return nil, err
// 			}
// 			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
// 		})
// 		if err != nil {
// 			err = fmt.Errorf("error parsing jwt: %w", err)
// 			http.Error(w, err.Error(), http.StatusUnauthorized)
// 			return
// 		}

// 		// Check if token is valid
// 		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
// 			// Set claims in context
// 			ctx := r.Context()
// 			ctx = context.WithValue(ctx, "claims", claims)
// 			r = r.WithContext(ctx)

// 			// Call next handler
// 			next.ServeHTTP(w, r)
// 		} else {
// 			err = fmt.Errorf("invalid jwt")
// 			http.Error(w, err.Error(), http.StatusUnauthorized)
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }
