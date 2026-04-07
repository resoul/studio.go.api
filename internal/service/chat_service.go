package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/resoul/studio.go.api/internal/domain"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type chatService struct {
	repo domain.ChatRepository
	hub  domain.PresenceHub
}

func NewChatService(repo domain.ChatRepository, hub domain.PresenceHub) domain.ChatService {
	return &chatService{
		repo: repo,
		hub:  hub,
	}
}

func (s *chatService) CreateChannel(ctx context.Context, workspaceID uuid.UUID, name, description string, isPrivate bool, creatorID string, participants []string) (*domain.Channel, error) {
	logrus.WithFields(logrus.Fields{
		"workspace_id": workspaceID,
		"name":         name,
		"creator_id":   creatorID,
	}).Info("Creating new channel")

	channel := &domain.Channel{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        name,
		Description: description,
		Type:        domain.ChannelTypePublic,
		CreatedBy:   creatorID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if isPrivate {
		channel.Type = domain.ChannelTypePrivate
	}

	if err := s.repo.CreateChannel(ctx, channel); err != nil {
		return nil, err
	}

	// Add creator as member
	_ = s.repo.AddChannelMember(ctx, &domain.ChannelMember{
		ChannelID:  channel.ID,
		UserID:     creatorID,
		JoinedAt:   time.Now(),
		LastReadAt: time.Now(),
	})

	// Add other participants
	for _, pID := range participants {
		if pID != creatorID && pID != "" {
			_ = s.repo.AddChannelMember(ctx, &domain.ChannelMember{
				ChannelID:  channel.ID,
				UserID:     pID,
				JoinedAt:   time.Now(),
				LastReadAt: time.Now(),
			})
		}
	}

	return channel, nil
}

func (s *chatService) ListChannels(ctx context.Context, workspaceID uuid.UUID, userID string) ([]domain.Channel, error) {
	channels, err := s.repo.ListChannels(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}

	if len(channels) == 0 {
		// If no channels exist for this user, check if any public channels exist in the workspace
		publicChannels, _ := s.repo.ListChannels(ctx, workspaceID, "") // Empty userID only gets public channels
		if len(publicChannels) == 0 {
			// Create default #general channel
			general, err := s.CreateChannel(ctx, workspaceID, "general", "Default channel for everyone", false, "system", nil)
			if err == nil {
				channels = append(channels, *general)
			}
		} else {
			// User is not a member of any private channel and there are public ones they should see?
			channels = publicChannels
		}
	}

	// Populate unread count
	for i := range channels {
		if member, err := s.repo.GetChannelMember(ctx, channels[i].ID, userID); err == nil {
			count, _ := s.repo.CountUnreadMessages(ctx, channels[i].ID, member.LastReadAt, userID)
			channels[i].UnreadCount = count
		} else {
			// If not a member (public channel not joined), it has 0 unread for this user
			channels[i].UnreadCount = 0
		}
	}

	return channels, nil
}

func (s *chatService) GetOrCreateConversation(ctx context.Context, workspaceID uuid.UUID, user1ID, user2ID string) (*domain.DirectMessageConversation, error) {
	conv, err := s.repo.FindConversation(ctx, workspaceID, user1ID, user2ID)
	if err == nil {
		return conv, nil
	}

	conv = &domain.DirectMessageConversation{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		User1ID:     user1ID,
		User2ID:     user2ID,
		Name:        "",
		IsGroup:     false,
		CreatedBy:   user1ID,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}

	// Create conversation members
	_ = s.repo.AddConversationMember(ctx, &domain.ConversationMember{
		ConversationID: conv.ID,
		UserID:         user1ID,
		JoinedAt:       time.Now(),
		LastReadAt:     time.Now(),
	})
	if user1ID != user2ID {
		_ = s.repo.AddConversationMember(ctx, &domain.ConversationMember{
			ConversationID: conv.ID,
			UserID:         user2ID,
			JoinedAt:       time.Now(),
			LastReadAt:     time.Now(),
		})
	}

	return conv, nil
}

func (s *chatService) ListConversations(ctx context.Context, workspaceID uuid.UUID, userID string) ([]domain.DirectMessageConversation, error) {
	convs, err := s.repo.ListConversations(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}

	// Populate unread count
	for i := range convs {
		if member, err := s.repo.GetConversationMember(ctx, convs[i].ID, userID); err == nil {
			count, _ := s.repo.CountUnreadMessages(ctx, convs[i].ID, member.LastReadAt, userID)
			convs[i].UnreadCount = count
		}
	}

	return convs, nil
}

func (s *chatService) CreateGroupConversation(ctx context.Context, workspaceID uuid.UUID, creatorID, name string, participantIDs []string) (*domain.DirectMessageConversation, error) {
	conv := &domain.DirectMessageConversation{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		User1ID:     creatorID,
		User2ID:     "",
		Name:        name,
		IsGroup:     true,
		CreatedBy:   creatorID,
		CreatedAt:   time.Now(),
	}
	if err := s.repo.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}

	uniqueParticipants := map[string]struct{}{
		creatorID: {},
	}
	for _, userID := range participantIDs {
		if userID == "" {
			continue
		}
		uniqueParticipants[userID] = struct{}{}
	}

	for userID := range uniqueParticipants {
		_ = s.repo.AddConversationMember(ctx, &domain.ConversationMember{
			ConversationID: conv.ID,
			UserID:         userID,
			JoinedAt:       time.Now(),
			LastReadAt:     time.Now(),
		})
	}

	return conv, nil
}

