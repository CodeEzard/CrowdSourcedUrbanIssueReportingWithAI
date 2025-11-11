# Real-Time ML Integration - Complete Guide

## Overview

Your Urban Civic application now features **real-time ML-powered predictions** during the report creation process:

1. **Image Classification** - Auto-detects issue type when user uploads an image
2. **Urgency Prediction** - Auto-analyzes urgency level as user types description

Both predictions happen **instantly** on the frontend with **graceful fallback** if ML APIs are unavailable.

---

## Features

### 1️⃣ Image Classification (Real-Time)

**What happens:**
- User uploads image via Cloudinary
- Image classification ML API analyzes the image
- Predicted category is displayed
- Category dropdown auto-fills if a match is found

**User Experience:**
```
User uploads image (potholes.jpg)
  ↓
"🔍 Analyzing image..."
  ↓
"✅ Image classified as: potholes"
  ↓
Category auto-filled: "Road"
  ↓
Toast: "Category auto-filled: Road (for potholes)"
```

**Supported Auto-Mapping:**
| Predicted Class | Maps To Category | Icon |
|---|---|---|
| pothole, pothole* | Road | 🛣️ |
| pole, utility pole | Utilities | ⚡ |
| light, streetlight | Streetlight | 💡 |
| trash, garbage, waste | Garbage | 🗑️ |
| water, flood, waterlog | Waterlogging | 💧 |

### 2️⃣ Urgency Prediction (Real-Time)

**What happens:**
- User types description (min 10 characters)
- Urgency prediction ML API analyzes the text
- Predicted urgency level is displayed with color coding
- Updates as user continues typing

**User Experience:**
```
User types: "There is a dangerous pothole blocking the road"
  ↓
After 500ms debounce:
  ↓
"🤖 Analyzing urgency..."
  ↓
"✅ Predicted Urgency: Critical (3/3)"
(shown in red color)
  ↓
Backend will use urgency: 3 when posting
```

**Color Coding:**
| Urgency | Level | Color | Code |
|---|---|---|---|
| 1 | Low | 🔵 Blue | `#4a9eff` |
| 2 | Medium | 🟠 Orange | `#ffaa00` |
| 3 | Critical | 🔴 Red | `#ff4444` |

---

## Configuration

### Backend Environment Variables

To enable the ML endpoints, set these environment variables:

```bash
# Urgency Prediction API
export ML_API_URL="https://urgency-api-latest.onrender.com/predict"

# Image Classification API
export IMAGE_CLASSIFICATION_API_URL="https://issue-classification-api.onrender.com/predict"
```

### Deployment (Render.com)

1. Go to your Render service dashboard
2. Click "Environment" tab
3. Add these variables:
   - `ML_API_URL=https://urgency-api-latest.onrender.com/predict`
   - `IMAGE_CLASSIFICATION_API_URL=https://issue-classification-api.onrender.com/predict`
4. Click "Save" and service redeploys automatically

If variables are **not set**, the ML features gracefully disable with user-friendly messages.

---

## API Endpoints

### POST `/classify-image`

**Request:**
```json
{
  "image_url": "https://cdn.example.com/image.jpg"
}
```

**Response (Success):**
```json
{
  "predicted_class": "potholes",
  "error": ""
}
```

**Response (Error/Unavailable):**
```json
{
  "predicted_class": "",
  "error": "image_classification api returned non-2xx status"
}
```

**Usage from Frontend:**
```javascript
const response = await fetch('/classify-image', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ image_url: 'https://...' })
});

const result = await response.json();
console.log(result.predicted_class); // "potholes"
```

---

### POST `/predict-urgency`

**Request:**
```json
{
  "text": "There is a dangerous pothole on Main Street"
}
```

**Response (Success):**
```json
{
  "urgency": 3,
  "error": ""
}
```

**Response (Error/Unavailable):**
```json
{
  "urgency": 0,
  "error": "ml api returned non-2xx status"
}
```

**Usage from Frontend:**
```javascript
const response = await fetch('/predict-urgency', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ text: 'description' })
});

const result = await response.json();
console.log(result.urgency); // 1, 2, or 3
```

---

## How It Works

### Image Classification Flow

```
┌─ User uploads image ─────────────────┐
│ "potholes.jpg"                       │
└──────────────┬──────────────────────┘
               │
    ┌──────────↓──────────┐
    │ Cloudinary uploads  │
    │ Returns: secure_url │
    └──────────┬──────────┘
               │
    ┌──────────↓────────────────────────────┐
    │ POST /classify-image                   │
    │ Body: {image_url: "https://...jpg"}   │
    └──────────┬────────────────────────────┘
               │
    ┌──────────↓────────────────────────────┐
    │ ML API analyzes image                  │
    │ Returns: {predicted_class: "potholes"} │
    └──────────┬────────────────────────────┘
               │
    ┌──────────↓────────────────────────────┐
    │ Frontend displays:                      │
    │ "✅ Image classified as: potholes"     │
    │ Auto-fills: Category = "Road"           │
    └────────────────────────────────────────┘
```

