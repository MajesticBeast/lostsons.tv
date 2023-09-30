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
