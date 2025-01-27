package services

import (
	"context"
	"fmt"
	"time"

	"Groupchat-Service/internal/database/repositories"
	"Groupchat-Service/internal/models"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
)

type messageService struct {
	messageRepo         repositories.MessageRepository
	fcmTokenRepo        repositories.FCMTokenRepository
	notificationService NotificationService
	validationService   ValidationService
}

func NewMessageService(
	messageRepo repositories.MessageRepository,
	fcmTokenRepo repositories.FCMTokenRepository,
	notificationService NotificationService,
	validationService ValidationService,
) MessageService {
	return &messageService{
		messageRepo:         messageRepo,
		fcmTokenRepo:        fcmTokenRepo,
		notificationService: notificationService,
		validationService:   validationService,
	}
}

func (s *messageService) GetMessages(ctx context.Context, groupID uuid.UUID, query models.PaginationQuery) ([]models.MessageResponse, *models.PaginationResponse, error) {
	// Fetch messages from the repository
	messages, pagination, err := s.messageRepo.GetMessages(ctx, groupID, query)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting messages from repository: %w", err)
	}

	// If no messages are found, return an empty list and pagination response
	if len(messages) == 0 {
		return []models.MessageResponse{}, pagination, nil
	}

	var messageResponses []models.MessageResponse
	for _, message := range messages {
		messageResponses = append(messageResponses, models.MessageResponse{
			ID:         message.ID,
			GroupID:    message.GroupID,
			SenderID:   message.SenderID,
			SenderName: message.SenderName,
			Content:    message.Content,
			SentAt:     message.SentAt,
			IsPinned:   message.IsPinned,
		})
	}

	return messageResponses, pagination, nil
}

func (s *messageService) CreateMessage(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, userName string, create models.MessageCreate) (*models.Message, error) {
	// Sanitize message content
	p := bluemonday.UGCPolicy()
	sanitizedContent := p.Sanitize(create.Content)

	// Create message entity
	message := &models.Message{
		ID:         uuid.New(),
		GroupID:    groupID,
		SenderID:   userID,
		SenderName: userName,
		Content:    sanitizedContent,
		SentAt:     time.Now().UTC(),
		IsPinned:   false,
	}

	// Save to database
	if err := s.messageRepo.CreateMessage(ctx, groupID, message); err != nil {
		return nil, fmt.Errorf("error creating message: %w", err)
	}

	// Get FCM tokens for group members asynchronously
	go func() {
		tokens, err := s.fcmTokenRepo.GetGroupMemberTokens(ctx, groupID)
		if err != nil {
			// Log error but don't fail the message creation
			fmt.Printf("Error getting FCM tokens: %v\n", err)
			return
		}

		// Send notification asynchronously
		go func() {
			_, err := s.notificationService.SendGroupMessage(Message{
				SenderID:   message.SenderID.String(),
				SenderName: userName,
				Content:    message.Content,
				GroupID:    message.GroupID.String(),
				Timestamp:  message.SentAt.Unix(),
			}, tokens)
			if err != nil {
				fmt.Printf("Error sending notification: %v\n", err)
			}
		}()
	}()

	return message, nil
}

func (s *messageService) ToggleMessagePin(ctx context.Context, groupID uuid.UUID, messageID uuid.UUID) (*models.Message, error) {
	message, err := s.messageRepo.GetMessageByID(ctx, groupID, messageID)
	if err != nil {
		return nil, fmt.Errorf("error getting message: %v", err)
	}

	message.IsPinned = !message.IsPinned
	updatedMessage, err := s.messageRepo.ToggleMessagePin(ctx, groupID, messageID)
	if err != nil {
		return nil, fmt.Errorf("error updating message pin status: %v", err)
	}

	return updatedMessage, nil
}
