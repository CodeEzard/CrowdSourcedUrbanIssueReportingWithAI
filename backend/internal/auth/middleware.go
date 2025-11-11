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

			// Extract token candidates in preferred order and validate the first that works.
			// This makes the middleware robust if, for example, a stale Authorization header
			// exists but a valid cookie is also present.
			type candidate struct{ src, val string }
			candidates := make([]candidate, 0, 3)
			if auth := r.Header.Get("Authorization"); auth != "" {
				parts := strings.SplitN(auth, " ", 2)
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					candidates = append(candidates, candidate{src: "authorization_header", val: parts[1]})
				}
			}
			if c, err := r.Cookie("access_token"); err == nil && c.Value != "" {
				candidates = append(candidates, candidate{src: "cookie", val: c.Value})
			}
			if t := r.URL.Query().Get("token"); t != "" {
				candidates = append(candidates, candidate{src: "query_param", val: t})
			}

			if len(candidates) == 0 {
				log.Printf("auth: missing access token for %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
				writeJSONError(w, "missing access token", http.StatusUnauthorized)
				return
			}

			var validated bool
			for _, cnd := range candidates {
				// Check blacklist in redis first (if provided)
				if rdb != nil {
					if ok, err := cache.IsTokenBlacklisted(context.Background(), rdb, cnd.val); err == nil && ok {
						// If one candidate is explicitly revoked, fail fast
						writeJSONError(w, "token revoked", http.StatusUnauthorized)
						return
					}
				}
				if uid, err := jwtSvc.ValidateToken(cnd.val); err == nil {
					log.Printf("auth: accepted token source=%s method=%s path=%s remote=%s", cnd.src, r.Method, r.URL.Path, r.RemoteAddr)
					// inject into context and proceed
					ctx := context.WithValue(r.Context(), ContextUserID, uid)
					next.ServeHTTP(w, r.WithContext(ctx))
					validated = true
					break
				} else {
					log.Printf("auth: token from %s invalid: %v", cnd.src, err)
					// try next candidate
				}
			}
			if !validated {
				writeJSONError(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}
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
