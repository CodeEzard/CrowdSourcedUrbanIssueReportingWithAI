package handlers

import (
	"context"
	"encoding/base64"
	"time"

	config "crowdsourcedurbanissuereportingwithai/backend/configs"
	"crowdsourcedurbanissuereportingwithai/backend/internal/auth"
	"crowdsourcedurbanissuereportingwithai/backend/internal/cache"
	"crowdsourcedurbanissuereportingwithai/backend/internal/services"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/redis/go-redis/v9"
)

type AuthHandler struct {
	AuthService *services.AuthService
	JWTService  *auth.JWTService
	RedisClient *redis.Client
}

func NewAuthHandler(authSvc *services.AuthService, jwtSvc *auth.JWTService, rdb *redis.Client) *AuthHandler {
	return &AuthHandler{AuthService: authSvc, JWTService: jwtSvc, RedisClient: rdb}
}

type registerReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResp struct {
	AccessToken string `json:"access_token"`
}

// GoogleLoginRequest allows sign-in/up using a trusted identity provider (Google)
type googleLoginReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GoogleLogin finds or creates a user by email (from verified Google identity)
// and returns a JWT. This avoids password handling for social sign-in.
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Two forms accepted:
	// 1) application/json: { name, email }
	// 2) form POST from Google redirect: credential=<JWT>
	var req googleLoginReq
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.Email) == "" {
			http.Error(w, "email is required", http.StatusBadRequest)
			return
		}
	} else {
		// Assume form post
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		if idToken := r.FormValue("credential"); idToken != "" {
			// Decode JWT payload (unverified) to extract email/name for demo purposes
			// Format: header.payload.signature (base64url)
			parts := strings.Split(idToken, ".")
			if len(parts) >= 2 {
				payloadB64 := parts[1]
				// base64url decode
				payloadB64 = strings.ReplaceAll(payloadB64, "-", "+")
				payloadB64 = strings.ReplaceAll(payloadB64, "_", "/")
				// add padding
				switch len(payloadB64) % 4 {
				case 2:
					payloadB64 += "=="
				case 3:
					payloadB64 += "="
				}
				if decoded, err := base64.StdEncoding.DecodeString(payloadB64); err == nil {
					var claims map[string]any
					if json.Unmarshal(decoded, &claims) == nil {
						if v, ok := claims["email"].(string); ok {
							req.Email = v
						}
						if v, ok := claims["name"].(string); ok {
							req.Name = v
						}
					}
				}
			}
			if strings.TrimSpace(req.Email) == "" {
				http.Error(w, "failed to extract email from credential", http.StatusBadRequest)
				return
			}
		} else {
			http.Error(w, "missing credential", http.StatusBadRequest)
			return
		}
	}

	// Try to find existing user; if not found, create one with a random password
	user, err := h.AuthService.UserRepo.GetByEmail(req.Email)
	if err != nil {
		// Not found -> register a new user with a generated password
		// Use the existing Register flow to handle hashing and persistence
		if req.Name == "" {
			req.Name = req.Email
		}
		// Use a time-based random string as password; it won't be used directly
		genPass := time.Now().UTC().Format(time.RFC3339Nano)
		user, err = h.AuthService.Register(req.Name, req.Email, genPass)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Issue JWT for the user with appropriate role
	role := "user"
	// Check if user is admin (hardcoded list for now)
	adminEmails := map[string]bool{
		"admin@example.com":   true,
		"231b322@juetguna.in": true,
		"231b328@juetguna.in": true,
		"231b315@juetguna.in": true,
		// Add more admin emails as needed
	}
	if adminEmails[user.Email] {
		role = "admin"
	}
	token, err := h.JWTService.GenerateTokenWithRole(user.ID, role)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	// Set cookie similar to Login/Register
	cookie := &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
	}
	if ao := config.GetAllowedOrigin(); ao != "" {
		cookie.SameSite = http.SameSiteNoneMode
		if !strings.Contains(ao, "localhost") && !strings.Contains(ao, "127.0.0.1") {
			cookie.Secure = true
		}
	} else {
		cookie.SameSite = http.SameSiteLaxMode
	}
	http.SetCookie(w, cookie)
	// If this was a JSON/XHR request, return JSON. If it was a redirect form
	// submission from Google, redirect to the app.
	if strings.Contains(ct, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenResp{AccessToken: token})
		return
	}
	// Redirect to index within the same server (serving frontend)
	http.Redirect(w, r, "/index.html", http.StatusFound)
}
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	user, err := h.AuthService.Register(req.Name, req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// generate token with user role (default)
	role := "user"
	// Check if user is admin
	adminEmails := map[string]bool{
		"admin@example.com":   true,
		"231b322@juetguna.in": true,
		"231b328@juetguna.in": true,
		"231b315@juetguna.in": true,
	}
	if adminEmails[user.Email] {
		role = "admin"
	}
	token, err := h.JWTService.GenerateTokenWithRole(user.ID, role)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}
	// set cookie for convenience so browsers will send it automatically
	// If ALLOWED_ORIGIN is configured (cross-origin), set SameSite=None and Secure
	// so browsers will include the cookie for cross-site requests when the client
	// uses credentials: 'include'. For same-origin development keep Lax for ease.
	cookie := &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
	}
	if ao := config.GetAllowedOrigin(); ao != "" {
		cookie.SameSite = http.SameSiteNoneMode
		// For local development (localhost/127.0.0.1) Secure should be false so
		// browsers accept the cookie over HTTP. Only set Secure=true for non-local
		// origins which are expected to be HTTPS in production.
		if !strings.Contains(ao, "localhost") && !strings.Contains(ao, "127.0.0.1") {
			cookie.Secure = true
		}
	} else {
		cookie.SameSite = http.SameSiteLaxMode
	}
	http.SetCookie(w, cookie)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResp{AccessToken: token})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	user, err := h.AuthService.Authenticate(req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	role := "user"
	// Check if user is admin
	adminEmails := map[string]bool{
		"admin@example.com":   true,
		"231b322@juetguna.in": true,
		"231b328@juetguna.in": true,
		"231b315@juetguna.in": true,
	}
	if adminEmails[user.Email] {
		role = "admin"
	}
	token, err := h.JWTService.GenerateTokenWithRole(user.ID, role)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}
	// set cookie so subsequent requests include the token. Use SameSite=None+Secure
	// when ALLOWED_ORIGIN is configured for cross-origin clients.
	cookie2 := &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
	}
	if ao := config.GetAllowedOrigin(); ao != "" {
		cookie2.SameSite = http.SameSiteNoneMode
		if !strings.Contains(ao, "localhost") && !strings.Contains(ao, "127.0.0.1") {
			cookie2.Secure = true
		}
	} else {
		cookie2.SameSite = http.SameSiteLaxMode
	}
	http.SetCookie(w, cookie2)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResp{AccessToken: token})
}

