package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DBConnString string
	JWTSecret    string
}

func Load() *Config {
	err := godotenv.Load()

	if err != nil {
		panic("Error loading .env file")
	}

	return &Config{
		Port:         getEnv("PORT", "8080"),
		DBConnString: getEnv("DB_CONN_STRING", "user:password@tcp(localhost:3306)/dbname"),
		JWTSecret:    getEnv("JWT_SECRET", "SecretKey"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
