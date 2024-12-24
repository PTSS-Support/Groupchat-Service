package services

import (
	"context"
	"fmt"
	"time"

	"Groupchat-Service/internal/database/repositories"
	"Groupchat-Service/internal/models"
	"github.com/google/uuid"
)

type MessageService interface {
	GetMessages(ctx context.Context, groupID uuid.UUID, query models.PaginationQuery) ([]models.Message, *models.PaginationResponse, error)
	CreateMessage(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, userName string, create models.MessageCreate) (*models.Message, error)
	ToggleMessagePin(ctx context.Context, messageID uuid.UUID) (*models.Message, error)
}

type messageService struct {
	messageRepo         repositories.MessageRepository
	fcmTokenRepo        *repositories.FCMTokenRepository
	notificationService *NotificationService
}

func NewMessageService(
	messageRepo repositories.MessageRepository,
	fcmTokenRepo *repositories.FCMTokenRepository,
	notificationService *NotificationService,
) MessageService {
	return &messageService{
		messageRepo:         messageRepo,
		fcmTokenRepo:        fcmTokenRepo,
		notificationService: notificationService,
	}
}

func (s *messageService) GetMessages(ctx context.Context, groupID uuid.UUID, query models.PaginationQuery) ([]models.Message, *models.PaginationResponse, error) {
	messages, pagination, err := s.messageRepo.GetMessages(ctx, groupID, query)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting messages: %v", err)
	}
	return messages, pagination, nil
}

func (s *messageService) CreateMessage(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, userName string, create models.MessageCreate) (*models.Message, error) {
	// Create message entity
	message := &models.Message{
		ID:         uuid.New(),
		GroupID:    groupID,
		SenderID:   userID,
		SenderName: userName,
		Content:    create.Content,
		SentAt:     time.Now().UTC(),
		IsPinned:   false,
	}

	// Save to database
	if err := s.messageRepo.CreateMessage(ctx, groupID, message); err != nil {
		return nil, fmt.Errorf("error creating message: %w", err)
	}

	// Get FCM tokens for group members
	tokens, err := s.fcmTokenRepo.GetGroupMemberTokens(ctx, groupID)
	if err != nil {
		// Log error but don't fail the message creation
		fmt.Printf("Error getting FCM tokens: %v\n", err)
		return message, nil
	}

	// Send notification in a goroutine
	go func() {
		_, err := s.notificationService.SendGroupMessage(Message{
			SenderID:   message.SenderID.String(),
			SenderName: message.SenderName,
			Content:    message.Content,
			GroupID:    message.GroupID.String(),
			Timestamp:  message.SentAt.Unix(),
		}, tokens)
		if err != nil {
			fmt.Printf("Error sending notification: %v\n", err)
		}
	}()

	return message, nil
}

func (s *messageService) ToggleMessagePin(ctx context.Context, messageID uuid.UUID) (*models.Message, error) {
	message, err := s.messageRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("error getting message: %v", err)
	}

	message.IsPinned = !message.IsPinned
	message, err = s.messageRepo.ToggleMessagePin(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("error updating message pin status: %v", err)
	}

	return message, nil
}
