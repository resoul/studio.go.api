package domain

import (
	"context"
	"io"
	"time"
)

type Profile struct {
	ID         string     `gorm:"primaryKey;type:uuid" json:"id"` // Identity ID from Kratos
	FirstName  string     `json:"first_name"`
	LastName   string     `json:"last_name"`
	AvatarURL  string     `json:"avatar_url"`
	Completed  bool       `gorm:"default:false" json:"completed"`
	LastSeenAt *time.Time `json:"last_seen_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type ProfileRepository interface {
	FindByID(ctx context.Context, id string) (*Profile, error)
	Create(ctx context.Context, profile *Profile) error
	Update(ctx context.Context, profile *Profile) error
	UpdateLastSeen(ctx context.Context, id string, lastSeenAt time.Time) error
}

type UpdateProfileInput struct {
	FirstName  string
	LastName   string
	Avatar     io.Reader
	AvatarSize int64
	AvatarType string
}

type ProfileService interface {
	GetProfile(ctx context.Context, userID string) (*Profile, error)
	UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*Profile, error)
	EnsureProfileExists(ctx context.Context, userID string) (*Profile, error)
	MarkLastSeen(ctx context.Context, userID string) error
}
