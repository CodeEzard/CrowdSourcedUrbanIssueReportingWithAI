package handlers

import (
	"crowdsourcedurbanissuereportingwithai/backend/internal/services"
	"encoding/json"
	"net/http"
)

type FeedHandler struct {
	FeedService *services.FeedService
}

func NewFeedHandler(feedService *services.FeedService) *FeedHandler {
	return &FeedHandler{FeedService: feedService}
}

func (h *FeedHandler) ServeFeed(w http.ResponseWriter, r *http.Request) {
	posts, err := h.FeedService.GetFeed()
	if err != nil {
		http.Error(w, "Failed to fetch feed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
