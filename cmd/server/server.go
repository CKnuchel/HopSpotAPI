package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hopSpotAPI/pkg/cache"
	"hopSpotAPI/pkg/logger"

	"gorm.io/gorm"
)

func startServer(srv *http.Server) {
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal().Err(err).Msg("Server failed")
	}
}

func waitForShutdown(srv *http.Server, db *gorm.DB, redisClient *cache.RedisClient) {
	quit := make(chan os.Signal, 1)

	// SIGINT (Ctrl+C) || SIGTERM signals from docker/kubernetes
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	logger.Info().Str("signal", sig.String()).Msg("Received signal - shutting down gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	closeDatabase(db)

	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			logger.Error().Err(err).Msg("Error closing Redis")
		} else {
			logger.Info().Msg("Redis connection closed")
		}
	}

	logger.Info().Msg("Server stopped")
}

func closeDatabase(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error().Err(err).Msg("Error getting database instance")
		return
	}

	if err := sqlDB.Close(); err != nil {
		logger.Error().Err(err).Msg("Error closing database")
		return
	}

	logger.Info().Msg("Database connection closed")
}
