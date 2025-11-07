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
		// ensure an ID is generated at the application level so tests using sqlite
		// (or DBs without uuid_generate_v4()) get a valid UUID primary key.
		newID := uuid.New()
		issue = models.Issue{
			ID:          newID,
			Name:        issueName,
			Description: issueDesc,
			Category:    issueCat,
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

// AddComment creates a comment on a post by a user
func (r *PostRepository) AddComment(userID, postID uuid.UUID, content string) (*models.Comment, error) {
	comment := models.Comment{
		PostID:  postID,
		UserID:  userID,
		Content: content,
	}
	if err := r.DB.Create(&comment).Error; err != nil {
		return nil, err
	}
	// preload user info
	if err := r.DB.Preload("User").First(&comment, "id = ?", comment.ID).Error; err != nil {
		// non-fatal: return created comment anyway
		return &comment, nil
	}
	return &comment, nil
}

// ToggleUpvote creates an upvote if missing, or removes an existing upvote.
// Returns true if upvote was created, false if removed.
func (r *PostRepository) ToggleUpvote(userID, postID uuid.UUID) (bool, error) {
	var existing models.Upvote
	err := r.DB.Where("post_id = ? AND user_id = ?", postID, userID).First(&existing).Error
	if err == nil {
		// found existing - remove it
		if err := r.DB.Delete(&existing).Error; err != nil {
			return false, err
		}
		return false, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}
	// create new upvote
	up := models.Upvote{
		PostID: postID,
		UserID: userID,
	}
	if err := r.DB.Create(&up).Error; err != nil {
		return false, err
	}
	return true, nil
}
