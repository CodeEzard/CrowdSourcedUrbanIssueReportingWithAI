# RBAC System Architecture Diagram

## System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                       URBAN CIVIC ADMIN SYSTEM                      │
└─────────────────────────────────────────────────────────────────────┘

┌──────────────────────────┐            ┌──────────────────────────┐
│     FRONTEND (Browser)   │            │    BACKEND (Go Server)   │
├──────────────────────────┤            ├──────────────────────────┤
│                          │            │                          │
│  1. login2.html          │            │  1. AuthHandler          │
│     - Get JWT            │────────────│     - Login/Register     │
│     - Extract role       │◄───────────│     - Check admin list   │
│     - Store localStorage │            │     - GenerateToken      │
│                          │            │       + role claim       │
│  2. common.js            │            │                          │
│     - updateAuthUI()     │            │  2. JWTService           │
│     - Show/hide admin    │            │     - GenerateTokenWith  │
│       link               │            │       Role()             │
│                          │            │     - ValidateToken()    │
│  3. header.html          │            │     - GetRoleFromToken() │
│     - Admin nav link     │            │                          │
│     - Hidden by default  │            │  3. Middleware           │
│                          │            │     - AuthMiddleware:    │
│  4. admin-dashboard.html │            │       + Validate JWT     │
│     - Check JWT          │────────────│       + Extract role     │
│     - Check role         │◄───────────│       + Inject context   │
│     - Access control     │            │     - AdminMiddleware:   │
│     - Fetch/update       │────────────│       + Check role=="ad" │
│       issues             │◄───────────│       + 403 if not admin │
│                          │            │                          │
└──────────────────────────┘            │  4. Router               │
                                        │     - Protect routes     │
                                        │     with middleware      │
                                        │                          │
                                        └──────────────────────────┘
```

## Authentication Flow

```
User Login/Register
    │
    ├─→ [Email validation]
    │
    ├─→ Backend receives credentials
    │
    ├─→ [Hash password & verify]
    │
    ├─→ Check admin email list
    │   │
    │   ├─→ Email in list? → role = "admin"
    │   │
    │   └─→ Email NOT in list? → role = "user"
    │
    ├─→ Generate JWT with role claim
    │   │ Claims: {
    │   │   "user_id": "uuid",
    │   │   "role": "admin" | "user",
    │   │   "exp": timestamp,
    │   │   "iat": timestamp
    │   │ }
    │
    ├─→ Return JWT to frontend
    │
    └─→ Frontend:
        │
        ├─→ Store JWT in localStorage
        │
        ├─→ Decode JWT payload
        │
        ├─→ Extract role claim
        │
        ├─→ Store role in localStorage.uc_role
        │
        ├─→ Call updateAuthUI()
        │
        ├─→ Show admin link if role == "admin"
        │
        └─→ Redirect to index.html
```

## API Request Flow (Admin Endpoint)

```
Frontend Request: GET /api/admin/issues
    │ Include JWT in Authorization header
    │
    ▼
Backend Router receives request
    │
    ▼ Step 1: AuthMiddleware
    ├─→ Extract JWT from Authorization header
    │
    ├─→ Validate JWT signature
    │   (Check: secret matches, token not expired)
    │
    ├─→ Parse JWT claims
    │   (Extract: user_id, role, exp, iat)
    │
    ├─→ Inject into request context:
    │   ├─ ContextUserID = user_id
    │   └─ ContextUserRole = role
    │
    ├─→ If JWT invalid → 401 Unauthorized
    │
    └─→ Pass to next middleware
    
    ▼ Step 2: AdminMiddleware
    ├─→ Get role from context
    │
    ├─→ Check: role == "admin"?
    │
    ├─→ If YES → Continue to handler
    │
    ├─→ If NO → Return 403 Forbidden
    │           {"error": "admin access required"}
    │
    └─→ Pass to handler
    
    ▼ Step 3: Handler (e.g., ServeAdminFeed)
    ├─→ Fetch all posts from database
    │
    ├─→ Apply optional filters (?status=open)
    │
    ├─→ Return JSON response
    │
    └─→ 200 OK with issues list

