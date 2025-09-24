
package repository

import (
	"crowdsourcedurbanissuereportingwithai/backend/models"
	"gorm.io/gorm"
)

type PostRepository struct {
	DB *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{DB: db}
}

// Feed: Get posts ordered by timeline (newest first)
func (r *PostRepository) GetFeedPosts() ([]models.Post, error) {
	var posts []models.Post
	err := r.DB.Preload("User").Preload("Issue").Order("created_at DESC").Find(&posts).Error
	return posts, err
}
