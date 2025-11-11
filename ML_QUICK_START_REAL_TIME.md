# ML Real-Time Integration - Quick Start

## What You Now Have

Your Urban Civic application automatically:

1. **Detects Issue Type from Image** 🖼️
   - User uploads image
   - ML analyzes it
   - Shows predicted category (e.g., "pothole")
   - Auto-fills the category dropdown

2. **Predicts Urgency from Description** 📝
   - User types description
   - ML analyzes sentiment/keywords
   - Shows urgency level: Low (1) | Medium (2) | Critical (3)
   - Color-coded: Blue | Orange | Red

---

## Enable the ML Features

### Option 1: Local Development

```bash
cd backend
export ML_API_URL="https://urgency-api-latest.onrender.com/predict"
export IMAGE_CLASSIFICATION_API_URL="https://issue-classification-api.onrender.com/predict"
go run .
```

### Option 2: Render.com Deployment

1. Dashboard → Your Service → Environment
2. Add two variables:
   ```
   ML_API_URL = https://urgency-api-latest.onrender.com/predict
   IMAGE_CLASSIFICATION_API_URL = https://issue-classification-api.onrender.com/predict
   ```
3. Click Save (auto-redeploys)

### Option 3: Without ML (Will Still Work)

If you don't set the variables, the app works fine without ML:
- User manually selects category
- User selects urgency level
- No auto-predictions shown

---

## Test It Out

### Test Image Classification

1. Go to your website → **Report** tab
2. Click **"Next Step"** past the description
3. Upload an image (e.g., pothole, broken pole, etc.)
4. Wait for upload to complete
5. Look for: **"✅ Image classified as: [type]"**
6. Category should auto-fill

**Example:**
```
Uploaded: potholes.jpg
  ↓
✅ Image classified as: potholes
✅ Category auto-filled: Road
```

### Test Urgency Prediction

1. Go to **Report** tab
2. In "Description" field, type something (min 10 characters)
3. Stop typing
4. Wait 0.5 seconds
5. Look for: **"✅ Predicted Urgency: Low/Medium/Critical (1-3)"**

**Example:**
```
Typed: "There is a dangerous pothole on Main Street"
  ↓
✅ Predicted Urgency: Critical (3/3)
(shown in RED)
```

---

## How It Works

### Image Upload Flow
```
User selects image
  ↓
Upload to Cloudinary (our image hosting)
  ↓
Send image URL to /classify-image endpoint
  ↓
ML API analyzes the image
  ↓
Returns predicted category
  ↓
Frontend auto-fills category dropdown
  ↓
Shows toast notification
```

### Description Input Flow
```
User types description (min 10 chars)
  ↓
Wait 500ms (debounce to avoid spam)
  ↓
Send text to /predict-urgency endpoint
  ↓
ML API analyzes urgency
  ↓
Returns urgency level (1, 2, or 3)
  ↓
Shows prediction with color coding
```

---

## Endpoints

### POST `/classify-image`
Classifies an image by URL

```bash
curl -X POST http://localhost:8080/classify-image \
  -H "Content-Type: application/json" \
  -d '{"image_url": "https://example.com/image.jpg"}'
```

Response:
```json
{
  "predicted_class": "potholes",
  "error": ""
}
```

### POST `/predict-urgency`
Predicts urgency from description text

```bash
curl -X POST http://localhost:8080/predict-urgency \
  -H "Content-Type: application/json" \
  -d '{"text": "There is a dangerous pothole"}'
```

Response:
```json
{
  "urgency": 3,
  "error": ""
}
```

---

## Auto-Fill Mapping

When image is classified, the category dropdown auto-fills if a match is found:

| ML Predicts | Auto-Fills To | Match Words |
|---|---|---|
| Pothole | Road | pothole, hole, pavement |
| Pole | Utilities | pole, utility, power, electric |
| Light | Streetlight | light, streetlight, lamp |
| Trash | Garbage | trash, garbage, waste, litter |
| Water | Waterlogging | water, flood, waterlog, wet |

