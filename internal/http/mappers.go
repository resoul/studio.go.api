package http

import (
	"github.com/football.manager.api/internal/usecase"
)

func toRegisterDTO(req RegisterRequest) usecase.RegisterDTO {
	return usecase.RegisterDTO{
		Username: req.Username,
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

func toUserResponse(dto *usecase.UserDTO) UserResponse {
	if dto == nil {
		return UserResponse{}
	}

	return UserResponse{
		ID:              dto.ID,
		Username:        dto.Username,
		FullName:        dto.FullName,
		Email:           dto.Email,
		EmailVerified:   dto.EmailVerified,
		EmailVerifiedAt: dto.EmailVerifiedAt,
		CreatedAt:       dto.CreatedAt,
		UpdatedAt:       dto.UpdatedAt,
	}
}
