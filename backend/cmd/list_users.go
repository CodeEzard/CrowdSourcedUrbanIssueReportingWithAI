package main

import (
	"fmt"
	"log"
	"os"

	"crowdsourcedurbanissuereportingwithai/backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Read DSN from environment if provided, otherwise use default from main.go
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=post4321 dbname=Civicissue port=5432 sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	var users []models.User
	if err := db.Order("created_at desc").Limit(50).Find(&users).Error; err != nil {
		log.Fatalf("query error: %v", err)
	}
	if len(users) == 0 {
		fmt.Println("No users found")
		return
	}
	fmt.Printf("Found %d users:\n", len(users))
	for _, u := range users {
		fmt.Printf("- %s  %s  %s\n", u.ID.String(), u.Email, u.Name)
	}
}
