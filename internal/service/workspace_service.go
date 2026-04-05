package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/resoul/studio.go.api/internal/domain"
	"github.com/resoul/studio.go.api/internal/infrastructure/rabbitmq"
	"github.com/sirupsen/logrus"
)

const inviteQueue = "workspace.invites"

type workspaceService struct {
	repo    domain.WorkspaceRepository
	storage domain.Storage
	rbmq    *rabbitmq.Client // optional — nil when RabbitMQ is unavailable
}

func NewWorkspaceService(repo domain.WorkspaceRepository, storage domain.Storage, rbmq *rabbitmq.Client) domain.WorkspaceService {
	return &workspaceService{
		repo:    repo,
		storage: storage,
		rbmq:    rbmq,
	}
}

func (s *workspaceService) CreateWorkspace(ctx context.Context, input domain.CreateWorkspaceInput) (*domain.Workspace, error) {
	wsID := uuid.New()
	slug := strings.ToLower(strings.ReplaceAll(input.Name, " ", "-"))

	var logoURL string
	if input.Logo != nil {
		objectName := fmt.Sprintf("logos/%s/%s", wsID.String(), "logo")
		if err := s.storage.Upload(ctx, "workspaces", objectName, input.Logo, input.LogoSize, input.LogoType); err != nil {
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
		if err != nil {
			logrus.WithError(err).WithField("object", ws.LogoURL).Warn("failed to presign workspace logo")
		} else {
			ws.LogoURL = fmt.Sprintf("%s?v=%d", presigned, ws.UpdatedAt.Unix())
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
			if err != nil {
				logrus.WithError(err).WithField("object", workspaces[i].LogoURL).Warn("failed to presign workspace logo")
			} else {
				workspaces[i].LogoURL = fmt.Sprintf("%s?v=%d", presigned, workspaces[i].UpdatedAt.Unix())
			}
		}
	}
	return workspaces, nil
}

// InviteUser persists the invite record and publishes an event to RabbitMQ.
// Email delivery is handled asynchronously by InviteWorker.
// If RabbitMQ is unavailable the invite is still saved — email just won't be sent.
func (s *workspaceService) InviteUser(ctx context.Context, input domain.CreateInviteInput) (*domain.WorkspaceInvite, error) {
	token, err := generateRandomToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invite token: %w", err)
	}

	invite := &domain.WorkspaceInvite{
		Token:       token,
		WorkspaceID: input.WorkspaceID,
		Email:       input.Email,
		Role:        input.Role,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		return nil, err
	}

	if input.SendEmail {
		s.publishInviteEvent(ctx, invite, input)
	}

	return invite, nil
}

// publishInviteEvent enqueues the email delivery task.
// Failures are non-fatal — the invite record already exists.
func (s *workspaceService) publishInviteEvent(ctx context.Context, invite *domain.WorkspaceInvite, input domain.CreateInviteInput) {
	if s.rbmq == nil {
		logrus.WithField("token", invite.Token).Warn("RabbitMQ unavailable, invite email will not be sent")
		return
	}

	ws, err := s.repo.FindByID(ctx, invite.WorkspaceID)
	if err != nil {
		logrus.WithError(err).WithField("workspace_id", invite.WorkspaceID).Warn("could not load workspace for invite event")
		return
	}

	// InviteEvent mirrors worker.InviteEvent — defined here to avoid an import cycle.
	// Both structs must stay in sync; consider moving to domain/ if they drift.
	type inviteEvent struct {
		Token         string `json:"token"`
		WorkspaceID   string `json:"workspace_id"`
		WorkspaceName string `json:"workspace_name"`
		Email         string `json:"email"`
		Role          string `json:"role"`
		ExpiresAt     string `json:"expires_at"`
		InviteBaseURL string `json:"invite_base_url"`
	}

	payload, err := json.Marshal(inviteEvent{
		Token:         invite.Token,
		WorkspaceID:   invite.WorkspaceID.String(),
		WorkspaceName: ws.Name,
		Email:         invite.Email,
		Role:          string(invite.Role),
		ExpiresAt:     invite.ExpiresAt.Format(time.RFC3339),
		InviteBaseURL: input.InviteBaseURL,
	})
	if err != nil {
		logrus.WithError(err).Warn("could not marshal invite event")
		return
	}

	if err := s.rbmq.Publish(ctx, "", inviteQueue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         payload,
	}); err != nil {
		logrus.WithError(err).
			WithField("token", invite.Token).
			WithField("email", invite.Email).
			Warn("invite created but event publish failed")
	}
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
		if err != nil {
			logrus.WithError(err).WithField("object", ws.LogoURL).Warn("failed to presign invite workspace logo")
		} else {
			ws.LogoURL = fmt.Sprintf("%s?v=%d", presigned, ws.UpdatedAt.Unix())
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

	return s.repo.DeleteInvite(ctx, token)
}

func (s *workspaceService) SetCurrentWorkspace(ctx context.Context, userID string, workspaceID uuid.UUID) error {
	config := &domain.UserWorkspaceConfig{
		UserID:      userID,
		WorkspaceID: workspaceID,
		UpdatedAt:   time.Now(),
	}
	return s.repo.SetCurrentWorkspace(ctx, config)
}

func (s *workspaceService) GetCurrentWorkspace(ctx context.Context, userID string) (*domain.Workspace, error) {
	config, err := s.repo.GetCurrentWorkspace(ctx, userID)
	if err != nil {
		workspaces, err := s.repo.ListForUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		if len(workspaces) == 0 {
			return nil, fmt.Errorf("user has no workspaces")
		}
		firstWS := workspaces[0]
		if err := s.SetCurrentWorkspace(ctx, userID, firstWS.ID); err != nil {
			return nil, err
		}
		return &firstWS, nil
	}
	return s.GetWorkspace(ctx, config.WorkspaceID)
}

func (s *workspaceService) UpdateConfig(ctx context.Context, userID string, workspaceID uuid.UUID, language, theme string) error {
	config, err := s.repo.GetCurrentWorkspace(ctx, userID)
	if err != nil || config.WorkspaceID != workspaceID {
		config = &domain.UserWorkspaceConfig{
			UserID:      userID,
			WorkspaceID: workspaceID,
		}
	}
	config.Language = language
	config.Theme = theme
	config.UpdatedAt = time.Now()
	return s.repo.UpdateConfig(ctx, config)
}

func (s *workspaceService) GetCurrentConfig(ctx context.Context, userID string) (*domain.UserWorkspaceConfig, error) {
	return s.repo.GetCurrentWorkspace(ctx, userID)
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
		if err := s.storage.Upload(ctx, "workspaces", objectName, input.Logo, input.LogoSize, input.LogoType); err != nil {
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
