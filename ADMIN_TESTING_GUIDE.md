# Admin Dashboard Testing Guide

## Quick Start for Testing

### Prerequisites
- Backend running and compiled
- Frontend served (e.g., via HTTP server on port 8080 or similar)
- PostgreSQL database populated with test data
- Google OAuth configured (or use register/login endpoints)

---

## Testing Scenarios

### 1. Test Regular User (No Admin Access)

#### Steps:
1. Register a new user with email: `user@example.com`
   - Go to `login2.html`
   - Click "Register" or use Google Sign-In with a non-admin email
2. Verify localStorage:
   - Open DevTools (F12) → Application → LocalStorage
   - Check: `uc_role` should be `"user"`
3. Check header:
   - The "Admin" link should NOT appear in the navigation
4. Try to access admin dashboard:
   - Navigate to `admin-dashboard.html` directly
   - Should see alert: "You do not have admin access"
   - Should be redirected to `profile.html`

#### Expected Result:
✅ Regular users cannot access admin dashboard
✅ Admin link hidden from navigation

---

### 2. Test Admin User (Full Access)

#### Steps:
1. Login with admin email: `admin@example.com`
   - Register with `admin@example.com` and any password
   - Or use Google Sign-In with the admin email
2. Verify localStorage:
   - Open DevTools → Application → LocalStorage
   - Check: `uc_role` should be `"admin"`
3. Check header:
   - The "Admin" link should appear in the navigation
   - Click it to navigate to admin dashboard
4. Test admin dashboard:
   - Page loads successfully
   - Filter buttons work (All, Open, In Progress, Closed)
   - Click "View & Update" on an issue
   - Modal opens with issue details, comments, status dropdown
   - Change status and add optional notes
   - Click "Update Status" button
   - Issue status should update immediately

#### Expected Result:
✅ Admin user can access dashboard
✅ Admin link visible in header
✅ Status updates work
✅ Admin notes are added as comments

---

### 3. Test Backend RBAC Enforcement

#### Using curl (from terminal):

##### Get Regular User JWT:
```bash
# Register regular user
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"testuser@example.com","password":"password123"}'

# Response contains: {"access_token":"eyJ..."}
# Save this as: USER_JWT
```

##### Try to Access Admin Endpoint (Should Fail with 403):
```bash
curl -X GET http://localhost:8080/api/admin/issues \
  -H "Authorization: Bearer $USER_JWT"

# Expected response: {"error":"admin access required"} with 403 status
```

##### Get Admin JWT:
```bash
# Register admin user
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Admin User","email":"admin@example.com","password":"password123"}'

# Response contains JWT token
# Save this as: ADMIN_JWT
```

##### Access Admin Endpoint (Should Succeed with 200):
```bash
curl -X GET http://localhost:8080/api/admin/issues \
  -H "Authorization: Bearer $ADMIN_JWT"

# Expected response: JSON array of all posts
```

##### Update Issue Status (Admin Only):
```bash
curl -X POST http://localhost:8080/api/admin/post-status \
  -H "Authorization: Bearer $ADMIN_JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "post_id":"<EXISTING_POST_UUID>",
    "status":"closed",
    "notes":"Issue resolved by admin"
  }'

# Expected response: Updated post with status changed
```

#### Expected Results:
✅ Regular user JWT returns 403 on admin endpoints
✅ Admin user JWT returns 200 and valid data
✅ Status updates work with admin JWT
✅ Non-admin users cannot update status (403 Forbidden)

---

### 4. Test JWT Role Extraction

#### In Browser Console:
```javascript
// After login, check JWT payload
const jwt = localStorage.getItem('jwt');
const parts = jwt.split('.');
const payload = JSON.parse(atob(parts[1]));
console.log(payload); // Should show: { user_id: "...", role: "admin" or "user", exp: ..., iat: ... }
console.log(localStorage.getItem('uc_role')); // Should match payload.role
```

#### Expected Results:
✅ JWT contains `role` claim
✅ localStorage `uc_role` matches JWT role
✅ Admin users have `role: "admin"`
✅ Regular users have `role: "user"`

---

### 5. Test Filter Functionality

#### Steps:
1. Login as admin
2. Go to admin dashboard
3. Click filter buttons:
   - "All Issues" - shows all issues
   - "Open" - shows only status="open"
   - "In Progress" - shows only status="inprogress"
   - "Closed" - shows only status="closed"
4. Verify issue counts update
5. Verify URL changes: `?status=open`, `?status=inprogress`, etc.

#### Expected Results:
✅ Filters work correctly
✅ Issue list updates when filter changes
✅ URL params update (helps with bookmarking filtered views)

---

### 6. Test Session Persistence

