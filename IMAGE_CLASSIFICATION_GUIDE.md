# Image Classification Integration Guide

## Overview
Your backend now integrates with an external image classification API that analyzes images of issues and predicts their category (e.g., "potholes", "broken poles", etc.).

## APIs Integrated

### 1. **Urgency Prediction API** (Already Working ✅)
- **Endpoint:** `https://urgency-api-latest.onrender.com/predict`
- **Input:** `{"text": "issue description"}`
- **Output:** `{"label": "critical/moderate/low", "confidence": 0.99}`
- **Maps to urgency:** 1, 2, or 3

### 2. **Image Classification API** (Newly Integrated ✅)
- **Endpoint:** `https://issue-classification-api.onrender.com/predict`
- **Input:** multipart form with `image_url=<URL>`
- **Output:** `{"predicted_class": "potholes/broken_pole/etc"}`
- **Stored in:** Post.ClassifiedAs field

---

## Complete Integration Flow

When a user reports an issue with an image:

```
User submits report with:
- description: "There is a pothole on the road"
- media_url: "https://anonomz.com/wp-content/uploads/2014/04/potholes.jpg"
    ↓
Backend calls ReportIssueViaPost()
    ↓
[Parallel calls]
├─ PredictUrgency(description)
│  └─ ML API returns: {"label": "critical", ...}
│     └─ Mapped to urgency: 3
│
└─ ClassifyImage(media_url)
   └─ Image Classification API returns: {"predicted_class": "potholes"}
      └─ Stored as classified_as: "potholes"
    ↓
Report saved with:
- urgency: 3 (from text analysis)
- classified_as: "potholes" (from image analysis)
    ↓
Response sent to frontend with both predictions
```

---

## Testing Instructions

### Step 1: Start Backend with Both APIs Enabled

**Windows CMD:**
```cmd
cd backend
set DISABLE_AUTH=true
set ML_API_URL=https://urgency-api-latest.onrender.com/predict
set IMAGE_CLASSIFICATION_API_URL=https://issue-classification-api.onrender.com/predict
go run .
```

**Windows PowerShell:**
```powershell
cd backend
$env:DISABLE_AUTH="true"
$env:ML_API_URL="https://urgency-api-latest.onrender.com/predict"
$env:IMAGE_CLASSIFICATION_API_URL="https://issue-classification-api.onrender.com/predict"
go run .
```

**Unix/Linux/macOS:**
```bash
cd backend
export DISABLE_AUTH=true
export ML_API_URL="https://urgency-api-latest.onrender.com/predict"
export IMAGE_CLASSIFICATION_API_URL="https://issue-classification-api.onrender.com/predict"
go run .
```

### Step 2: Submit Test Report with Image

**Test 1: Pothole Report**
```bash
curl -X POST "http://localhost:8080/report" \
  -H "Content-Type: application/json" \
  -d '{
    "issue_name": "Pothole on Main Street",
    "issue_desc": "Large pothole affecting traffic",
    "issue_category": "Road",
    "post_desc": "There is a dangerous pothole on Main Street near downtown",
    "status": "open",
    "urgency": 1,
    "lat": 40.7128,
    "lng": -74.0060,
    "media_url": "https://anonomz.com/wp-content/uploads/2014/04/potholes.jpg"
  }'
```

**Expected Response:**
```json
{
  "id": "uuid-here",
  "description": "There is a dangerous pothole on Main Street near downtown",
  "status": "open",
  "urgency": 3,
  "classified_as": "potholes",
  "lat": 40.7128,
  "lng": -74.0060,
  "media_url": "https://anonomz.com/wp-content/uploads/2014/04/potholes.jpg",
  "created_at": "2025-11-11T..."
}
```

Notice:
- ✅ `urgency` changed from 1 → 3 (text analysis: "dangerous" = "critical")
- ✅ `classified_as` set to "potholes" (image analysis)

---

**Test 2: Broken Pole Report**
```bash
curl -X POST "http://localhost:8080/report" \
  -H "Content-Type: application/json" \
  -d '{
    "issue_name": "Broken Utility Pole",
    "issue_desc": "Damaged utility infrastructure",
    "issue_category": "Utilities",
    "post_desc": "There is a dangerous broken pole near the road that could collapse",
    "status": "open",
    "urgency": 1,
    "lat": 40.7200,
    "lng": -74.0080,
    "media_url": "https://en.wikipedia.org/wiki/File:Utility_pole_lean.jpg"
  }'
```

**Expected Response:**
```json
{
  "urgency": 3,
  "classified_as": "broken_pole"
}
```

---

## Backend Logs

Watch your terminal for these logs:

**Urgency Prediction:**
```
ml: urgency prediction - label: critical -> urgency: 3
```

**Image Classification:**
```
image_classification: predicted_class: potholes
image_classification: classification: broken_pole
```

---

## Response Format

The Post model now includes:

```json
{
  "id": "uuid",
  "issue_id": "uuid",
  "issue": {...},
  "user_id": "uuid",
  "user": {...},
  "description": "Issue description text",
  "status": "open",
  "urgency": 3,
  "classified_as": "potholes",
  "lat": 40.7128,
  "lng": -74.0060,
  "media_url": "https://...",
  "created_at": "2025-11-11T...",
  "updated_at": "2025-11-11T..."
}
```

