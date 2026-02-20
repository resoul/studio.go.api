package data

import (
	"context"
	"fmt"

	"github.com/football.manager.api/internal/domain"
	"gorm.io/gorm"
)

type careerRepository struct {
	db *gorm.DB
}

func NewCareerRepository(db *gorm.DB) domain.CareerRepository {
	return &careerRepository{db: db}
}

func (r *careerRepository) Create(ctx context.Context, career *domain.Career) error {
	model := toCareerModel(career)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create career: %w", err)
	}

	career.ID = model.ID
	career.CreatedAt = model.CreatedAt
	career.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *careerRepository) ListByManagerID(ctx context.Context, managerID uint) ([]*domain.Career, error) {
	var models []CareerModel
	if err := r.db.WithContext(ctx).Where("manager_id = ?", managerID).Order("id DESC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list careers by manager id: %w", err)
	}

	careers := make([]*domain.Career, 0, len(models))
	for i := range models {
		careers = append(careers, toCareerDomain(&models[i]))
	}
	return careers, nil
}

func (r *careerRepository) ListByManagerIDs(ctx context.Context, managerIDs []uint) ([]*domain.Career, error) {
	if len(managerIDs) == 0 {
		return []*domain.Career{}, nil
	}

	var models []CareerModel
	if err := r.db.WithContext(ctx).Where("manager_id IN ?", managerIDs).Order("id DESC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list careers by manager ids: %w", err)
	}

	careers := make([]*domain.Career, 0, len(models))
	for i := range models {
		careers = append(careers, toCareerDomain(&models[i]))
	}
	return careers, nil
}
