
package services

import (
	"crowdsourcedurbanissuereportingwithai/backend/internal/repository"
	"crowdsourcedurbanissuereportingwithai/backend/models"
)

type FeedService struct {
	PostRepo *repository.PostRepository
}

func NewFeedService(postRepo *repository.PostRepository) *FeedService {
	return &FeedService{PostRepo: postRepo}
}

func (s *FeedService) GetFeed() ([]models.Post, error) {
	return s.PostRepo.GetFeedPosts()
}
