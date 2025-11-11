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
func PredictUrgency(text string) (int, error) {
	mlURL := config.GetMLAPIURL()
	if mlURL == "" {
		return 0, nil
	}

	// request body
	reqBody := map[string]string{"text": text}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", mlURL, bytes.NewReader(b))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, errors.New("ml api returned non-2xx status")
	}

	var parsed map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return 0, err
	}

	// Handle "label" field (primary response format from the model)
	if v, ok := parsed["label"]; ok {
		if s, ok := v.(string); ok {
			switch s {
			case "critical", "urgent":
				log.Printf("ml: urgency prediction - label: %s -> urgency: 3\n", s)
				return 3, nil
			case "moderate", "medium":
				log.Printf("ml: urgency prediction - label: %s -> urgency: 2\n", s)
				return 2, nil
			case "low", "minor":
				log.Printf("ml: urgency prediction - label: %s -> urgency: 1\n", s)
				return 1, nil
			}
		}
	}

	// Fallback: try numeric urgency field
	if v, ok := parsed["urgency"]; ok {
		switch t := v.(type) {
		case float64:
			urgency := int(t)
			log.Printf("ml: urgency prediction - numeric urgency: %d\n", urgency)
			return urgency, nil
		case int:
			log.Printf("ml: urgency prediction - numeric urgency: %d\n", t)
			return t, nil
		}
	}

	// Fallback: try score/confidence field
	if scoreV, ok := parsed["score"]; ok {
		if sc, ok := scoreV.(float64); ok {
			if sc >= 0.8 {
				log.Printf("ml: urgency prediction - score: %.2f -> urgency: 3\n", sc)
				return 3, nil
			} else if sc >= 0.5 {
				log.Printf("ml: urgency prediction - score: %.2f -> urgency: 2\n", sc)
				return 2, nil
			}
			log.Printf("ml: urgency prediction - score: %.2f -> urgency: 1\n", sc)
			return 1, nil
		}
	}

	// No recognized field found
	log.Println("ml: could not extract urgency from response", parsed)
	return 0, nil
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 12 * time.Second}
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