---

## Troubleshooting

### Image Not Classifying?
- Check if `IMAGE_CLASSIFICATION_API_URL` is set
- Check if API service is running
- Try a clearer image
- Check backend logs: `ml: image_classification ...`

### Urgency Not Predicting?
- Check if `ML_API_URL` is set
- Check description is at least 10 characters
- Wait 0.5 seconds after typing stops
- Check backend logs: `ml: urgency prediction ...`

### Forms Still Work Without ML?
Yes! The form always works. ML is optional enhancement. You can:
- Always manually select category and urgency
- Submit report without any ML predictions
- ML predictions are just helpful suggestions

---

## Data Storage

When you submit a report:

### User Provides:
- Description
- Category (or auto-filled)
- Image URL
- Location
- Urgency (or auto-predicted)

### Backend Also Calls ML:
- `PredictUrgency(description)` → gets urgency from description
- `ClassifyImage(image_url)` → gets category from image

### Stored in Database:
- `urgency`: Final urgency (from ML or user input)
- `classified_as`: What ML thinks the issue is
- All user inputs preserved

### Frontend Shows:
- All user inputs (what they entered)
- ML predictions (what we predicted)
- Final stored values (what was saved)

---

## Common Questions

**Q: Will my reports still work if I don't set the ML URLs?**
A: Yes, 100%. The form works perfectly without ML. Just no auto-predictions.

**Q: Can I use different ML APIs?**
A: Yes! As long as they return JSON with fields like `label`, `urgency`, `predicted_class`, etc.

**Q: What happens if the API is slow?**
A: There's a timeout. If it takes too long, we skip the prediction and let user continue.

**Q: Is the ML prediction sent with the report?**
A: The backend's ML predictions are stored. Frontend predictions are just for UX feedback.

---

## Next Steps

### To Enable ML on Your Deployment:

1. **Identify your Render service URL**
   - Example: `https://civic-issue-app.onrender.com`

2. **Add Environment Variables**
   - Go to Render Dashboard
   - Find your service
   - Click "Environment"
   - Add two variables (copy-paste below):
   ```
   ML_API_URL
   https://urgency-api-latest.onrender.com/predict
   
   IMAGE_CLASSIFICATION_API_URL
   https://issue-classification-api.onrender.com/predict
   ```

3. **Save and Deploy**
   - Click "Save"
   - Service automatically redeploys (~2 minutes)

4. **Test**
   - Visit your website
   - Create a report
   - Test image classification and urgency prediction

---

## Architecture

```
Frontend (Browser)
  ↓
report.html
  ├─ Image Upload
  │  └─→ POST /classify-image
  │      └─→ ML API (image classification)
  │          └─→ Returns predicted class
  │              └─→ Auto-fill category
  │
  └─ Description Input
     └─→ POST /predict-urgency
         └─→ ML API (urgency prediction)
             └─→ Returns urgency level (1-3)
                 └─→ Show colored feedback

Backend (When Submitting)
  ↓
POST /report
  ├─ Calls PredictUrgency(desc)
  │  └─→ ML API
  │      └─→ Store as urgency
  │
  └─ Calls ClassifyImage(url)
     └─→ ML API
         └─→ Store as classified_as

Database
  ├─ urgency (from ML or user)
  ├─ classified_as (from ML)
  ├─ description (user input)
  ├─ media_url (Cloudinary image)
  └─ category (user selected or auto-filled)
```

---

## Performance Metrics

- **Image Classification:** ~2-5 seconds
- **Urgency Prediction:** ~1-3 seconds
- **Timeouts:** 10s (image), 5s (urgency)
- **Debounce:** 500ms (urgency only)
- **User Impact:** No blocking, no delays

---

**Status:** ✅ Ready to Deploy  
**Last Updated:** November 11, 2025
