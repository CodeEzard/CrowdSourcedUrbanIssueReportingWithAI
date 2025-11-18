# ML Model Integration - Where It's Called

## Overview
The Urban Civic platform integrates with two external ML APIs for intelligent issue classification and urgency prediction. The models are called through service functions that are invoked at various points in the reporting workflow.

---

## ML Endpoints Configured

### 1. Urgency Prediction API
**Purpose**: Predict issue urgency level (1-3) from text description

**Endpoint**: 
```
https://urgency-api-latest.onrender.com/predict
```

**Configured Via**: `.env` file
```
ML_API_URL="https://urgency-api-latest.onrender.com/predict"
```

**Expected Input**:
```json
{
  "text": "Issue description here"
}
```

**Expected Output** (supports multiple formats):
```json
{
  "score": 0.85,
  "label": "critical",
  "urgency": 3,
  "confidence": 0.92
}
```

---

### 2. Image Classification API
**Purpose**: Classify issue category from image (Lighting, Road, Sanitation, etc.)

**Endpoint**:
```
https://issue-classification-api.onrender.com/predict
```

**Configured Via**: `.env` file
```
IMAGE_CLASSIFICATION_API_URL="https://issue-classification-api.onrender.com/predict"
```

**Expected Input**: Multipart form with `image_url` field
```
POST /predict
Content-Type: multipart/form-data

image_url=https://...
```

**Expected Output** (supports multiple formats):
```json
{
  "predicted_class": "Road",
  "classification": "Sanitation",
  "class": "Lighting"
}
```

---

## Service Functions

### File: `backend/internal/services/ml.go`

#### Function 1: `PredictUrgency(text string) (int, error)`
**Purpose**: Predict urgency bucket (1-3) from text

**Location**: `ml.go:23-26`

**How it works**:
1. Calls ML API endpoint configured in `ML_API_URL`
2. Sends POST request with `{"text": description}`
3. Receives score/label from API
4. Maps to urgency level (1=low, 2=moderate, 3=critical)
5. Returns (urgency, error)

**Fallback**: If ML API unavailable/fails, uses heuristic scoring

---

#### Function 2: `PredictUrgencyDetailed(text string) (int, float64, error)`
**Purpose**: Predict urgency AND continuous confidence score

**Location**: `ml.go:28-174`

**How it works**:
1. Checks if ML API is configured (`.env` `ML_API_URL`)
2. If no URL → uses local heuristic scoring
3. If URL exists → makes HTTP POST to ML API
4. Parses response flexibly (supports `score`, `label`, `urgency`, `confidence` fields)
5. Returns (urgencyLevel, confidenceScore, error)

**Supported Response Formats**:
- Direct score field: `{"score": 0.85}`
- Label mapping: `{"label": "critical"}` → maps to urgency 3
- Numeric urgency: `{"urgency": 3}`
- Confidence field: `{"confidence": 0.92}`

**Heuristic Fallback**:
- Critical keywords: "emergency", "danger", "fire", "explosion", "injury" → score 0.85
- Moderate keywords: "broken", "delay", "blocked", "leak", "issue", "problem" → score 0.6
- Default: 0.3

---

#### Function 3: `ClassifyImage(imageURL string) (string, error)`
**Purpose**: Classify issue category from image URL

**Location**: `ml.go:179-248`

**How it works**:
1. Checks if image classification API is configured (`.env` `IMAGE_CLASSIFICATION_API_URL`)
2. If no URL → returns empty string (feature disabled)
3. If URL exists → makes multipart POST to API with image_url field
4. Parses response (supports `predicted_class`, `classification`, `class` fields)
5. Returns (classificationLabel, error)

**Supported Response Formats**:
- `{"predicted_class": "Road"}`
- `{"classification": "Sanitation"}`
- `{"class": "Lighting"}`

---

## Where the ML Models Are Called

### 1. When Creating a New Report

**File**: `backend/internal/handlers/handlers.go`
**Function**: `ServeReport()` (report submission endpoint)

**Flow**:
```go
1. User submits report via POST /report
2. Handler validates JWT
3. Handler calls reportService.SavePost() with image URL
4. SavePost() calls:
   - PredictUrgencyDetailed(description) → get urgency score
   - ClassifyImage(mediaURL) → get issue category
5. Results stored in database
6. Response returned to frontend
```

