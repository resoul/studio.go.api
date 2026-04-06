package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/resoul/studio.go.api/internal/domain"
	"github.com/resoul/studio.go.api/internal/infrastructure/rabbitmq"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const inviteQueue = "workspace.invites"

type workspaceService struct {
	repo        domain.WorkspaceRepository
	profileRepo domain.ProfileRepository
	userRepo    domain.UserRepository
	storage     domain.Storage
	rbmq        *rabbitmq.Client // optional — nil when RabbitMQ is unavailable
}

func NewWorkspaceService(
	repo domain.WorkspaceRepository,
	profileRepo domain.ProfileRepository,
	userRepo domain.UserRepository,
	storage domain.Storage,
	rbmq *rabbitmq.Client,
) domain.WorkspaceService {
	return &workspaceService{
		repo:        repo,
		profileRepo: profileRepo,
		userRepo:    userRepo,
		storage:     storage,
		rbmq:        rbmq,
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("workspace %s: %w", id, domain.ErrNotFound)
		}
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

// InviteUser persists the invite record and publishes a domain.InviteEvent to RabbitMQ.
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

func (s *workspaceService) ListInvites(ctx context.Context, workspaceID uuid.UUID) ([]domain.WorkspaceInvite, error) {
	return s.repo.ListInvites(ctx, workspaceID)
}

// publishInviteEvent enqueues the email delivery task using domain.InviteEvent.
// Failures are non-fatal — the invite record already exists in the DB.
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

	// domain.InviteEvent is the single source of truth — no local struct needed.
	payload, err := json.Marshal(domain.InviteEvent{
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

func (s *workspaceService) PreviewInvite(ctx context.Context, token string) (*domain.Workspace, int64, *domain.WorkspaceInvite, error) {
	invite, err := s.repo.GetInvite(ctx, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, nil, fmt.Errorf("invite %s: %w", token, domain.ErrNotFound)
		}
		return nil, 0, nil, err
	}
	if time.Now().After(invite.ExpiresAt) {
		return nil, 0, nil, domain.ErrInviteExpired
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
	return ws, count, invite, nil
}

func (s *workspaceService) AcceptInvite(ctx context.Context, token string, userID string) error {
	invite, err := s.repo.GetInvite(ctx, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("invite %s: %w", token, domain.ErrNotFound)
		}
		return err
	}
	if time.Now().After(invite.ExpiresAt) {
		return domain.ErrInviteExpired
	}

	// Check if already a member to make this idempotent
	_, err = s.repo.GetMember(ctx, invite.WorkspaceID, userID)
	if err == nil {
		// User is already a member, just cleanup the invite and return success
		return s.repo.DeleteInvite(ctx, token)
	}

	// Ensure profile exists for the joining user so they appear in members list
	if _, err := s.profileRepo.FindByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = s.profileRepo.Create(ctx, &domain.Profile{
				ID:        userID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}
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
			return nil, fmt.Errorf("user has no workspaces: %w", domain.ErrNotFound)
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("workspace %s: %w", id, domain.ErrNotFound)
		}
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

func (s *workspaceService) ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]domain.MemberInfo, error) {
	members, err := s.repo.ListMembers(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.MemberInfo, 0, len(members))
	for _, m := range members {
		info := domain.MemberInfo{WorkspaceMember: m}

		if profile, err := s.profileRepo.FindByID(ctx, m.UserID); err == nil {
			info.FirstName = profile.FirstName
			info.LastName = profile.LastName
			info.AvatarURL = profile.AvatarURL
		}

		if user, err := s.userRepo.GetIdentity(ctx, m.UserID); err == nil {
			info.Email = user.Email
		}

		result = append(result, info)
	}

	return result, nil
}

func (s *workspaceService) RemoveMember(ctx context.Context, workspaceID uuid.UUID, userID string) error {
	ws, err := s.repo.FindByID(ctx, workspaceID)
	if err != nil {
		return err
	}
	if ws.OwnerID == userID {
		return domain.ErrOwnerCannotBeRemoved
	}
	return s.repo.DeleteMember(ctx, workspaceID, userID)
}

func (s *workspaceService) ResendInvite(ctx context.Context, workspaceID uuid.UUID, email string, baseURL string) (*domain.WorkspaceInvite, error) {
	_ = s.repo.DeleteInviteByEmail(ctx, workspaceID, email)

	input := domain.CreateInviteInput{
		WorkspaceID:   workspaceID,
		Email:         email,
		Role:          domain.RoleMember,
		SendEmail:     true,
		InviteBaseURL: baseURL,
	}

	invites, _ := s.repo.ListInvites(ctx, workspaceID)
	for _, inv := range invites {
		if inv.Email == email {
			input.Role = inv.Role
			break
		}
	}

	return s.InviteUser(ctx, input)
}

func (s *workspaceService) RevokeInvite(ctx context.Context, workspaceID uuid.UUID, email string) error {
	return s.repo.DeleteInviteByEmail(ctx, workspaceID, email)
}

func generateRandomToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
