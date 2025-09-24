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

	// Clean up previous test data in correct order
	// Find test user, issue, and post IDs
	var testUser models.User
	db.Where("email = ?", "testuser@example.com").First(&testUser)
	var testIssue models.Issue
	db.Where("name = ?", "Test Issue").First(&testIssue)
	var testPost models.Post
	db.Where("description = ?", "Test Post Description").First(&testPost)

	// Delete upvotes referencing test post
	if testPost.ID != uuid.Nil {
		db.Where("post_id = ?", testPost.ID).Delete(&models.Upvote{})
	}
	// Delete comments referencing test post
	if testPost.ID != uuid.Nil {
		db.Where("post_id = ?", testPost.ID).Delete(&models.Comment{})
	}
	// Delete posts referencing test issue and user
	if testIssue.ID != uuid.Nil {
		db.Where("issue_id = ?", testIssue.ID).Delete(&models.Post{})
	}
	if testUser.ID != uuid.Nil {
		db.Where("user_id = ?", testUser.ID).Delete(&models.Post{})
	}
	// Delete issue
	if testIssue.ID != uuid.Nil {
		db.Where("id = ?", testIssue.ID).Delete(&models.Issue{})
	}
	// Delete user
	if testUser.ID != uuid.Nil {
		db.Where("id = ?", testUser.ID).Delete(&models.User{})
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

	t.Logf("\n--- Users ---")
	for _, u := range users {
		t.Logf("ID: %s | Name: %s | Email: %s | PasswordHash: %s | CreatedAt: %s | UpdatedAt: %s", u.ID, u.Name, u.Email, u.PasswordHash, u.CreatedAt.Format(time.RFC3339), u.UpdatedAt.Format(time.RFC3339))
	}

	t.Logf("\n--- Issues ---")
	for _, i := range issues {
		t.Logf("ID: %s | Name: %s | Description: %s | Category: %s | CreatedAt: %s | UpdatedAt: %s", i.ID, i.Name, i.Description, i.Category, i.CreatedAt.Format(time.RFC3339), i.UpdatedAt.Format(time.RFC3339))
	}

	t.Logf("\n--- Posts ---")
	for _, p := range posts {
		t.Logf("ID: %s | IssueID: %s | UserID: %s | Description: %s | Status: %s | Urgency: %d | Lat: %.2f | Lng: %.2f | MediaURL: %s | CreatedAt: %s | UpdatedAt: %s", p.ID, p.IssueID, p.UserID, p.Description, p.Status, p.Urgency, p.Lat, p.Lng, p.MediaURL, p.CreatedAt.Format(time.RFC3339), p.UpdatedAt.Format(time.RFC3339))
	}

	t.Logf("\n--- Comments ---")
	for _, c := range comments {
		t.Logf("ID: %s | PostID: %s | UserID: %s | Content: %s | CreatedAt: %s", c.ID, c.PostID, c.UserID, c.Content, c.CreatedAt.Format(time.RFC3339))
	}

	t.Logf("\n--- Upvotes ---")
	for _, u := range upvotes {
		t.Logf("ID: %s | PostID: %s | UserID: %s | CreatedAt: %s", u.ID, u.PostID, u.UserID, u.CreatedAt.Format(time.RFC3339))
	}
}
