package config

import (
	"log"
	"os"

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
