package auth

import (
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestGenerateAndValidateToken(t *testing.T) {
	// ensure secret is set for predictable tokens
	os.Setenv("JWT_SECRET", "test-secret")
	svc := NewJWTService()

	uid := uuid.New()
	tok, err := svc.GenerateToken(uid)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	got, err := svc.ValidateToken(tok)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if got != uid {
		t.Fatalf("expected uid %s, got %s", uid.String(), got.String())
	}
}
