package main

import (
	"log"
	"pethealth/internal/app"
	"pethealth/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	// load .env file (if exists) - useful for local development and testing
	// vars in prod can be set in K8s)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// load configuration
	cfg := config.Load()

	// initialize
	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// run the application (starts HTTP server, etc.)
	if err := application.Run(); err != nil {
		log.Fatalf("Application stopped with error: %v", err)
	}
}
