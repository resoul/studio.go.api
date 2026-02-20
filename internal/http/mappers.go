package http

import (
	"github.com/football.manager.api/internal/usecase"
)

func toRegisterDTO(req RegisterRequest) usecase.RegisterDTO {
	return usecase.RegisterDTO{
		FullName: req.FullName,
		Email:    req.Email,
		Password: req.Password,
	}
}

func toVerifyEmailDTO(req VerifyEmailRequest) usecase.VerifyEmailDTO {
	return usecase.VerifyEmailDTO{
		Email: req.Email,
		Code:  req.Code,
	}
}

func toResendVerificationDTO(req ResetPasswordRequest) usecase.ResendVerificationDTO {
	return usecase.ResendVerificationDTO{
		Email: req.Email,
	}
}

func toLoginDTO(req LoginRequest) usecase.LoginDTO {
	return usecase.LoginDTO{
		Email:    req.Email,
		Password: req.Password,
	}
}

func toResetPasswordRequestDTO(req ResetPasswordRequest) usecase.ResetPasswordRequestDTO {
	return usecase.ResetPasswordRequestDTO{
		Email: req.Email,
	}
}

func toResetPasswordDTO(req ConfirmResetPasswordRequest) usecase.ResetPasswordDTO {
	return usecase.ResetPasswordDTO{
		Email:       req.Email,
		Code:        req.Code,
		NewPassword: req.NewPassword,
	}
}

func toCreateManagerDTO(req CreateManagerRequest) usecase.CreateManagerDTO {
	return usecase.CreateManagerDTO{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Birthday:  req.Birthday,
	}
}

func toUserResponse(dto *usecase.UserDTO) UserResponse {
	if dto == nil {
		return UserResponse{}
	}

	return UserResponse{
		ID:              dto.UUID,
		FullName:        dto.FullName,
		Email:           dto.Email,
		EmailVerified:   dto.EmailVerified,
		EmailVerifiedAt: dto.EmailVerifiedAt,
		CreatedAt:       dto.CreatedAt,
		UpdatedAt:       dto.UpdatedAt,
	}
}

func toManagerResponse(dto *usecase.ManagerDTO) ManagerResponse {
	if dto == nil {
		return ManagerResponse{}
	}

	return ManagerResponse{
		ID:        dto.ID,
		UserID:    dto.UserID,
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		Birthday:  dto.Birthday,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}
