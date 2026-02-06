package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"hopSpotAPI/internal/domain"
	"hopSpotAPI/pkg/apperror"
	"hopSpotAPI/pkg/utils"
)

type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Reading the header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			apperror.AbortWithError(c, apperror.AppErrInvalidToken.WithDetails("Authorization header missing"))
			return
		}

		// Removing the Bearer prefix
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			apperror.AbortWithError(c, apperror.AppErrInvalidToken.WithDetails("Invalid token format"))
			return
		}

		// Validating the token
		claims, err := utils.ValidateJWT(tokenString, m.jwtSecret)
		if err != nil {
			apperror.AbortWithError(c, apperror.AppErrTokenExpired)
			return
		}

		// Parsing user ID
		userID, err := strconv.ParseUint(claims.RegisteredClaims.Subject, 10, 64)
		if err != nil {
			apperror.AbortWithError(c, apperror.AppErrInvalidToken.WithDetails("Invalid user ID in token"))
			return
		}

		// Storing user information in context
		c.Set(ContextKeyUserEmail, claims.Email)
		c.Set(ContextKeyUserRole, claims.Role)
		c.Set(ContextKeyUserID, uint(userID))

		c.Next()
	}
}

func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		roleValue, ok := role.(domain.Role)
		if !exists || !ok || roleValue != domain.RoleAdmin {
			apperror.AbortWithError(c, apperror.AppErrAdminRequired)
			return
		}

		c.Next()
	}
}
