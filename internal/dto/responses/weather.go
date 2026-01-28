package responses

type WeatherResponse struct {
	Latitude       float64                `json:"latitude"`
	Longitude      float64                `json:"longitude"`
	CurrentWeather CurrentWeatherResponse `json:"current_weather"`
}

type CurrentWeatherResponse struct {
	Temperature   float64 `json:"temperature"`
	Windspeed     float64 `json:"windspeed"`
	Winddirection int     `json:"winddirection"`
	Weathercode   int     `json:"weathercode"`
	Time          string  `json:"time"`
}
