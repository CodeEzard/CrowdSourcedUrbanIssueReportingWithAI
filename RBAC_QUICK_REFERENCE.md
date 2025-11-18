# RBAC Quick Reference Card

## Roles

| Role | Permissions | NavBar | Dashboard |
|------|-------------|--------|-----------|
| `user` | Report, comment, upvote | Regular links | ❌ Access denied |
| `admin` | All user + status updates, issue management | + Admin link | ✅ Full access |

## Key Endpoints

### Public (No Auth Required)
- `GET /feed` - Get issue feed
- `POST /register` - Register new user
- `POST /login` - Login (email/password)
- `POST /google-login` - Google OAuth login

### Protected (JWT Required, Any User)
- `POST /report` - Submit new issue
- `POST /comment` - Add comment
- `POST /upvote` - Upvote issue
- `POST /logout` - Logout

### Admin Only (JWT + Role="admin")
- `GET /api/admin/issues` - List all issues
- `GET /api/admin/issues?status=open` - Filter by status
- `POST /api/admin/post-status` - Update issue status

## JWT Claims

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "role": "admin",
  "exp": 1234567890,
  "iat": 1234567800
}
```

## Admin Email List (Hardcoded)

Located in: `backend/internal/handlers/auth_handlers.go`

```go
adminEmails := map[string]bool{
    "admin@example.com": true,
}
```

**To add more admins**: Add email to map and redeploy

## Frontend Checklist

- [ ] Login → JWT stored in `localStorage.jwt`
- [ ] Login → Role extracted and stored in `localStorage.uc_role`
- [ ] updateAuthUI() → Admin link shows if role=="admin"
- [ ] updateAuthUI() → Admin link hidden if role=="user"
- [ ] Admin nav link → Links to `admin-dashboard.html`
- [ ] Admin dashboard → Checks JWT (redirects to login if missing)
- [ ] Admin dashboard → Checks role (redirects to profile if not admin)

## Backend Checklist

- [ ] JWT includes `role` claim
- [ ] AuthMiddleware extracts role
- [ ] AdminMiddleware checks role=="admin"
- [ ] Non-admin gets 403 on `/api/admin/*`
- [ ] Admin gets 200 on `/api/admin/*`
- [ ] Admin can update issue status
- [ ] Admin notes saved as comments

## Testing Commands

### Check User Role (Browser Console)
```javascript
const jwt = localStorage.getItem('jwt');
const payload = JSON.parse(atob(jwt.split('.')[1]));
console.log(payload.role); // "admin" or "user"
```

### Test Regular User (curl)
```bash
# Register
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"User","email":"user@test.com","password":"pass"}'

# Try admin endpoint (should get 403)
curl -X GET http://localhost:8080/api/admin/issues \
  -H "Authorization: Bearer $JWT"
```

### Test Admin User (curl)
```bash
# Register as admin
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Admin","email":"admin@example.com","password":"pass"}'

# Try admin endpoint (should get 200 with issues list)
curl -X GET http://localhost:8080/api/admin/issues \
  -H "Authorization: Bearer $JWT"
```

### Update Issue Status (curl)
```bash
curl -X POST http://localhost:8080/api/admin/post-status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_JWT" \
  -d '{
    "post_id":"<UUID>",
    "status":"closed",
    "notes":"Resolved by admin"
  }'
```

## Troubleshooting

| Problem | Check |
|---------|-------|
| Admin link not showing | Is `uc_role` = "admin" in localStorage? |
| Cannot access dashboard | Is `uc_role` checked in admin-dashboard.html? |
| API returns 403 | Is user's role "admin"? Is email in admin list? |
| JWT has no role | Did you use `GenerateTokenWithRole()`? |
| Role not in localStorage | Did login.html extract role from JWT? |
| updateAuthUI not called | Is `common.js` loaded? Is DOMContentLoaded fired? |

## Common Issues

### "You do not have admin access"
- User is not in admin email list
- Add email to `adminEmails` map in auth_handlers.go
- User must logout and login again

### Admin link appears but dashboard says "access denied"
- Frontend's `uc_role` is not "admin"
- Clear localStorage and login again
- Check browser console for JWT decoding errors

### API returns 401 instead of 403
- JWT is invalid or expired
- User must re-login
- Check JWT expiration in browser console

### API returns 200 but no issues shown
- No issues in database yet
- Create an issue first via the report form
- Or populate test data in database

## Files to Modify (If Changing Roles)

### Add Admin Email:
1. `backend/internal/handlers/auth_handlers.go`
   - Update `adminEmails` map in GoogleLogin()
   - Update `adminEmails` map in Login()
   - Update `adminEmails` map in Register()

### Add New Role:
1. `backend/internal/auth/jwt_service.go`
   - Update role constants (if needed)
2. `backend/internal/auth/middleware.go`
   - Create new middleware (e.g., ModeratorMiddleware)
3. `backend/router.go`
   - Wrap routes with new middleware
4. `backend/internal/handlers/*`
   - Update role assignment logic
5. `frontend/js/common.js`
   - Update updateAuthUI() to show/hide role-specific UI

## Environment Variables

None needed for RBAC (admin list is hardcoded).

**Recommended for production**:
```bash
ADMIN_EMAILS=admin@example.com,superadmin@example.com
```

Then parse and use in code instead of hardcoding.

## Deployment Steps

1. Compile backend: `go build -o backend ./backend`
2. Verify no errors: `go build ./backend` should succeed
3. Deploy backend (new endpoints work, JWT changes backward-compatible)
4. Deploy frontend (new admin link won't break anything)
5. Test with both regular and admin users
6. Monitor logs for any 403 Forbidden errors (expected for non-admins)

## Security Notes

⚠️ **WARNING**: Admin email list is currently hardcoded. Consider moving to:
- Environment variables (better)
- Database (best)
- Configuration file (secure)

⚠️ **JWT Secret**: Must be set in `.env` and kept secure
- If compromised, all JWTs can be forged
- Change secret after suspected compromise (all users must re-login)

⚠️ **HTTPS**: Recommended for production
- JWTs should only be transmitted over HTTPS
- Current implementation works over HTTP (dev-friendly)

## Next Steps

1. ✅ Implementation complete
2. ✅ Testing ready
3. 📋 Deploy and verify
4. 📋 Gather feedback from admins
5. 📋 Consider database-backed roles
6. 📋 Add audit logging
7. 📋 Implement refresh tokens (for longer sessions)

## Support

For detailed documentation, see:
- `RBAC_IMPLEMENTATION.md` - Full technical details
- `RBAC_ARCHITECTURE.md` - System design and flows
- `ADMIN_TESTING_GUIDE.md` - Testing procedures
- `RBAC_SUMMARY.md` - Implementation summary
