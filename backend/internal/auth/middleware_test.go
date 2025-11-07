package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestAuthMiddleware_SucceedsAndFails(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	jwt := NewJWTService()
	uid := uuid.New()
	tok, err := jwt.GenerateToken(uid)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	// Protected handler echoes user id
	rc := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 1})
	defer rc.Close()
	handler := AuthMiddleware(jwt, rc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, ok := GetUserIDFromContext(r.Context()); ok {
			w.Write([]byte(u.String()))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))

	// Success case
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d, body: %s", rr.Code, rr.Body.String())
	}
	if rr.Body.String() != uid.String() {
		t.Fatalf("expected body %s got %s", uid.String(), rr.Body.String())
	}

	// Missing token case
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for missing token, got %d", rr2.Code)
	}
}
