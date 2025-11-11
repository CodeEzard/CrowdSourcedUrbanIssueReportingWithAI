#!/bin/bash

# Image Classification Integration Test Script

echo "==================================="
echo "Image Classification Integration Test"
echo "==================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if server is running
echo "Checking if backend server is running..."
if ! curl -s http://localhost:8080/feed > /dev/null 2>&1; then
    echo -e "${YELLOW}Server not responding. Start it with:${NC}"
    echo "cd backend"
    echo "DISABLE_AUTH=true ML_API_URL='https://urgency-api-latest.onrender.com/predict' IMAGE_CLASSIFICATION_API_URL='https://issue-classification-api.onrender.com/predict' go run ."
    exit 1
fi

echo -e "${GREEN}✓ Server is running${NC}"
echo ""

# Test 1: Pothole Report
echo "Test 1: Submitting Pothole Report..."
echo "---"

RESPONSE=$(curl -s -X POST "http://localhost:8080/report" \
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
  }')

echo "Response:"
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"

# Extract and check values
URGENCY=$(echo "$RESPONSE" | jq '.urgency' 2>/dev/null)
CLASSIFIED=$(echo "$RESPONSE" | jq '.classified_as' 2>/dev/null)

echo ""
echo "Verification:"
if [ "$URGENCY" == "3" ]; then
    echo -e "${GREEN}✓ Urgency prediction working (1 → 3)${NC}"
else
    echo -e "${YELLOW}⚠ Urgency: $URGENCY (expected 3)${NC}"
fi

if [ "$CLASSIFIED" == '"potholes"' ] || [ "$CLASSIFIED" == '"pothole"' ]; then
    echo -e "${GREEN}✓ Image classification working (potholes detected)${NC}"
else
    echo -e "${YELLOW}⚠ Classification: $CLASSIFIED${NC}"
fi

echo ""
echo "---"
echo ""

# Test 2: Broken Pole Report
echo "Test 2: Submitting Broken Pole Report..."
echo "---"

RESPONSE2=$(curl -s -X POST "http://localhost:8080/report" \
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
  }')

echo "Response:"
echo "$RESPONSE2" | jq '.' 2>/dev/null || echo "$RESPONSE2"

URGENCY2=$(echo "$RESPONSE2" | jq '.urgency' 2>/dev/null)
CLASSIFIED2=$(echo "$RESPONSE2" | jq '.classified_as' 2>/dev/null)

echo ""
echo "Verification:"
if [ "$URGENCY2" == "3" ]; then
    echo -e "${GREEN}✓ Urgency prediction working${NC}"
else
    echo -e "${YELLOW}⚠ Urgency: $URGENCY2${NC}"
fi

if [[ "$CLASSIFIED2" == *"pole"* ]] || [[ "$CLASSIFIED2" == *"utility"* ]]; then
    echo -e "${GREEN}✓ Image classification working${NC}"
else
    echo -e "${YELLOW}⚠ Classification: $CLASSIFIED2${NC}"
fi

echo ""
echo "---"
echo ""
echo -e "${GREEN}✓ Integration tests complete!${NC}"
echo ""
echo "Check backend logs for:"
echo "  - 'ml: urgency prediction' messages"
echo "  - 'image_classification:' messages"
