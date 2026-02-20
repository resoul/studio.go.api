package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/football.manager.api/internal/domain"
)

type CareerUseCase interface {
	CreateForUser(ctx context.Context, userID uint, dto CreateCareerDTO) (*CareerDTO, error)
	ListForUser(ctx context.Context, userID uint) ([]*CareerDTO, error)
}

type careerUseCase struct {
	managerRepo domain.ManagerRepository
	careerRepo  domain.CareerRepository
}

func NewCareerUseCase(managerRepo domain.ManagerRepository, careerRepo domain.CareerRepository) CareerUseCase {
	return &careerUseCase{
		managerRepo: managerRepo,
		careerRepo:  careerRepo,
	}
}

func (uc *careerUseCase) CreateForUser(ctx context.Context, userID uint, dto CreateCareerDTO) (*CareerDTO, error) {
	manager, err := uc.managerRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(dto.Name)
	if name == "" {
		return nil, fmt.Errorf("career name is required")
	}

	career := &domain.Career{
		ManagerID: manager.ID,
		Name:      name,
	}

	if err := uc.careerRepo.Create(ctx, career); err != nil {
		return nil, err
	}

	return mapCareerToDTO(career), nil
}

func (uc *careerUseCase) ListForUser(ctx context.Context, userID uint) ([]*CareerDTO, error) {
	manager, err := uc.managerRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	careers, err := uc.careerRepo.ListByManagerID(ctx, manager.ID)
	if err != nil {
		return nil, err
	}

	result := make([]*CareerDTO, 0, len(careers))
	for _, c := range careers {
		result = append(result, mapCareerToDTO(c))
	}
	return result, nil
}
