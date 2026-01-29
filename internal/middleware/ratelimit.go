package middleware

import (
	"fmt"
	"hopSpotAPI/pkg/cache"
	"hopSpotAPI/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimitMiddleware struct {
	redisClient *cache.RedisClient
	limit       int
}

func NewRateLimitMiddleware(redisClient *cache.RedisClient, limit int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		redisClient: redisClient,
		limit:       limit,
	}
}

// Global Rate Limiter (für alle Requests)
func (m *RateLimitMiddleware) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// If Redis is not available, skip rate limiting
		if m.redisClient == nil {
			c.Next()
			return
		}

		ipAddr := c.ClientIP()
		key := fmt.Sprintf("ratelimit:global:%s", ipAddr)

		count, err := m.redisClient.Increment(c.Request.Context(), key, time.Hour)
		if err != nil {
			// Redis error - log and allow request (graceful degradation)
			logger.Warn().Err(err).Str("ip", ipAddr).Msg("Rate limit check failed")
			c.Next()
			return
		}

		if count > int64(m.limit) {
			c.Header("Retry-After", "3600")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}

// Login Rate Limiter (nur für Login-Endpoint)
func (m *RateLimitMiddleware) LimitLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.redisClient == nil {
			c.Next()
			return
		}

		ipAddr := c.ClientIP()
		key := fmt.Sprintf("ratelimit:login:%s", ipAddr)

		count, err := m.redisClient.Increment(c.Request.Context(), key, time.Hour)
		if err != nil {
			logger.Warn().Err(err).Str("ip", ipAddr).Msg("Login rate limit check failed")
			c.Next()
			return
		}

		if count > int64(m.limit) {
			c.Header("Retry-After", "3600")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many login attempts. Please try again later.",
			})
			return
		}

		c.Next()
	}
}
