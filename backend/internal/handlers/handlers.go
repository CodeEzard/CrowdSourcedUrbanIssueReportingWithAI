package handlers

import (
	"crowdsourcedurbanissuereportingwithai/backend/internal/auth"
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

// DevTestUserID is an optional UUID used in development when auth is disabled.
// When non-nil, ServeReport will use this user ID as the reporter when no
// authenticated user is present.
var DevTestUserID uuid.UUID

type ReportRequest struct {
	UserID    string  `json:"user_id"`
	IssueName string  `json:"issue_name"`
	IssueDesc string  `json:"issue_desc"`
	IssueCat  string  `json:"issue_cat"`
	PostDesc  string  `json:"post_desc"`
	Status    string  `json:"status"`
	Urgency   int     `json:"urgency"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	MediaURL  string  `json:"media_url"`
}

func (h *ReportHandler) ServeReport(w http.ResponseWriter, r *http.Request) {
	var req ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	// Prefer authenticated user from context (set by AuthMiddleware).
	if uidCtx, ok := auth.GetUserIDFromContext(r.Context()); ok {
		req.UserID = uidCtx.String()
	}
	// If no authenticated user but a dev fallback is configured use it.
	if req.UserID == "" && DevTestUserID != uuid.Nil {
		req.UserID = DevTestUserID.String()
	}
	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "Invalid or missing user ID", http.StatusBadRequest)
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

// CommentRequest is the payload to add a comment to a post
type CommentRequest struct {
	PostID  string `json:"post_id"`
	Content string `json:"content"`
}

func (h *ReportHandler) ServeComment(w http.ResponseWriter, r *http.Request) {
	var req CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	// user must be authenticated
	uidCtx, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	comment, err := h.ReportService.AddComment(uidCtx.String(), req.PostID, req.Content)
	if err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

// UpvoteRequest toggles an upvote for a post
type UpvoteRequest struct {
	PostID string `json:"post_id"`
}

func (h *ReportHandler) ServeUpvote(w http.ResponseWriter, r *http.Request) {
	var req UpvoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	uidCtx, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	created, err := h.ReportService.ToggleUpvote(uidCtx.String(), req.PostID)
	if err != nil {
		http.Error(w, "Failed to toggle upvote", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"upvoted": created})
}

// Admin: Update post status (open -> closed, etc)
type StatusUpdateRequest struct {
	PostID  string `json:"post_id"`
	Status  string `json:"status"` // "open", "inprogress", "closed"
	Notes   string `json:"notes"`  // optional notes from admin
}

func (h *ReportHandler) ServeUpdateStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var req StatusUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	if req.PostID == "" || req.Status == "" {
		http.Error(w, "Missing post_id or status", http.StatusBadRequest)
		return
	}
	
	// Verify user is authenticated (middleware ensures this for protected routes)
	_, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	post, err := h.ReportService.UpdatePostStatus(req.PostID, req.Status, req.Notes)
	if err != nil {
		http.Error(w, "Failed to update status: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(post)
}

// Admin: Get all issues for admin dashboard
func (h *FeedHandler) ServeAdminFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	
	// Get filter params from query (optional)
	status := r.URL.Query().Get("status") // "open", "inprogress", "closed"
	limit := 100
	
	posts, err := h.FeedService.GetAllPostsForAdmin(status, limit)
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(posts)
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
	// Ensure clients do not cache the feed; we want fresh data on every refresh
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	json.NewEncoder(w).Encode(posts)
}
