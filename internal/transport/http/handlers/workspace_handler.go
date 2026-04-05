package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ory "github.com/ory/client-go"
	"github.com/resoul/studio.go.api/internal/config"
	"github.com/resoul/studio.go.api/internal/domain"
	"github.com/resoul/studio.go.api/internal/transport/http/utils"
)

// WorkspaceHandler holds shared dependencies for all workspace sub-handlers.
// Route wiring is done in router.go; the three handler files share this struct.
type WorkspaceHandler struct {
	service domain.WorkspaceService
	cfg     *config.Config
}

func NewWorkspaceHandler(service domain.WorkspaceService, cfg *config.Config) *WorkspaceHandler {
	return &WorkspaceHandler{service: service, cfg: cfg}
}

func (h *WorkspaceHandler) Create(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	name := c.PostForm("name")
	if name == "" {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Name is required")
		return
	}

	input := domain.CreateWorkspaceInput{
		Name:        name,
		Description: c.PostForm("description"),
		OwnerID:     oryIdentity.Id,
	}
	if file, header, err := c.Request.FormFile("logo"); err == nil {
		defer file.Close()
		input.Logo = file
		input.LogoSize = header.Size
		input.LogoType = header.Header.Get("Content-Type")
	}

	ws, err := h.service.CreateWorkspace(c.Request.Context(), input)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}
	utils.RespondOK(c, ws)
}

func (h *WorkspaceHandler) List(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	workspaces, err := h.service.ListForUser(c.Request.Context(), oryIdentity.Id)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}
	utils.RespondOK(c, workspaces)
}

func (h *WorkspaceHandler) Update(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid workspace id")
		return
	}

	input := domain.UpdateWorkspaceInput{
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
	}
	if file, header, err := c.Request.FormFile("logo"); err == nil {
		defer file.Close()
		input.Logo = file
		input.LogoSize = header.Size
		input.LogoType = header.Header.Get("Content-Type")
	}

	ws, err := h.service.UpdateWorkspace(c.Request.Context(), wsID, input)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}
	utils.RespondOK(c, ws)
}

func (h *WorkspaceHandler) GetCurrent(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	ws, err := h.service.GetCurrentWorkspace(c.Request.Context(), oryIdentity.Id)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}
	utils.RespondOK(c, ws)
}

func (h *WorkspaceHandler) SetCurrent(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	wsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid workspace id")
		return
	}

	if err = h.service.SetCurrentWorkspace(c.Request.Context(), oryIdentity.Id, wsID); err != nil {
		utils.RespondMapped(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *WorkspaceHandler) GetConfig(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	cfg, err := h.service.GetCurrentConfig(c.Request.Context(), oryIdentity.Id)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}
	utils.RespondOK(c, cfg)
}

func (h *WorkspaceHandler) UpdateConfig(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	wsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid workspace id")
		return
	}

	var req struct {
		Language string `json:"language"`
		Theme    string `json:"theme"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	if err = h.service.UpdateConfig(c.Request.Context(), oryIdentity.Id, wsID, req.Language, req.Theme); err != nil {
		utils.RespondMapped(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
