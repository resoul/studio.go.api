package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ChannelType string

const (
	ChannelTypePublic  ChannelType = "public"
	ChannelTypePrivate ChannelType = "private"
)

type Channel struct {
	ID          uuid.UUID   `gorm:"primaryKey;type:uuid" json:"id"`
	WorkspaceID uuid.UUID   `gorm:"type:uuid;index" json:"workspace_id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        ChannelType `gorm:"default:'public'" json:"type"`
	CreatedBy   string      `json:"created_by"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	UnreadCount int         `gorm:"-" json:"unread_count,omitempty"`
}

type DirectMessageConversation struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;index" json:"workspace_id"`
	User1ID     string    `gorm:"index" json:"user1_id"`
	User2ID     string    `gorm:"index" json:"user2_id"`
	Name        string    `json:"name"`
	IsGroup     bool      `gorm:"default:false" json:"is_group"`
	CreatedBy   string    `gorm:"index" json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UnreadCount int       `gorm:"-" json:"unread_count,omitempty"`
}

type Message struct {
	ID               uuid.UUID                `gorm:"primaryKey;type:uuid" json:"id"`
	ChannelID        *uuid.UUID               `gorm:"type:uuid;index" json:"channel_id,omitempty"`
	ConversationID   *uuid.UUID               `gorm:"type:uuid;index" json:"conversation_id,omitempty"`
	ParentMessageID  *uuid.UUID               `gorm:"type:uuid;index" json:"parent_message_id,omitempty"`
	SenderID         string                   `gorm:"index" json:"sender_id"`
	Content          string                   `json:"content"`
	CreatedAt        time.Time                `json:"created_at"`
	ThreadReplyCount int                      `gorm:"-" json:"thread_reply_count,omitempty"`
	Reactions        []MessageReactionSummary `gorm:"-" json:"reactions,omitempty"`
}

type ChannelMember struct {
	ChannelID  uuid.UUID `gorm:"primaryKey;type:uuid" json:"channel_id"`
	UserID     string    `gorm:"primaryKey" json:"user_id"`
	JoinedAt   time.Time `json:"joined_at"`
	LastReadAt time.Time `gorm:"default:now()" json:"last_read_at"`
}

type ConversationMember struct {
	ConversationID uuid.UUID `gorm:"primaryKey;type:uuid" json:"conversation_id"`
	UserID         string    `gorm:"primaryKey" json:"user_id"`
	JoinedAt       time.Time `gorm:"default:now()" json:"joined_at"`
	LastReadAt     time.Time `gorm:"default:now()" json:"last_read_at"`
}

type MessageReaction struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	MessageID uuid.UUID `gorm:"type:uuid;index;uniqueIndex:idx_message_user_reaction" json:"message_id"`
	UserID    string    `gorm:"index;uniqueIndex:idx_message_user_reaction" json:"user_id"`
	Emoji     string    `gorm:"size:32" json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

type MessageReactionSummary struct {
	Emoji       string `json:"emoji"`
	Count       int    `json:"count"`
	ReactedByMe bool   `json:"reacted_by_me"`
}

type ChatReactionEvent struct {
	MessageID      uuid.UUID  `json:"message_id"`
	ChannelID      *uuid.UUID `json:"channel_id,omitempty"`
	ConversationID *uuid.UUID `json:"conversation_id,omitempty"`
	UserID         string     `json:"user_id"`
	Emoji          string     `json:"emoji"`
	Action         string     `json:"action"` // "ADDED" | "REMOVED"
}

type ChatReadEvent struct {
	ChannelID      *uuid.UUID `json:"channel_id,omitempty"`
	ConversationID *uuid.UUID `json:"conversation_id,omitempty"`
	UserID         string     `json:"user_id"`
	ReadAt         time.Time  `json:"read_at"`
}