// Logout reads the token from Authorization header/cookie/query and blacklists it in Redis
// so subsequent requests using the same token are rejected.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// extract token similar to middleware
	var token string
	authHdr := r.Header.Get("Authorization")
	if authHdr != "" {
		parts := strings.SplitN(authHdr, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			token = parts[1]
		}
	}
	if token == "" {
		if c, err := r.Cookie("access_token"); err == nil {
			token = c.Value
		}
	}
	if token == "" {
		if t := r.URL.Query().Get("token"); t != "" {
			token = t
		}
	}
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	if h.RedisClient == nil {
		// Nothing to blacklist against - accept logout as successful
		// delete cookie client-side by setting expired cookie; preserve SameSite/Secure
		del := &http.Cookie{Name: "access_token", Value: "", Path: "/", Expires: time.Unix(0, 0), MaxAge: -1}
		if ao := config.GetAllowedOrigin(); ao != "" {
			del.SameSite = http.SameSiteNoneMode
			if !strings.Contains(ao, "localhost") && !strings.Contains(ao, "127.0.0.1") {
				del.Secure = true
			}
		} else {
			del.SameSite = http.SameSiteLaxMode
		}
		http.SetCookie(w, del)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "logout successful (no redis configured)"})
		return
	}

	ttl, err := h.JWTService.GetTokenExpiry(token)
	if err != nil || ttl <= 0 {
		// fallback TTL if expiry can't be determined
		ttl = 1 * time.Hour
	}
	// blacklist token
	if err := cache.BlacklistToken(context.Background(), h.RedisClient, token, ttl); err != nil {
		http.Error(w, "failed to revoke token", http.StatusInternalServerError)
		return
	}
	// delete cookie (preserve SameSite/Secure handling)
	del2 := &http.Cookie{Name: "access_token", Value: "", Path: "/", Expires: time.Unix(0, 0), MaxAge: -1}
	if ao := config.GetAllowedOrigin(); ao != "" {
		del2.SameSite = http.SameSiteNoneMode
		if !strings.Contains(ao, "localhost") && !strings.Contains(ao, "127.0.0.1") {
			del2.Secure = true
		}
	} else {
		del2.SameSite = http.SameSiteLaxMode
	}
	http.SetCookie(w, del2)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "logout successful"})
}
