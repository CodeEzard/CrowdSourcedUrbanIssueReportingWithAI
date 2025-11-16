package models

import (
	"time"

	"github.com/google/uuid"
)

// Issue category constants
const (
	CategoryLighting   = "Lighting"
	CategoryRoad       = "Road"
	CategorySanitation = "Sanitation"
	CategoryUtilities  = "Utilities"
	CategoryVandalism  = "Vandalism"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Email        string    `gorm:"unique;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Issue struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name        string    `gorm:"unique;not null" json:"name"`
	Description string    `json:"description,omitempty"`
	Category    string    `gorm:"not null" json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Post struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	IssueID      uuid.UUID `gorm:"type:uuid;not null;index:idx_post_issue" json:"issue_id"`
	Issue        Issue     `gorm:"foreignKey:IssueID" json:"issue"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_post_user" json:"user_id"`
	User         User      `gorm:"foreignKey:UserID" json:"user"`
	Description  string    `json:"description,omitempty"`
	Status       string    `gorm:"default:'open';not null;index:idx_post_status" json:"status"`
	Urgency      int       `gorm:"not null;index:idx_post_urgency" json:"urgency"`
	ClassifiedAs string    `json:"classified_as,omitempty"`
	Lat          float64   `gorm:"not null" json:"lat"`
	Lng          float64   `gorm:"not null" json:"lng"`
	MediaURL     string    `gorm:"not null" json:"media_url"`
	Comments     []Comment `gorm:"foreignKey:PostID" json:"comments"`
	Upvotes      []Upvote  `gorm:"foreignKey:PostID" json:"upvotes"`
	// Persistent incremental scoring fields
	ScoreSum     float64   `gorm:"default:0" json:"-"`
	ScoreCount   int       `gorm:"default:0" json:"-"`
	// Transient, computed at request time for ranking the feed
	Score            float64 `gorm:"-" json:"score,omitempty"`
	ComputedUrgency  int     `gorm:"-" json:"computed_urgency,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Comment struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	PostID    uuid.UUID `gorm:"type:uuid;not null" json:"post_id"`
	Post      Post      `gorm:"foreignKey:PostID" json:"post"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	Content   string    `gorm:"not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Upvote struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	PostID    uuid.UUID `gorm:"type:uuid;not null" json:"post_id"`
	Post      Post      `gorm:"foreignKey:PostID" json:"post"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	CreatedAt time.Time `json:"created_at"`
}
