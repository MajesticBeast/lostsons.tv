package main

import (
	"time"
)

type Clip struct {
	ID            string    `json:"id"`
	PlaybackID    string    `json:"playback_id"`
	AssetID       string    `json:"asset_id"`
	DateUploaded  time.Time `json:"date_uploaded"`
	Description   string    `json:"description"`
	UserID        string    `json:"user_id"`
	GameID        string    `json:"game_id"`
	Tags          string    `json:"tags"`
	FeaturedUsers string    `json:"featured_users"`
	Game          string    `json:"game"`
	Username      string    `json:"username"`
}

type NewClipForm struct {
	Username      string
	Description   string
	Game          string
	Tags          string
	FeaturedUsers string
}

type User struct {
	ID       string
	Username string
	Email    string
	Role     string
}

type NewUserForm struct {
	Username string
	Email    string
}

type Game struct {
	ID   string
	Name string
}

type NewGameForm struct {
	Name string
}
