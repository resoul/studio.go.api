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

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	model := toUserModel(user)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			constraint := strings.ToLower(pgErr.ConstraintName)
			if strings.Contains(constraint, "email") {
				return domain.ErrUserAlreadyExists
			}
			if strings.Contains(constraint, "username") {
				return domain.ErrUsernameTaken
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = model.ID
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return toUserDomain(&model), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return toUserDomain(&model), nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return toUserDomain(&model), nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	model := toUserModel(user)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	user.UpdatedAt = model.UpdatedAt
	return nil
}
