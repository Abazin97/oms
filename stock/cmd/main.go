package main

import (
	log "log/slog"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Error("Error loading .env file")
	}

	url := os.Getenv("POSTGRES_URL")
	if url == "" {
		log.Error("url not set")
	}
}
