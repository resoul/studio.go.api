package http

import (
	"net/http"
	"strconv"

	"github.com/football.manager.api/internal/domain"
	platformauth "github.com/football.manager.api/internal/platform/auth"
	"github.com/football.manager.api/internal/platform/httpx"
	"github.com/football.manager.api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	userUC usecase.UserUseCase
}

func NewUserHandler(userUC usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "invalid_id", "Invalid user ID format")
		return
	}

	user, err := h.userUC.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if err == domain.ErrUserNotFound {
			httpx.RespondError(c, http.StatusNotFound, "not_found", "User not found")
			return
		}
		logrus.WithError(err).Error("Failed to get user")
		httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to get user")
		return
	}

	httpx.RespondOK(c, toUserResponse(user))
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID, ok := platformauth.GetUserIDFromContext(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	user, err := h.userUC.GetByID(c.Request.Context(), userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			httpx.RespondError(c, http.StatusNotFound, "not_found", "User not found")
			return
		}
		logrus.WithError(err).Error("Failed to get current user")
		httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to get current user")
		return
	}

	httpx.RespondOK(c, gin.H{
		"user": toUserResponse(user),
	})
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	result, err := h.userUC.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		logrus.WithError(err).Error("Failed to list users")
		httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to list users")
		return
	}

	httpx.RespondOK(c, result)
}
