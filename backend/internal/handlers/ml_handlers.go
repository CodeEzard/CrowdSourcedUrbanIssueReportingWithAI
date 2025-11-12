package handlers

import (
	"crowdsourcedurbanissuereportingwithai/backend/internal/services"
	"encoding/json"
	"net/http"
)

type MLHandler struct {
	// Empty for now - we'll call service functions directly
}

func NewMLHandler() *MLHandler {
	return &MLHandler{}
}

// ClassifyImageRequest contains the image URL to classify
type ClassifyImageRequest struct {
	ImageURL string `json:"image_url"`
}

// ClassifyImageResponse contains the predicted class
type ClassifyImageResponse struct {
	PredictedClass string `json:"predicted_class"`
	Error          string `json:"error,omitempty"`
}

// ServeClassifyImage handles image classification requests from the frontend
// This allows real-time image classification as the user uploads
func (h *MLHandler) ServeClassifyImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ClassifyImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.ImageURL == "" {
		http.Error(w, "Image URL is required", http.StatusBadRequest)
		return
	}

	// Call the ML service to classify the image
	classified, err := services.ClassifyImage(req.ImageURL)
	if err != nil {
		// Return 200 with error message so frontend can degrade gracefully without console 500s
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ClassifyImageResponse{
			PredictedClass: "",
			Error:          err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ClassifyImageResponse{
		PredictedClass: classified,
		Error:          "",
	})
}

// PredictUrgencyRequest contains the text to analyze
type PredictUrgencyRequest struct {
	Text string `json:"text"`
}

// PredictUrgencyResponse contains the predicted urgency
type PredictUrgencyResponse struct {
	Urgency int    `json:"urgency"`
	Error   string `json:"error,omitempty"`
}

// ServePredictUrgency handles urgency prediction requests from the frontend
// This allows real-time urgency prediction as the user types
func (h *MLHandler) ServePredictUrgency(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PredictUrgencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

    // Call the ML service to predict urgency. Service now internally falls back to heuristic.
    urgency, err := services.PredictUrgency(req.Text)
    if err != nil {
        // Graceful heuristic fallback if error surfaces (e.g. legacy code path)
        scoreUrg, _, _ := services.PredictUrgencyDetailed(req.Text)
        urgency = scoreUrg
    }

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(PredictUrgencyResponse{
		Urgency: urgency,
		Error:   "", // Always blank; fallback prevents 500
	})
}
