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

	// "crowdsourcedurbanissuereportingwithai/backend/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// Create a simple users table compatible with sqlite for tests.
	create := `CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME,
		updated_at DATETIME
	);`
	if err := db.Exec(create).Error; err != nil {
		t.Fatalf("create users table: %v", err)
	}
	return db
}

func TestRegisterAndLoginHandlers(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authSvc := services.NewAuthService(userRepo)
	jwtSvc := auth.NewJWTService()
	authHandler := NewAuthHandler(authSvc, jwtSvc, nil)

	// Register
	regBody := map[string]string{"name": "Alice", "email": "alice@example.com", "password": "pass1234"}
	b, _ := json.Marshal(regBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(b))
	rr := httptest.NewRecorder()
	authHandler.Register(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("register expected 200 got %d body: %s", rr.Code, rr.Body.String())
	}
	var tr map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&tr); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	if tr["access_token"] == "" {
		t.Fatalf("expected access_token in register response")
	}

	// Login
	loginBody := map[string]string{"email": "alice@example.com", "password": "pass1234"}
	lb, _ := json.Marshal(loginBody)
	req2 := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(lb))
	rr2 := httptest.NewRecorder()
	authHandler.Login(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("login expected 200 got %d body: %s", rr2.Code, rr2.Body.String())
	}
	var tr2 map[string]string
	if err := json.NewDecoder(rr2.Body).Decode(&tr2); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if tr2["access_token"] == "" {
		t.Fatalf("expected access_token in login response")
	}
}