func (s *chatService) GetConversationMessages(ctx context.Context, userID string, convID uuid.UUID, limit, offset int) ([]domain.Message, error) {
	messages, err := s.repo.ListMessages(ctx, convID, limit, offset)
	if err != nil {
		return nil, err
	}
	s.enrichMessages(ctx, userID, messages)
	return messages, nil
}

func (s *chatService) GetChannelMessages(ctx context.Context, userID string, channelID uuid.UUID, limit, offset int) ([]domain.Message, error) {
	return s.GetConversationMessages(ctx, userID, channelID, limit, offset)
}

func (s *chatService) GetThreadMessages(ctx context.Context, userID string, parentMessageID uuid.UUID, limit, offset int) ([]domain.Message, error) {
	messages, err := s.repo.ListThreadMessages(ctx, parentMessageID, limit, offset)
	if err != nil {
		return nil, err
	}
	s.enrichMessages(ctx, userID, messages)
	return messages, nil
}

func (s *chatService) SendMessage(ctx context.Context, senderID string, targetID uuid.UUID, content string, isChannel bool, parentMessageID *uuid.UUID) (*domain.Message, error) {
	msg := &domain.Message{
		ID:              uuid.New(),
		SenderID:        senderID,
		Content:         content,
		ParentMessageID: parentMessageID,
		CreatedAt:       time.Now(),
	}

	if isChannel {
		msg.ChannelID = &targetID
	} else {
		msg.ConversationID = &targetID
	}

	if err := s.repo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	// Broadcast message via Hub
	logrus.WithFields(logrus.Fields{
		"message_id": msg.ID,
		"sender_id":  senderID,
		"target_id":  targetID,
		"is_channel": isChannel,
	}).Info("Broadcasting chat message via Hub")
	s.hub.Broadcast(ctx, msg)

	return msg, nil
}

