package repository

import (
	"errors"
	config "crowdsourcedurbanissuereportingwithai/backend/configs"
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

// Feed: Get posts ordered by timeline (newest first), including comments and upvotes
func (r *PostRepository) GetFeedPosts() ([]models.Post, error) {
	var posts []models.Post
	// Limit feed size for performance if configured
	limit := config.GetFeedLimit()
	if limit <= 0 { limit = 50 }
	err := r.DB.
		Preload("User").
		Preload("Issue").
		Preload("Comments", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User").Order("created_at DESC")
		}).
		Preload("Upvotes").
		Order("created_at DESC").
		Limit(limit).
		Find(&posts).Error
	return posts, err
}

// ReportIssueViaPost: Create an issue if not exists, then create a post for it
func (r *PostRepository) ReportIssueViaPost(userID, issueName, issueDesc, issueCat, postDesc, status string, urgency int, lat, lng float64, mediaURL string, classifiedAs string) (*models.Post, error) {
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
		IssueID:      issue.ID,
		UserID:       userUUID,
		Description:  postDesc,
		Status:       status,
		Urgency:      urgency,
		ClassifiedAs: classifiedAs,
		Lat:          lat,
		Lng:          lng,
		MediaURL:     mediaURL,
	}
	if err := r.DB.Create(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

// UpdatePostScoreAdd increments (or decrements) score_sum and score_count atomically for a post.
// deltaCount can be negative (e.g., when removing an upvote) but will not reduce below 1 if description initialized.
func (r *PostRepository) UpdatePostScoreAdd(postID uuid.UUID, deltaScore float64, deltaCount int) error {
	var p models.Post
	if err := r.DB.First(&p, "id = ?", postID).Error; err != nil {
		return err
	}
	// Initialize baseline if not set
	sum := p.ScoreSum + deltaScore
	cnt := p.ScoreCount + deltaCount
	if cnt < 1 {
		cnt = 1
	}
	if sum < 0 {
		sum = 0
	}
	return r.DB.Model(&models.Post{}).Where("id = ?", postID).Updates(map[string]interface{}{
		"score_sum":   sum,
		"score_count": cnt,
	}).Error
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
	switch {
	case err == nil:
		// found existing - remove it
		if err := r.DB.Delete(&existing).Error; err != nil {
			return false, err
		}
		return false, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		// continue to create new upvote
	default:
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

// GetPostComments fetches all comments for a post
func (r *PostRepository) GetPostComments(postID uuid.UUID) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.DB.Where("post_id = ?", postID).Preload("User").Order("created_at ASC").Find(&comments).Error
	return comments, err
}

// UpdatePostUrgency updates the urgency field of a post
func (r *PostRepository) UpdatePostUrgency(postID uuid.UUID, newUrgency int) error {
	return r.DB.Model(&models.Post{}).Where("id = ?", postID).Update("urgency", newUrgency).Error
}

// UpdatePostStatus updates the status of a post (e.g., "open" -> "closed")
func (r *PostRepository) UpdatePostStatus(postID uuid.UUID, status string) (*models.Post, error) {
	post := &models.Post{}
	if err := r.DB.Model(post).Where("id = ?", postID).Update("status", status).Error; err != nil {
		return nil, err
	}
	// Reload the post with preloads
	if err := r.DB.Preload("User").Preload("Issue").Preload("Comments").Preload("Upvotes").First(post, "id = ?", postID).Error; err != nil {
		return nil, err
	}
	return post, nil
}

// GetPost fetches a single post by ID
func (r *PostRepository) GetPost(postID uuid.UUID) (*models.Post, error) {
	var post models.Post
	err := r.DB.Preload("User").Preload("Issue").Preload("Comments").First(&post, "id = ?", postID).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}
