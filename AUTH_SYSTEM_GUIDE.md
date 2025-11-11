# Authentication System - Complete Guide

## Overview

The Urban Civic platform now has a complete authentication system with JWT tokens, login/register forms, and automatic token management.

## How It Works

### 1. User Registration

**Endpoint:** `POST /register`

```javascript
const response = await fetch('/register', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    name: 'John Doe',
    email: 'john@example.com',
    password: 'securePassword'
  })
});

const data = await response.json();
// Response: { "access_token": "eyJhbGc..." }
```

**UI:** `frontend/login2.html` (Registration Form)

### 2. User Login

**Endpoint:** `POST /login`

```javascript
const response = await fetch('/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'john@example.com',
    password: 'securePassword'
  })
});

const data = await response.json();
// Response: { "access_token": "eyJhbGc..." }
```

**UI:** `frontend/login2.html` (Login Form)

### 3. Token Storage

After successful login/register, the JWT is stored in **localStorage**:

```javascript
localStorage.setItem('jwt', data.access_token);
localStorage.setItem('uc_user', email); // Store user identifier
```

### 4. Using Protected Endpoints

All protected endpoints require the JWT token in the **Authorization header**:

```javascript
const token = localStorage.getItem('jwt');
const response = await fetch('/report', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({...})
});
```

### 5. Logout

**Endpoint:** `POST /logout` (Protected - requires token)

```javascript
async function logout() {
  const jwt = localStorage.getItem('jwt');
  const response = await fetch('/logout', {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${jwt}` }
  });
  
  // Clear local storage
  localStorage.removeItem('jwt');
  localStorage.removeItem('uc_user');
  
  // Redirect to login
  window.location.href = 'login2.html';
}
```

## UI Components

### Login/Register Page

**File:** `frontend/login2.html`

Features:
- ✅ Login form (email + password)
- ✅ Register form (name + email + password)
- ✅ Toggle between login and register
- ✅ Error messages with 5-second auto-clear
- ✅ Success messages with redirect
- ✅ Beautiful gradient background with floating bubbles
- ✅ Responsive design

### Header Auth Status

**File:** `frontend/includes/header.html`

Shows:
- 👤 **Logged In:** User name + Logout button
- 🔐 **Logged Out:** Login button link

Updates automatically via `updateAuthUI()` function.

## Frontend API (JavaScript)

### updateAuthUI()

Updates the header to show login/logout state based on localStorage token.

```javascript
updateAuthUI();
// Shows Login button if no token
// Shows Logout button + username if token exists
```

**Called automatically on:**
- Page load (via DOMContentLoaded)
- After successful login
- After logout

### logout()

Clears token from localStorage and redirects to login page.

```javascript
await logout();
// Calls /logout endpoint
// Clears localStorage
// Redirects to login2.html
```

### requireAuth(message)

Helper to enforce authentication on protected pages.

```javascript
if (!requireAuth('You need to log in to submit reports')) {
  return; // User redirected to login
}
// Continue with protected action
```

## Protected Endpoints

These endpoints require a valid JWT token:

| Endpoint | Method | Requires Token |
|----------|--------|---|
| /report | POST | ✅ Yes |
| /comment | POST | ✅ Yes |
| /upvote | POST | ✅ Yes |
| /logout | POST | ✅ Yes |
| /feed | GET | ❌ No |
| /login | POST | ❌ No |
| /register | POST | ❌ No |

## Token Mechanism

### JWT Token Generation

Backend generates tokens using HS256 algorithm:

```go
token, err := h.JWTService.GenerateToken(user.ID)
// Returns: "eyJhbGc..."
// Signed with JWT_SECRET environment variable
// Expires in 15 minutes (configurable)
```

### Token Validation Middleware

The `AuthMiddleware` validates tokens from multiple sources:

1. **Authorization Header** (Bearer token)
   ```
   Authorization: Bearer eyJhbGc...
   ```

2. **HTTP Cookie** (HttpOnly)
   ```
   Cookie: access_token=eyJhbGc...
   ```

3. **Query Parameter** (fallback)
   ```
   GET /report?token=eyJhbGc...
   ```

The middleware accepts the **first valid** token found (in order above).

### Token Blacklisting

Logout blacklists tokens in Redis (if configured) so they can't be reused.

## Error Handling

### 401 Unauthorized

**Cause:** Missing or invalid token

```javascript
// Error response
{
  "error": "missing access token"
}
// OR
{
  "error": "invalid or expired token"
}
```

**Solution:** 
1. Check if `localStorage.getItem('jwt')` exists
2. Ensure token is included in Authorization header
3. Re-login if token expired

### 400 Bad Request (Login/Register)

**Cause:** Invalid input (bad email, weak password, etc.)

```javascript
{
  "error": "invalid credentials"
  // OR
  "error": "invalid request"
}
```

**Solution:** Check form validation before submit

## Security Best Practices

✅ **Implemented:**
- JWT tokens signed with secret key
- HttpOnly cookies (immune to XSS)
- CORS with SameSite=None/Lax
- Tokens auto-include in Authorization header
- Logout blacklists tokens in Redis
- Token expiry (15 minutes)

⚠️ **For Production:**
- Use HTTPS (Secure flag on cookies)
- Store JWT_SECRET in environment variables (never in code)
- Implement token refresh mechanism
- Add rate limiting on login attempts
- Enable CORS only for trusted origins

## Flow Diagrams

### Login Flow

```
┌─────────────────────────────────────────────┐
│ User opens login2.html                      │
└──────────────┬──────────────────────────────┘
               │
               ↓
