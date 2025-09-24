package models

import (
	"time"

	"github.com/google/uuid"
)

// Issue category constants
const (
	CategoryLighting    = "Lighting"
	CategoryRoad        = "Road"
	CategorySanitation  = "Sanitation"
	CategoryUtilities   = "Utilities"
	CategoryVandalism   = "Vandalism"
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
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	IssueID     uuid.UUID `gorm:"type:uuid;not null" json:"issue_id"`
	Issue       Issue     `gorm:"foreignKey:IssueID" json:"issue"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"user"`
	Description string    `json:"description,omitempty"`
	Status      string    `gorm:"type:status;default:'open';not null" json:"status"`
	Urgency     int       `gorm:"not null" json:"urgency"`
	Lat         float64   `gorm:"not null" json:"lat"`
	Lng         float64   `gorm:"not null" json:"lng"`
	MediaURL    string    `gorm:"not null" json:"media_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
