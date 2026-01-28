package main

//	@title			HopSpot API
//	@version		1.0
//	@description	REST API for the HopSpot bench management app

//	@contact.name	Christoph Knuchel
//	@contact.email	christoph.knuchel@gmail.com

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				JWT Token in the format: Bearer {token}

import (
	"context"
	"hopSpotAPI/internal/config"
	"hopSpotAPI/internal/database"
	"hopSpotAPI/internal/handler"
	"hopSpotAPI/internal/middleware"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/internal/router"
	"hopSpotAPI/internal/service"
	"hopSpotAPI/pkg/notification"
	"hopSpotAPI/pkg/storage"
	"hopSpotAPI/pkg/weather"
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

	// MinIo client setup
	minioClient, err := storage.NewMinioClient(*cfg)
	if err != nil {
		panic("Failed to create MinIO client: " + err.Error())
	}

	// Ensure the bucket exists
	if err := minioClient.EnsureBucket(context.Background()); err != nil {
		panic("Failed to ensure MinIO bucket exists: " + err.Error())
	}

	// Weather Client
	weatherClient := weather.NewWeatherClient()

	// Nofification Service Setup
	fcmClient, err := notification.NewFCMClient(*cfg)
	if err != nil {
		panic("Failed to create FCM client: " + err.Error())
	}

	// Repositorys
	userRepo := repository.NewUserRepository(db)
	benchRepo := repository.NewBenchRepository(db)
	visitRepo := repository.NewVisitRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)
	photoRepo := repository.NewPhotoRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// Services
	authService := service.NewAuthService(userRepo, invitationRepo, refreshTokenRepo, *cfg)
	userService := service.NewUserService(userRepo, *cfg)
	notificationService := service.NewNotificationService(fcmClient, userRepo)
	benchService := service.NewBenchService(benchRepo, notificationService)
	visitService := service.NewVisitService(visitRepo)
	adminService := service.NewAdminService(userRepo, invitationRepo)
	photoService := service.NewPhotoService(photoRepo, benchRepo, minioClient)
	weatherService := service.NewWeatherService(weatherClient)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	benchHandler := handler.NewBenchHandler(benchService)
	visitHandler := handler.NewVisitHandler(visitService)
	adminHandler := handler.NewAdminHandler(adminService)
	photoHandler := handler.NewPhotoHandler(photoService)
	weatherHandler := handler.NewWeatherHandler(weatherService)

	// Middlewares
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)

	// Router
	r := router.Setup(authHandler, userHandler, benchHandler,
		visitHandler, adminHandler, photoHandler, weatherHandler,
		authMiddleware)

	// Server mit Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go startServer(srv)
	waitForShutdown(srv, db)
}
