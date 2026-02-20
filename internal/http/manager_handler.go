package http

import (
	"net/http"
	"strings"

	"github.com/football.manager.api/internal/domain"
	platformauth "github.com/football.manager.api/internal/platform/auth"
	"github.com/football.manager.api/internal/platform/httpx"
	"github.com/football.manager.api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ManagerHandler struct {
	managerUC usecase.ManagerUseCase
}

func NewManagerHandler(managerUC usecase.ManagerUseCase) *ManagerHandler {
	return &ManagerHandler{managerUC: managerUC}
}

func (h *ManagerHandler) CreateMe(c *gin.Context) {
	userID, ok := platformauth.GetUserIDFromContext(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	var req CreateManagerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	manager, err := h.managerUC.Create(c.Request.Context(), userID, toCreateManagerDTO(req))
	if err != nil {
		switch err {
		case domain.ErrManagerExists:
			httpx.RespondError(c, http.StatusConflict, "manager_exists", "Manager already exists")
		default:
			if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "YYYY-MM-DD") {
				httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
				return
			}
			logrus.WithError(err).Error("Failed to create manager")
			httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to create manager")
		}
		return
	}

	httpx.RespondCreated(c, gin.H{
		"message": "Manager created",
		"manager": toManagerResponse(manager),
	})
}

func (h *ManagerHandler) GetMe(c *gin.Context) {
	userID, ok := platformauth.GetUserIDFromContext(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	manager, err := h.managerUC.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		if err == domain.ErrManagerNotFound {
			httpx.RespondError(c, http.StatusNotFound, "not_found", "Manager not found")
			return
		}
		logrus.WithError(err).Error("Failed to get current manager")
		httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to get manager")
		return
	}

	httpx.RespondOK(c, gin.H{
		"manager": toManagerResponse(manager),
	})
}
