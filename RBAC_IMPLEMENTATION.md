# Role-Based Access Control (RBAC) Implementation

## Overview
This document describes the implementation of role-based access control (RBAC) for the Urban Civic Issue Reporting platform. The system now supports two roles:
- **`user`** (default) - Regular platform users who can report issues, comment, and upvote
- **`admin`** - Administrative users who can manage issue status and view the admin dashboard

## Backend Implementation

### 1. JWT Service Enhancements (`backend/internal/auth/jwt_service.go`)

#### New Methods:
- **`GenerateTokenWithRole(userID uuid.UUID, role string) (string, error)`**
  - Generates JWT tokens with both user ID and role claim
  - Adds `role` field to JWT claims alongside `user_id`, `exp`, and `iat`

- **`GetRoleFromToken(tokenStr string) (string, error)`**
  - Extracts role claim from JWT token
  - Returns `"user"` as default if role claim is missing (backward compatibility)

#### Modified Methods:
- **`GenerateToken(userID uuid.UUID) (string, error)`**
  - Now delegates to `GenerateTokenWithRole()` with default `"user"` role
  - Maintains backward compatibility for existing code

### 2. Middleware Enhancements (`backend/internal/auth/middleware.go`)

#### New Constants:
```go
const ContextUserRole contextKey = "user_role"
```

#### New Middleware:
- **`AdminMiddleware(next http.Handler) http.Handler`**
  - Checks if user has `admin` role in context
  - Returns `403 Forbidden` if user is not admin
  - Must be used after `AuthMiddleware` to ensure authentication

#### Updated AuthMiddleware:
- Now extracts role from JWT using `jwtSvc.GetRoleFromToken()`
- Injects both user ID and role into request context
- Role is extracted from every validated token

#### New Helper Function:
- **`GetUserRoleFromContext(ctx context.Context) (string, bool)`**
  - Retrieves user role from request context
  - Returns role string and boolean indicating presence

### 3. Authentication Handlers (`backend/internal/handlers/auth_handlers.go`)

#### Updated Handlers:
All authentication endpoints now include admin role checking:

1. **`GoogleLogin()`**
   - Checks hardcoded admin email list
   - Sets role to `"admin"` for admin emails, `"user"` otherwise
   - Calls `GenerateTokenWithRole()` instead of `GenerateToken()`

2. **`Login()`**
   - Added role checking logic
   - Uses admin email list to determine role
   - Generates JWT with appropriate role

3. **`Register()`**
   - New users default to `"user"` role
   - Checks admin list to assign `"admin"` role if applicable
   - Generates JWT with role information

#### Admin Email List:
Currently hardcoded in each handler:
```go
adminEmails := map[string]bool{
    "admin@example.com": true,
    // Add more admin emails as needed
}
```

**Future Enhancement**: Move to database or environment configuration for dynamic admin management.

### 4. Route Registration (`backend/router.go`)

#### Admin Routes:
Both admin endpoints are now wrapped with both authentication and authorization middleware:

```go
// Admin endpoints require both JWT auth AND admin role
adminMw := auth.AdminMiddleware(http.HandlerFunc(reportHandler.ServeUpdateStatus))
http.Handle("/api/admin/post-status", authMw(adminMw))

adminMw2 := auth.AdminMiddleware(http.HandlerFunc(feedHandler.ServeAdminFeed))
http.Handle("/api/admin/issues", authMw(adminMw2))
```

**Security Model**:
1. First, `authMw` validates JWT and extracts user ID + role
2. Then, `adminMw` checks if role is `"admin"`
3. If either fails, request is rejected with appropriate error

## Frontend Implementation

### 1. Common Utilities (`frontend/js/common.js`)

#### Updated `updateAuthUI()` Function:
```javascript
function updateAuthUI() {
  const jwt = localStorage.getItem('jwt');
  const user = localStorage.getItem('uc_user');
  const role = localStorage.getItem('uc_role');
  const adminLink = document.getElementById('nav-admin');

  if (jwt && user) {
    // Show admin link only if user is admin
    if (adminLink) {
      adminLink.style.display = role === 'admin' ? 'inline-block' : 'none';
    }
  } else {
    if (adminLink) adminLink.style.display = 'none';
  }
}
```

**Features**:
- Shows/hides `#nav-admin` link based on `uc_role` localStorage value
- Called on page load and after authentication changes
- Gracefully handles missing role (hides admin link)

### 2. Login Flow (`frontend/login2.html`)

#### JWT Parsing and Role Storage:
```javascript
if (data.access_token) {
  localStorage.setItem('jwt', data.access_token);
  localStorage.setItem('uc_user', googleName);
  
  // Extract role from JWT payload
  try {
    const parts = data.access_token.split('.');
    if (parts.length === 3) {
      const payload = JSON.parse(atob(parts[1]));
      localStorage.setItem('uc_role', payload.role || 'user');
    }
  } catch (e) {
    localStorage.setItem('uc_role', 'user');
  }
}
```

**Process**:
1. After successful authentication, JWT is stored
2. JWT payload is base64-decoded to extract claims
3. `role` claim is extracted and stored in `localStorage.uc_role`
4. Defaults to `"user"` if role claim is missing

### 3. Header Navigation (`frontend/includes/header.html`)

#### Admin Link:
```html
<a href="admin-dashboard.html" id="nav-admin" style="display: none;">Admin</a>
```

