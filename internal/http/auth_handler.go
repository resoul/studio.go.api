package http

import (
	"net/http"

	"github.com/football.manager.api/internal/domain"
	platformauth "github.com/football.manager.api/internal/platform/auth"
	"github.com/football.manager.api/internal/platform/httpx"
	"github.com/football.manager.api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	authUC usecase.AuthUseCase
}

func NewAuthHandler(authUC usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUC: authUC,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	user, err := h.authUC.Register(c.Request.Context(), toRegisterDTO(req), c.ClientIP(), c.Request.UserAgent(), getLocale(c))
	if err != nil {
		if err == domain.ErrUserAlreadyExists {
			httpx.RespondError(c, http.StatusConflict, "user_exists", "User with this email already exists")
			return
		}
		if err == domain.ErrUsernameTaken {
			httpx.RespondError(c, http.StatusConflict, "username_exists", "User with this username already exists")
			return
		}
		if err == domain.ErrInvalidUsername {
			httpx.RespondError(c, http.StatusBadRequest, "invalid_username", "Username must be 3-32 chars and contain only a-z, 0-9, dot, dash or underscore")
			return
		}
		logrus.WithError(err).Error("Failed to register user")
		httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to register user")
		return
	}

	httpx.RespondCreated(c, AuthSuccessResponse{
		Message: "Registration successful. Check email for verification code",
		User:    toUserResponse(user),
	})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	user, token, onboardingRequired, err := h.authUC.VerifyEmail(c.Request.Context(), toVerifyEmailDTO(req), getLocale(c))
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			httpx.RespondError(c, http.StatusNotFound, "not_found", "User not found")
		case domain.ErrInvalidCode:
			httpx.RespondError(c, http.StatusBadRequest, "invalid_code", "Verification code is invalid")
		case domain.ErrCodeExpired:
			httpx.RespondError(c, http.StatusBadRequest, "code_expired", "Verification code expired")
		default:
			logrus.WithError(err).Error("Failed to verify email")
			httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to verify email")
		}
		return
	}

	httpx.RespondOK(c, AuthSuccessResponse{
		Message:            "Email verified successfully",
		User:               toUserResponse(user),
		Token:              token,
		OnboardingRequired: onboardingRequired,
	})
}

func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if err := h.authUC.ResendVerificationCode(c.Request.Context(), toResendVerificationDTO(req), getLocale(c)); err != nil {
		logrus.WithError(err).Error("Failed to resend verification code")
		httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to resend verification code")
		return
	}

	httpx.RespondOK(c, AuthSuccessResponse{Message: "If account exists and email is not verified, verification code was sent"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	user, token, onboardingRequired, err := h.authUC.Login(c.Request.Context(), toLoginDTO(req), c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			httpx.RespondError(c, http.StatusUnauthorized, "invalid_credentials", "Email or password is invalid")
		case domain.ErrEmailNotVerified:
			httpx.RespondError(c, http.StatusForbidden, "email_not_verified", "Please verify your email first")
		default:
			logrus.WithError(err).Error("Failed to login")
			httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to login")
		}
		return
	}

	httpx.RespondOK(c, AuthSuccessResponse{
		Message:            "Login successful",
		User:               toUserResponse(user),
		Token:              token,
		OnboardingRequired: onboardingRequired,
	})
}

func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	user, token, onboardingRequired, err := h.authUC.AdminLogin(c.Request.Context(), toLoginDTO(req), c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			httpx.RespondError(c, http.StatusUnauthorized, "invalid_credentials", "Email or password is invalid")
		case domain.ErrEmailNotVerified:
			httpx.RespondError(c, http.StatusForbidden, "email_not_verified", "Please verify your email first")
		case domain.ErrAdminAccessDenied:
			httpx.RespondError(c, http.StatusForbidden, "forbidden", "Admin access required")
		default:
			logrus.WithError(err).Error("Failed to login admin")
			httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to login admin")
		}
		return
	}

	httpx.RespondOK(c, AuthSuccessResponse{
		Message:            "Admin login successful",
		User:               toUserResponse(user),
		Token:              token,
		Role:               platformauth.RoleAdmin,
		OnboardingRequired: onboardingRequired,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	clearCookie(c, "fm-auth-token")
	clearCookie(c, "token")
	clearCookie(c, "access_token")
	clearCookie(c, "refresh_token")

	httpx.RespondOK(c, AuthSuccessResponse{Message: "Logout successful"})
}

func (h *AuthHandler) RequestResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if err := h.authUC.RequestPasswordReset(c.Request.Context(), toResetPasswordRequestDTO(req), getLocale(c)); err != nil {
		logrus.WithError(err).Error("Failed to request reset password")
		httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to request reset password")
		return
	}

	httpx.RespondOK(c, AuthSuccessResponse{Message: "If email exists, reset code was sent"})
}

func (h *AuthHandler) ConfirmResetPassword(c *gin.Context) {
	var req ConfirmResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	err := h.authUC.ResetPassword(c.Request.Context(), toResetPasswordDTO(req), getLocale(c))
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			httpx.RespondError(c, http.StatusNotFound, "not_found", "User not found")
		case domain.ErrInvalidCode:
			httpx.RespondError(c, http.StatusBadRequest, "invalid_code", "Reset code is invalid")
		case domain.ErrCodeExpired:
			httpx.RespondError(c, http.StatusBadRequest, "code_expired", "Reset code expired")
		default:
			logrus.WithError(err).Error("Failed to reset password")
			httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to reset password")
		}
		return
	}

	httpx.RespondOK(c, AuthSuccessResponse{Message: "Password updated successfully"})
}

func (h *AuthHandler) CheckAuth(c *gin.Context) {
	if _, ok := platformauth.GetUserIDFromContext(c); !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	role, _ := platformauth.GetUserRoleFromContext(c)
	httpx.RespondOK(c, gin.H{"status": "ok", "role": role})
}

func clearCookie(c *gin.Context, name string) {
	c.SetCookie(name, "", -1, "/", "", false, true)
}

func getLocale(c *gin.Context) string {
	header := c.GetHeader("Accept-Language")
	if len(header) < 2 {
		return "en"
	}
	switch header[:2] {
	case "ru", "RU":
		return "ru"
	default:
		return "en"
	}
}
