package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	config "crowdsourcedurbanissuereportingwithai/backend/configs"
)

// PredictUrgency calls the configured ML API with the provided text and
// attempts to extract an integer urgency score. Returns (0, nil) if no
// ML API is configured. If the call fails, returns an error.
// PredictUrgency retains backward-compatibility (integer only) by delegating
// to PredictUrgencyDetailed and discarding the score.
func PredictUrgency(text string) (int, error) {
	urg, _, err := PredictUrgencyDetailed(text)
	return urg, err
}

// PredictUrgencyDetailed returns both a discrete urgency bucket (1-3) and a
// continuous score in [0,1]. Threshold mapping (>=0.8->3, >=0.5->2, else 1).
// If ML API disabled it attempts heuristic scoring from text content.
func PredictUrgencyDetailed(text string) (int, float64, error) {
	mlURL := config.GetMLAPIURL()
	if mlURL == "" {
		// Heuristic fallback when ML disabled
		score := heuristicScore(text)
		return mapScoreToUrgency(score), score, nil
	}

	reqBody := map[string]string{"text": text}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return 0, 0, err
	}

	// Use configured timeout for text ML
	ctx, cancel := context.WithTimeout(context.Background(), config.GetMLTextTimeout())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", mlURL, bytes.NewReader(b))
	if err != nil {
		// Fallback to heuristic if request cannot be created
		score := heuristicScore(text)
		return mapScoreToUrgency(score), score, nil
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: config.GetMLTextTimeout() + (2 * time.Second)}
	resp, err := client.Do(req)
	if err != nil {
		// Network error -> fallback to heuristic
		score := heuristicScore(text)
		return mapScoreToUrgency(score), score, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Non-2xx -> fallback to heuristic
		score := heuristicScore(text)
		return mapScoreToUrgency(score), score, nil
	}

	var parsed map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		// Malformed response -> fallback to heuristic
		score := heuristicScore(text)
		return mapScoreToUrgency(score), score, nil
	}

	// 1. Direct score field
	if scoreV, ok := parsed["score"]; ok {
		if sc, ok := scoreV.(float64); ok {
			urg := mapScoreToUrgency(sc)
			log.Printf("ml: urgency prediction - score: %.2f -> urgency: %d\n", sc, urg)
			return urg, sc, nil
		}
	}

	// 2. Label mapping
	if v, ok := parsed["label"]; ok {
		if s, ok := v.(string); ok {
			label := strings.ToLower(strings.TrimSpace(s))
			switch label {
			case "critical", "urgent":
				return 3, 0.9, nil
			case "moderate", "medium":
				return 2, 0.65, nil
			case "low", "minor":
				return 1, 0.3, nil
			}
		}
	}

	// 3. Numeric urgency -> mapped score
	if v, ok := parsed["urgency"]; ok {
		switch t := v.(type) {
		case float64:
			urg := int(t)
			sc := mapNumericUrgencyToScore(urg)
			return urg, sc, nil
		case int:
			sc := mapNumericUrgencyToScore(t)
			return t, sc, nil
		}
	}

	// 4. Classification / confidence alternative fields
	if confV, ok := parsed["confidence"]; ok {
		if cf, ok := confV.(float64); ok {
			urg := mapScoreToUrgency(cf)
			return urg, cf, nil
		}
	}

	log.Println("ml: could not extract urgency from response", parsed)
	// If ML can't provide, fallback to heuristic instead of returning zero
	score := heuristicScore(text)
	return mapScoreToUrgency(score), score, nil
}

// helper: map score to urgency bucket
func mapScoreToUrgency(score float64) int {
	if score >= 0.8 {
		return 3
	} else if score >= 0.5 {
		return 2
	}
	return 1
}

// helper: map integer urgency to representative score
func mapNumericUrgencyToScore(urg int) float64 {
	switch urg {
	case 3:
		return 0.85
	case 2:
		return 0.6
	case 1:
		return 0.3
	default:
		return 0.0
	}
}

// heuristicScore provides a lightweight local estimation when ML API disabled.
// Very naive keyword scoring; can be improved later.
func heuristicScore(text string) float64 {
	lower := strings.ToLower(text)
	criticalTerms := []string{"emergency", "danger", "fire", "explosion", "injury", "critical", "urgent"}
	moderateTerms := []string{"broken", "delay", "blocked", "leak", "issue", "problem", "trash"}
	for _, w := range criticalTerms {
		if strings.Contains(lower, w) {
			return 0.85
		}
	}
	for _, w := range moderateTerms {
		if strings.Contains(lower, w) {
			return 0.6
		}
	}
	if strings.TrimSpace(lower) == "" {
		return 0.0
	}
	return 0.3
}

// ClassifyImage calls the configured image classification API with the provided image URL
// and attempts to extract the predicted class. Returns empty string if no API is configured.
// If the call fails, returns an error but allows graceful fallback.
func ClassifyImage(imageURL string) (string, error) {
	apiURL := config.GetImageClassificationAPIURL()
	if apiURL == "" {
		return "", nil // feature disabled
	}

	if imageURL == "" {
		return "", nil // no image to classify
	}

	// Create multipart form with image_url field
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add image_url as form field
	if err := writer.WriteField("image_url", imageURL); err != nil {
		return "", err
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	// Use configured timeout for image classification (larger default)
	ctx, cancel := context.WithTimeout(context.Background(), config.GetMLImageTimeout())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: config.GetMLImageTimeout() + (2 * time.Second)}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", errors.New("image classification api returned non-2xx status: " + string(bodyBytes))
	}

	var parsed map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}

	// Handle "predicted_class" field (primary response format from the model)
	if v, ok := parsed["predicted_class"]; ok {
		if s, ok := v.(string); ok {
			classified := strings.TrimSpace(s)
			log.Printf("image_classification: predicted_class: %s\n", classified)
			return classified, nil
		}
	}

	// Fallback: try "classification" field
	if v, ok := parsed["classification"]; ok {
		if s, ok := v.(string); ok {
			classified := strings.TrimSpace(s)
			log.Printf("image_classification: classification: %s\n", classified)
			return classified, nil
		}
	}

	// Fallback: try "class" field
	if v, ok := parsed["class"]; ok {
		if s, ok := v.(string); ok {
			classified := strings.TrimSpace(s)
			log.Printf("image_classification: class: %s\n", classified)
			return classified, nil
		}
	}

	// No recognized field found
	log.Println("image_classification: could not extract predicted class from response", parsed)
	return "", nil
}
