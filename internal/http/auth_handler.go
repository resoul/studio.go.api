package http

import (
	"net/http"

	"github.com/football.manager.api/internal/domain"
	"github.com/football.manager.api/internal/infrastructure"
	"github.com/football.manager.api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	authUC       usecase.AuthUseCase
	tokenManager *infrastructure.UserTokenManager
}

func NewAuthHandler(authUC usecase.AuthUseCase, tokenManager *infrastructure.UserTokenManager) *AuthHandler {
	return &AuthHandler{
		authUC:       authUC,
		tokenManager: tokenManager,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	user, err := h.authUC.Register(c.Request.Context(), toRegisterDTO(req))
	if err != nil {
		if err == domain.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "user_exists", Message: "User with this email already exists"})
			return
		}
		if err == domain.ErrUsernameTaken {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "username_exists", Message: "User with this username already exists"})
			return
		}
		if err == domain.ErrInvalidUsername {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_username", Message: "Username must be 3-32 chars and contain only a-z, 0-9, dot, dash or underscore"})
			return
		}
		logrus.WithError(err).Error("Failed to register user")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, AuthSuccessResponse{
		Message: "Registration successful. Check email for verification code",
		User:    toUserResponse(user),
	})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	err := h.authUC.VerifyEmail(c.Request.Context(), toVerifyEmailDTO(req))
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "User not found"})
		case domain.ErrInvalidCode:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_code", Message: "Verification code is invalid"})
		case domain.ErrCodeExpired:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "code_expired", Message: "Verification code expired"})
		default:
			logrus.WithError(err).Error("Failed to verify email")
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to verify email"})
		}
		return
	}

	c.JSON(http.StatusOK, AuthSuccessResponse{Message: "Email verified successfully"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	user, err := h.authUC.Login(c.Request.Context(), toLoginDTO(req))
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid_credentials", Message: "Email or password is invalid"})
		case domain.ErrEmailNotVerified:
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "email_not_verified", Message: "Please verify your email first"})
		default:
			logrus.WithError(err).Error("Failed to login")
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to login"})
		}
		return
	}

	token, err := h.tokenManager.Generate(user.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate token")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to login"})
		return
	}

	c.JSON(http.StatusOK, AuthSuccessResponse{
		Message: "Login successful",
		User:    toUserResponse(user),
		Token:   token,
	})
}

func (h *AuthHandler) RequestResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	if err := h.authUC.RequestPasswordReset(c.Request.Context(), toResetPasswordRequestDTO(req)); err != nil {
		logrus.WithError(err).Error("Failed to request reset password")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to request reset password"})
		return
	}

	c.JSON(http.StatusOK, AuthSuccessResponse{Message: "If email exists, reset code was sent"})
}

func (h *AuthHandler) ConfirmResetPassword(c *gin.Context) {
	var req ConfirmResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	err := h.authUC.ResetPassword(c.Request.Context(), toResetPasswordDTO(req))
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "User not found"})
		case domain.ErrInvalidCode:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_code", Message: "Reset code is invalid"})
		case domain.ErrCodeExpired:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "code_expired", Message: "Reset code expired"})
		default:
			logrus.WithError(err).Error("Failed to reset password")
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to reset password"})
		}
		return
	}

	c.JSON(http.StatusOK, AuthSuccessResponse{Message: "Password updated successfully"})
}

func (h *AuthHandler) CheckAuth(c *gin.Context) {
	if _, ok := infrastructure.GetUserIDFromContext(c); !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
