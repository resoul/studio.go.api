package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ory "github.com/ory/client-go"
	"github.com/resoul/studio.go.api/internal/domain"
	"github.com/resoul/studio.go.api/internal/transport/http/utils"
)

type WorkspaceHandler struct {
	service domain.WorkspaceService
}

func NewWorkspaceHandler(service domain.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{service: service}
}

func (h *WorkspaceHandler) Create(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "SNAKE_CASE_UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	name := c.PostForm("name")
	if name == "" {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_INPUT", "Name is required")
		return
	}
	description := c.PostForm("description")

	file, header, err := c.Request.FormFile("logo")
	var logoInput domain.CreateWorkspaceInput
	logoInput.Name = name
	logoInput.Description = description
	logoInput.OwnerID = oryIdentity.Id

	if err == nil {
		defer file.Close()
		logoInput.Logo = file
		logoInput.LogoSize = header.Size
		logoInput.LogoType = header.Header.Get("Content-Type")
	}

	ws, err := h.service.CreateWorkspace(c.Request.Context(), logoInput)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "SNAKE_CASE_INTERNAL_ERROR", err.Error())
		return
	}

	utils.RespondOK(c, ws)
}

func (h *WorkspaceHandler) List(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "SNAKE_CASE_UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	workspaces, err := h.service.ListForUser(c.Request.Context(), oryIdentity.Id)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "SNAKE_CASE_INTERNAL_ERROR", err.Error())
		return
	}

	utils.RespondOK(c, workspaces)
}

type InvitePreviewResponse struct {
	ID           uuid.UUID `json:"id"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	LogoURL      string    `json:"logo_url"`
	MembersCount int64     `json:"members_count"`
}

func (h *WorkspaceHandler) GetInvitePreview(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_INPUT", "Token is required")
		return
	}

	ws, membersCount, err := h.service.PreviewInvite(c.Request.Context(), token)
	if err != nil {
		utils.RespondError(c, http.StatusNotFound, "SNAKE_CASE_NOT_FOUND", "Invite not found or expired")
		return
	}

	utils.RespondOK(c, InvitePreviewResponse{
		ID:           ws.ID,
		Slug:         ws.Slug,
		Name:         ws.Name,
		LogoURL:      ws.LogoURL,
		MembersCount: membersCount,
	})
}

func (h *WorkspaceHandler) AcceptInvite(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "SNAKE_CASE_UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	token := c.Param("token")
	if token == "" {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_INPUT", "Token is required")
		return
	}

	err := h.service.AcceptInvite(c.Request.Context(), token, oryIdentity.Id)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "SNAKE_CASE_INTERNAL_ERROR", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *WorkspaceHandler) CreateInvite(c *gin.Context) {
	wsIDStr := c.Param("id")
	wsID, err := uuid.Parse(wsIDStr)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_INPUT", "Invalid workspace id")
		return
	}

	var req struct {
		Email string               `json:"email"`
		Role  domain.WorkspaceRole `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_INPUT", err.Error())
		return
	}

	invite, err := h.service.InviteUser(c.Request.Context(), wsID, req.Email, req.Role)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "SNAKE_CASE_INTERNAL_ERROR", err.Error())
		return
	}

	utils.RespondOK(c, invite)
}

func (h *WorkspaceHandler) GetCurrent(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "SNAKE_CASE_UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	ws, err := h.service.GetCurrentWorkspace(c.Request.Context(), oryIdentity.Id)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "SNAKE_CASE_INTERNAL_ERROR", err.Error())
		return
	}

	utils.RespondOK(c, ws)
}

func (h *WorkspaceHandler) SetCurrent(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "SNAKE_CASE_UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	wsIDStr := c.Param("id")
	wsID, err := uuid.Parse(wsIDStr)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_INPUT", "Invalid workspace id")
		return
	}

	err = h.service.SetCurrentWorkspace(c.Request.Context(), oryIdentity.Id, wsID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "SNAKE_CASE_INTERNAL_ERROR", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *WorkspaceHandler) Update(c *gin.Context) {
	wsIDStr := c.Param("id")
	wsID, err := uuid.Parse(wsIDStr)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_INPUT", "Invalid workspace id")
		return
	}

	name := c.PostForm("name")
	description := c.PostForm("description")

	file, header, err := c.Request.FormFile("logo")
	var input domain.UpdateWorkspaceInput
	input.Name = name
	input.Description = description

	if err == nil {
		defer file.Close()
		input.Logo = file
		input.LogoSize = header.Size
		input.LogoType = header.Header.Get("Content-Type")
	}

	ws, err := h.service.UpdateWorkspace(c.Request.Context(), wsID, input)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "SNAKE_CASE_INTERNAL_ERROR", err.Error())
		return
	}

	utils.RespondOK(c, ws)
}
