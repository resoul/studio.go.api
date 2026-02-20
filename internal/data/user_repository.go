package data

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/football.manager.api/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uint, ip, userAgent string) error {
	model := &UserLastLoginModel{
		UserID:             userID,
		LastLoginAt:        timePtr(time.Now().UTC()),
		LastLoginIP:        ip,
		LastLoginUserAgent: userAgent,
	}

	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_login_at":         model.LastLoginAt,
			"last_login_ip":         model.LastLoginIP,
			"last_login_user_agent": model.LastLoginUserAgent,
			"updated_at":            time.Now().UTC(),
		}),
	}).Create(model).Error; err != nil {
		return fmt.Errorf("failed to update last login metadata: %w", err)
	}

	return nil
}

func (r *userRepository) SetRole(ctx context.Context, userID uint, role string) error {
	if err := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", userID).Update("role", role).Error; err != nil {
		return fmt.Errorf("failed to set user role: %w", err)
	}
	return nil
}

func (r *userRepository) ListAll(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	if err := r.db.WithContext(ctx).Model(&UserModel{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	var models []UserModel
	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Order("id DESC").Offset(offset).Limit(pageSize).Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]*domain.User, 0, len(models))
	for i := range models {
		users = append(users, toUserDomain(&models[i]))
	}
	return users, total, nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}
