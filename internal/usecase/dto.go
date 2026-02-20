package usecase

import (
	"time"

	"github.com/football.manager.api/internal/domain"
)

type RegisterDTO struct {
	FullName string
	Email    string
	Password string
}

type LoginDTO struct {
	Email    string
	Password string
}

type VerifyEmailDTO struct {
	Email string
	Code  string
}

type ResendVerificationDTO struct {
	Email string
}

type ResetPasswordRequestDTO struct {
	Email string
}

type ResetPasswordDTO struct {
	Email       string
	Code        string
	NewPassword string
}

type UserDTO struct {
	ID              uint   `json:"id"`
	UUID            string `json:"uuid"`
	FullName        string `json:"full_name"`
	Email           string `json:"email"`
	EmailVerified   bool   `json:"email_verified"`
	EmailVerifiedAt *int64 `json:"email_verified_at,omitempty"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
}

type CreateManagerDTO struct {
	FirstName string
	LastName  string
	Birthday  string
}

type ManagerDTO struct {
	ID        uint
	UserID    uint
	FirstName string
	LastName  string
	Birthday  string
	CreatedAt int64
	UpdatedAt int64
}

func mapManagerToDTO(manager *domain.Manager) *ManagerDTO {
	if manager == nil {
		return nil
	}

	return &ManagerDTO{
		ID:        manager.ID,
		UserID:    manager.UserID,
		FirstName: manager.FirstName,
		LastName:  manager.LastName,
		Birthday:  manager.Birthday.Format(time.DateOnly),
		CreatedAt: manager.CreatedAt.Unix(),
		UpdatedAt: manager.UpdatedAt.Unix(),
	}
}
