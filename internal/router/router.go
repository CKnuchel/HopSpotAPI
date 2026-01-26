package router

import (
	"hopSpotAPI/internal/handler"
	"hopSpotAPI/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	authMiddleware *middleware.AuthMiddleware) *gin.Engine {
	router := gin.Default()

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

		}
	}

	// Default routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	return router
}
