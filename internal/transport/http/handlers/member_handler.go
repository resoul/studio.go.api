package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/resoul/studio.go.api/internal/transport/http/utils"
)

func (h *WorkspaceHandler) ListMembers(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid workspace id")
		return
	}

	members, err := h.service.ListMembers(c.Request.Context(), wsID)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}
	utils.RespondOK(c, members)
}

func (h *WorkspaceHandler) RemoveMember(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid workspace id")
		return
	}

	userID := c.Param("user_id")
	if userID == "" {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "User ID is required")
		return
	}

	if err := h.service.RemoveMember(c.Request.Context(), wsID, userID); err != nil {
		utils.RespondMapped(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