func (s *chatService) MarkAsRead(ctx context.Context, userID string, targetID uuid.UUID, isChannel bool) error {
	readAt := time.Now()

	if isChannel {
		member, err := s.repo.GetChannelMember(ctx, targetID, userID)
		if err != nil {
			// If user is not a member of a public channel, add them as they read it?
			// Actually let's just create the membership record if it doesn't exist for public channels
			ch, err := s.repo.GetChannel(ctx, targetID)
			if err != nil {
				return err
			}
			if ch.Type == domain.ChannelTypePublic {
				member = &domain.ChannelMember{
					ChannelID:  targetID,
					UserID:     userID,
					JoinedAt:   time.Now(),
					LastReadAt: readAt,
				}
				if addErr := s.repo.AddChannelMember(ctx, member); addErr != nil {
					return addErr
				}
				s.hub.Broadcast(ctx, domain.WebsocketEvent{
					Type: domain.WebsocketEventChatRead,
					Payload: domain.ChatReadEvent{
						ChannelID: &targetID,
						UserID:    userID,
						ReadAt:    readAt,
					},
				})
				return nil
			}
			return err
		}
		member.LastReadAt = readAt
		if updateErr := s.repo.UpdateChannelMember(ctx, member); updateErr != nil {
			return updateErr
		}
		s.hub.Broadcast(ctx, domain.WebsocketEvent{
			Type: domain.WebsocketEventChatRead,
			Payload: domain.ChatReadEvent{
				ChannelID: &targetID,
				UserID:    userID,
				ReadAt:    readAt,
			},
		})
		return nil
	}

	member, err := s.repo.GetConversationMember(ctx, targetID, userID)
	if err != nil {
		// This shouldn't happen if the conversation exists, but just in case
		return err
	}
	member.LastReadAt = readAt
	if updateErr := s.repo.UpdateConversationMember(ctx, member); updateErr != nil {
		return updateErr
	}

	s.hub.Broadcast(ctx, domain.WebsocketEvent{
		Type: domain.WebsocketEventChatRead,
		Payload: domain.ChatReadEvent{
			ConversationID: &targetID,
			UserID:         userID,
			ReadAt:         readAt,
		},
	})

	return nil
}

func (s *chatService) ToggleReaction(ctx context.Context, userID string, messageID uuid.UUID, emoji string) ([]domain.MessageReactionSummary, error) {
	message, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	existingReaction, err := s.repo.GetMessageReaction(ctx, messageID, userID)
	if err == nil {
		// Toggle off if the same emoji was selected again.
		if existingReaction.Emoji == emoji {
			if removeErr := s.repo.RemoveMessageReaction(ctx, messageID, userID); removeErr != nil {
				return nil, removeErr
			}
			s.hub.Broadcast(ctx, domain.WebsocketEvent{
				Type: domain.WebsocketEventChatReaction,
				Payload: domain.ChatReactionEvent{
					MessageID:      messageID,
					ChannelID:      message.ChannelID,
					ConversationID: message.ConversationID,
					UserID:         userID,
					Emoji:          emoji,
					Action:         "REMOVED",
				},
			})
		} else {
			// Swap reaction: remove previous emoji then add the new one.
			previousEmoji := existingReaction.Emoji
			if removeErr := s.repo.RemoveMessageReaction(ctx, messageID, userID); removeErr != nil {
				return nil, removeErr
			}
			if addErr := s.repo.AddMessageReaction(ctx, &domain.MessageReaction{
				ID:        uuid.New(),
				MessageID: messageID,
				UserID:    userID,
				Emoji:     emoji,
				CreatedAt: time.Now(),
			}); addErr != nil {
				return nil, addErr
			}

			s.hub.Broadcast(ctx, domain.WebsocketEvent{
				Type: domain.WebsocketEventChatReaction,
				Payload: domain.ChatReactionEvent{
					MessageID:      messageID,
					ChannelID:      message.ChannelID,
					ConversationID: message.ConversationID,
					UserID:         userID,
					Emoji:          previousEmoji,
					Action:         "REMOVED",
				},
			})
			s.hub.Broadcast(ctx, domain.WebsocketEvent{
				Type: domain.WebsocketEventChatReaction,
				Payload: domain.ChatReactionEvent{
					MessageID:      messageID,
					ChannelID:      message.ChannelID,
					ConversationID: message.ConversationID,
					UserID:         userID,
					Emoji:          emoji,
					Action:         "ADDED",
				},
			})
		}
	} else {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		if addErr := s.repo.AddMessageReaction(ctx, &domain.MessageReaction{
			ID:        uuid.New(),
			MessageID: messageID,
			UserID:    userID,
			Emoji:     emoji,
			CreatedAt: time.Now(),
		}); addErr != nil {
			return nil, addErr
		}
		s.hub.Broadcast(ctx, domain.WebsocketEvent{
			Type: domain.WebsocketEventChatReaction,
			Payload: domain.ChatReactionEvent{
				MessageID:      messageID,
				ChannelID:      message.ChannelID,
				ConversationID: message.ConversationID,
				UserID:         userID,
				Emoji:          emoji,
				Action:         "ADDED",
			},
		})
	}

	reactionSummaries, buildErr := s.buildReactionSummaries(ctx, messageID, userID)
	if buildErr != nil {
		return nil, buildErr
	}

	return reactionSummaries, nil
}