---

## Data Model Changes

### Post Model Updated
```go
type Post struct {
    // ... existing fields ...
    Urgency       int       `gorm:"not null" json:"urgency"`
    ClassifiedAs  string    `json:"classified_as,omitempty"`    // ← NEW
    MediaURL      string    `gorm:"not null" json:"media_url"`
}
```

### Database Migration
The `ClassifiedAs` field will be automatically added to the posts table by GORM AutoMigrate (if enabled) or add manually:

```sql
ALTER TABLE posts ADD COLUMN classified_as VARCHAR(255);
```

---

## Configuration

### Environment Variables

| Variable | Required | Example | Purpose |
|----------|----------|---------|---------|
| `ML_API_URL` | No | `https://urgency-api-latest.onrender.com/predict` | Text urgency prediction |
| `IMAGE_CLASSIFICATION_API_URL` | No | `https://issue-classification-api.onrender.com/predict` | Image classification |
| `DISABLE_AUTH` | No | `true` | Skip authentication (dev only) |

Both APIs are **optional**. If not configured, those features are skipped gracefully.

---

## Error Handling

Both ML APIs are **non-blocking**:

1. **If image classification fails:**
   - Logs the error
   - Continues with report creation
   - `classified_as` remains empty
   - Report is still created and saved

2. **If urgency prediction fails:**
   - Logs the error
   - Uses original urgency value
   - Report is still created and saved

3. **If both fail:**
   - Both use fallback values
   - Report is created with submitted values
   - Errors are logged for debugging

---

## Code Changes Summary

### Files Modified

1. **`backend/models/models.go`**
   - Added `ClassifiedAs string` field to Post struct

2. **`backend/configs/config.go`**
   - Added `GetImageClassificationAPIURL()` function

3. **`backend/internal/services/ml.go`**
   - Added `ClassifyImage(imageURL string)` function
   - Handles multipart form POST
   - Parses `predicted_class`, `classification`, or `class` fields
   - Non-blocking error handling

4. **`backend/internal/services/services.go`**
   - Modified `ReportIssueViaPost()` to call `ClassifyImage()`
   - Passes `classifiedAs` to repository

5. **`backend/internal/repository/repository.go`**
   - Updated `ReportIssueViaPost()` signature to accept `classifiedAs`
   - Stores `ClassifiedAs` in Post struct before saving

---

## Verification Checklist

- [ ] Backend builds successfully
- [ ] Environment variables are set correctly
- [ ] Backend server starts without errors
- [ ] Can submit report with image URL
- [ ] Response includes both `urgency` and `classified_as` fields
- [ ] Backend logs show "ml: urgency prediction" messages
- [ ] Backend logs show "image_classification:" messages
- [ ] Database stores `classified_as` value correctly
- [ ] Feed returns posts with classification data

---

## Testing Different Images

You can test with different image URLs:

### Pothole Images
```
https://anonomz.com/wp-content/uploads/2014/04/potholes.jpg
https://townsquare.media/site/696/files/2019/02/Potholes.jpg?w=980&q=75
```

### Utility Pole Images
```
https://en.wikipedia.org/wiki/File:Utility_pole_lean.jpg
https://upload.wikimedia.org/wikipedia/commons/thumb/d/d0/Utility_pole_lean.jpg
```

### Testing API Directly

```bash
# Test image classification API directly
curl -X POST -F "image_url=https://anonomz.com/wp-content/uploads/2014/04/potholes.jpg" \
  https://issue-classification-api.onrender.com/predict
```

Expected response:
```json
{"predicted_class":"potholes"}
```

---

## Frontend Integration

When the frontend fetches issues from `/feed`, it now receives:

```json
{
  "posts": [
    {
      "id": "uuid",
      "description": "...",
      "urgency": 3,
      "classified_as": "potholes",
      "media_url": "https://..."
    }
  ]
}
```

Frontend can display:
- Issue urgency level (1-3 or color-coded)
- Auto-detected issue type from ML prediction
- Show confidence in predictions
- Filter by classified category

---

## Monitoring

Check logs for:

✅ **Success (Image classified):**
```
image_classification: predicted_class: potholes
```

✅ **Success (Text analyzed):**
```
ml: urgency prediction - label: critical -> urgency: 3
```

⚠️ **Graceful fallback (API not configured):**
```
[No log, feature skipped, report continues]
```

❌ **Error (Still succeeds with fallback):**
```
[Error logged, report still created with submitted values]
```

---

## Deployment

For production:

1. **Set environment variables in your deployment platform:**
   - Docker: Add to `.env` file or pass as `ENV` in Dockerfile
   - Kubernetes: Add to ConfigMap or Secrets
   - Cloud Platform: Set in environment configuration

2. **Database migration:**
   - Ensure `ClassifiedAs` column exists in posts table
   - GORM will auto-create if AutoMigrate is enabled

3. **No code changes needed:**
   - Just set the environment variables
   - Both APIs are optional
   - All existing functionality continues to work

---

## Next Steps

1. Test locally with the provided curl examples
2. Verify logs show both urgency and classification predictions
3. Check database to confirm `classified_as` is stored
4. Update frontend to display classification results
5. Deploy with both ML_API_URL and IMAGE_CLASSIFICATION_API_URL set