Frontend receives response
    │
    ├─→ If 200 → Parse JSON and render issues
    │
    ├─→ If 403 → Show error message
    │
    └─→ If 401 → Redirect to login
```

## Middleware Chain Diagram

```
Request comes in
    │
    ▼
────────────────────────────────────────
│  Stage 1: Authentication (AuthMiddleware)
├──────────────────────────────────────────
│  Input:  Request with JWT
│  Output: Request with user_id + role in context
│  Actions:
│    • Extract JWT from header
│    • Validate signature
│    • Parse claims
│    • Inject context values
│    • Pass to next middleware or reject (401)
────────────────────────────────────────
    │
    ▼
────────────────────────────────────────
│  Stage 2: Authorization (AdminMiddleware)
├──────────────────────────────────────────
│  Input:  Request with authenticated context
│  Output: Request if authorized, error otherwise
│  Actions:
│    • Extract role from context
│    • Check: role == "admin"
│    • Pass to handler or reject (403)
────────────────────────────────────────
    │
    ▼
────────────────────────────────────────
│  Stage 3: Handler
├──────────────────────────────────────────
│  Input:  Authenticated + authorized request
│  Output: Response data
│  Actions:
│    • Business logic
│    • Database queries
│    • Return response (200, 400, 500, etc.)
────────────────────────────────────────
    │
    ▼
