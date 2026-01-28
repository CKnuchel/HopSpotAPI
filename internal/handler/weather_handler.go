package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"hopSpotAPI/internal/service"
	"hopSpotAPI/pkg/utils"
)

type WeatherHandler struct {
	weatherService service.WeatherService
}

func NewWeatherHandler(weatherService service.WeatherService) *WeatherHandler {
	return &WeatherHandler{weatherService: weatherService}
}

// GetCurrentWeather godoc
// @Summary      Aktuelles Wetter abrufen
// @Description  Gibt das aktuelle Wetter für die angegebenen Koordinaten zurück
// @Tags         Weather
// @Produce      json
// @Param        lat  query     number  true  "Breitengrad"
// @Param        lon  query     number  true  "Längengrad"
// @Success      200  {object}  responses.WeatherResponse
// @Failure      400
// @Failure      500
// @Security     BearerAuth
// @Router       /weather [get]
func (wh *WeatherHandler) GetCurrentWeather(c *gin.Context) {
	lat := c.Query("lat")
	lon := c.Query("lon")

	// Validate input
	if lat == "" || lon == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat and lon are required"})
		return
	}

	// Convert lat and lon to float64
	latitude, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude"})
		return
	}

	longitude, err := strconv.ParseFloat(lon, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid longitude"})
		return
	}

	if err := utils.ValidateCoordinates(latitude, longitude); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the weather service
	weatherResponse, err := wh.weatherService.GetCurrentWeather(c.Request.Context(), latitude, longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get weather data"})
		return
	}

	c.JSON(http.StatusOK, weatherResponse)
}
