package main

import (
	"log"

	"github.com/amillerrr/jwt-rbac-cors-app/internal/config"
	"github.com/amillerrr/jwt-rbac-cors-app/internal/database"
	"github.com/amillerrr/jwt-rbac-cors-app/internal/server"
)

func main() {
	// Load configuration from environment variables
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize and start the HTTP server
	srv := server.New(cfg, db)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
