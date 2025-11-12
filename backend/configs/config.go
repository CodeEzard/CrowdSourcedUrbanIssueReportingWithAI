package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load("backend/.env")
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}
}

func GetDSN() string {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")
	return "host=" + host + " user=" + user + " password=" + password + " dbname=" + dbname + " port=" + port + " sslmode=" + sslmode
}

// GetJWTSecret returns the JWT signing secret from environment.
func GetJWTSecret() string {
	return os.Getenv("JWT_SECRET")
}

// GetRedisAddr returns the Redis address (HOST:PORT) from environment.
func GetRedisAddr() string {
	return os.Getenv("REDIS_ADDR")
}

// GetRedisPassword returns the Redis password from environment.
func GetRedisPassword() string {
	return os.Getenv("REDIS_PASSWORD")
}

// GetAllowedOrigin returns the configured ALLOWED_ORIGIN for CORS
func GetAllowedOrigin() string {
	return os.Getenv("ALLOWED_ORIGIN")
}

// GetMLAPIURL returns the configured ML prediction API URL (optional)
func GetMLAPIURL() string {
	return os.Getenv("ML_API_URL")
}

// GetImageClassificationAPIURL returns the configured image classification API URL (optional)
func GetImageClassificationAPIURL() string {
	return os.Getenv("IMAGE_CLASSIFICATION_API_URL")
}

// GetFeedScoringMode controls how /feed computes priority scores.
// Values: "ml" (call external ML), "heuristic" (local keywords), "none" (map from stored urgency only).
// Default: "none" for performance.
func GetFeedScoringMode() string {
	m := strings.ToLower(strings.TrimSpace(os.Getenv("FEED_SCORING_MODE")))
	if m == "" {
		return "none"
	}
	switch m {
	case "ml", "heuristic", "none":
		return m
	default:
		return "none"
	}
}

// GetFeedLimit optionally caps number of posts returned by /feed to improve latency.
// Default 50 if unset or invalid.
func GetFeedLimit() int {
	v := strings.TrimSpace(os.Getenv("FEED_LIMIT"))
	if v == "" {
		return 50
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 50
	}
	return n
}
