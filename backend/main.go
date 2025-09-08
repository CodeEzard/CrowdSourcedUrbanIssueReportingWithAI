package main

import (
	"fmt"
	"log"
	"time"

	"crowdsourcedurbanissuereportingwithai/backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=post4321 dbname=Civicissue port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	_, err = sqlDB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
	if err != nil {
		log.Fatal("Failed to enable uuid-ossp extension:", err)
	}

	// AutoMigrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Issue{},
		&models.Post{},
		&models.Comment{},
		&models.Upvote{},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Clean up previous test data in correct order
	db.Where("created_at > ?", time.Now().Add(-time.Hour)).Delete(&models.Upvote{})
	db.Where("content = ?", "Test Comment").Delete(&models.Comment{})
	db.Where("description = ?", "Test Post Description").Delete(&models.Post{})
	db.Where("name = ?", "Test Issue").Delete(&models.Issue{})
	db.Where("email = ?", "testuser@example.com").Delete(&models.User{})

	// Comprehensive test: Create all models with all fields
	user := models.User{
		Name:         "Test User",
		Email:        "testuser@example.com",
		PasswordHash: "hashedpassword",
	}
	if err := db.Create(&user).Error; err != nil {
		log.Fatal("Failed to create user:", err)
	}

	issue := models.Issue{
		Name:        "Test Issue",
		Description: "Test Description",
		Category:    "Test Category",
	}
	if err := db.Create(&issue).Error; err != nil {
		log.Fatal("Failed to create issue:", err)
	}

	post := models.Post{
		IssueID:     issue.ID,
		UserID:      user.ID,
		Description: "Test Post Description",
		Status:      "open",
		Urgency:     5,
		Lat:         12.34,
		Lng:         56.78,
		MediaURL:    "http://example.com/media.jpg",
	}
	if err := db.Create(&post).Error; err != nil {
		log.Fatal("Failed to create post:", err)
	}

	comment := models.Comment{
		PostID:  post.ID,
		UserID:  user.ID,
		Content: "Test Comment",
	}
	if err := db.Create(&comment).Error; err != nil {
		log.Fatal("Failed to create comment:", err)
	}

	upvote := models.Upvote{
		PostID: post.ID,
		UserID: user.ID,
	}
	if err := db.Create(&upvote).Error; err != nil {
		log.Fatal("Failed to create upvote:", err)
	}

	// Query and print all models
	var users []models.User
	var issues []models.Issue
	var posts []models.Post
	var comments []models.Comment
	var upvotes []models.Upvote

	db.Find(&users)
	db.Find(&issues)
	db.Find(&posts)
	db.Find(&comments)
	db.Find(&upvotes)

	fmt.Println("\n--- Users ---")
	for _, u := range users {
		fmt.Printf("ID: %s | Name: %s | Email: %s | PasswordHash: %s | CreatedAt: %s | UpdatedAt: %s\n", u.ID, u.Name, u.Email, u.PasswordHash, u.CreatedAt.Format(time.RFC3339), u.UpdatedAt.Format(time.RFC3339))
	}

	fmt.Println("\n--- Issues ---")
	for _, i := range issues {
		fmt.Printf("ID: %s | Name: %s | Description: %s | Category: %s | CreatedAt: %s | UpdatedAt: %s\n", i.ID, i.Name, i.Description, i.Category, i.CreatedAt.Format(time.RFC3339), i.UpdatedAt.Format(time.RFC3339))
	}

	fmt.Println("\n--- Posts ---")
	for _, p := range posts {
		fmt.Printf("ID: %s | IssueID: %s | UserID: %s | Description: %s | Status: %s | Urgency: %d | Lat: %.2f | Lng: %.2f | MediaURL: %s | CreatedAt: %s | UpdatedAt: %s\n", p.ID, p.IssueID, p.UserID, p.Description, p.Status, p.Urgency, p.Lat, p.Lng, p.MediaURL, p.CreatedAt.Format(time.RFC3339), p.UpdatedAt.Format(time.RFC3339))
	}

	fmt.Println("\n--- Comments ---")
	for _, c := range comments {
		fmt.Printf("ID: %s | PostID: %s | UserID: %s | Content: %s | CreatedAt: %s\n", c.ID, c.PostID, c.UserID, c.Content, c.CreatedAt.Format(time.RFC3339))
	}

	fmt.Println("\n--- Upvotes ---")
	for _, u := range upvotes {
		fmt.Printf("ID: %s | PostID: %s | UserID: %s | CreatedAt: %s\n", u.ID, u.PostID, u.UserID, u.CreatedAt.Format(time.RFC3339))
	}

	log.Println("\nAll models migrated and test data inserted successfully!")
}
