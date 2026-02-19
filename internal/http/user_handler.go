package http

import (
	"net/http"
	"strconv"

	"github.com/football.manager.api/internal/domain"
	"github.com/football.manager.api/internal/infrastructure"
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid user ID format"})
		return
	}

	user, err := h.userUC.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if err == domain.ErrUserNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "User not found"})
			return
		}
		logrus.WithError(err).Error("Failed to get user")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID, ok := infrastructure.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Not authenticated"})
		return
	}

	user, err := h.userUC.GetByID(c.Request.Context(), userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "User not found"})
			return
		}
		logrus.WithError(err).Error("Failed to get current user")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to get current user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": toUserResponse(user),
	})
}
