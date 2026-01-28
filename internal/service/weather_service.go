package service

import (
	"context"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/pkg/weather"
)

type WeatherService interface {
	GetCurrentWeather(ctx context.Context, lat, lon float64) (*responses.WeatherResponse, error)
}

type weatherService struct {
	weatherClient *weather.WeatherClient
}

func NewWeatherService(weatherClient *weather.WeatherClient) WeatherService {
	return &weatherService{
		weatherClient: weatherClient,
	}
}

func (s *weatherService) GetCurrentWeather(ctx context.Context, lat float64, lon float64) (*responses.WeatherResponse, error) {
	return s.weatherClient.GetCurrentWeather(ctx, lat, lon)
}
