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

	server := NewAPIServer(store)
	server.Run()
}
