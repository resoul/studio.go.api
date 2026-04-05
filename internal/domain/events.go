package domain

// InviteEvent is the message contract published to RabbitMQ by WorkspaceService
// and consumed by InviteWorker.
//
// Keeping it in domain/ avoids the import cycle between service/ and worker/
// and makes the contract the single source of truth.
type InviteEvent struct {
	Token         string `json:"token"`
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	ExpiresAt     string `json:"expires_at"` // RFC3339
	InviteBaseURL string `json:"invite_base_url"`
}
