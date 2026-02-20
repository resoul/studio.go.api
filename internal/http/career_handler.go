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

type CareerHandler struct {
	careerUC usecase.CareerUseCase
}

func NewCareerHandler(careerUC usecase.CareerUseCase) *CareerHandler {
	return &CareerHandler{careerUC: careerUC}
}

func (h *CareerHandler) CreateMe(c *gin.Context) {
	userID, ok := platformauth.GetUserIDFromContext(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	var req CreateCareerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	career, err := h.careerUC.CreateForUser(c.Request.Context(), userID, toCreateCareerDTO(req))
	if err != nil {
		switch err {
		case domain.ErrManagerNotFound:
			httpx.RespondError(c, http.StatusNotFound, "manager_not_found", "Manager not found")
		default:
			if strings.Contains(strings.ToLower(err.Error()), "required") {
				httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
				return
			}
			logrus.WithError(err).Error("Failed to create career")
			httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to create career")
		}
		return
	}

	httpx.RespondCreated(c, gin.H{
		"message": "Career created",
		"career":  toCareerResponse(career),
	})
}

func (h *CareerHandler) ListMe(c *gin.Context) {
	userID, ok := platformauth.GetUserIDFromContext(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	careers, err := h.careerUC.ListForUser(c.Request.Context(), userID)
	if err != nil {
		if err == domain.ErrManagerNotFound {
			httpx.RespondError(c, http.StatusNotFound, "manager_not_found", "Manager not found")
			return
		}

		logrus.WithError(err).Error("Failed to list careers")
		httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to list careers")
		return
	}

	items := make([]CareerResponse, 0, len(careers))
	for _, crr := range careers {
		items = append(items, toCareerResponse(crr))
	}

	httpx.RespondOK(c, gin.H{
		"careers": items,
	})
}
