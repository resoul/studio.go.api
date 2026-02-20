package usecase

import (
	"context"

	"github.com/football.manager.api/internal/domain"
)

type UserListDTO struct {
	Users    []*AdminUserDTO `json:"users"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

type AdminUserDTO struct {
	User    *UserDTO     `json:"user"`
	Manager *ManagerDTO  `json:"manager,omitempty"`
	Careers []*CareerDTO `json:"careers"`
}

type UserUseCase interface {
	GetByID(ctx context.Context, id uint) (*UserDTO, error)
	ListUsers(ctx context.Context, page, pageSize int) (*UserListDTO, error)
}

type userUseCase struct {
	userRepo    domain.UserRepository
	managerRepo domain.ManagerRepository
	careerRepo  domain.CareerRepository
}

func NewUserUseCase(userRepo domain.UserRepository, managerRepo domain.ManagerRepository, careerRepo domain.CareerRepository) UserUseCase {
	return &userUseCase{
		userRepo:    userRepo,
		managerRepo: managerRepo,
		careerRepo:  careerRepo,
	}
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

	userIDs := make([]uint, 0, len(users))
	for _, u := range users {
		userIDs = append(userIDs, u.ID)
	}

	managers, err := uc.managerRepo.ListByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	managerByUserID := make(map[uint]*domain.Manager, len(managers))
	managerIDs := make([]uint, 0, len(managers))
	for _, manager := range managers {
		managerByUserID[manager.UserID] = manager
		managerIDs = append(managerIDs, manager.ID)
	}

	careers, err := uc.careerRepo.ListByManagerIDs(ctx, managerIDs)
	if err != nil {
		return nil, err
	}

	careersByManagerID := make(map[uint][]*domain.Career, len(managerIDs))
	for _, career := range careers {
		careersByManagerID[career.ManagerID] = append(careersByManagerID[career.ManagerID], career)
	}

	dtos := make([]*AdminUserDTO, 0, len(users))
	for _, user := range users {
		manager := managerByUserID[user.ID]
		managerDTO := mapManagerToDTO(manager)
		careerDTOs := make([]*CareerDTO, 0)
		if manager != nil {
			for _, career := range careersByManagerID[manager.ID] {
				careerDTOs = append(careerDTOs, mapCareerToDTO(career))
			}
		}

		dtos = append(dtos, &AdminUserDTO{
			User:    mapUserToDTO(user),
			Manager: managerDTO,
			Careers: careerDTOs,
		})
	}

	return &UserListDTO{
		Users:    dtos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
