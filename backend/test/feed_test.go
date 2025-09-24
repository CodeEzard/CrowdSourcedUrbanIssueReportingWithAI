package test

import (
	"crowdsourcedurbanissuereportingwithai/backend/internal/repository"
	"crowdsourcedurbanissuereportingwithai/backend/internal/services"
	"testing"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"fmt"
)

func TestFeedService(t *testing.T) {
	dsn := "host=localhost user=postgres password=post4321 dbname=Civicissue port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	repo := repository.NewPostRepository(db)
	service := services.NewFeedService(repo)

	// Setup: create a user and issue for fake posts
	db.Exec("DELETE FROM upvotes")
	db.Exec("DELETE FROM comments")
	db.Exec("DELETE FROM posts")
	db.Exec("DELETE FROM issues")
	db.Exec("DELETE FROM users")

	user := &struct {
		ID   string
		Name string
		Email string
		PasswordHash string
	}{
		ID:   "11111111-1111-1111-1111-111111111111",
		Name: "FeedTestUser",
		Email: "feedtestuser@example.com",
		PasswordHash: "hashed",
	}
	db.Exec("INSERT INTO users (id, name, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())", user.ID, user.Name, user.Email, user.PasswordHash)

	issue := &struct {
		ID   string
		Name string
		Description string
		Category string
	}{
		ID:   "22222222-2222-2222-2222-222222222222",
		Name: "FeedTestIssue",
		Description: "Feed test issue desc",
		Category: "FeedTestCat",
	}
	db.Exec("INSERT INTO issues (id, name, description, category, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())", issue.ID, issue.Name, issue.Description, issue.Category)

	// Add fake posts
	for i := 1; i <= 5; i++ {
		desc := fmt.Sprintf("Fake Post #%d", i)
		postID := fmt.Sprintf("33333333-3333-3333-3333-33333333333%d", i)
		interval := fmt.Sprintf("%d minutes", i*10)
		db.Exec(fmt.Sprintf("INSERT INTO posts (id, issue_id, user_id, description, status, urgency, lat, lng, media_url, created_at, updated_at) VALUES (?, ?, ?, ?, 'open', ?, ?, ?, ?, NOW() - INTERVAL '%s', NOW() - INTERVAL '%s')", interval, interval), postID, issue.ID, user.ID, desc, i, 10.0+i, 20.0+i, "http://example.com/fake.jpg")
	}

	posts, err := service.GetFeed()
	if err != nil {
		t.Fatalf("Failed to get feed: %v", err)
	}

	if len(posts) < 5 {
		t.Fatalf("Feed should have at least 5 fake posts, got %d", len(posts))
	}

	t.Logf("Feed posts (newest first):")
	for _, p := range posts {
		t.Logf("ID: %s | Description: %s | CreatedAt: %s | User: %s | Issue: %s", p.ID, p.Description, p.CreatedAt.Format("2006-01-02 15:04:05"), p.User.Name, p.Issue.Name)
	}
}
