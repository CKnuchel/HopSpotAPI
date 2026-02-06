package router

import (
	"hopSpotAPI/internal/handler"
	"hopSpotAPI/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "hopSpotAPI/docs"
)

func Setup(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	benchHandler *handler.BenchHandler,
	visitHandler *handler.VisitHandler,
	adminHandler *handler.AdminHandler,
	photoHandler *handler.PhotoHandler,
	weatherHandler *handler.WeatherHandler,
	favoriteHandler *handler.FavoriteHandler,
	activityHandler *handler.ActivityHandler,
	authMiddleware *middleware.AuthMiddleware,
	globalRateLimiter *middleware.RateLimitMiddleware,
	loginRateLimiter *middleware.RateLimitMiddleware,
) *gin.Engine {
	router := gin.Default()

	// Global Rate Limiting (all Requests)
	router.Use(globalRateLimiter.Limit())

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", loginRateLimiter.LimitLogin(), authHandler.Register)
			auth.POST("/login", loginRateLimiter.LimitLogin(), authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(authMiddleware.Authenticate())
		{
			// Auth routes
			protectedAuth := protected.Group("/auth")
			{
				protectedAuth.POST("refresh-fcm-token", authHandler.RefreshFCMToken)
			}

			// User routes
			user := protected.Group("/users")
			{
				user.GET("/me", userHandler.GetProfile)
				user.PATCH("/me", userHandler.UpdateProfile)
				user.POST("/me/change-password", userHandler.ChangePassword)
			}

			// Bench routes
			bench := protected.Group("/benches")
			{
				bench.GET("", benchHandler.List)
				bench.GET("/random", benchHandler.GetRandom)
				bench.GET("/:id", benchHandler.GetByID)
				bench.POST("", benchHandler.Create)
				bench.PATCH("/:id", benchHandler.Update)
				bench.DELETE("/:id", benchHandler.Delete)

				// Visit count by bench ID
				bench.GET("/:id/visits/count", visitHandler.GetVisitCountByBenchID)

				// Favorite routes unter /benches/:id
				bench.GET("/:id/favorite", favoriteHandler.Check)
				bench.POST("/:id/favorite", favoriteHandler.Add)
				bench.DELETE("/:id/favorite", favoriteHandler.Remove)

				// Photo routes unter /benches/:id
				bench.POST("/:id/photos", photoHandler.Upload)
				bench.GET("/:id/photos", photoHandler.GetByBenchID)
			}

			// Visit routes
			visits := protected.Group("/visits")
			{
				visits.GET("", visitHandler.ListVisits)
				visits.POST("", visitHandler.CreateVisit)
				visits.DELETE("/:id", visitHandler.DeleteVisit)
			}

			// Favorites routes
			favorites := protected.Group("/favorites")
			{
				favorites.GET("", favoriteHandler.List)
			}

			// Photo routes
			photos := protected.Group("/photos")
			{
				photos.DELETE("/:id", photoHandler.Delete)
				photos.PATCH("/:id/main", photoHandler.SetMainPhoto)
				photos.GET("/:id/url", photoHandler.GetPresignedURL)
			}

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(authMiddleware.RequireAdmin())
			{
				admin.GET("/users", adminHandler.ListUsers)
				admin.PATCH("/users/:id", adminHandler.UpdateUser)
				admin.DELETE("/users/:id", adminHandler.DeleteUser)
				admin.GET("/invitation-codes", adminHandler.ListInvitationCodes)
				admin.POST("/invitation-codes", adminHandler.CreateInvitationCode)
				admin.DELETE("/invitation-codes/:id", adminHandler.DeleteInvitationCode)
			}

			// Weather routes
			weather := protected.Group("/weather")
			{
				weather.GET("", weatherHandler.GetCurrentWeather)
			}

			// Activity routes
			activities := protected.Group("/activities")
			{
				activities.GET("", activityHandler.List)
			}
		}
	}

	// Default routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	return router
}