func (s *chatService) enrichMessages(ctx context.Context, userID string, messages []domain.Message) {
	if len(messages) == 0 {
		return
	}

	messageIDs := make([]uuid.UUID, 0, len(messages))
	for _, msg := range messages {
		messageIDs = append(messageIDs, msg.ID)
	}

	reactions, _ := s.repo.ListMessageReactions(ctx, messageIDs)
	replyCounts, _ := s.repo.CountThreadReplies(ctx, messageIDs)

	type reactionCounter struct {
		count       int
		reactedByMe bool
	}

	reactionMap := make(map[uuid.UUID]map[string]reactionCounter)
	for _, reaction := range reactions {
		if _, ok := reactionMap[reaction.MessageID]; !ok {
			reactionMap[reaction.MessageID] = make(map[string]reactionCounter)
		}
		current := reactionMap[reaction.MessageID][reaction.Emoji]
		current.count++
		if reaction.UserID == userID {
			current.reactedByMe = true
		}
		reactionMap[reaction.MessageID][reaction.Emoji] = current
	}

	for i := range messages {
		messages[i].ThreadReplyCount = replyCounts[messages[i].ID]
		perMessageReactions := reactionMap[messages[i].ID]
		if len(perMessageReactions) == 0 {
			messages[i].Reactions = []domain.MessageReactionSummary{}
			continue
		}

		summary := make([]domain.MessageReactionSummary, 0, len(perMessageReactions))
		for emoji, counter := range perMessageReactions {
			summary = append(summary, domain.MessageReactionSummary{
				Emoji:       emoji,
				Count:       counter.count,
				ReactedByMe: counter.reactedByMe,
			})
		}
		messages[i].Reactions = summary
	}
}

func (s *chatService) buildReactionSummaries(ctx context.Context, messageID uuid.UUID, userID string) ([]domain.MessageReactionSummary, error) {
	reactions, err := s.repo.ListMessageReactions(ctx, []uuid.UUID{messageID})
	if err != nil {
		return nil, err
	}

	type reactionCounter struct {
		count       int
		reactedByMe bool
	}

	perEmoji := make(map[string]reactionCounter)
	for _, reaction := range reactions {
		if reaction.MessageID != messageID {
			continue
		}
		current := perEmoji[reaction.Emoji]
		current.count++
		if reaction.UserID == userID {
			current.reactedByMe = true
		}
		perEmoji[reaction.Emoji] = current
	}

	result := make([]domain.MessageReactionSummary, 0, len(perEmoji))
	for emoji, counter := range perEmoji {
		result = append(result, domain.MessageReactionSummary{
			Emoji:       emoji,
			Count:       counter.count,
			ReactedByMe: counter.reactedByMe,
		})
	}
	return result, nil
}
