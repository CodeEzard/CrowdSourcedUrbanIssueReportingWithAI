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
	return s.PostRepo.ReportIssueViaPost(uid.String(), issueName, issueDesc, issueCat, postDesc, status, urgency, lat, lng, mediaURL)
}
func (s *FeedService) GetFeed() ([]models.Post, error) {
	return s.PostRepo.GetFeedPosts()
}
