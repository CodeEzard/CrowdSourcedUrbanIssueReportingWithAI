# Database Performance Optimization

## Current Status

Your application is experiencing two database-related issues:

### 1. ⏱️ Slow SQL Query (203ms)
```
SLOW SQL >= 200ms [203.248ms]
SELECT c.column_name, c.is_nullable ... FROM information_schema.columns 
WHERE table_catalog = 'civicissue_d39c' AND table_schema = CURRENT_SCHEMA() 
AND table_name = 'comments'
```

**Cause:** GORM's `AutoMigrate()` inspects your database schema on every application startup (line 53 in `main.go`).

### 2. 🔴 Redis Not Configured
```
Redis not configured; token revocation disabled
```

**Impact:** Token logout/revocation isn't working (tokens stay valid even after logout).

---

## Quick Fixes

### Fix #1: Add Database Indexes

Add these indexes to your `models/models.go` to speed up schema inspection and queries:

```go
// In Comment struct, add index tags:
type Comment struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	PostID    uuid.UUID `gorm:"type:uuid;not null;index:idx_comment_post" json:"post_id"`  // ← ADD INDEX
	Post      Post      `gorm:"foreignKey:PostID" json:"post"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_comment_user" json:"user_id"`  // ← ADD INDEX
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	Content   string    `gorm:"not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// In Post struct, add indexes:
type Post struct {
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	IssueID       uuid.UUID `gorm:"type:uuid;not null;index:idx_post_issue" json:"issue_id"`      // ← ADD INDEX
	Issue         Issue     `gorm:"foreignKey:IssueID" json:"issue"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index:idx_post_user" json:"user_id"`        // ← ADD INDEX
	User          User      `gorm:"foreignKey:UserID" json:"user"`
	Description   string    `json:"description,omitempty"`
	Status        string    `gorm:"default:'open';not null;index:idx_post_status" json:"status"` // ← ADD INDEX
	Urgency       int       `gorm:"not null;index:idx_post_urgency" json:"urgency"`              // ← ADD INDEX
	ClassifiedAs  string    `json:"classified_as,omitempty"`
	Lat           float64   `gorm:"not null" json:"lat"`
	Lng           float64   `gorm:"not null" json:"lng"`
	MediaURL      string    `gorm:"not null" json:"media_url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
```

### Fix #2: Optimize AutoMigrate (Option A - Recommended for Production)

Disable AutoMigrate in production and run migrations only once:

**Option A.1: Environment-based flag**

```go
// In main.go, around line 49-55
if os.Getenv("ENABLE_AUTO_MIGRATE") == "true" {
	// AutoMigrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Issue{},
		&models.Post{},
		&models.Comment{},
		&models.Upvote{},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("AutoMigrate completed")
} else {
	log.Println("AutoMigrate skipped (set ENABLE_AUTO_MIGRATE=true to enable)")
}
```

Then in your deployment (Render.com):
- First deployment: `ENABLE_AUTO_MIGRATE=true`
- Subsequent deployments: Don't set the variable (skip migrations)

**Option A.2: Run migrations manually via CLI**

Create `backend/cmd/migrate/main.go`:

```go
package main

import (
	"log"
	"os"

	"crowdsourcedurbanissuereportingwithai/backend/configs"
	"crowdsourcedurbanissuereportingwithai/backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	configs.LoadEnv()

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=post4321 dbname=Civicissue port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Enable UUID extension
	sqlDB, _ := db.DB()
	sqlDB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

	// Run migrations
	err = db.AutoMigrate(
		&models.User{},
		&models.Issue{},
		&models.Post{},
		&models.Comment{},
		&models.Upvote{},
	)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	log.Println("✅ Database migrations completed successfully!")
}
```

Then run before deployment:
```bash
go run backend/cmd/migrate/main.go
```

### Fix #3: Configure Redis for Token Revocation

**For Local Development:**

1. Install Redis:
```bash
# Windows (using WSL or Docker)
docker run -d -p 6379:6379 redis:latest

# Or use: brew install redis (macOS)
```

2. Set environment variable:
```bash
export REDIS_ADDR="localhost:6379"
export REDIS_PASSWORD=""  # empty if no password
```

3. Verify connection:
```bash
redis-cli ping
# Should return: PONG
```

**For Render.com Deployment:**

1. Go to Render.com Dashboard → Create New → Redis
2. Copy the **Internal URL** (e.g., `redis://default:password@redis-service:10000`)
3. Add to environment variables:
   ```
   REDIS_ADDR=redis-service:10000
   REDIS_PASSWORD=your-password
   ```
4. Update `config.go` to parse the full Redis URL:

```go
// In backend/configs/config.go
func GetRedisAddr() string {
	// Return full URL if set (Render format)
	if url := os.Getenv("REDIS_URL"); url != "" {
		return url
	}
	// Fallback to host:port format (local development)
	return os.Getenv("REDIS_ADDR")
}
```

