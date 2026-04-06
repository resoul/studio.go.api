package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ory "github.com/ory/client-go"
	"github.com/resoul/studio.go.api/internal/domain"
	"github.com/resoul/studio.go.api/internal/transport/http/utils"
)

// createInviteRequest is the typed JSON body for POST /workspaces/:id/invites.
type createInviteRequest struct {
	Email     string               `json:"email"      binding:"required,email"`
	Role      domain.WorkspaceRole `json:"role"       binding:"required"`
	SendEmail bool                 `json:"send_email"`
}

// InvitePreviewResponse is the public shape for GET /workspaces/invites/:token/preview.
type InvitePreviewResponse struct {
	ID           uuid.UUID `json:"id"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	LogoURL      string    `json:"logo_url"`
	MembersCount int64     `json:"members_count"`
	Email        string    `json:"email"`
}

func (h *WorkspaceHandler) GetInvitePreview(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Token is required")
		return
	}

	ws, membersCount, invite, err := h.service.PreviewInvite(c.Request.Context(), token)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}

	utils.RespondOK(c, InvitePreviewResponse{
		ID:           ws.ID,
		Slug:         ws.Slug,
		Name:         ws.Name,
		LogoURL:      ws.LogoURL,
		MembersCount: membersCount,
		Email:        invite.Email,
	})
}

func (h *WorkspaceHandler) AcceptInvite(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found in context")
		return
	}

	oryIdentity, ok := identity.(*ory.Identity)
	if !ok {
		utils.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid identity type")
		return
	}

	token := c.Param("token")
	if token == "" {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Token is required")
		return
	}

	if err := h.service.AcceptInvite(c.Request.Context(), token, oryIdentity.Id); err != nil {
		utils.RespondMapped(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *WorkspaceHandler) CreateInvite(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid workspace id")
		return
	}

	var req createInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	invite, err := h.service.InviteUser(c.Request.Context(), domain.CreateInviteInput{
		WorkspaceID:   wsID,
		Email:         req.Email,
		Role:          req.Role,
		SendEmail:     req.SendEmail,
		InviteBaseURL: h.cfg.Server.DashboardURL,
	})
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}
	utils.RespondOK(c, invite)
}

func (h *WorkspaceHandler) ListInvites(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusUnprocessableEntity, "INVALID_INPUT", "Invalid workspace id")
		return
	}

	invites, err := h.service.ListInvites(c.Request.Context(), wsID)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}
	utils.RespondOK(c, invites)
}

func (h *WorkspaceHandler) ResendInvite(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid workspace id")
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	invite, err := h.service.ResendInvite(c.Request.Context(), wsID, req.Email, h.cfg.Server.DashboardURL)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}
	utils.RespondOK(c, invite)
}

func (h *WorkspaceHandler) RevokeInvite(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid workspace id")
		return
	}

	email := c.Param("email")
	if email == "" {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Email is required")
		return
	}

	if err := h.service.RevokeInvite(c.Request.Context(), wsID, email); err != nil {
		utils.RespondMapped(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
