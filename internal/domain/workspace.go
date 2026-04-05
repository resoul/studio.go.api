package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

type Workspace struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	Slug        string    `gorm:"uniqueIndex" json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	LogoURL     string    `json:"logo_url"`
	OwnerID     string    `json:"owner_id"`
	Metadata    string    `json:"metadata"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WorkspaceRole string

const (
	RoleAdmin  WorkspaceRole = "admin"
	RoleMember WorkspaceRole = "member"
)

type WorkspaceMember struct {
	WorkspaceID uuid.UUID     `gorm:"primaryKey;type:uuid" json:"workspace_id"`
	UserID      string        `gorm:"primaryKey" json:"user_id"`
	Role        WorkspaceRole `json:"role"`
	JoinedAt    time.Time     `json:"joined_at"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

type WorkspaceInvite struct {
	Token       string        `gorm:"primaryKey" json:"token"`
	WorkspaceID uuid.UUID     `gorm:"type:uuid" json:"workspace_id"`
	Email       string        `json:"email"`
	Role        WorkspaceRole `json:"role"`
	ExpiresAt   time.Time     `json:"expires_at"`
	CreatedAt   time.Time     `json:"created_at"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

type UserWorkspaceConfig struct {
	UserID             string    `gorm:"primaryKey" json:"user_id"`
	CurrentWorkspaceID uuid.UUID `gorm:"type:uuid" json:"current_workspace_id"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type WorkspaceRepository interface {
	Create(ctx context.Context, ws *Workspace) error
	FindByID(ctx context.Context, id uuid.UUID) (*Workspace, error)
	FindBySlug(ctx context.Context, slug string) (*Workspace, error)
	ListForUser(ctx context.Context, userID string) ([]Workspace, error)

	AddMember(ctx context.Context, member *WorkspaceMember) error
	GetMember(ctx context.Context, workspaceID uuid.UUID, userID string) (*WorkspaceMember, error)
	CountMembers(ctx context.Context, workspaceID uuid.UUID) (int64, error)

	CreateInvite(ctx context.Context, invite *WorkspaceInvite) error
	GetInvite(ctx context.Context, token string) (*WorkspaceInvite, error)
	DeleteInvite(ctx context.Context, token string) error
	Update(ctx context.Context, ws *Workspace) error

	SetCurrentWorkspace(ctx context.Context, config *UserWorkspaceConfig) error
	GetCurrentWorkspace(ctx context.Context, userID string) (*UserWorkspaceConfig, error)
}

type CreateWorkspaceInput struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Logo        io.Reader `json:"-"`
	LogoSize    int64     `json:"-"`
	LogoType    string    `json:"-"`
	OwnerID     string    `json:"-"`
}

type UpdateWorkspaceInput struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Logo        io.Reader `json:"-"`
	LogoSize    int64     `json:"-"`
	LogoType    string    `json:"-"`
}

type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, input CreateWorkspaceInput) (*Workspace, error)
	GetWorkspace(ctx context.Context, id uuid.UUID) (*Workspace, error)
	ListForUser(ctx context.Context, userID string) ([]Workspace, error)

	InviteUser(ctx context.Context, workspaceID uuid.UUID, email string, role WorkspaceRole) (*WorkspaceInvite, error)
	PreviewInvite(ctx context.Context, token string) (*Workspace, int64, error)
	AcceptInvite(ctx context.Context, token string, userID string) error

	SetCurrentWorkspace(ctx context.Context, userID string, workspaceID uuid.UUID) error
	GetCurrentWorkspace(ctx context.Context, userID string) (*Workspace, error)
	UpdateWorkspace(ctx context.Context, id uuid.UUID, input UpdateWorkspaceInput) (*Workspace, error)
}

type Storage interface {
	Upload(ctx context.Context, bucketName, objectName string, reader io.Reader, size int64, contentType string) error
	GetPresignedURL(ctx context.Context, bucketName, objectName string, expires time.Duration) (string, error)
}
