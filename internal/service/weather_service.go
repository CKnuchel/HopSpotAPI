package service

import (
	"context"
	"fmt"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/pkg/cache"
	"hopSpotAPI/pkg/weather"
	"log"
	"time"
)

type WeatherService interface {
	GetCurrentWeather(ctx context.Context, lat, lon float64) (*responses.WeatherResponse, error)
}

type weatherService struct {
	weatherClient *weather.WeatherClient
	redisClient   *cache.RedisClient
	cacheTTL      time.Duration
}

func NewWeatherService(weatherClient *weather.WeatherClient, redisClient *cache.RedisClient, cacheTTL time.Duration) WeatherService {
	return &weatherService{
		weatherClient: weatherClient,
		redisClient:   redisClient,
		cacheTTL:      cacheTTL,
	}
}

func (s *weatherService) GetCurrentWeather(ctx context.Context, lat float64, lon float64) (*responses.WeatherResponse, error) {
	cacheKey := s.generateCacheKey(lat, lon)

	// Try cache first (if Redis available)
	if s.redisClient != nil {
		var cachedResponse responses.WeatherResponse
		found, err := s.redisClient.Get(ctx, cacheKey, &cachedResponse)
		if err != nil {
			// Log but don't fail - just skip cache
			log.Printf("Redis get error: %v", err)
		}
		if found {
			return &cachedResponse, nil
		}
	}

	// Cache miss or no Redis → fetch from API
	weatherData, err := s.weatherClient.GetCurrentWeather(ctx, lat, lon)
	if err != nil {
		return nil, err
	}

	// Store in cache (if Redis available)
	if s.redisClient != nil {
		if err := s.redisClient.Set(ctx, cacheKey, weatherData, s.cacheTTL); err != nil {
			// Log but don't fail - data was fetched successfully
			log.Printf("Redis set error: %v", err)
		}
	}

	return weatherData, nil
}

func (s *weatherService) generateCacheKey(lat, lon float64) string {
	return fmt.Sprintf("weather:%.2f:%.2f", lat, lon)
}
