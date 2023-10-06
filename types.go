package main

import (
	"time"
)

type Clip struct {
	ID            string
	PlaybackID    string
	AssetID       string
	DateUploaded  time.Time
	Description   string
	UserID        string
	GameID        string
	Tags          string
	FeaturedUsers string
	Game          string
	Username      string
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
