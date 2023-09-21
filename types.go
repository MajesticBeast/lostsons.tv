package main

import (
	"time"
)

type Clip struct {
	Id             string
	Playback_id    string
	Asset_id       string
	Date_uploaded  time.Time
	Description    string
	User           string
	Game           string
	Tags           string
	Featured_users string
	GameName       string
	Username       string
}
