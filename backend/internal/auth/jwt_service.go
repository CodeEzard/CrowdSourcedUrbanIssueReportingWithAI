package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	config "crowdsourcedurbanissuereportingwithai/backend/configs"
)

// JWTService handles generation and validation of JWT access tokens.
type JWTService struct {
	secret string
	// access token expiry in minutes
	expiryMinutes int
}

// NewJWTService creates a JWTService using the secret from configs or environment.
func NewJWTService() *JWTService {
	secret := config.GetJWTSecret()
	if secret == "" {
		// use a fallback but this should be configured in env for production
		secret = "dev-secret"
	}
	return &JWTService{
		secret:        secret,
		expiryMinutes: 15,
	}
}

// GenerateToken returns a signed JWT containing the user ID.
func (s *JWTService) GenerateToken(userID uuid.UUID) (string, error) {
	return s.GenerateTokenWithRole(userID, "user")
}

// GenerateTokenWithRole returns a signed JWT containing the user ID and role.
func (s *JWTService) GenerateTokenWithRole(userID uuid.UUID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    role,
		"exp":     time.Now().Add(time.Duration(s.expiryMinutes) * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

// ValidateToken parses and validates a token, returning the user UUID if valid.
func (s *JWTService) ValidateToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.secret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	if !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid token claims")
	}
	uidRaw, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, errors.New("user_id claim missing")
	}
	uid, err := uuid.Parse(uidRaw)
	if err != nil {
		return uuid.Nil, err
	}
	return uid, nil
}

// GetRoleFromToken extracts the role claim from a JWT token.
func (s *JWTService) GetRoleFromToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.secret), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}
	role, ok := claims["role"].(string)
	if !ok {
		return "user", nil // default to "user" if role claim is missing
	}
	return role, nil
}

// GetTokenExpiry returns the duration until token expiry. If token is invalid or
// has no exp claim, an error is returned.
func (s *JWTService) GetTokenExpiry(tokenStr string) (time.Duration, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.secret), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}
	expRaw, ok := claims["exp"]
	if !ok {
		return 0, errors.New("exp claim missing")
	}
	// exp can be float64 (from JSON number) or json.Number
	var expInt int64
	switch v := expRaw.(type) {
	case float64:
		expInt = int64(v)
	case int64:
		expInt = v
	default:
		return 0, errors.New("invalid exp claim type")
	}
	expTime := time.Unix(expInt, 0)
	return time.Until(expTime), nil
}
