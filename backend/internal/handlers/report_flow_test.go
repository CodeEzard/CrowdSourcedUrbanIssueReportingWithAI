package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"crowdsourcedurbanissuereportingwithai/backend/internal/auth"
	"crowdsourcedurbanissuereportingwithai/backend/internal/repository"
	"crowdsourcedurbanissuereportingwithai/backend/internal/services"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Full flow: register -> login -> report (protected) -> feed
func TestRegisterLoginReportFeedFlow(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")

	// setup in-memory DB
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// create minimal tables compatible with tests
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL,
        created_at DATETIME,
        updated_at DATETIME
    );`).Error; err != nil {
		t.Fatalf("create users table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS issues (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL UNIQUE,
        description TEXT,
        category TEXT NOT NULL,
        created_at DATETIME,
        updated_at DATETIME
    );`).Error; err != nil {
		t.Fatalf("create issues table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS posts (
        id TEXT PRIMARY KEY,
        issue_id TEXT NOT NULL,
        user_id TEXT NOT NULL,
        description TEXT,
        status TEXT NOT NULL,
        urgency INTEGER NOT NULL,
        lat REAL NOT NULL,
        lng REAL NOT NULL,
        media_url TEXT NOT NULL,
        created_at DATETIME,
        updated_at DATETIME
    );`).Error; err != nil {
		t.Fatalf("create posts table: %v", err)
	}

	// repos/services/handlers
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	authSvc := services.NewAuthService(userRepo)
	feedSvc := services.NewFeedService(postRepo)
	reportSvc := services.NewReportService(postRepo)

	jwtSvc := auth.NewJWTService()
	authHandler := NewAuthHandler(authSvc, jwtSvc, nil)
	feedHandler := NewFeedHandler(feedSvc)
	reportHandler := NewReportHandler(reportSvc)

	// Register user
	regBody := map[string]string{"name": "E2EUser", "email": "e2e@example.com", "password": "password"}
	rb, _ := json.Marshal(regBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(rb))
	rr := httptest.NewRecorder()
	authHandler.Register(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("register failed: %d %s", rr.Code, rr.Body.String())
	}
	var regResp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&regResp); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	token := regResp["access_token"]
	if token == "" {
		t.Fatalf("no token returned on register")
	}

	// Use token to call protected report handler
	reportReq := map[string]interface{}{
		"user_id":    "", // will be filled by middleware from token
		"issue_name": "Test Issue",
		"issue_desc": "desc",
		"issue_cat":  "Road",
		"post_desc":  "reported",
		"status":     "open",
		"urgency":    2,
		"lat":        1.23,
		"lng":        4.56,
		"media_url":  "http://example.com/p.jpg",
	}
	prb, _ := json.Marshal(reportReq)
	rreq := httptest.NewRequest(http.MethodPost, "/report", bytes.NewReader(prb))
	rreq.Header.Set("Authorization", "Bearer "+token)
	rrr := httptest.NewRecorder()
	// Wrap report handler with middleware
	mw := auth.AuthMiddleware(jwtSvc, nil)
	wrapped := mw(http.HandlerFunc(reportHandler.ServeReport))
	wrapped.ServeHTTP(rrr, rreq)
	if rrr.Code != http.StatusOK {
		t.Fatalf("report failed: %d %s", rrr.Code, rrr.Body.String())
	}

	// Call feed handler - should include the new post
	fr := httptest.NewRequest(http.MethodGet, "/feed", nil)
	frr := httptest.NewRecorder()
	feedHandler.ServeFeed(frr, fr)
	if frr.Code != http.StatusOK {
		t.Fatalf("feed failed: %d %s", frr.Code, frr.Body.String())
	}
	var posts []map[string]interface{}
	if err := json.NewDecoder(frr.Body).Decode(&posts); err != nil {
		t.Fatalf("decode feed response: %v", err)
	}
	if len(posts) == 0 {
		t.Fatalf("expected at least one post in feed")
	}
	// stronger checks: ensure description, issue.name, and user.email exist and match
	first := posts[0]
	if desc, ok := first["description"].(string); !ok || desc == "" {
		t.Fatalf("post missing or empty description: %v", first)
	}
	issueObj, ok := first["issue"].(map[string]interface{})
	if !ok {
		t.Fatalf("post missing issue object: %v", first)
	}
	if name, ok := issueObj["name"].(string); !ok || name == "" {
		t.Fatalf("unexpected or missing issue name: got %v", issueObj["name"])
	}
	userObj, ok := first["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("post missing user object: %v", first)
	}
	if email, ok := userObj["email"].(string); !ok || email != "e2e@example.com" {
		t.Fatalf("unexpected user email: got %v want %v", userObj["email"], "e2e@example.com")
	}
}

func TestReportWithoutTokenReturns401(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// minimal tables
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME,
		updated_at DATETIME
	);`).Error; err != nil {
		t.Fatalf("create users table: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	authSvc := services.NewAuthService(userRepo)
	reportSvc := services.NewReportService(postRepo)
	jwtSvc := auth.NewJWTService()
	authHandler := NewAuthHandler(authSvc, jwtSvc, nil)
	reportHandler := NewReportHandler(reportSvc)

	// register a user so DB has a user (not strictly necessary for middleware)
	regBody := map[string]string{"name": "NoToken", "email": "notoken@example.com", "password": "password"}
	rb, _ := json.Marshal(regBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(rb))
	rr := httptest.NewRecorder()
	authHandler.Register(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("register failed: %d %s", rr.Code, rr.Body.String())
	}

	// call report without Authorization header
	reportReq := map[string]interface{}{
		"user_id":    "",
		"issue_name": "NoToken Issue",
		"issue_desc": "desc",
		"issue_cat":  "Road",
		"post_desc":  "reported",
		"status":     "open",
		"urgency":    2,
		"lat":        1.23,
		"lng":        4.56,
		"media_url":  "http://example.com/p.jpg",
	}
	prb, _ := json.Marshal(reportReq)
	rreq := httptest.NewRequest(http.MethodPost, "/report", bytes.NewReader(prb))
	rrr := httptest.NewRecorder()
	mw := auth.AuthMiddleware(jwtSvc, nil)
	wrapped := mw(http.HandlerFunc(reportHandler.ServeReport))
	wrapped.ServeHTTP(rrr, rreq)
	if rrr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when calling /report without token, got %d", rrr.Code)
	}
}
