package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ory "github.com/ory/client-go"
	"github.com/resoul/studio.go.api/internal/domain"
	"github.com/resoul/studio.go.api/internal/transport/http/utils"
)

type ProfileHandler struct {
	profileService   domain.ProfileService
	workspaceService domain.WorkspaceService
}

type userMeResponse struct {
	Identity         *ory.Identity                   `json:"identity"`
	Profile          *domain.Profile                 `json:"profile"`
	Workspaces       []domain.Workspace              `json:"workspaces"`
	Onboarded        bool                            `json:"onboarded"`
	ProfileCompleted bool                            `json:"profile_completed"`
	HasWorkspaces    bool                            `json:"has_workspaces"`
	PendingInvites   []domain.PendingWorkspaceInvite `json:"pending_invites"`
}

func NewProfileHandler(profileService domain.ProfileService, workspaceService domain.WorkspaceService) *ProfileHandler {
	return &ProfileHandler{
		profileService:   profileService,
		workspaceService: workspaceService,
	}
}

func (h *ProfileHandler) GetMe(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "SNAKE_CASE_UNAUTHORIZED", "User not found in context")
		return
	}

	oryIdentity, ok := identity.(*ory.Identity)
	if !ok {
		utils.RespondError(c, http.StatusInternalServerError, "SNAKE_CASE_INTERNAL_ERROR", "Invalid identity type in context")
		return
	}

	profile, err := h.profileService.GetProfile(c.Request.Context(), oryIdentity.Id)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "SNAKE_CASE_INTERNAL_ERROR", err.Error())
		return
	}

	if profile == nil {
		utils.RespondError(c, http.StatusInternalServerError, "SNAKE_CASE_INTERNAL_ERROR", "Profile not found")
		return
	}

	workspaces, _ := h.workspaceService.ListForUser(c.Request.Context(), oryIdentity.Id)
	pendingInvites, err := h.workspaceService.ListPendingInvitesForUser(c.Request.Context(), oryIdentity.Id)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "SNAKE_CASE_INTERNAL_ERROR", err.Error())
		return
	}

	hasWorkspaces := len(workspaces) > 0
	profileCompleted := profile.Completed

	utils.RespondOK(c, userMeResponse{
		Identity:         oryIdentity,
		Profile:          profile,
		Workspaces:       workspaces,
		Onboarded:        profileCompleted && hasWorkspaces,
		ProfileCompleted: profileCompleted,
		HasWorkspaces:    hasWorkspaces,
		PendingInvites:   pendingInvites,
	})
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	identity, _ := c.Get("user")
	oryIdentity := identity.(*ory.Identity)

	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")

	if firstName == "" || lastName == "" {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_INPUT", "first_name and last_name are required")
		return
	}

	input := domain.UpdateProfileInput{
		FirstName: firstName,
		LastName:  lastName,
	}

	file, header, err := c.Request.FormFile("avatar")
	if err == nil {
		defer file.Close()
		input.Avatar = file
		input.AvatarSize = header.Size
		input.AvatarType = header.Header.Get("Content-Type")
	}

	profile, err := h.profileService.UpdateProfile(c.Request.Context(), oryIdentity.Id, input)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "SNAKE_CASE_INTERNAL_ERROR", err.Error())
		return
	}

	utils.RespondOK(c, profile)
}
