
package repository

import (
	"crowdsourcedurbanissuereportingwithai/backend/models"
	"github.com/google/uuid"
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

// ReportIssueViaPost: Create an issue if not exists, then create a post for it
func (r *PostRepository) ReportIssueViaPost(userID, issueName, issueDesc, issueCat, postDesc, status string, urgency int, lat, lng float64, mediaURL string) (*models.Post, error) {
	var issue models.Issue
	err := r.DB.Where("name = ?", issueName).First(&issue).Error
	if err == gorm.ErrRecordNotFound {
		issue = models.Issue{
			Name:        issueName,
			Description: issueDesc,
			Category:    "Miscellaneous",
		}
		if err := r.DB.Create(&issue).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	post := models.Post{
		IssueID:     issue.ID,
		UserID:      userUUID,
		Description: postDesc,
		Status:      status,
		Urgency:     urgency,
		Lat:         lat,
		Lng:         lng,
		MediaURL:    mediaURL,
	}
	if err := r.DB.Create(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}