### Urgency Prediction Flow

```
┌─ User types description ──────────────┐
│ "There is a dangerous pothole..."     │
└──────────────┬──────────────────────┘
               │
    ┌──────────↓──────────────────────┐
    │ After 500ms debounce             │
    │ (if text length > 10)             │
    └──────────┬──────────────────────┘
               │
    ┌──────────↓────────────────────────────┐
    │ POST /predict-urgency                  │
    │ Body: {text: "...dangerous pothole..."} │
    └──────────┬────────────────────────────┘
               │
    ┌──────────↓────────────────────────────┐
    │ ML API analyzes text                   │
    │ Returns: {label: "critical", conf: 0.99} │
    │ Maps to: urgency = 3                   │
    └──────────┬────────────────────────────┘
               │
    ┌──────────↓────────────────────────────┐
    │ Frontend displays:                     │
    │ "✅ Predicted Urgency: Critical (3/3)" │
    │ (in red color)                         │
    └────────────────────────────────────────┘
```

---

## Code Structure

### Backend

**File:** `backend/internal/handlers/ml_handlers.go`

```go
type MLHandler struct {}

// ServeClassifyImage handles POST /classify-image
func (h *MLHandler) ServeClassifyImage(w http.ResponseWriter, r *http.Request)

// ServePredictUrgency handles POST /predict-urgency  
func (h *MLHandler) ServePredictUrgency(w http.ResponseWriter, r *http.Request)
```

**File:** `backend/internal/services/ml.go`

```go
// Existing functions (already implemented)
func PredictUrgency(text string) (int, error)
func ClassifyImage(imageURL string) (string, error)
```

### Frontend

**File:** `frontend/report.html`

```javascript
// Image classification on file upload
qs('#r-photo').addEventListener('change', async e => {
  // ... upload to Cloudinary ...
  classifyImageWithML(imageUrl); // New function
});

async function classifyImageWithML(imageUrl) {
  // Calls /classify-image endpoint
  // Auto-fills category dropdown
}

// Urgency prediction as user types
qs('#r-desc').addEventListener('input', () => {
  // Debounced call to /predict-urgency
  // Shows urgency level with color coding
});
```

---

## Testing

### Test Image Classification

1. Go to **Report** page
2. Click **"Next Step"** to see the image upload
3. Select any image file
4. Wait for Cloudinary upload to complete
5. You should see:
   ```
   ✅ Image classified as: [category]
   Category auto-filled: [Category Name]
   ```

**Test Cases:**
- Upload pothole image → Should classify as "pothole" → Auto-fill "Road"
- Upload utility pole image → Should classify as "pole" → Auto-fill "Utilities"
- Upload streetlight image → Should classify as "light" → Auto-fill "Streetlight"

### Test Urgency Prediction

1. Go to **Report** page
2. In the description field, type a description (min 10 characters)
3. Stop typing for 500ms
4. You should see:
   ```
   ✅ Predicted Urgency: Low/Medium/Critical (1-3)
   ```

**Test Cases:**
```
Low urgency:       "There is a minor issue on the road"        → 1 (Blue)
Medium urgency:    "There is damage that needs fixing"         → 2 (Orange)
Critical urgency:  "There is a dangerous pothole blocking way" → 3 (Red)
```

---

## Error Handling

### Graceful Fallback

If ML APIs are unavailable or misconfigured:

**Image Classification:**
```
⚠️ Could not classify image, please select category manually
⚠️ Classification service unavailable, please select category manually
```

**Urgency Prediction:**
```
⚠️ Could not analyze urgency
⚠️ Urgency service unavailable
```

In both cases, the report form still works normally. The user can:
1. Manually select the category
2. Continue with report submission
3. Backend will use the submitted urgency value (or default 1 if not set)

### API Down Scenarios

| Scenario | Image Upload | Report Submit |
|---|---|---|
| Image API down | Shows warning, form continues | Works normally |
| Urgency API down | N/A | Works normally |
| Both down | Shows warnings, form continues | Works normally |
| Both up | Auto-classifies & predicts | Uses ML predictions |

---

## Performance

### Image Classification
- **Timeout:** 10 seconds (+ 2 second buffer = 12 second total)
- **Trigger:** After Cloudinary upload completes
- **Async:** Non-blocking, user can continue filling form

