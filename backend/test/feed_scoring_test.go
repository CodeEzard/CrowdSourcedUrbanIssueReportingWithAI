package test

import (
    "crowdsourcedurbanissuereportingwithai/backend/internal/repository"
    "crowdsourcedurbanissuereportingwithai/backend/internal/services"
    "crowdsourcedurbanissuereportingwithai/backend/models"
    "testing"
    "time"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// TestFeedScoring verifies that the transient Score field is populated and posts
// are ordered by descending score (when heuristic produces distinct scores).
func TestFeedScoring(t *testing.T) {
    dsn := "host=localhost user=postgres password=post4321 dbname=Civicissue port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        t.Skipf("Skipping: cannot connect to postgres (%v)", err)
    }

    // Ensure schema
    if err := db.AutoMigrate(&models.User{}, &models.Issue{}, &models.Post{}, &models.Comment{}, &models.Upvote{}); err != nil {
        t.Fatalf("Failed automigrate: %v", err)
    }

    // Clean tables
    db.Exec("DELETE FROM upvotes")
    db.Exec("DELETE FROM comments")
    db.Exec("DELETE FROM posts")
    db.Exec("DELETE FROM issues")
    db.Exec("DELETE FROM users")

    // Create a user
    user := &models.User{ID: models.User{}.ID, Name: "ScoreUser", Email: "scoreuser@example.com", PasswordHash: "x"}
    if err := db.Create(user).Error; err != nil {
        t.Fatalf("create user: %v", err)
    }

    // Create issues with varying description urgency heuristics
    // emergency -> high (0.85), broken -> medium (0.6), baseline -> low (0.3)
    issueHigh := models.Issue{Name: "HighIssue", Description: "Emergency fire near transformer", Category: "Utilities"}
    issueMed := models.Issue{Name: "MedIssue", Description: "Broken pipeline causing minor leak", Category: "Utilities"}
    issueLow := models.Issue{Name: "LowIssue", Description: "Leaves gathered at corner", Category: "Garbage"}
    if err := db.Create(&issueHigh).Error; err != nil { t.Fatalf("issue high: %v", err) }
    if err := db.Create(&issueMed).Error; err != nil { t.Fatalf("issue med: %v", err) }
    if err := db.Create(&issueLow).Error; err != nil { t.Fatalf("issue low: %v", err) }

    // Create posts referencing those issues
    posts := []models.Post{
        {IssueID: issueLow.ID, UserID: user.ID, Description: issueLow.Description, Status: "open", Urgency: 1, Lat: 0, Lng: 0, MediaURL: "", CreatedAt: time.Now()},
        {IssueID: issueMed.ID, UserID: user.ID, Description: issueMed.Description, Status: "open", Urgency: 1, Lat: 0, Lng: 0, MediaURL: "", CreatedAt: time.Now()},
        {IssueID: issueHigh.ID, UserID: user.ID, Description: issueHigh.Description, Status: "open", Urgency: 1, Lat: 0, Lng: 0, MediaURL: "", CreatedAt: time.Now()},
    }
    for _, p := range posts { if err := db.Create(&p).Error; err != nil { t.Fatalf("create post: %v", err) } }

    repo := repository.NewPostRepository(db)
    feedSvc := services.NewFeedService(repo)

    feed, err := feedSvc.GetFeed()
    if err != nil { t.Fatalf("GetFeed error: %v", err) }
    if len(feed) < 3 { t.Fatalf("expected at least 3 posts, got %d", len(feed)) }

    // Ensure ordering by Score descending (first >= second >= third)
    first, second, third := feed[0], feed[1], feed[2]
    if !(first.Score >= second.Score && second.Score >= third.Score) {
        t.Fatalf("expected descending score ordering, got first=%.2f second=%.2f third=%.2f", first.Score, second.Score, third.Score)
    }
    if first.Score < 0.6 || first.Score > 0.9 {
        t.Fatalf("expected high score in range [0.6,0.9] for first, got %.2f", first.Score)
    }
    if third.Score > 0.4 { // low baseline ~0.3
        t.Fatalf("expected low score <= 0.4 for third, got %.2f", third.Score)
    }
}
