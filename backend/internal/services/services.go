package services

import (
	"crowdsourcedurbanissuereportingwithai/backend/internal/repository"
	"crowdsourcedurbanissuereportingwithai/backend/models"

	"github.com/google/uuid"
)

type FeedService struct {
	PostRepo *repository.PostRepository
}

func NewFeedService(postRepo *repository.PostRepository) *FeedService {
	return &FeedService{PostRepo: postRepo}
}

type ReportService struct {
	PostRepo *repository.PostRepository
}

func NewReportService(postRepo *repository.PostRepository) *ReportService {
	return &ReportService{PostRepo: postRepo}
}

func (s *ReportService) ReportIssueViaPost(userID, issueName, issueDesc, issueCat, postDesc, status string, urgency int, lat, lng float64, mediaURL string) (*models.Post, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	// If an ML API is configured, attempt to predict urgency from the post description.
	if pred, err := PredictUrgency(postDesc); err == nil && pred != 0 {
		urgency = pred
	} else if err != nil {
		// non-fatal: log and continue with provided urgency
		// PredictUrgency logs details; we don't block reporting if ML fails.
	}
	
	// If an image classification API is configured, attempt to classify the image.
	classifiedAs := ""
	if classified, err := ClassifyImage(mediaURL); err == nil && classified != "" {
		classifiedAs = classified
	} else if err != nil {
		// non-fatal: log and continue without classification
		// ClassifyImage logs details; we don't block reporting if classification fails.
	}
	
	return s.PostRepo.ReportIssueViaPost(uid.String(), issueName, issueDesc, issueCat, postDesc, status, urgency, lat, lng, mediaURL, classifiedAs)
}
func (s *FeedService) GetFeed() ([]models.Post, error) {
	return s.PostRepo.GetFeedPosts()
}

// AddComment wraps repository call to add a comment
func (s *ReportService) AddComment(userID, postID, content string) (*models.Comment, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	pid, err := uuid.Parse(postID)
	if err != nil {
		return nil, err
	}
	return s.PostRepo.AddComment(uid, pid, content)
}

// ToggleUpvote wraps repository call to toggle an upvote
func (s *ReportService) ToggleUpvote(userID, postID string) (bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}
	pid, err := uuid.Parse(postID)
	if err != nil {
		return false, err
	}
	return s.PostRepo.ToggleUpvote(uid, pid)
}
