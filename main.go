package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	dbConnStr := os.Getenv("DBCONNSTR")

	// Initialize the database connection
	store, err := NewPostgresStore(dbConnStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	// Test CreateClip
	// clip := Clip{
	// 	Playback_id:   "987",
	// 	Asset_id:      "789",
	// 	Date_uploaded: time.Now(),
	// 	User:          "majestic",
	// 	Game:          "Ready or Not",
	// 	Description:   "Third clip",
	// }

	// if err := store.CreateClip(clip); err != nil {
	// 	log.Fatal(err)
	// }
	server := NewAPIServer(store)
	server.Run()
}
