package data

import (
	"github.com/football.manager.api/internal/domain"
)

func toUserDomain(model *UserModel) *domain.User {
	if model == nil {
		return nil
	}
	return &domain.User{
		ID:                     model.ID,
		UUID:                   model.UUID,
		FullName:               model.FullName,
		Email:                  model.Email,
		PasswordHash:           model.PasswordHash,
		Role:                   model.Role,
		RegistrationIP:         model.RegistrationIP,
		RegistrationUserAgent:  model.RegistrationUserAgent,
		LoginCount:             model.LoginCount,
		EmailVerifiedAt:        model.EmailVerifiedAt,
		VerificationCode:       model.VerificationCode,
		VerificationExpiresAt:  model.VerificationExpiresAt,
		ResetPasswordCode:      model.ResetPasswordCode,
		ResetPasswordExpiresAt: model.ResetPasswordExpiresAt,
		CreatedAt:              model.CreatedAt,
		UpdatedAt:              model.UpdatedAt,
	}
}

func toUserModel(entity *domain.User) *UserModel {
	if entity == nil {
		return nil
	}
	return &UserModel{
		ID:                     entity.ID,
		UUID:                   entity.UUID,
		FullName:               entity.FullName,
		Email:                  entity.Email,
		PasswordHash:           entity.PasswordHash,
		Role:                   entity.Role,
		RegistrationIP:         entity.RegistrationIP,
		RegistrationUserAgent:  entity.RegistrationUserAgent,
		LoginCount:             entity.LoginCount,
		EmailVerifiedAt:        entity.EmailVerifiedAt,
		VerificationCode:       entity.VerificationCode,
		VerificationExpiresAt:  entity.VerificationExpiresAt,
		ResetPasswordCode:      entity.ResetPasswordCode,
		ResetPasswordExpiresAt: entity.ResetPasswordExpiresAt,
		CreatedAt:              entity.CreatedAt,
		UpdatedAt:              entity.UpdatedAt,
	}
}

func toManagerDomain(model *ManagerModel) *domain.Manager {
	if model == nil {
		return nil
	}
	return &domain.Manager{
		ID:        model.ID,
		UserID:    model.UserID,
		FirstName: model.FirstName,
		LastName:  model.LastName,
		Birthday:  model.Birthday,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func toManagerModel(entity *domain.Manager) *ManagerModel {
	if entity == nil {
		return nil
	}
	return &ManagerModel{
		ID:        entity.ID,
		UserID:    entity.UserID,
		FirstName: entity.FirstName,
		LastName:  entity.LastName,
		Birthday:  entity.Birthday,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}

func toCareerDomain(model *CareerModel) *domain.Career {
	if model == nil {
		return nil
	}
	return &domain.Career{
		ID:        model.ID,
		ManagerID: model.ManagerID,
		Name:      model.Name,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func toCareerModel(entity *domain.Career) *CareerModel {
	if entity == nil {
		return nil
	}
	return &CareerModel{
		ID:        entity.ID,
		ManagerID: entity.ManagerID,
		Name:      entity.Name,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}
