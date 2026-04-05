package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/resoul/studio.go.api/internal/domain"
	"gorm.io/gorm"
)

type workspaceRepository struct {
	db *gorm.DB
}

func NewWorkspaceRepository(db *gorm.DB) domain.WorkspaceRepository {
	return &workspaceRepository{db: db}
}

func (r *workspaceRepository) Create(ctx context.Context, ws *domain.Workspace) error {
	return r.db.WithContext(ctx).Create(ws).Error
}

func (r *workspaceRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Workspace, error) {
	var ws domain.Workspace
	if err := r.db.WithContext(ctx).First(&ws, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &ws, nil
}

func (r *workspaceRepository) FindBySlug(ctx context.Context, slug string) (*domain.Workspace, error) {
	var ws domain.Workspace
	if err := r.db.WithContext(ctx).First(&ws, "slug = ?", slug).Error; err != nil {
		return nil, err
	}
	return &ws, nil
}

func (r *workspaceRepository) ListForUser(ctx context.Context, userID string) ([]domain.Workspace, error) {
	var workspaces []domain.Workspace
	err := r.db.WithContext(ctx).
		Table("workspaces").
		Joins("JOIN workspace_members ON workspace_members.workspace_id = workspaces.id").
		Where("workspace_members.user_id = ?", userID).
		Find(&workspaces).Error
	return workspaces, err
}

func (r *workspaceRepository) AddMember(ctx context.Context, member *domain.WorkspaceMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *workspaceRepository) GetMember(ctx context.Context, workspaceID uuid.UUID, userID string) (*domain.WorkspaceMember, error) {
	var member domain.WorkspaceMember
	if err := r.db.WithContext(ctx).First(&member, "workspace_id = ? AND user_id = ?", workspaceID, userID).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *workspaceRepository) CountMembers(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.WorkspaceMember{}).Where("workspace_id = ?", workspaceID).Count(&count).Error
	return count, err
}

func (r *workspaceRepository) CreateInvite(ctx context.Context, invite *domain.WorkspaceInvite) error {
	return r.db.WithContext(ctx).Create(invite).Error
}

func (r *workspaceRepository) GetInvite(ctx context.Context, token string) (*domain.WorkspaceInvite, error) {
	var invite domain.WorkspaceInvite
	if err := r.db.WithContext(ctx).Preload("Workspace").First(&invite, "token = ?", token).Error; err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *workspaceRepository) DeleteInvite(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Delete(&domain.WorkspaceInvite{}, "token = ?", token).Error
}

func (r *workspaceRepository) SetCurrentWorkspace(ctx context.Context, config *domain.UserWorkspaceConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *workspaceRepository) GetCurrentWorkspace(ctx context.Context, userID string) (*domain.UserWorkspaceConfig, error) {
	var config domain.UserWorkspaceConfig
	if err := r.db.WithContext(ctx).First(&config, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *workspaceRepository) Update(ctx context.Context, ws *domain.Workspace) error {
	return r.db.WithContext(ctx).Save(ws).Error
}
