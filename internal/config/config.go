package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port     string
	LogLevel string

	// JWT
	JWTSecret string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// TODO: MinIo / S3 Configs

	// TODO: Firebase Configs
}

func Load() *Config {
	err := godotenv.Load()

	if err != nil {
		panic("Error loading .env file")
	}

	return &Config{
		Port:       getEnv("PORT", "8080"),
		JWTSecret:  getEnv("JWT_SECRET", "SecretKey"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "hopspotdb"),
		LogLevel:   getEnv("LOG_LEVEL", "INFO"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
