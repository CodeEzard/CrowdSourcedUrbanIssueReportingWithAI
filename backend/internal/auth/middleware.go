package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"crowdsourcedurbanissuereportingwithai/backend/internal/cache"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type contextKey string

const ContextUserID contextKey = "user_id"

type errorResp struct {
	Error string `json:"error"`
}

// writeJSONError writes a JSON error response with given status code.
func writeJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(errorResp{Error: msg})
}

// AuthMiddleware returns an http middleware that validates the Bearer token
// and injects the user UUID into the request context under `ContextUserID`.
// It supports tokens in the Authorization header (Bearer), a cookie named
// `access_token`, or a `token` query parameter. OPTIONS requests are skipped
// for CORS preflight.
func AuthMiddleware(jwtSvc *JWTService, rdb *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip preflight
			if strings.ToUpper(r.Method) == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from Authorization header, cookie, or query param
			var tokenStr string
			tokenSource := "none"
			auth := r.Header.Get("Authorization")
			if auth != "" {
				parts := strings.SplitN(auth, " ", 2)
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenStr = parts[1]
					tokenSource = "authorization_header"
				}
			}
			if tokenStr == "" {
				if c, err := r.Cookie("access_token"); err == nil {
					tokenStr = c.Value
					tokenSource = "cookie"
				}
			}
			if tokenStr == "" {
				if t := r.URL.Query().Get("token"); t != "" {
					tokenStr = t
					tokenSource = "query_param"
				}
			}

			if tokenStr == "" {
				log.Printf("auth: missing access token for %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
				writeJSONError(w, "missing access token", http.StatusUnauthorized)
				return
			}
			log.Printf("auth: token source=%s method=%s path=%s remote=%s", tokenSource, r.Method, r.URL.Path, r.RemoteAddr)
			// Check blacklist in redis first (if provided)
			if rdb != nil {
				if ok, err := cache.IsTokenBlacklisted(context.Background(), rdb, tokenStr); err == nil && ok {
					writeJSONError(w, "token revoked", http.StatusUnauthorized)
					return
				}
			}

			uid, err := jwtSvc.ValidateToken(tokenStr)
			if err != nil {
				writeJSONError(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}
			// inject into context
			ctx := context.WithValue(r.Context(), ContextUserID, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext retrieves the user UUID from the request context.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(ContextUserID)
	if v == nil {
		return uuid.Nil, false
	}
	uid, ok := v.(uuid.UUID)
	return uid, ok
}