**Code Reference** (in `services.go`):
```go
func (s *ReportService) SavePost(...) {
    // Line 37-42: Predict urgency
    if u, sc, err := PredictUrgencyDetailed(postDesc); err == nil {
        post.Urgency = u
        post.ScoreSum = sc
    }
    
    // Line 48-52: Classify image
    if classified, err := ClassifyImage(mediaURL); err == nil && classified != "" {
        post.ClassifiedAs = classified
    }
    
    // Save to database
    ...
}
```

---

### 2. When Computing Feed Scores

**File**: `backend/internal/services/services.go`
**Function**: `GetFeedPosts()` (fetch issues with scoring)

**Controlled By**: `.env` setting
```
FEED_SCORING_MODE=incremental  # "ml", "heuristic", "none", or "incremental"
```

**How it works** (see `services.go:100-150`):
```go
for each post in database:
    if FEED_SCORING_MODE == "incremental":
        // Use stored urgency + upvote count (no new ML call)
        score = post.ScoreSum * 0.8 + upvoteCount * 0.2
    
    else if FEED_SCORING_MODE == "heuristic":
        // Re-score using heuristic keywords
        score = heuristicScore(post.Description)
    
    else if FEED_SCORING_MODE == "ml":
        // Re-score by calling ML API again (expensive!)
        score = PredictUrgencyDetailed(post.Description)
    
    post.ComputedScore = score
    
sort posts by score descending
return top N posts
```

**Note**: Default mode is "incremental" for performance (no new ML calls per feed request)

---

### 3. When User Comments

**File**: `backend/internal/services/services.go`
**Function**: `SaveComment()` (save comment to post)

**Flow** (Line 114-144):
```go
1. User submits comment via POST /comment
2. Handler calls SaveComment()
3. SaveComment() calls PredictUrgencyDetailed(comment.Content)
4. If ML returns higher urgency, post urgency is updated
5. Comment stored with computed score
6. Feed re-scored if enabled
```

---

### 4. When User Upvotes

**File**: `backend/internal/services/services.go`
**Function**: `ToggleUpvote()` (upvote/remove upvote)

**Flow** (Line 220-260):
```go
1. User toggles upvote via POST /upvote
2. Handler calls ToggleUpvote()
3. Score updated: ScoreSum += upvoteCount * UPVOTE_SCORE
4. Feed re-scores all posts using incremental mode (if enabled)
```

---

## Frontend Integration

### Endpoint 1: Real-Time Urgency Prediction
**Path**: `POST /predict-urgency`
**Handler**: `backend/internal/handlers/ml_handlers.go:ServeP redictUrgency()`

**Used in**: Report form as user types description

```javascript
// frontend/report.html
async function predictUrgency(text) {
  const res = await fetch('/predict-urgency', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ text })
  });
  const result = await res.json();
  displayUrgency(result.urgency); // Show to user
}
```

---

### Endpoint 2: Real-Time Image Classification
**Path**: `POST /classify-image`
**Handler**: `backend/internal/handlers/ml_handlers.go:ServeClassifyImage()`

**Used in**: Report form after image upload

```javascript
// frontend/report.html
async function classifyImage(imageUrl) {
  const res = await fetch('/classify-image', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ image_url: imageUrl })
  });
  const result = await res.json();
  displayCategory(result.predicted_class); // Show to user
}
```

---

## Configuration & Timeouts

### File: `backend/configs/config.go`

**ML Text Timeout** (for urgency prediction):
```go
GetMLTextTimeout() → defaults to 15000ms (15 seconds)
Configured via: ML_TEXT_TIMEOUT_MS=15000
```

**ML Image Timeout** (for image classification):
```go
GetMLImageTimeout() → defaults to 20000ms (20 seconds)
Configured via: ML_IMAGE_TIMEOUT_MS=20000
```

**Feed Scoring Mode**:
```go
GetFeedScoringMode() → "incremental" (default)
Options: "ml", "heuristic", "none", "incremental"
```

---

## Call Sequence Diagram

### Workflow 1: User Submits Report