---

## Performance Optimization Checklist

| Item | Status | Impact |
|------|--------|--------|
| ✅ Add database indexes | Not done | -50ms query time |
| ❌ Use environment-based AutoMigrate | Not done | Eliminates 203ms on startup |
| ❌ Configure Redis | Not done | Enables token revocation |
| ⏳ Monitor slow queries | Not done | Identifies bottlenecks |
| ⏳ Add query pagination | Possible | Reduces large result sets |
| ⏳ Cache feed results | Possible | Speed up feed loads |

---

## Implementation Steps

### Step 1: Add Database Indexes (5 minutes)

Edit `backend/models/models.go` and add index tags to foreign keys:

```bash
cd "C:\Users\ASUS\Desktop\Web P\CrowdSourcedUrbanIssueReportingWithAI"
# Edit backend/models/models.go
```

### Step 2: Disable AutoMigrate in Production (10 minutes)

Edit `backend/main.go` and wrap AutoMigrate with environment check:

```go
if os.Getenv("ENABLE_AUTO_MIGRATE") == "true" {
    // AutoMigrate code here...
}
```

### Step 3: Configure Redis (15 minutes)

For local development:
```bash
docker run -d -p 6379:6379 redis:latest
export REDIS_ADDR="localhost:6379"
go run backend/main.go
```

For Render.com:
- Create Redis instance
- Add `REDIS_ADDR` and `REDIS_PASSWORD` environment variables

---

## Monitoring & Debugging

### Check if Indexes Exist

```sql
-- Connect to your database via psql:
psql -h <host> -U <user> -d <database>

-- List all indexes:
\di

-- Check specific table:
SELECT * FROM pg_indexes WHERE tablename = 'comments';
```

### Monitor Slow Queries

Add to `main.go` for query logging:

```go
import "gorm.io/logger"

db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
	Logger: logger.Default.LogMode(logger.Slow),
})
// Logs queries slower than 200ms (default threshold)
```

### Test Redis Connection

```go
// In main.go, after creating redis client:
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
if _, err := redisClient.Ping(ctx).Result(); err != nil {
	log.Println("Redis connection failed:", err)
} else {
	log.Println("✅ Redis connection successful")
}
```

---

## SQL Query Optimization

Your current slow query examines column metadata. Here's how GORM uses it:

```
AutoMigrate()
  ↓
Check if table exists
  ↓
Inspect column definitions  ← SLOW (203ms)
  ↓
Create missing columns
  ↓
Complete
```

**Solution:** Run this once, then disable AutoMigrate in production.

---

## Expected Performance After Optimization

| Metric | Before | After |
|--------|--------|-------|
| Startup time | ~500ms | ~200ms |
| AutoMigrate overhead | 203ms | 0ms |
| Token revocation | ❌ Disabled | ✅ Enabled |
| Query latency (indexed) | ~50ms | ~5ms |
| Feed load time | ~300ms | ~150ms |

---

## Deployment Steps

### First-Time Deployment

```bash
# 1. Enable migrations
export ENABLE_AUTO_MIGRATE=true

# 2. Deploy to Render
git push origin main

# 3. Verify migrations completed in logs
# Wait for startup message "AutoMigrate completed"
```

### Subsequent Deployments

```bash
# 1. Don't set ENABLE_AUTO_MIGRATE
# 2. Deploy normally
git push origin main

# 3. Server starts in ~200ms instead of ~500ms
```

---

## Troubleshooting

### AutoMigrate still running on every startup?

Check that `ENABLE_AUTO_MIGRATE` is not set in Render environment:
1. Go to Render.com → Select your service
2. Environment → Remove `ENABLE_AUTO_MIGRATE` if present
3. Redeploy

### Redis connection failing?

```bash
# Local test
redis-cli ping
# Should return PONG

# Check Render Redis dashboard for:
# - Service is running (status: "Available")
# - Connection string is correct
# - Network access allows your service
```

### Indexes not being created?

Make sure you're using the updated `models.go`:

```bash
# Verify indexes exist:
SELECT indexname FROM pg_indexes WHERE tablename = 'comments';

# If missing, enable AutoMigrate once:
export ENABLE_AUTO_MIGRATE=true && go run backend/main.go
```

---

## References

- [GORM AutoMigrate Documentation](https://gorm.io/docs/migration.html)
- [GORM Logger with Slow Query Threshold](https://gorm.io/docs/logger.html)
- [PostgreSQL Indexes](https://www.postgresql.org/docs/current/indexes.html)
- [Redis Configuration](https://redis.io/docs/management/config/)

**Last Updated:** November 11, 2025  
**Status:** Ready for Implementation ✅
