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
	"math/rand"
	"net/http"

	"hopSpotAPI/internal/config"
	"hopSpotAPI/internal/database"
	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/handler"
	"hopSpotAPI/internal/middleware"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/internal/router"
	"hopSpotAPI/internal/service"
	"hopSpotAPI/pkg/cache"
	"hopSpotAPI/pkg/logger"
	"hopSpotAPI/pkg/notification"
	"hopSpotAPI/pkg/storage"
	"hopSpotAPI/pkg/weather"
)

func main() {
	// Load configuration from environment variables
	cfg := config.Load()

	// Initialize logger FIRST (before validation)
	logger.Init(cfg.LogLevel, cfg.LogFormat)
	logger.Info().Str("version", "1.0.0").Msg("Starting HopSpot API")

	// Validate config
	if err := cfg.Validate(); err != nil {
		logger.Fatal().Err(err).Msg("Configuration error")
	}

	// Initialize database connection
	db, err := database.Connect(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Run database migrations
	if err := database.Migrate(db); err != nil {
		logger.Fatal().Err(err).Msg("Database migration failed")
	}

	// MinIO client setup
	minioClient, err := storage.NewMinioClient(*cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create MinIO client")
	}

	// Ensure the bucket exists
	if err := minioClient.EnsureBucket(context.Background()); err != nil {
		logger.Fatal().Err(err).Msg("Failed to ensure MinIO bucket exists")
	}

	// Weather Client
	weatherClient := weather.NewWeatherClient()

	// Notification Service Setup (optional - graceful degradation)
	var fcmClient *notification.FCMClient
	if cfg.FirebaseAuthKey != "" {
		var err error
		fcmClient, err = notification.NewFCMClient(*cfg)
		if err != nil {
			logger.Warn().Err(err).Msg("FCM client not available - push notifications disabled")
		} else {
			logger.Info().Msg("FCM connected - push notifications enabled")
		}
	} else {
		logger.Info().Msg("Firebase not configured - push notifications disabled")
	}

	// Redis Client Setup
	redisClient := cache.NewRedisClient(*cfg)
	if redisClient != nil {
		logger.Info().Msg("Redis connected - caching enabled")
	} else {
		logger.Warn().Msg("Redis not available - caching disabled")
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	benchRepo := repository.NewBenchRepository(db)
	visitRepo := repository.NewVisitRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)
	photoRepo := repository.NewPhotoRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)

	// Bootstrap: Create initial invitation code if no users exist
	userCount, err := userRepo.Count(context.Background())
	if err != nil {
		logger.Warn().Err(err).Msg("Could not check user count")
	} else if userCount == 0 {
		code := generateBootstrapCode()
		invitationCode := &domain.InvitationCode{
			Code:    code,
			Comment: "Bootstrap - First Admin",
		}
		if err := invitationRepo.Create(context.Background(), invitationCode); err != nil {
			logger.Warn().Err(err).Msg("Could not create bootstrap code")
		} else {
			logger.Info().
				Str("code", code).
				Msg("═══════════════════════════════════════════════════════\n" +
					"  🎉 FIRST TIME SETUP\n" +
					"  Use this invitation code to register the first admin:\n" +
					"  👉  " + code + "\n" +
					"═══════════════════════════════════════════════════════")
		}
	}

	// Services
	authService := service.NewAuthService(userRepo, invitationRepo, refreshTokenRepo, *cfg)
	userService := service.NewUserService(userRepo, *cfg)
	notificationService := service.NewNotificationService(fcmClient, userRepo)
	benchService := service.NewBenchService(benchRepo, photoRepo, minioClient, notificationService)
	visitService := service.NewVisitService(visitRepo, photoRepo, minioClient)
	adminService := service.NewAdminService(userRepo, invitationRepo)
	photoService := service.NewPhotoService(photoRepo, benchRepo, minioClient)
	weatherService := service.NewWeatherService(weatherClient, redisClient, cfg.WeatherCacheTTL)
	favoriteService := service.NewFavoriteService(favoriteRepo, benchRepo, photoRepo, minioClient)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	benchHandler := handler.NewBenchHandler(benchService)
	visitHandler := handler.NewVisitHandler(visitService)
	adminHandler := handler.NewAdminHandler(adminService)
	photoHandler := handler.NewPhotoHandler(photoService)
	weatherHandler := handler.NewWeatherHandler(weatherService)
	favoriteHandler := handler.NewFavoriteHandler(favoriteService)

	// Middlewares
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)
	globalRateLimiter := middleware.NewRateLimitMiddleware(redisClient, cfg.RateLimitGlobal)
	loginRateLimiter := middleware.NewRateLimitMiddleware(redisClient, cfg.RateLimitLogin)

	// Router
	r := router.Setup(authHandler, userHandler, benchHandler,
		visitHandler, adminHandler, photoHandler, weatherHandler,
		favoriteHandler, authMiddleware, globalRateLimiter, loginRateLimiter)

	// Server mit Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	logger.Info().Str("port", cfg.Port).Msg("Server starting")
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
