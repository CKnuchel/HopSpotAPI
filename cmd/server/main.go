package main

import (
	"hopSpotAPI/internal/config"
	"hopSpotAPI/internal/database"
)

func main() {
	// Load configuration from environment variables
	cfg := config.Load()

	// Initialize database connection
	db, err := database.Connect(cfg)

	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	// Run database migrations
	if err := database.Migrate(db); err != nil {
		panic("Database migration failed: " + err.Error())
	}

	_ = db // TODO: to avoid unused variable error
	println("Server starting on port:", cfg.Port)
}
