package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/resoul/studio.go.api/internal/domain"
	"gorm.io/gorm"
)

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) domain.ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) CreateChannel(ctx context.Context, channel *domain.Channel) error {
	return r.db.WithContext(ctx).Create(channel).Error
}

func (r *chatRepository) GetChannel(ctx context.Context, id uuid.UUID) (*domain.Channel, error) {
	var channel domain.Channel
	if err := r.db.WithContext(ctx).First(&channel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &channel, nil
}

func (r *chatRepository) ListChannels(ctx context.Context, workspaceID uuid.UUID, userID string) ([]domain.Channel, error) {
	var channels []domain.Channel
	// List all public channels in workspace OR private channels where user is a member
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND (type = ? OR id IN (SELECT channel_id FROM channel_members WHERE user_id = ?))",
			workspaceID, domain.ChannelTypePublic, userID).
		Find(&channels).Error
	if err != nil {
		return nil, err
	}
	return channels, nil
}

func (r *chatRepository) DeleteChannel(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Channel{}, "id = ?", id).Error
}

func (r *chatRepository) CreateConversation(ctx context.Context, conv *domain.DirectMessageConversation) error {
	return r.db.WithContext(ctx).Create(conv).Error
}

func (r *chatRepository) GetConversation(ctx context.Context, id uuid.UUID) (*domain.DirectMessageConversation, error) {
	var conv domain.DirectMessageConversation
	if err := r.db.WithContext(ctx).First(&conv, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *chatRepository) FindConversation(ctx context.Context, workspaceID uuid.UUID, user1ID, user2ID string) (*domain.DirectMessageConversation, error) {
	var conv domain.DirectMessageConversation
	// Search for conversation where both users are participants
	err := r.db.WithContext(ctx).Where("workspace_id = ? AND ((user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?))",
		workspaceID, user1ID, user2ID, user2ID, user1ID).First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *chatRepository) ListConversations(ctx context.Context, workspaceID uuid.UUID, userID string) ([]domain.DirectMessageConversation, error) {
	var convs []domain.DirectMessageConversation
	if err := r.db.WithContext(ctx).
		Table("direct_message_conversations").
		Joins("JOIN conversation_members ON conversation_members.conversation_id = direct_message_conversations.id").
		Where("direct_message_conversations.workspace_id = ? AND conversation_members.user_id = ?", workspaceID, userID).
		Find(&convs).Error; err != nil {
		return nil, err
	}
	return convs, nil
}

func (r *chatRepository) SaveMessage(ctx context.Context, msg *domain.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *chatRepository) GetMessageByID(ctx context.Context, messageID uuid.UUID) (*domain.Message, error) {
	var msg domain.Message
	if err := r.db.WithContext(ctx).First(&msg, "id = ?", messageID).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *chatRepository) ListMessages(ctx context.Context, targetID uuid.UUID, limit, offset int) ([]domain.Message, error) {
	var messages []domain.Message
	err := r.db.WithContext(ctx).
		Where("(channel_id = ? OR conversation_id = ?) AND parent_message_id IS NULL", targetID, targetID).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *chatRepository) ListThreadMessages(ctx context.Context, parentMessageID uuid.UUID, limit, offset int) ([]domain.Message, error) {
	var messages []domain.Message
	err := r.db.WithContext(ctx).
		Where("parent_message_id = ?", parentMessageID).
		Order("created_at asc").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *chatRepository) AddChannelMember(ctx context.Context, member *domain.ChannelMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *chatRepository) IsMember(ctx context.Context, channelID uuid.UUID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.ChannelMember{}).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *chatRepository) ListChannelMembers(ctx context.Context, channelID uuid.UUID) ([]domain.ChannelMember, error) {
	var members []domain.ChannelMember
	err := r.db.WithContext(ctx).Where("channel_id = ?", channelID).Find(&members).Error
	return members, err
}

func (r *chatRepository) GetChannelMember(ctx context.Context, channelID uuid.UUID, userID string) (*domain.ChannelMember, error) {
	var member domain.ChannelMember
	if err := r.db.WithContext(ctx).First(&member, "channel_id = ? AND user_id = ?", channelID, userID).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *chatRepository) UpdateChannelMember(ctx context.Context, member *domain.ChannelMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

func (r *chatRepository) AddConversationMember(ctx context.Context, member *domain.ConversationMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *chatRepository) GetConversationMember(ctx context.Context, convID uuid.UUID, userID string) (*domain.ConversationMember, error) {
	var member domain.ConversationMember
	if err := r.db.WithContext(ctx).First(&member, "conversation_id = ? AND user_id = ?", convID, userID).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *chatRepository) UpdateConversationMember(ctx context.Context, member *domain.ConversationMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

func (r *chatRepository) CountUnreadMessages(ctx context.Context, targetID uuid.UUID, lastReadAt time.Time, userID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Message{}).
		Where("(channel_id = ? OR conversation_id = ?) AND parent_message_id IS NULL AND created_at > ? AND sender_id <> ?", targetID, targetID, lastReadAt, userID).
		Count(&count).Error
	return int(count), err
}

func (r *chatRepository) GetMessageReaction(ctx context.Context, messageID uuid.UUID, userID string) (*domain.MessageReaction, error) {
	var reaction domain.MessageReaction
	if err := r.db.WithContext(ctx).First(&reaction, "message_id = ? AND user_id = ?", messageID, userID).Error; err != nil {
		return nil, err
	}
	return &reaction, nil
}

func (r *chatRepository) AddMessageReaction(ctx context.Context, reaction *domain.MessageReaction) error {
	return r.db.WithContext(ctx).Create(reaction).Error
}

func (r *chatRepository) RemoveMessageReaction(ctx context.Context, messageID uuid.UUID, userID string) error {
	return r.db.WithContext(ctx).
		Delete(&domain.MessageReaction{}, "message_id = ? AND user_id = ?", messageID, userID).
		Error
}

func (r *chatRepository) ListMessageReactions(ctx context.Context, messageIDs []uuid.UUID) ([]domain.MessageReaction, error) {
	if len(messageIDs) == 0 {
		return []domain.MessageReaction{}, nil
	}

	var reactions []domain.MessageReaction
	err := r.db.WithContext(ctx).
		Where("message_id IN ?", messageIDs).
		Order("created_at asc").
		Find(&reactions).Error
	return reactions, err
}

func (r *chatRepository) CountThreadReplies(ctx context.Context, messageIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(messageIDs) == 0 {
		return map[uuid.UUID]int{}, nil
	}

	type resultRow struct {
		ParentMessageID uuid.UUID `gorm:"column:parent_message_id"`
		Count           int       `gorm:"column:count"`
	}

	var rows []resultRow
	err := r.db.WithContext(ctx).
		Model(&domain.Message{}).
		Select("parent_message_id, COUNT(*) as count").
		Where("parent_message_id IN ?", messageIDs).
		Group("parent_message_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		result[row.ParentMessageID] = row.Count
	}
	return result, nil
}
