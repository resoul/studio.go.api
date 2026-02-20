package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/football.manager.api/internal/domain"
)

type ManagerUseCase interface {
	Create(ctx context.Context, userID uint, dto CreateManagerDTO) (*ManagerDTO, error)
	GetByUserID(ctx context.Context, userID uint) (*ManagerDTO, error)
	ExistsByUserID(ctx context.Context, userID uint) (bool, error)
}

type managerUseCase struct {
	managerRepo domain.ManagerRepository
}

func NewManagerUseCase(managerRepo domain.ManagerRepository) ManagerUseCase {
	return &managerUseCase{managerRepo: managerRepo}
}

func (uc *managerUseCase) Create(ctx context.Context, userID uint, dto CreateManagerDTO) (*ManagerDTO, error) {
	firstName := strings.TrimSpace(dto.FirstName)
	lastName := strings.TrimSpace(dto.LastName)
	if firstName == "" || lastName == "" || strings.TrimSpace(dto.Birthday) == "" {
		return nil, fmt.Errorf("first name, last name and birthday are required")
	}

	birthday, err := time.Parse(time.DateOnly, dto.Birthday)
	if err != nil {
		return nil, fmt.Errorf("birthday must be in YYYY-MM-DD format")
	}

	manager := &domain.Manager{
		UserID:    userID,
		FirstName: firstName,
		LastName:  lastName,
		Birthday:  birthday,
	}

	if err := uc.managerRepo.Create(ctx, manager); err != nil {
		return nil, err
	}

	return mapManagerToDTO(manager), nil
}

func (uc *managerUseCase) GetByUserID(ctx context.Context, userID uint) (*ManagerDTO, error) {
	manager, err := uc.managerRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return mapManagerToDTO(manager), nil
}

func (uc *managerUseCase) ExistsByUserID(ctx context.Context, userID uint) (bool, error) {
	return uc.managerRepo.ExistsByUserID(ctx, userID)
}