Response to client
```

## Role Decision Tree

```
                           User Registration/Login
                                   │
                                   ▼
                        Get email from request
                                   │
                                   ▼
                    ┌──────────────────────────┐
                    │ Email in admin list?     │
                    └──────────────────────────┘
                           │          │
                          YES        NO
                           │          │
                    ┌──────▼──┐  ┌───▼─────┐
                    │ Admin   │  │ User    │
                    │ role    │  │ role    │
                    └──────┬──┘  └───┬─────┘
                           │         │
                           └────┬────┘
                                │
                                ▼
                  Generate JWT with role claim
                                │
                                ▼
                   Send token to frontend
                                │
                    ┌───────────┴───────────┐
                    │                       │
              ┌─────▼────┐          ┌──────▼────┐
              │ role:    │          │ role:    │
              │ "admin"  │          │ "user"   │
              └─────┬────┘          └──────┬────┘
                    │                      │
          ┌─────────▼────┐       ┌────────▼──────┐
          │ Show admin   │       │ No admin link │
          │ link in nav  │       │ in header     │
          └─────┬────────┘       └────────┬──────┘
                │                        │
                │                        │
          ┌─────▼──────────┐    ┌───────▼────────┐
          │ Can access:    │    │ Can access:    │
          │ /api/admin/*   │    │ /feed          │
          │ (200 OK)       │    │ /report        │
          │                │    │ /comment       │
          │ Cannot access: │    │ /upvote        │
          │ Regular routes │    │ (200 OK)       │
          │ (403 Forbidden)│    │                │
          │                │    │ Cannot access: │
          │                │    │ /api/admin/*   │
          │                │    │ (403 Forbidden)│
          └────────────────┘    └────────────────┘
```

## Storage & Context Flow

```
JWT Token
└─ Contains:
   ├─ user_id (subject)
   ├─ role (claim)
   ├─ exp (expiration)
   └─ iat (issued at)
        │
        │ (transmitted in request)
        │
        ▼
Request Context (Server)
├─ ContextUserID → user_id
├─ ContextUserRole → role
└─ Used by middleware & handlers
        │
        │ (not sent back)
        │
        (JWT kept with response)


Frontend localStorage (Browser)
├─ jwt → Full JWT token
├─ uc_user → Username (for display)
├─ uc_role → Extracted role (for UI logic)
├─ uc_email → User email
├─ google_id → Google ID (if OAuth)
└─ api_base → Backend URL
        │
        │ (used for page rendering)
        │
        ▼
UI Decision Logic
├─ show admin link? → if uc_role == "admin"
├─ show admin dashboard? → if uc_role == "admin"
├─ allow status update? → if uc_role == "admin"
└─ etc...
```

## Error Response Flowchart

```
API Request to admin endpoint
    │
    ▼
┌─────────────────────────────┐
│ Valid JWT?                  │
├─────────────────────────────┤
│ NO  → 401 Unauthorized      │
│       "invalid or expired   │
│        token"               │
│ YES → Continue              │
└──────────┬──────────────────┘
           │
           ▼
┌─────────────────────────────┐
│ Authenticated?              │
│ (JWT signature valid?)      │
├─────────────────────────────┤
│ NO  → 401 Unauthorized      │
│       "invalid or expired   │
│        token"               │
│ YES → Continue              │
└──────────┬──────────────────┘
           │
           ▼
┌─────────────────────────────┐
│ Extract role from JWT       │
├─────────────────────────────┤
│ role = claims["role"]       │
│ (defaults to "user" if none)│
└──────────┬──────────────────┘
           │
           ▼
┌─────────────────────────────┐
│ role == "admin"?            │
├─────────────────────────────┤
│ NO  → 403 Forbidden         │
│       "admin access required"
│ YES → Continue to handler   │
└──────────┬──────────────────┘
           │
           ▼
Handler processes request
    │
    ├─ Success → 200 OK (+ data)
    ├─ Bad input → 400 Bad Request
    └─ Server error → 500 Internal Server Error
```

## Frontend UI Flow

```
Page Load
    │
    ▼
Load common.js
    │
    ├─ DOMContentLoaded event
    │
    ▼
updateAuthUI() called
    │
    ├─ Check localStorage.jwt
    │
    ├─ Check localStorage.uc_role
    │
    ▼
┌─────────────────────────┐
│ JWT exists?             │
├─────────────────────────┤
│ NO  → User not logged   │
│       ├─ Hide logout    │
│       ├─ Show login     │
│       └─ Hide admin     │
│ YES → User logged in    │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ role == "admin"?        │
├─────────────────────────┤
│ NO  → Regular user      │
│       ├─ Show user menu │
│       ├─ Hide admin     │
│ YES → Admin user        │
│       ├─ Show user menu │
│       ├─ Show admin ✓   │
└─────────────────────────┘
         │
         ▼
Navigation bar updated
    │
    ├─ Links visible/hidden correctly
    └─ Ready for interaction
```

## Summary: Defense in Depth

```
Layer 1: FRONTEND
┌─────────────────────────────────────┐
│ • JavaScript checks localStorage    │
│ • Hide admin UI from non-admins     │
│ • Prevent navigation to admin pages │
│ • Redirect if role not "admin"      │
│ Protection: User Experience         │
└────────────┬────────────────────────┘
             │ (Cannot be bypassed,
             │  but user can try)
             │
Layer 2: NETWORK
┌─────────────────────────────────────┐
│ • JWT transmitted in HTTP header    │
│ • HTTPS recommended (not enforced)  │
│ • JWT is signed, not encrypted      │
│ Protection: Token Integrity         │
└────────────┬────────────────────────┘
             │ (User can see role
             │  in JWT payload but
             │  cannot modify it)
             │
Layer 3: API/BACKEND
┌─────────────────────────────────────┐
│ • Validate JWT signature (must work)│
│ • Extract user from JWT claims      │
│ • Check role == "admin" (must work) │
│ • Reject if not admin (403)         │
│ Protection: Access Control          │
└─────────────────────────────────────┘
             │
             ▼ (Even if user
             │  forges JWT or
             │  modifies claims,
    SECURITY backend rejects it)
    ✓✓✓✓✓✓✓✓✓
```

This ensures:
- Even if JavaScript is disabled → Backend still enforces
- Even if user modifies localStorage → JWT validation fails
- Even if user forges JWT → Signature verification fails
- Even if user changes JWT claims → Role verification fails
