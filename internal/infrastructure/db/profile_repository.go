package db

import (
	"context"
	"time"

	"github.com/resoul/studio.go.api/internal/domain"
	"gorm.io/gorm"
)

type profileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) domain.ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) FindByID(ctx context.Context, id string) (*domain.Profile, error) {
	var profile domain.Profile
	if err := r.db.WithContext(ctx).First(&profile, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *profileRepository) Create(ctx context.Context, profile *domain.Profile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *profileRepository) Update(ctx context.Context, profile *domain.Profile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

func (r *profileRepository) UpdateLastSeen(ctx context.Context, id string, lastSeenAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&domain.Profile{}).
		Where("id = ?", id).
		Update("last_seen_at", lastSeenAt).
		Error
}