```
User fills form in report.html
  ↓
Real-time: /predict-urgency called (shows suggested urgency)
Real-time: /classify-image called (shows suggested category)
  ↓
User clicks "Submit Report"
  ↓
POST /report sent to backend
  ↓
ServeReport() handler receives request
  ↓
Calls SavePost() service
  ↓
SavePost() calls:
  ├─ PredictUrgencyDetailed() → ML API (if configured)
  └─ ClassifyImage() → Image Classification API (if configured)
  ↓
Results stored in database
  ↓
Response sent to frontend
  ↓
Frontend updates issue grid with new post
```

### Workflow 2: Feed is Requested

```
User navigates to /map.html or views feed
  ↓
GET /feed requested
  ↓
GetFeedPosts() service called
  ↓
Check FEED_SCORING_MODE setting
  ├─ if "incremental": Use stored scores (no ML call)
  ├─ if "heuristic": Re-score with keywords (no ML call)
  ├─ if "ml": Call ML API for EACH post (expensive!)
  └─ if "none": No scoring, return all posts
  ↓
Posts sorted by score
  ↓
Top N posts returned (N = FEED_LIMIT)
  ↓
Frontend renders issue grid
```

---

## Error Handling

### When ML API is Unavailable
```
If ML API returns error (timeout, 500, etc.)
  ↓
Function catches error
  ↓
Falls back to heuristic scoring
  ↓
Returns computed score (never returns error to user)
  ↓
Report is still created/feed still works
```

**Example** (from `ml.go`):
```go
// Network error → fallback to heuristic
if err != nil {
    score := heuristicScore(text)
    return mapScoreToUrgency(score), score, nil
}

// Non-2xx status → fallback to heuristic
if resp.StatusCode < 200 || resp.StatusCode >= 300 {
    score := heuristicScore(text)
    return mapScoreToUrgency(score), score, nil
}
```

---

## Summary: Where ML is Called

| Where | What | When | Required |
|-------|------|------|----------|
| `SavePost()` | Urgency prediction | Report submitted | No (heuristic fallback) |
| `SavePost()` | Image classification | Report submitted | No (returns empty if disabled) |
| `GetFeedPosts()` | Re-score posts | Feed requested | No (depends on FEED_SCORING_MODE) |
| `SaveComment()` | Update urgency | Comment posted | No (only if urgency increases) |
| `ToggleUpvote()` | Update score | User upvotes | No (incremental scoring built-in) |
| `/predict-urgency` | Real-time suggestion | User typing | No (frontend only) |
| `/classify-image` | Real-time suggestion | After image upload | No (frontend only) |

---

## Key Files

```
backend/
├── internal/
│   ├── handlers/
│   │   ├── ml_handlers.go          ← Frontend endpoints for real-time prediction
│   │   └── handlers.go             ← ServeReport() calls SavePost()
│   └── services/
│       ├── ml.go                   ← PredictUrgency(), ClassifyImage()
│       └── services.go             ← SavePost(), GetFeedPosts(), SaveComment()
└── configs/
    └── config.go                   ← ML endpoint URLs and timeouts
.env                                 ← ML_API_URL, IMAGE_CLASSIFICATION_API_URL
```

---

## Testing ML Integration

### Test Urgency Prediction (curl):
```bash
curl -X POST http://localhost:8080/predict-urgency \
  -H "Content-Type: application/json" \
  -d '{"text":"Fire in building, people trapped!"}'

# Response: {"urgency":3,"error":""}
```

### Test Image Classification (curl):
```bash
curl -X POST http://localhost:8080/classify-image \
  -H "Content-Type: application/json" \
  -d '{"image_url":"https://example.com/image.jpg"}'

# Response: {"predicted_class":"Road","error":""}
```

### Disable ML (use heuristics only):
```bash
# In .env, comment out or remove:
# ML_API_URL="..."
# IMAGE_CLASSIFICATION_API_URL="..."

# System will use heuristic scoring instead
```

---

## Notes

1. **ML APIs are optional** - System works with or without them
2. **Heuristic fallback** - If ML APIs unavailable, local keyword scoring used
3. **Real-time frontend endpoints** - `/predict-urgency` and `/classify-image` for UX suggestions
4. **Async server-side** - SavePost/SaveComment/ToggleUpvote call ML in background
5. **Performance** - FEED_SCORING_MODE="incremental" avoids expensive ML calls on every feed request
6. **Flexible response parsing** - ML APIs can return different JSON structures, all supported
