package main

import (
	"time"

	"github.com/google/uuid"
)

type Clip struct {
	Playback_id   string
	Asset_id      string
	Date_uploaded time.Time
	User_id       uuid.UUID
	Game_id       uuid.UUID
	Description   string
}
