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
	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/handler"
	"hopSpotAPI/internal/middleware"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/internal/router"
	"hopSpotAPI/internal/service"
	"hopSpotAPI/pkg/cache"
	"hopSpotAPI/pkg/notification"
	"hopSpotAPI/pkg/storage"
	"hopSpotAPI/pkg/weather"
	"log"
	"math/rand"
	"net/http"
)

func main() {
	// Load configuration from environment variables
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

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

	// Notification Service Setup (optional - graceful degradation)
	var fcmClient *notification.FCMClient
	if cfg.FirebaseAuthKey != "" {
		var err error
		fcmClient, err = notification.NewFCMClient(*cfg)
		if err != nil {
			log.Printf("FCM client not available: %v - push notifications disabled", err)
		} else {
			log.Println("FCM connected - push notifications enabled")
		}
	} else {
		log.Println("Firebase not configured - push notifications disabled")
	}

	// Redis Client Setup
	redisClient := cache.NewRedisClient(*cfg)
	if redisClient != nil {
		log.Println("Redis connected - caching enabled")
	} else {
		log.Println("Redis not available - caching disabled")
	}

	// Repositorys
	userRepo := repository.NewUserRepository(db)
	benchRepo := repository.NewBenchRepository(db)
	visitRepo := repository.NewVisitRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)
	photoRepo := repository.NewPhotoRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// Bootstrap: Create initial invitation code if no users exist
	userCount, err := userRepo.Count(context.Background())
	if err != nil {
		log.Printf("Warning: Could not check user count: %v", err)
	} else if userCount == 0 {
		code := generateBootstrapCode()
		invitationCode := &domain.InvitationCode{
			Code:    code,
			Comment: "Bootstrap - First Admin",
		}
		if err := invitationRepo.Create(context.Background(), invitationCode); err != nil {
			log.Printf("Warning: Could not create bootstrap code: %v", err)
		} else {
			log.Println("═══════════════════════════════════════════════════════")
			log.Println("  🎉 FIRST TIME SETUP")
			log.Println("  Use this invitation code to register the first admin:")
			log.Printf("  👉  %s", code)
			log.Println("═══════════════════════════════════════════════════════")
		}
	}

	// Services
	authService := service.NewAuthService(userRepo, invitationRepo, refreshTokenRepo, *cfg)
	userService := service.NewUserService(userRepo, *cfg)
	notificationService := service.NewNotificationService(fcmClient, userRepo)
	benchService := service.NewBenchService(benchRepo, notificationService)
	visitService := service.NewVisitService(visitRepo)
	adminService := service.NewAdminService(userRepo, invitationRepo)
	photoService := service.NewPhotoService(photoRepo, benchRepo, minioClient)
	weatherService := service.NewWeatherService(weatherClient, redisClient, cfg.WeatherCacheTTL)

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
	globalRateLimiter := middleware.NewRateLimitMiddleware(redisClient, cfg.RateLimitGlobal)
	loginRateLimiter := middleware.NewRateLimitMiddleware(redisClient, cfg.RateLimitLogin)

	// Router
	r := router.Setup(authHandler, userHandler, benchHandler,
		visitHandler, adminHandler, photoHandler, weatherHandler,
		authMiddleware, globalRateLimiter, loginRateLimiter)

	// Server mit Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go startServer(srv)
	waitForShutdown(srv, db, redisClient)
}

func generateBootstrapCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
