# ML Integration Quick Reference

## 🚀 Start Backend (Fastest Way to Test)

```bash
cd backend
DISABLE_AUTH=true ML_API_URL="https://urgency-api-latest.onrender.com/predict" go run .
```

## 📤 Submit Test Report

```bash
curl -X POST "http://localhost:8080/report" \
  -H "Content-Type: application/json" \
  -d '{
    "issue_name": "Broken Pole",
    "issue_desc": "Dangerous pole",
    "issue_category": "Utilities",
    "post_desc": "There is a dangerous broken pole near the road",
    "status": "open",
    "urgency": 1,
    "lat": 40.7128,
    "lng": -74.0060,
    "media_url": ""
  }'
```

## 📊 Response Mapping

| Input Urgency | ML Label | Output Urgency |
|---|---|---|
| 1 | critical | **3** ✅ |
| 1 | moderate | **2** ✅ |
| 1 | low | **1** ✅ |

## 🔍 What to Look For

**Backend Terminal Output:**
```
ml: urgency prediction - label: critical -> urgency: 3
```

**Response JSON:**
```json
{
  "urgency": 3
}
```

## 🛠️ Implementation Details

- **Request:** `POST https://urgency-api-latest.onrender.com/predict`
- **Body:** `{"text": "issue description"}`
- **Response:** `{"label": "critical", "confidence": 0.99}`
- **Timeout:** 5 seconds
- **Fallback:** Uses submitted urgency if ML fails

## 📁 Key Files

| File | Purpose |
|------|---------|
| `backend/internal/services/ml.go` | ML HTTP client |
| `backend/internal/services/services.go` | Service layer integration |
| `backend/configs/config.go` | Configuration reader |

## ✅ Verification

```bash
# Check build
go build ./backend

# Test with curl (see above)

# Check logs for "ml: urgency prediction"
```

## 🚨 Troubleshooting

| Issue | Solution |
|-------|----------|
| 401 errors | Add `DISABLE_AUTH=true` |
| ML not called | Check `ML_API_URL` is set |
| Timeout errors | ML API may be slow; check network |
| Report fails | Check backend logs |

## 📚 Full Documentation

- `ML_INTEGRATION_COMPLETE.md` — Full summary
- `ML_INTEGRATION_TEST.md` — Comprehensive testing guide
- `ML_INTEGRATION_EXAMPLES.md` — Practical examples
