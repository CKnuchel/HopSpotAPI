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
	JWTSecret          string
	JWTExpire          time.Duration
	JWTIssuer          string
	JWTAudience        string
	RefreshTokenExpire time.Duration

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

	// Redis
	RedisHost       string
	RedisPort       string
	RedisPassword   string
	RedisDB         int
	WeatherCacheTTL time.Duration

	// Rate Limiting
	RateLimitGlobal int // Requests per hour per IP
	RateLimitLogin  int // Login attempts per hour per IP
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

	// Refresh Token
	refreshDays, err := strconv.Atoi(getEnv("REFRESH_TOKEN_EXPIRE_DAYS", "90"))
	if err != nil {
		refreshDays = 90
	}

	// Redis
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		redisDB = 0
	}

	weatherTTL, err := strconv.Atoi(getEnv("WEATHER_CACHE_TTL_MINUTES", "15"))
	if err != nil {
		weatherTTL = 15
	}

	// Rate Limiting
	rateLimitGlobal, err := strconv.Atoi(getEnv("RATE_LIMIT_GLOBAL", "1000"))
	if err != nil {
		rateLimitGlobal = 1000
	}

	rateLimitLogin, err := strconv.Atoi(getEnv("RATE_LIMIT_LOGIN", "10"))
	if err != nil {
		rateLimitLogin = 10
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
		JWTSecret:          getEnv("JWT_SECRET", "supersecretkey"),
		JWTExpire:          time.Duration(jwtSeconds) * time.Second,
		JWTAudience:        getEnv("JWT_AUDIENCE", "yourapp.com"),
		JWTIssuer:          getEnv("JWT_ISSUER", "yourapp.com"),
		RefreshTokenExpire: time.Duration(refreshDays) * 24 * time.Hour,

		// MinIO
		MinioEndpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey:  getEnv("MINIO_ACCESS_KEY", ""),
		MinioSecretKey:  getEnv("MINIO_SECRET_KEY", ""),
		MinioUseSSL:     getEnv("MINIO_USE_SSL", "false") == "true",
		MinioBucketName: getEnv("MINIO_BUCKET_NAME", "hopspot-photos"),

		// Firebase
		FirebaseAuthKey: getEnv("FIREBASE_AUTH_KEY", ""),

		// Redis
		RedisHost:       getEnv("REDIS_HOST", "localhost"),
		RedisPort:       getEnv("REDIS_PORT", "6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:         redisDB,
		WeatherCacheTTL: time.Duration(weatherTTL) * time.Minute,

		// Rate Limiting
		RateLimitGlobal: rateLimitGlobal,
		RateLimitLogin:  rateLimitLogin,
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