#### Steps:
1. Login as admin
2. Open browser DevTools → Application → LocalStorage
3. Note the `jwt` and `uc_role` values
4. Refresh the page (Ctrl+R)
5. Check that:
   - Admin link still visible
   - Dashboard still loads
   - Can still update statuses
6. Close and reopen browser tab
7. Verify admin access persists (until JWT expires at 15 minutes)

#### Expected Results:
✅ Admin access persists across page refreshes
✅ LocalStorage correctly restored on page load
✅ JWT remains valid for 15 minutes
✅ After 15 minutes, user is logged out (requires re-login)

---

### 7. Test Error Scenarios

#### Missing JWT:
1. Clear localStorage: `localStorage.clear()`
2. Navigate to `admin-dashboard.html`
3. Should redirect to `login2.html`

#### Non-Admin User on Dashboard:
1. Login as regular user
2. Manually navigate to `admin-dashboard.html` (type in URL)
3. Should see alert and redirect to `profile.html`

#### Invalid/Expired JWT:
```bash
curl -X GET http://localhost:8080/api/admin/issues \
  -H "Authorization: Bearer invalid-token"

# Expected: 401 Unauthorized with "invalid or expired token"
```

#### Wrong HTTP Method:
```bash
curl -X POST http://localhost:8080/api/admin/issues \
  -H "Authorization: Bearer $ADMIN_JWT"

# Expected: 405 Method Not Allowed
```

#### Missing Required Fields in Update:
```bash
curl -X POST http://localhost:8080/api/admin/post-status \
  -H "Authorization: Bearer $ADMIN_JWT" \
  -H "Content-Type: application/json" \
  -d '{}'

# Expected: 400 Bad Request (missing post_id)
```

#### Expected Results:
✅ All error cases return appropriate HTTP status codes
✅ Clear error messages in response
✅ No sensitive information leaked in errors

---

## Manual Testing Checklist

### Frontend:
- [ ] Regular user cannot see Admin link
- [ ] Admin user sees Admin link in header
- [ ] Admin user can navigate to dashboard
- [ ] Non-admin cannot access dashboard (redirect)
- [ ] Dashboard loads issue grid correctly
- [ ] Filters work (All/Open/InProgress/Closed)
- [ ] Click "View & Update" opens modal
- [ ] Modal shows issue details and comments
- [ ] Status dropdown has 3 options (Open, In Progress, Closed)
- [ ] Can add admin notes in textarea
- [ ] "Update Status" button works
- [ ] Modal closes after update
- [ ] Issue grid refreshes with new status

### Backend:
- [ ] Regular user 403 on `/api/admin/issues`
- [ ] Admin user 200 on `/api/admin/issues`
- [ ] Admin user can POST to `/api/admin/post-status`
- [ ] Status update reflects in database
- [ ] Admin notes are stored as comment with [Admin] prefix
- [ ] Invalid status value rejected (400 Bad Request)
- [ ] Missing JWT returns 401
- [ ] Expired JWT returns 401

### Security:
- [ ] JWT contains role claim
- [ ] Role cannot be modified by client
- [ ] Backend always validates role (defense in depth)
- [ ] Admin list is checked at authentication time
- [ ] Role stored in localStorage (for UI only)
- [ ] Backend enforces even if frontend bypassed

---

## Troubleshooting

### Admin link not showing:
1. Check `localStorage.getItem('uc_role')` in console
2. If empty, try logging out and logging back in
3. If set to "user", user is not admin (check email in admin list)
4. Refresh page if updateAuthUI() wasn't called

### Cannot access admin dashboard:
1. Check JWT is stored: `localStorage.getItem('jwt')`
2. Check role: `localStorage.getItem('uc_role')`
3. If role is "user", user is not admin
4. If JWT missing, must login first
5. Check browser console for JavaScript errors

### API returns 403 Forbidden:
1. Verify JWT is valid: check it's not expired
2. Check role in JWT: `JSON.parse(atob(jwt.split('.')[1]))`
3. If role is not "admin", user doesn't have permission
4. Verify admin email is in hardcoded list (or update code)

### Status update doesn't work:
1. Check network tab in DevTools for POST response
2. Verify issue UUID is valid
3. Verify status is one of: "open", "inprogress", "closed"
4. Check server logs for any errors
5. Try with curl to isolate frontend vs backend issue

---

## Next Steps

After successful testing:

1. **Deploy to production**: Build Docker image and push
2. **Configure admin list**: Move from hardcoded map to env vars/database
3. **Add audit logging**: Log all admin actions with timestamps
4. **Implement role management**: UI for admins to grant/revoke roles
5. **Add more roles**: Consider "moderator", "department_staff" roles
6. **Implement token refresh**: Allow longer sessions with refresh tokens

