package test

import (
	"crowdsourcedurbanissuereportingwithai/backend/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestModels(t *testing.T) {
	dsn := "host=localhost user=postgres password=post4321 dbname=Civicissue port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Issue{},
		&models.Post{},
		&models.Comment{},
		&models.Upvote{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	user := models.User{
		ID:           uuid.New(),
		Name:         "Test User",
		Email:        "testuser@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	issue := models.Issue{
		ID:          uuid.New(),
		Name:        "Test Issue",
		Description: "Test Description",
		Category:    "Test Category",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := db.Create(&issue).Error; err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	post := models.Post{
		ID:          uuid.New(),
		IssueID:     issue.ID,
		UserID:      user.ID,
		Description: "Test Post Description",
		Status:      "open",
		Urgency:     5,
		Lat:         12.34,
		Lng:         56.78,
		MediaURL:    "http://example.com/media.jpg",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := db.Create(&post).Error; err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	comment := models.Comment{
		ID:        uuid.New(),
		PostID:    post.ID,
		UserID:    user.ID,
		Content:   "Test Comment",
		CreatedAt: time.Now(),
	}
	if err := db.Create(&comment).Error; err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	upvote := models.Upvote{
		ID:        uuid.New(),
		PostID:    post.ID,
		UserID:    user.ID,
		CreatedAt: time.Now(),
	}
	if err := db.Create(&upvote).Error; err != nil {
		t.Fatalf("Failed to create upvote: %v", err)
	}

	// Query and check all models
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

	if len(users) == 0 || len(issues) == 0 || len(posts) == 0 || len(comments) == 0 || len(upvotes) == 0 {
		t.Fatalf("One or more models were not created correctly")
	}
}
