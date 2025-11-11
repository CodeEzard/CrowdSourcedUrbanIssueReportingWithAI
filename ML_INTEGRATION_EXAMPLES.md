# ML Integration Examples

## Quick Start

### Start Backend with ML Enabled
```bash
cd backend
DISABLE_AUTH=true ML_API_URL="https://urgency-api-latest.onrender.com/predict" go run .
```

### Test 1: Report a Critical Issue
```bash
curl -X POST "http://localhost:8080/report" \
  -H "Content-Type: application/json" \
  -d '{
    "issue_name": "Dangerous Broken Pole",
    "issue_desc": "A broken utility pole near the main road",
    "issue_category": "Utilities",
    "post_desc": "There is a dangerous broken pole near the road that could collapse",
    "status": "open",
    "urgency": 1,
    "lat": 40.7128,
    "lng": -74.0060,
    "media_url": ""
  }'
```

**Expected urgency in response: 3** (mapped from "critical")

---

### Test 2: Report a Minor Issue
```bash
curl -X POST "http://localhost:8080/report" \
  -H "Content-Type: application/json" \
  -d '{
    "issue_name": "Slight Path Crack",
    "issue_desc": "Small crack in the sidewalk",
    "issue_category": "Road",
    "post_desc": "There is a minor crack in the sidewalk",
    "status": "open",
    "urgency": 1,
    "lat": 40.7128,
    "lng": -74.0060,
    "media_url": ""
  }'
```

**Expected urgency in response: 1** (mapped from "low")

---

### Test 3: Check Backend Logs
After running reports, check your terminal for logs like:
```
ml: urgency prediction - label: critical -> urgency: 3
ml: urgency prediction - label: low -> urgency: 1
```

---

## Understanding the Flow

1. **User submits report with description:** `"There is a dangerous broken pole near the road"`
2. **Backend calls `ReportIssueViaPost()`** in services.go
3. **Service calls `PredictUrgency(postDesc)`** from ml.go
4. **ML client POSTs to external API:** 
   ```json
   {
     "text": "There is a dangerous broken pole near the road"
   }
   ```
5. **API responds with:**
   ```json
   {
     "label": "critical",
     "confidence": 0.9948945045471191
   }
   ```
6. **Backend maps "critical" → urgency 3**
7. **Report is saved with urgency: 3** (overriding the submitted urgency: 1)
8. **Frontend displays urgency: 3** to users

---

## Failure Scenarios

### Scenario: ML API is down
- Request times out after 5 seconds
- Backend logs the error
- Report is saved with the **original urgency value**
- Issue reporting continues without interruption

### Scenario: ML_API_URL not set
- `PredictUrgency()` returns 0 immediately
- Report uses the **original urgency value**
- Feature is gracefully disabled

### Scenario: ML response format unexpected
- Response is logged for debugging
- Report uses the **original urgency value**
- No crash or data loss

---

## Integration Points

**Files involved:**
- `backend/configs/config.go` — Reads ML_API_URL environment variable
- `backend/internal/services/ml.go` — HTTP client and response parsing
- `backend/internal/services/services.go` — Calls PredictUrgency in ReportIssueViaPost
- `backend/models/models.go` — Post model has Urgency field
- `backend/internal/repository/repository.go` — Persists urgency to database

**Frontend display:**
- Frontend calls `/feed` or `/report` endpoint
- Gets back Post objects with urgency field
- Displays/renders urgency as part of issue details

---

## Configuration

### Environment Variables

| Variable | Required | Example | Purpose |
|----------|----------|---------|---------|
| `ML_API_URL` | No | `https://urgency-api-latest.onrender.com/predict` | External ML prediction endpoint |
| `DISABLE_AUTH` | No | `true` | Skip authentication (dev only) |

### Setting Variables

**Windows CMD:**
```cmd
set ML_API_URL=https://urgency-api-latest.onrender.com/predict
set DISABLE_AUTH=true
go run ./backend
```

**Windows PowerShell:**
```powershell
$env:ML_API_URL="https://urgency-api-latest.onrender.com/predict"
$env:DISABLE_AUTH="true"
go run ./backend
```

**Unix/Linux/macOS:**
```bash
export ML_API_URL="https://urgency-api-latest.onrender.com/predict"
export DISABLE_AUTH="true"
go run ./backend
```

---

## Testing with Frontend

1. Open `frontend/report.html` in browser
2. Ensure backend is running with ML_API_URL set
3. Fill out the form (choose a high-risk description)
4. Submit the report
5. Frontend calls `/report` endpoint
6. Backend calls ML API and stores prediction
7. Response includes `urgency` field with ML prediction
8. Frontend displays the urgency from the server response

---

## Monitoring

Look for these log messages in your backend terminal:

✅ **Success:**
```
ml: urgency prediction - label: critical -> urgency: 3
```

⚠️ **Fallback (ML disabled):**
```
[No log, returns 0]
```

❌ **Error (still continues):**
```
[Logged error but report still succeeds with original urgency]
```

