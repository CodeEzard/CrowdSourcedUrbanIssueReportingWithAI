# Image Classification Quick Reference

## 🚀 Start Backend (With Both ML APIs)

```bash
cd backend
DISABLE_AUTH=true ML_API_URL="https://urgency-api-latest.onrender.com/predict" IMAGE_CLASSIFICATION_API_URL="https://issue-classification-api.onrender.com/predict" go run .
```

## 📤 Test with Pothole Image

```bash
curl -X POST "http://localhost:8080/report" \
  -H "Content-Type: application/json" \
  -d '{
    "issue_name": "Pothole",
    "issue_desc": "Road damage",
    "issue_category": "Road",
    "post_desc": "There is a dangerous pothole on the road",
    "status": "open",
    "urgency": 1,
    "lat": 40.7128,
    "lng": -74.0060,
    "media_url": "https://anonomz.com/wp-content/uploads/2014/04/potholes.jpg"
  }'
```

## 📊 Expected Response

```json
{
  "urgency": 3,
  "classified_as": "potholes"
}
```

## 🔍 What to Look For in Backend Logs

```
ml: urgency prediction - label: critical -> urgency: 3
image_classification: predicted_class: potholes
```

## 🛠️ Configuration

| Env Var | Value |
|---------|-------|
| `ML_API_URL` | `https://urgency-api-latest.onrender.com/predict` |
| `IMAGE_CLASSIFICATION_API_URL` | `https://issue-classification-api.onrender.com/predict` |
| `DISABLE_AUTH` | `true` |

## 📁 Key Files

| File | Purpose |
|------|---------|
| `backend/internal/services/ml.go` | `ClassifyImage()` function |
| `backend/internal/services/services.go` | Integration point |
| `backend/internal/repository/repository.go` | Database storage |
| `backend/models/models.go` | `ClassifiedAs` field |
| `backend/configs/config.go` | Config reader |

## ✅ Verification Steps

1. ✅ Backend builds: `go build ./backend`
2. ✅ Start server with env vars set
3. ✅ Submit report with image URL
4. ✅ Check response has both `urgency` and `classified_as`
5. ✅ Check backend logs for both ML messages

## 🎯 Data Flow

```
Request: media_url + description
  ↓
PredictUrgency(description) → urgency: 1-3
ClassifyImage(media_url) → classified_as: "potholes"
  ↓
Response: {urgency, classified_as}
  ↓
Stored in database
```

## 📚 Full Documentation

See:
- `IMAGE_CLASSIFICATION_GUIDE.md` — Complete guide
- `IMAGE_CLASSIFICATION_COMPLETE.md` — Technical summary
- `ML_INTEGRATION_COMPLETE.md` — Urgency prediction

## 🧪 Test Image URLs

**Potholes:**
```
https://anonomz.com/wp-content/uploads/2014/04/potholes.jpg
```

**Utility Poles:**
```
https://en.wikipedia.org/wiki/File:Utility_pole_lean.jpg
```

