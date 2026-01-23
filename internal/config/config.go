package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Api Port
	Port string

	// Database Connection String
	DBConnString string
}

func Load() *Config {
	err := godotenv.Load()

	if err != nil {
		panic("Error loading .env file")
	}

	return &Config{
		Port:         getEnv("PORT", "8080"),
		DBConnString: getEnv("DB_CONN_STRING", "user:password@tcp(localhost:3306)/dbname"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
