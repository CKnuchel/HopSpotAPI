package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"hopSpotAPI/internal/dto/responses"
	"io"
	"net/http"
	"time"
)

const DefaultBaseURL = "https://api.open-meteo.com/v1/forecast"

type WeatherClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewWeatherClient() *WeatherClient {
	return &WeatherClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: DefaultBaseURL,
	}
}

func (wc *WeatherClient) GetCurrentWeather(ctx context.Context, lat, lon float64) (*responses.WeatherResponse, error) {
	requestURL := wc.baseURL +
		"?latitude=" + fmt.Sprintf("%f", lat) +
		"&longitude=" + fmt.Sprintf("%f", lon) +
		"&current_weather=true" +
		"&timezone=Europe/Zurich"

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send HTTP request
	resp, err := wc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	// Status Code prüfen
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal JSON response
	var result responses.WeatherResponse
	if err = json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}
