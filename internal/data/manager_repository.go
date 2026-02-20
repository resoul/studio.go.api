package data

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/football.manager.api/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type managerRepository struct {
	db *gorm.DB
}

func NewManagerRepository(db *gorm.DB) domain.ManagerRepository {
	return &managerRepository{db: db}
}

func (r *managerRepository) Create(ctx context.Context, manager *domain.Manager) error {
	model := toManagerModel(manager)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			constraint := strings.ToLower(pgErr.ConstraintName)
			if strings.Contains(constraint, "user") {
				return domain.ErrManagerExists
			}
		}
		return fmt.Errorf("failed to create manager: %w", err)
	}

	manager.ID = model.ID
	manager.CreatedAt = model.CreatedAt
	manager.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *managerRepository) GetByUserID(ctx context.Context, userID uint) (*domain.Manager, error) {
	var model ManagerModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrManagerNotFound
		}
		return nil, fmt.Errorf("failed to get manager by user id: %w", err)
	}
	return toManagerDomain(&model), nil
}

func (r *managerRepository) ExistsByUserID(ctx context.Context, userID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&ManagerModel{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check manager existence: %w", err)
	}
	return count > 0, nil
}

func (r *managerRepository) ListByUserIDs(ctx context.Context, userIDs []uint) ([]*domain.Manager, error) {
	if len(userIDs) == 0 {
		return []*domain.Manager{}, nil
	}

	var models []ManagerModel
	if err := r.db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list managers by user ids: %w", err)
	}

	managers := make([]*domain.Manager, 0, len(models))
	for i := range models {
		managers = append(managers, toManagerDomain(&models[i]))
	}
	return managers, nil
}
