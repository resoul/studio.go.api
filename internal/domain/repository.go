package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdateLastLogin(ctx context.Context, userID uint, ip, userAgent string) error
	SetRole(ctx context.Context, userID uint, role string) error
	ListAll(ctx context.Context, page, pageSize int) ([]*User, int64, error)
}

type ManagerRepository interface {
	Create(ctx context.Context, manager *Manager) error
	GetByUserID(ctx context.Context, userID uint) (*Manager, error)
	ExistsByUserID(ctx context.Context, userID uint) (bool, error)
}
