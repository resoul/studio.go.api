package usecase

import (
	"context"

	"github.com/football.manager.api/internal/domain"
)

type UserUseCase interface {
	GetByID(ctx context.Context, id uint) (*UserDTO, error)
}

type userUseCase struct {
	userRepo domain.UserRepository
}

func NewUserUseCase(userRepo domain.UserRepository) UserUseCase {
	return &userUseCase{userRepo: userRepo}
}

func (uc *userUseCase) GetByID(ctx context.Context, id uint) (*UserDTO, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapUserToDTO(user), nil
}