type ChatRepository interface {
	CreateChannel(ctx context.Context, channel *Channel) error
	GetChannel(ctx context.Context, id uuid.UUID) (*Channel, error)
	ListChannels(ctx context.Context, workspaceID uuid.UUID, userID string) ([]Channel, error)
	DeleteChannel(ctx context.Context, id uuid.UUID) error

	CreateConversation(ctx context.Context, conv *DirectMessageConversation) error
	GetConversation(ctx context.Context, id uuid.UUID) (*DirectMessageConversation, error)
	FindConversation(ctx context.Context, workspaceID uuid.UUID, user1ID, user2ID string) (*DirectMessageConversation, error)
	ListConversations(ctx context.Context, workspaceID uuid.UUID, userID string) ([]DirectMessageConversation, error)

	SaveMessage(ctx context.Context, msg *Message) error
	GetMessageByID(ctx context.Context, messageID uuid.UUID) (*Message, error)
	ListMessages(ctx context.Context, targetID uuid.UUID, limit, offset int) ([]Message, error)
	ListThreadMessages(ctx context.Context, parentMessageID uuid.UUID, limit, offset int) ([]Message, error)

	AddChannelMember(ctx context.Context, member *ChannelMember) error
	IsMember(ctx context.Context, channelID uuid.UUID, userID string) (bool, error)
	ListChannelMembers(ctx context.Context, channelID uuid.UUID) ([]ChannelMember, error)
	GetChannelMember(ctx context.Context, channelID uuid.UUID, userID string) (*ChannelMember, error)
	UpdateChannelMember(ctx context.Context, member *ChannelMember) error

	AddConversationMember(ctx context.Context, member *ConversationMember) error
	GetConversationMember(ctx context.Context, convID uuid.UUID, userID string) (*ConversationMember, error)
	UpdateConversationMember(ctx context.Context, member *ConversationMember) error
	CountUnreadMessages(ctx context.Context, targetID uuid.UUID, lastReadAt time.Time, userID string) (int, error)

	GetMessageReaction(ctx context.Context, messageID uuid.UUID, userID string) (*MessageReaction, error)
	AddMessageReaction(ctx context.Context, reaction *MessageReaction) error
	RemoveMessageReaction(ctx context.Context, messageID uuid.UUID, userID string) error
	ListMessageReactions(ctx context.Context, messageIDs []uuid.UUID) ([]MessageReaction, error)
	CountThreadReplies(ctx context.Context, messageIDs []uuid.UUID) (map[uuid.UUID]int, error)
}

type ChatService interface {
	CreateChannel(ctx context.Context, workspaceID uuid.UUID, name, description string, isPrivate bool, creatorID string, participants []string) (*Channel, error)
	CreateGroupConversation(ctx context.Context, workspaceID uuid.UUID, creatorID, name string, participantIDs []string) (*DirectMessageConversation, error)
	ListChannels(ctx context.Context, workspaceID uuid.UUID, userID string) ([]Channel, error)
	GetChannelMessages(ctx context.Context, userID string, channelID uuid.UUID, limit, offset int) ([]Message, error)
	GetThreadMessages(ctx context.Context, userID string, parentMessageID uuid.UUID, limit, offset int) ([]Message, error)

	GetOrCreateConversation(ctx context.Context, workspaceID uuid.UUID, user1ID, user2ID string) (*DirectMessageConversation, error)
	ListConversations(ctx context.Context, workspaceID uuid.UUID, userID string) ([]DirectMessageConversation, error)
	GetConversationMessages(ctx context.Context, userID string, convID uuid.UUID, limit, offset int) ([]Message, error)

	SendMessage(ctx context.Context, senderID string, targetID uuid.UUID, content string, isChannel bool, parentMessageID *uuid.UUID) (*Message, error)
	MarkAsRead(ctx context.Context, userID string, targetID uuid.UUID, isChannel bool) error
	ToggleReaction(ctx context.Context, userID string, messageID uuid.UUID, emoji string) ([]MessageReactionSummary, error)
}
