package handlers

import (
	"crowdsourcedurbanissuereportingwithai/backend/internal/services"
	"encoding/json"
	"net/http"
 "github.com/google/uuid"
)
type ReportHandler struct {
	ReportService *services.ReportService
}

func NewReportHandler(reportService *services.ReportService) *ReportHandler {
	return &ReportHandler{ReportService: reportService}
}

type ReportRequest struct {
	UserID      string  `json:"user_id"`
	IssueName   string  `json:"issue_name"`
	IssueDesc   string  `json:"issue_desc"`
	IssueCat    string  `json:"issue_cat"`
	PostDesc    string  `json:"post_desc"`
	Status      string  `json:"status"`
	Urgency     int     `json:"urgency"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	MediaURL    string  `json:"media_url"`
}

func (h *ReportHandler) ServeReport(w http.ResponseWriter, r *http.Request) {
	var req ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	post, err := h.ReportService.ReportIssueViaPost(uid.String(), req.IssueName, req.IssueDesc, req.IssueCat, req.PostDesc, req.Status, req.Urgency, req.Lat, req.Lng, req.MediaURL)
	if err != nil {
		http.Error(w, "Failed to report issue", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

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