**Features**:
- Hidden by default (`display: none`)
- Shown by `updateAuthUI()` only for admin users
- Styled consistently with other navigation links

### 4. Admin Dashboard (`frontend/admin-dashboard.html`)

#### Access Control:
```javascript
// Check if user is authenticated and is admin
if (!localStorage.getItem('jwt')) {
  window.location.href = 'login2.html';
}

// Check if user has admin role
const role = localStorage.getItem('uc_role');
if (role !== 'admin') {
  alert('You do not have admin access');
  window.location.href = 'profile.html';
}
```

**Protection Levels**:
1. Requires valid JWT (unauthenticated users redirected to login)
2. Requires `admin` role (non-admin users redirected to profile with alert)
3. Backend endpoint also requires admin role (defense in depth)

## Data Flow

### Authentication Flow:
```
User Login (Google/Email)
    ↓
Backend validates credentials
    ↓
Check admin email list
    ↓
Generate JWT with role claim
    ↓
Frontend stores JWT and extracts role
    ↓
updateAuthUI() shows/hides admin link
```

### API Request Flow (Admin Endpoints):
```
Frontend sends request to /api/admin/*
    ↓
AuthMiddleware validates JWT
    ↓
Extract user ID and role from token
    ↓
Inject both into request context
    ↓
AdminMiddleware checks role == "admin"
    ↓
If admin: proceed to handler
If not: return 403 Forbidden
```

## LocalStorage Keys

| Key | Purpose | Example |
|-----|---------|---------|
| `jwt` | JWT access token | Encoded JWT string |
| `uc_user` | Username | "John Doe" |
| `uc_role` | User role | "admin" or "user" |
| `uc_email` | User email | "user@example.com" |
| `google_id` | Google ID (if OAuth) | "118..." |

## Testing the Implementation

### For Regular Users:
1. Register/login with non-admin email
2. Verify `uc_role` is set to `"user"` in localStorage
3. Verify Admin link does NOT appear in header
4. Navigate to `admin-dashboard.html` and confirm redirect to profile

### For Admin Users:
1. Login/register with admin email (`admin@example.com`)
2. Verify `uc_role` is set to `"admin"` in localStorage
3. Verify Admin link appears in header navigation
4. Access `/api/admin/issues` - should return issues list
5. Access `/api/admin/post-status` - should allow status updates

### API Testing (curl):

#### As Regular User (403 Forbidden):
```bash
curl -X GET https://localhost:8080/api/admin/issues \
  -H "Authorization: Bearer USER_JWT"
```
Expected: `403 Forbidden` with message "admin access required"

#### As Admin User (200 OK):
```bash
curl -X GET https://localhost:8080/api/admin/issues \
  -H "Authorization: Bearer ADMIN_JWT"
```
Expected: `200 OK` with issues list

#### Status Update (Admin Only):
```bash
curl -X POST https://localhost:8080/api/admin/post-status \
  -H "Authorization: Bearer ADMIN_JWT" \
  -H "Content-Type: application/json" \
  -d '{"post_id":"uuid","status":"closed","notes":"Fixed"}'
```
Expected: `200 OK` with updated post

## Future Enhancements

### 1. Database-Backed Roles
- Add `role` column to `users` table
- Store roles in PostgreSQL instead of hardcoded list
- Allow dynamic role assignment/revocation

### 2. Multiple Role Types
- `moderator` - Can flag issues, hide inappropriate content
- `department_staff` - Can only manage issues in specific categories
- `superadmin` - Can manage users and roles

### 3. Role-Based Features
- Department-specific dashboards
- Issue assignment to specific teams
- Role-based notification preferences

### 4. Audit Logging
- Log all admin actions (status changes, notes)
- Track who made changes and when
- Compliance and accountability

### 5. Frontend Improvements
- Role-based UI variations (show/hide features)
- Permission-based feature flags
- Admin-only sections on report detail page

## Security Considerations

### Current Implementation:
✅ JWT includes role claim (tamper-evident if secret is secure)
✅ Backend validates role on every admin request (defense in depth)
✅ Frontend hides UI but backend enforces access control
✅ Admin email list controls who can become admin
✅ Role extraction has safe defaults

### Recommendations:
⚠️ Consider HTTPS-only for JWT transmission in production
⚠️ Implement JWT refresh tokens for long-lived sessions
⚠️ Add rate limiting on admin endpoints
⚠️ Log all admin actions for audit trail
⚠️ Consider moving admin list to secure configuration/database
⚠️ Implement CSRF protection for state-changing operations

## Files Modified

### Backend:
- `internal/auth/jwt_service.go` - Added role support to JWT
- `internal/auth/middleware.go` - Added AdminMiddleware, role context
- `internal/handlers/auth_handlers.go` - Updated auth endpoints for role assignment
- `router.go` - Wrapped admin routes with AdminMiddleware

### Frontend:
- `js/common.js` - Updated updateAuthUI() to show/hide admin link
- `login2.html` - Added JWT parsing and role extraction
- `includes/header.html` - Added hidden admin nav link
- `admin-dashboard.html` - Added role check for dashboard access

## Summary

The RBAC implementation provides a secure, layered approach to access control:
1. **Token Level**: Role is embedded in JWT claims
2. **Middleware Level**: Role is validated on every request
3. **UI Level**: Admin features are hidden from non-admin users
4. **Policy Level**: Hardcoded admin email list controls initial assignment

This ensures that even if frontend protections are bypassed, the backend remains secure.
