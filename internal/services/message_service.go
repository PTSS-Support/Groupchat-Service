package services

import (
	"context"
	"errors"

	"Groupchat-Service/internal/database/repository"
	"Groupchat-Service/internal/models"
	fcmclient "Groupchat-Service/pkg"
	"github.com/google/uuid"
)

type MessageService struct {
	messageRepo         repository.MessageRepository
	fcmTokenRepo        repository.FCMTokenRepository
	notificationService *NotificationService
}

func NewMessageService(
	messageRepo repository.MessageRepository,
	fcmTokenRepo repository.FCMTokenRepository,
	notificationService *NotificationService,
) *MessageService {
	return &MessageService{
		messageRepo:         messageRepo,
		fcmTokenRepo:        fcmTokenRepo,
		notificationService: notificationService,
	}
}

func (s *MessageService) SendGroupMessage(
	ctx context.Context,
	message *models.Message,
) error {
	// Validate message
	if err := s.validateMessage(message); err != nil {
		return err
	}

	// Save message to database
	if err := s.messageRepo.CreateMessage(ctx, message); err != nil {
		return err
	}

	// Fetch FCM tokens for group members
	tokens, err := s.fcmTokenRepo.GetGroupMemberTokens(ctx, message.GroupID)
	if err != nil {
		// Log error but don't stop message creation
		return nil
	}

	// Send push notifications
	go s.sendPushNotifications(tokens, message)

	return nil
}

func (s *MessageService) validateMessage(message *models.Message) error {
	if message.Content == "" {
		return errors.New("message content cannot be empty")
	}

	if message.SenderID == uuid.Nil {
		return errors.New("invalid sender ID")
	}

	// Add more validation as needed
	return nil
}

func (s *MessageService) sendPushNotifications(
	tokens []string,
	message *models.Message,
) {
	notification := &fcmclient.Notification{
		Title: message.SenderName,
		Body:  message.Content,
		Data: map[string]string{
			"message_id": message.ID.String(),
			"group_id":   message.GroupID.String(),
		},
	}

	// Send to all tokens, handling potential failures
	for _, token := range tokens {
		// Could implement retry mechanism or batch sending
		_ = s.notificationService.SendNotification(context.Background(), token, notification.Title, notification.Body, notification.Data)
	}
}

func (s *MessageService) GetGroupMessages(
	ctx context.Context,
	groupID uuid.UUID,
	opts repository.MessageQueryOptions,
) ([]models.Message, string, error) {
	// Implement message retrieval with pagination
	return s.messageRepo.GetGroupMessages(ctx, groupID, opts)
}
