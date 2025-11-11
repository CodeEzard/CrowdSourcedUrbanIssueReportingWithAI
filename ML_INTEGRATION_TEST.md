# ML Integration Testing Guide

## Overview
Your backend now integrates with the external ML urgency prediction API at `https://urgency-api-latest.onrender.com/predict`. When a user reports an issue, the report description is sent to the ML API, and the returned urgency level is stored with the issue.

## ML API Response Format
The API returns a JSON response with:
- `label`: String value ("critical", "moderate", "low", etc.)
- `confidence`: Float value (0.0 to 1.0)

**Example:**
```json
{
  "label": "critical",
  "confidence": 0.9948945045471191
}
```

## Urgency Mapping
The ML response is mapped to integer urgency levels:
- **"critical"** or **"urgent"** → **3** (High urgency)
- **"moderate"** or **"medium"** → **2** (Medium urgency)
- **"low"** or **"minor"** → **1** (Low urgency)

## Testing the Integration

### Option 1: DISABLE_AUTH Mode (Fastest for Local Testing)
This allows you to test without login credentials.

**1. Start the backend with DISABLE_AUTH and ML_API_URL:**
```bash
cd backend
DISABLE_AUTH=true ML_API_URL="https://urgency-api-latest.onrender.com/predict" go run .
```

**2. Test issue reporting with curl:**
```bash
curl -X POST "http://localhost:8080/report" \
  -H "Content-Type: application/json" \
  -d '{
    "issue_name": "Broken Pole",
    "issue_desc": "Dangerous pole in the street",
    "issue_category": "Utilities",
    "post_desc": "There is a dangerous broken pole near the road",
    "status": "open",
    "urgency": 1,
    "lat": 40.7128,
    "lng": -74.0060,
    "media_url": ""
  }'
```

**Expected Response:**
```json
{
  "id": "uuid-here",
  "issue": {
    "id": "uuid-here",
    "name": "Broken Pole",
    "description": "Dangerous pole in the street",
    "category": "Utilities"
  },
  "user": {
    "id": "uuid-here",
    "name": "Test User",
    "email": "test@example.com"
  },
  "description": "There is a dangerous broken pole near the road",
  "status": "open",
  "urgency": 3,
  "lat": 40.7128,
  "lng": -74.0060,
  "media_url": "",
  "created_at": "2025-11-11T...",
  "updated_at": "2025-11-11T..."
}
```

**Notice:** The `urgency` field in the response is **3** (mapped from "critical"), even though you sent **1** in the request. This shows the ML prediction successfully overrode the provided value.

### Option 2: With Authentication
If you want to test with proper authentication:

**1. Start the backend normally:**
```bash
cd backend
ML_API_URL="https://urgency-api-latest.onrender.com/predict" go run .
```

**2. Register a user:**
```bash
curl -X POST "http://localhost:8080/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "testuser@example.com",
    "password": "securepassword123"
  }'
```

**3. Login to get JWT token:**
```bash
curl -X POST "http://localhost:8080/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "securepassword123"
  }'
```

This returns:
```json
{
  "message": "Login successful",
  "token": "eyJhbGc..."
}
```

**4. Submit a report with the JWT token:**
```bash
curl -X POST "http://localhost:8080/report" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN_HERE>" \
  -d '{
    "issue_name": "Pothole on Main Street",
    "issue_desc": "Large pothole affecting traffic",
    "issue_category": "Road",
    "post_desc": "Dangerous pothole on Main Street near downtown",
    "status": "open",
    "urgency": 1,
    "lat": 40.7128,
    "lng": -74.0060,
    "media_url": ""
  }'
```

## Verification Checklist

- [ ] Backend starts without errors
- [ ] ML_API_URL environment variable is set
- [ ] Report submission endpoint responds with 200 OK
- [ ] Response contains `urgency` field with integer value (1-3)
- [ ] Check backend logs for "ml: urgency prediction" messages
- [ ] Urgency matches expected mapping (e.g., "critical" → 3)
- [ ] Frontend displays the urgency from the server response

## Backend Logs
When ML prediction is successful, you'll see logs like:
```
ml: urgency prediction - label: critical -> urgency: 3
```

## Troubleshooting

### ML API timeout
If the external API is slow or unavailable, the backend will:
- Log the error
- **Continue with the original urgency value** provided in the request
- **Not block** the issue creation

### Empty ML_API_URL
If you don't set `ML_API_URL`, the feature is disabled:
- `PredictUrgency()` returns immediately with urgency 0
- The original urgency from the request is used
- This is safe for production if you want to disable ML temporarily

### Check logs
Look for "ml:" prefixed log messages to debug ML integration issues.

## Code References
- ML prediction logic: `backend/internal/services/ml.go`
- Service integration: `backend/internal/services/services.go` (ReportIssueViaPost method)
- Config: `backend/configs/config.go` (GetMLAPIURL function)
