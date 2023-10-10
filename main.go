package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/majesticbeast/lostsons.tv/logger"
)

func main() {
	// Initialize logger
	log := &logger.StdLogger{}

	// Load environment variables
	godotenv.Load() // not likely needed with app platform env vars

	// Initialize the database connection
	dbConnStr := os.Getenv("DBCONNSTR")
	store, err := NewPostgresStore(dbConnStr)
	if err != nil {
		log.Error(err.Error())
	}

	// Initialize the database
	if err := store.Init(); err != nil {
		log.Error(err.Error())
	}

	// Initialize and run the API server
	server := NewAPIServer(store, log)
	server.Run()
}
