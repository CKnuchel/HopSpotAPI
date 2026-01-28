package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port     string
	LogLevel string

	// JWT
	JWTSecret   string
	JWTExpire   time.Duration
	JWTIssuer   string
	JWTAudience string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// MinIO
	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioUseSSL     bool
	MinioBucketName string

	// Firebase
	FirebaseAuthKey string
}

func Load() *Config {
	err := godotenv.Load()

	if err != nil {
		panic("Error loading .env file")
	}

	jwtSeconds, err := strconv.Atoi(getEnv("JWT_EXPIRE_SECONDS", "3600"))
	if err != nil {
		jwtSeconds = 3600
	}

	return &Config{
		Port:     getEnv("PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "INFO"),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "dbname"),

		// JWT
		JWTSecret:   getEnv("JWT_SECRET", "supersecretkey"),
		JWTExpire:   time.Duration(jwtSeconds) * time.Second,
		JWTAudience: getEnv("JWT_AUDIENCE", "yourapp.com"),
		JWTIssuer:   getEnv("JWT_ISSUER,", "yourapp.com"),

		// MinIO
		MinioEndpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey:  getEnv("MINIO_ACCESS_KEY", ""),
		MinioSecretKey:  getEnv("MINIO_SECRET_KEY", ""),
		MinioUseSSL:     getEnv("MINIO_USE_SSL", "false") == "true",
		MinioBucketName: getEnv("MINIO_BUCKET_NAME", "hopspot-photos"),

		// Firebase
		FirebaseAuthKey: getEnv("FIREBASE_AUTH_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
