package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/majesticbeast/lostsons.tv/logger"
)

func main() {
	// Initialize logger
	log := &logger.StdLogger{}

	godotenv.Load()
	dbConnStr := os.Getenv("DBCONNSTR")

	// Initialize the database connection
	store, err := NewPostgresStore(dbConnStr)
	if err != nil {
		log.Error(err.Error())
	}

	// Initialize the database
	if err := store.Init(); err != nil {
		log.Error(err.Error())
	}

	// Initialize and run the API server
	server := NewAPIServer(store)
	server.Run()
}
