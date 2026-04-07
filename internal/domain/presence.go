package domain

import (
	"context"
)

type PresenceEventType string

const (
	PresenceEventJoin  PresenceEventType = "JOIN"
	PresenceEventLeave PresenceEventType = "LEAVE"
	PresenceEventSync  PresenceEventType = "SYNC"
)

type PresenceEvent struct {
	Type   PresenceEventType `json:"type"`
	UserID string            `json:"user_id"`
}

type PresenceSyncEvent struct {
	Type    PresenceEventType `json:"type"`
	UserIDs []string          `json:"user_ids"`
}

type WebsocketEventType string

const (
	WebsocketEventPresence     WebsocketEventType = "PRESENCE"
	WebsocketEventChatMessage  WebsocketEventType = "CHAT_MESSAGE"
	WebsocketEventChatReaction WebsocketEventType = "CHAT_REACTION"
	WebsocketEventChatRead     WebsocketEventType = "CHAT_READ"
	WebsocketEventInvite       WebsocketEventType = "INVITE"
)

type WebsocketEvent struct {
	Type    WebsocketEventType `json:"type"`
	Payload interface{}        `json:"payload"`
}

// InviteNotificationEvent signals that a user's pending invites might have changed.
// Clients should refresh the current user payload to get latest pending_invites.
type InviteNotificationEvent struct {
	Type string `json:"type"`
}

type PresenceHub interface {
	Register(ctx context.Context, userID string, client PresenceClient)
	Unregister(ctx context.Context, userID string, client PresenceClient)
	GetOnlineUsers() []string
	Broadcast(ctx context.Context, event interface{})
}

type PresenceClient interface {
	Send(event interface{}) error
	Close() error
}