┌─────────────────────────────────────────────┐
│ Fills email + password                      │
└──────────────┬──────────────────────────────┘
               │
               ↓
┌─────────────────────────────────────────────┐
│ Submits form: POST /login                   │
└──────────────┬──────────────────────────────┘
               │
               ↓
┌─────────────────────────────────────────────┐
│ Backend validates credentials               │
│ Generates JWT token                         │
│ Sets HttpOnly cookie                        │
│ Returns { access_token: "..." }             │
└──────────────┬──────────────────────────────┘
               │
               ↓
┌─────────────────────────────────────────────┐
│ Frontend stores JWT in localStorage         │
│ Calls updateAuthUI()                        │
│ Redirects to index.html                     │
└─────────────────────────────────────────────┘
```

### Protected API Call Flow

```
┌──────────────────────────┐
│ User clicks "Report"     │
└───────────┬──────────────┘
            │
            ↓
┌──────────────────────────┐
│ Check localStorage.jwt   │
└───────────┬──────────────┘
            │
     ┌──────┴──────┐
     │             │
  Found         Not Found
     │             │
     ↓             ↓
  Add to      Redirect to
  Header      login2.html
     │
     ↓
┌──────────────────────────────────┐
│ POST /report                      │
│ Headers: {                        │
│   Authorization: Bearer eyJhb... │
│ }                                │
└───────────┬──────────────────────┘
            │
            ↓
┌──────────────────────────┐
│ AuthMiddleware validates │
│ token from header        │
└───────────┬──────────────┘
            │
    ┌───────┴────────┐
    │                │
 Valid            Invalid
    │                │
    ↓                ↓
 Proceed         401 Error
 Create post     Redirect
                 to login
```

## Testing Auth

### Manual Testing

1. **Test Registration:**
   ```bash
   curl -X POST http://localhost:8080/register \
     -H 'Content-Type: application/json' \
     -d '{"name":"Test","email":"test@example.com","password":"password123"}'
   ```

2. **Test Login:**
   ```bash
   curl -X POST http://localhost:8080/login \
     -H 'Content-Type: application/json' \
     -d '{"email":"test@example.com","password":"password123"}'
   ```

3. **Test Protected Endpoint:**
   ```bash
   TOKEN="eyJhbGc..." # From login response
   curl -X POST http://localhost:8080/report \
     -H 'Authorization: Bearer '$TOKEN \
     -H 'Content-Type: application/json' \
     -d '{...report payload...}'
   ```

### Automated Testing

See `backend/internal/handlers/report_flow_test.go` for full integration tests:

```bash
cd backend
go test ./internal/handlers -v
```

## Configuration

### Backend Environment Variables

```bash
# Required
JWT_SECRET=your-secret-key-here

# Optional (for CORS)
ALLOWED_ORIGIN=https://yourdomain.com

# Database
DATABASE_DSN=postgresql://...

# Redis (for token blacklisting)
REDIS_URL=redis://localhost:6379
```

### Token Expiry

Configured in `backend/internal/auth/auth_handlers.go`:

```go
Expires: time.Now().Add(15 * time.Minute)
```

Change the duration as needed for your security policy.

## Troubleshooting

### "401 missing access token" Error

**Cause:** Token not being sent to server

**Solutions:**
1. Check browser DevTools → Network → Request Headers
2. Verify `localStorage.getItem('jwt')` returns a token
3. Check if Authorization header is present: `Authorization: Bearer ...`
4. Ensure POST/GET method is correct

### "invalid or expired token" Error

**Cause:** Token is invalid or has expired

**Solutions:**
1. Clear localStorage and re-login
2. Check token expiry time (15 minutes default)
3. Verify JWT_SECRET matches between login generation and validation
4. Check backend logs for token validation errors

### Login Page Not Working

**Cause:** API endpoint not found or CORS error

**Solutions:**
1. Verify backend is running on correct port (8080)
2. Check `/login` endpoint exists in router.go
3. Test with curl command above
4. Check browser console for errors

---

**Last Updated:** 2025-11-11  
**Status:** Production Ready ✅
