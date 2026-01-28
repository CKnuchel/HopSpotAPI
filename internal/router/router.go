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
	authMiddleware *middleware.AuthMiddleware,
) *gin.Engine {
	router := gin.Default()

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
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
				bench.GET("/:id", benchHandler.GetByID)
				bench.POST("", benchHandler.Create)
				bench.PATCH("/:id", benchHandler.Update)
				bench.DELETE("/:id", benchHandler.Delete)

				// Visit count by bench ID
				bench.GET("/:id/visits/count", visitHandler.GetVisitCountByBenchID)

				// Photo routes unter /benches/:id
				bench.POST("/:id/photos", photoHandler.Upload)
				bench.GET("/:id/photos", photoHandler.GetByBenchID)
			}

			// Visit routes
			visits := protected.Group("/visits")
			{
				visits.GET("", visitHandler.ListVisits)
				visits.POST("", visitHandler.CreateVisit)
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
			}
		}
	}

	// Default routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	return router
}
