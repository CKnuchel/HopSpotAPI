package middleware

import (
	"fmt"
	"hopSpotAPI/pkg/cache"
	"log"
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
			log.Printf("Rate limit check failed: %v", err)
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
			log.Printf("Login rate limit check failed: %v", err)
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
