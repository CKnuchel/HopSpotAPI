package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/gorm"
)

func startServer(srv *http.Server) {
	log.Printf("Server starting on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

func waitForShutdown(srv *http.Server, db *gorm.DB) {
	quit := make(chan os.Signal, 1)

	// SIGINT (Ctrl+C) || SIGTERM signals from docker/kubernetes
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Printf("Received signal: %v. Shutting down gracefully...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	closeDatabase(db)
	log.Println("Server stopped")
}

func closeDatabase(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Error getting database instance: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
		return
	}

	log.Println("Database connection closed")
}
