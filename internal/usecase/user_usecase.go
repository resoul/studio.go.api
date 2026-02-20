package usecase

import (
	"context"

	"github.com/football.manager.api/internal/domain"
)

type UserListDTO struct {
	Users    []*UserDTO `json:"users"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

type UserUseCase interface {
	GetByID(ctx context.Context, id uint) (*UserDTO, error)
	ListUsers(ctx context.Context, page, pageSize int) (*UserListDTO, error)
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

func (uc *userUseCase) ListUsers(ctx context.Context, page, pageSize int) (*UserListDTO, error) {
	users, total, err := uc.userRepo.ListAll(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	dtos := make([]*UserDTO, 0, len(users))
	for _, u := range users {
		dtos = append(dtos, mapUserToDTO(u))
	}

	return &UserListDTO{
		Users:    dtos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
