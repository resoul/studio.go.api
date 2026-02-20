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
