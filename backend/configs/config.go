
package config

import (
	"github.com/joho/godotenv"
	"os"
	"log"
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
