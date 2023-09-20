package main

import (
	"time"
)

type Clip struct {
	Id            string
	Playback_id   string
	Asset_id      string
	Date_uploaded time.Time
	User          string
	Game          string
	Description   string
}