### Urgency Prediction
- **Debounce:** 500ms (waits for user to stop typing)
- **Timeout:** 5 seconds (+ 1 second buffer = 6 second total)
- **Async:** Non-blocking, user can continue typing/clicking

### Frontend Impact
- No UI blocking
- Instant visual feedback
- Graceful degradation if APIs slow/down
- User can always submit manually if needed

---

## Data Flow Through System

### Complete Report Submission with ML

```
User creates report:
┌─────────────────────────────────────┐
│ 1. Uploads image                     │
│    → Classified: "potholes"          │
│    → Category auto-filled: "Road"    │
│                                     │
│ 2. Types description                 │
│    → Analyzed: urgency = 3           │
│    → Shows: "Critical (3/3)"         │
│                                     │
│ 3. Submits report                    │
│    → POST /report with:              │
│       - image_url: from Cloudinary   │
│       - urgency: 1 (user default)    │
│       - description: typed text      │
│       - category: auto-filled or set │
└──────────────────┬──────────────────┘
                   │
           ┌───────↓────────┐
           │ Backend receives
           │ calls PredictUrgency(desc)
           │   ↓ returns 3
           │ calls ClassifyImage(url)
           │   ↓ returns "potholes"
           │ Stores in DB:
           │   urgency: 3 (from ML)
           │   classified_as: "potholes"
           └───────┬────────┘
                   │
        ┌──────────↓──────────┐
        │ Response to frontend │
        │ (urgency: 3,        │
        │  classified_as:     │
        │  "potholes")        │
        └─────────────────────┘
```

---

## Troubleshooting

### Image Not Classifying

**Symptoms:**
- Shows: "⚠️ Could not classify image"
- Category not auto-filled

**Causes:**
- Image API URL not configured (`IMAGE_CLASSIFICATION_API_URL`)
- API service is down
- Image URL is invalid
- Network connectivity issue

**Solutions:**
1. Check if `IMAGE_CLASSIFICATION_API_URL` is set in environment
2. Check if API service is running: `https://issue-classification-api.onrender.com/predict`
3. Try uploading a different image
4. Check backend logs for errors
5. Manually select category (form works without ML)

### Urgency Not Predicting

**Symptoms:**
- No urgency prediction shown while typing
- Shows: "⚠️ Could not analyze urgency"

**Causes:**
- ML API URL not configured (`ML_API_URL`)
- API service is down
- Description too short (min 10 chars required)
- Network connectivity issue

**Solutions:**
1. Check if `ML_API_URL` is set in environment
2. Check if API service is running: `https://urgency-api-latest.onrender.com/predict`
3. Make sure description is at least 10 characters
4. Check backend logs for errors
5. Report works fine with user-selected urgency level

### Backend Logs for Debugging

```bash
# Real-time backend logs
tail -f backend.log

# Look for these messages:
# Success:
ml: urgency prediction - label: critical -> urgency: 3
image_classification: predicted_class: potholes

# Errors:
ml api returned non-2xx status
image_classification: could not extract predicted class
```

---

## FAQ

**Q: Do I need to set the ML API URLs?**
A: No, they're optional. If not set, the form works normally without ML predictions.

**Q: What if the ML API is slow?**
A: There's a 5-10 second timeout. If it times out, graceful fallback is used (user continues normally).

**Q: Can I submit without ML predictions?**
A: Yes! ML predictions are optional. User can always manually select category/urgency.

**Q: Are ML predictions stored in the database?**
A: Yes! The backend also calls the ML APIs during report submission to get final predictions before storing.

**Q: Why does category sometimes not auto-fill?**
A: If the predicted class doesn't exactly match a category name, manual selection is required. Example: "sidewalk damage" won't match any category.

**Q: Can I customize the auto-fill mapping?**
A: Currently, it's hardcoded in report.html. You can modify the `classifyImageWithML()` function to add more mappings.

---

## Next Steps

### Potential Enhancements

1. **Save ML Predictions**
   - Store original user inputs AND ML predictions
   - Track prediction accuracy over time

2. **User Feedback Loop**
   - "Was the category correct?" buttons after auto-fill
   - Use feedback to improve ML model

3. **Advanced Analytics**
   - Dashboard showing accuracy of auto-predictions
   - Compare user input vs ML prediction

4. **Batch Processing**
   - Process multiple reports at once
   - Background job queue for heavy ML workloads

5. **Custom Models**
   - Train local models for your city
   - Fine-tune on local issue patterns

---

**Last Updated:** November 11, 2025  
**Status:** Production Ready ✅
