package main

import (
	"hopSpotAPI/internal/config"
	"hopSpotAPI/internal/database"
	"hopSpotAPI/internal/handler"
	"hopSpotAPI/internal/middleware"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/internal/router"
	"hopSpotAPI/internal/service"
	"net/http"
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

	// Repositorys
	userRepo := repository.NewUserRepository(db)
	invitation := repository.NewInvitationRepository(db)
	benchRepo := repository.NewBenchRepository(db)
	visitRepo := repository.NewVisitRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)

	// Services
	authService := service.NewAuthService(userRepo, invitation, *cfg)
	userService := service.NewUserService(userRepo, *cfg)
	benchService := service.NewBenchService(benchRepo)
	visitService := service.NewVisitService(visitRepo)
	adminService := service.NewAdminService(userRepo, invitationRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	benchHandler := handler.NewBenchHandler(benchService)
	visitHandler := handler.NewVisitHandler(visitService)
	adminHandler := handler.NewAdminHandler(adminService)

	// Middlewares
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)

	// Router
	r := router.Setup(authHandler, userHandler, benchHandler, visitHandler, adminHandler, authMiddleware)

	// Server mit Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go startServer(srv)
	waitForShutdown(srv, db)
}
