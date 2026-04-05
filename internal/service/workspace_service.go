package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/resoul/studio.go.api/internal/domain"
)

type workspaceService struct {
	repo    domain.WorkspaceRepository
	storage domain.Storage
}

func NewWorkspaceService(repo domain.WorkspaceRepository, storage domain.Storage) domain.WorkspaceService {
	return &workspaceService{
		repo:    repo,
		storage: storage,
	}
}

func (s *workspaceService) CreateWorkspace(ctx context.Context, input domain.CreateWorkspaceInput) (*domain.Workspace, error) {
	wsID := uuid.New()
	slug := strings.ToLower(strings.ReplaceAll(input.Name, " ", "-"))

	// In production, we should check if slug exists and handle collisions
	// For now, let's keep it simple as in tmp

	var logoURL string
	if input.Logo != nil {
		objectName := fmt.Sprintf("logos/%s/%s", wsID.String(), "logo")
		err := s.storage.Upload(ctx, "workspaces", objectName, input.Logo, input.LogoSize, input.LogoType)
		if err != nil {
			return nil, fmt.Errorf("failed to upload logo: %w", err)
		}
		logoURL = objectName
	}

	ws := &domain.Workspace{
		ID:          wsID,
		Name:        input.Name,
		Slug:        slug,
		Description: input.Description,
		LogoURL:     logoURL,
		OwnerID:     input.OwnerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, ws); err != nil {
		return nil, err
	}

	// Add owner as admin member
	member := &domain.WorkspaceMember{
		WorkspaceID: wsID,
		UserID:      input.OwnerID,
		Role:        domain.RoleAdmin,
		JoinedAt:    time.Now(),
	}

	if err := s.repo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	return ws, nil
}

func (s *workspaceService) GetWorkspace(ctx context.Context, id uuid.UUID) (*domain.Workspace, error) {
	ws, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if ws.LogoURL != "" {
		presigned, err := s.storage.GetPresignedURL(ctx, "workspaces", ws.LogoURL, time.Hour)
		if err == nil {
			ws.LogoURL = presigned
		}
	}

	return ws, nil
}

func (s *workspaceService) ListForUser(ctx context.Context, userID string) ([]domain.Workspace, error) {
	workspaces, err := s.repo.ListForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range workspaces {
		if workspaces[i].LogoURL != "" {
			presigned, err := s.storage.GetPresignedURL(ctx, "workspaces", workspaces[i].LogoURL, time.Hour)
			if err == nil {
				workspaces[i].LogoURL = presigned
			}
		}
	}

	return workspaces, nil
}

func (s *workspaceService) InviteUser(ctx context.Context, workspaceID uuid.UUID, email string, role domain.WorkspaceRole) (*domain.WorkspaceInvite, error) {
	token, _ := generateRandomToken(32)
	invite := &domain.WorkspaceInvite{
		Token:       token,
		WorkspaceID: workspaceID,
		Email:       email,
		Role:        role,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour), // 7 days TTL
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		return nil, err
	}

	return invite, nil
}

func (s *workspaceService) PreviewInvite(ctx context.Context, token string) (*domain.Workspace, int64, error) {
	invite, err := s.repo.GetInvite(ctx, token)
	if err != nil {
		return nil, 0, err
	}

	if time.Now().After(invite.ExpiresAt) {
		return nil, 0, fmt.Errorf("invite expired")
	}

	ws := &invite.Workspace
	if ws.LogoURL != "" {
		presigned, err := s.storage.GetPresignedURL(ctx, "workspaces", ws.LogoURL, time.Hour)
		if err == nil {
			ws.LogoURL = presigned
		}
	}

	count, _ := s.repo.CountMembers(ctx, ws.ID)

	return ws, count, nil
}

func (s *workspaceService) AcceptInvite(ctx context.Context, token string, userID string) error {
	invite, err := s.repo.GetInvite(ctx, token)
	if err != nil {
		return err
	}

	if time.Now().After(invite.ExpiresAt) {
		return fmt.Errorf("invite expired")
	}

	member := &domain.WorkspaceMember{
		WorkspaceID: invite.WorkspaceID,
		UserID:      userID,
		Role:        invite.Role,
		JoinedAt:    time.Now(),
	}

	if err := s.repo.AddMember(ctx, member); err != nil {
		return err
	}

	// Delete invite after acceptance
	return s.repo.DeleteInvite(ctx, token)
}

func (s *workspaceService) SetCurrentWorkspace(ctx context.Context, userID string, workspaceID uuid.UUID) error {
	config := &domain.UserWorkspaceConfig{
		UserID:             userID,
		CurrentWorkspaceID: workspaceID,
		UpdatedAt:          time.Now(),
	}
	return s.repo.SetCurrentWorkspace(ctx, config)
}

func (s *workspaceService) GetCurrentWorkspace(ctx context.Context, userID string) (*domain.Workspace, error) {
	config, err := s.repo.GetCurrentWorkspace(ctx, userID)
	if err != nil {
		// Fallback: pick the first workspace the user belongs to
		workspaces, err := s.repo.ListForUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		if len(workspaces) == 0 {
			return nil, fmt.Errorf("user has no workspaces")
		}

		// Save this as the current one
		firstWS := workspaces[0]
		err = s.SetCurrentWorkspace(ctx, userID, firstWS.ID)
		if err != nil {
			return nil, err
		}
		return &firstWS, nil
	}

	return s.GetWorkspace(ctx, config.CurrentWorkspaceID)
}

func (s *workspaceService) UpdateWorkspace(ctx context.Context, id uuid.UUID, input domain.UpdateWorkspaceInput) (*domain.Workspace, error) {
	ws, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		ws.Name = input.Name
		ws.Slug = strings.ToLower(strings.ReplaceAll(input.Name, " ", "-"))
	}
	if input.Description != "" {
		ws.Description = input.Description
	}

	if input.Logo != nil {
		objectName := fmt.Sprintf("logos/%s/%s", ws.ID.String(), "logo")
		err := s.storage.Upload(ctx, "workspaces", objectName, input.Logo, input.LogoSize, input.LogoType)
		if err != nil {
			return nil, fmt.Errorf("failed to upload logo: %w", err)
		}
		ws.LogoURL = objectName
	}

	ws.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, ws); err != nil {
		return nil, err
	}

	return s.GetWorkspace(ctx, ws.ID)
}

func generateRandomToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
