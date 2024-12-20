package services

import (
	"context"
	"firebase.google.com/go/v4/messaging"
)

// NotificationService handles all push notification operations
type NotificationService struct {
	client *messaging.Client
}

// NewNotificationService creates a new instance of NotificationService with Firebase Messaging Client
func NewNotificationService(client *messaging.Client) *NotificationService {
	return &NotificationService{client: client}
}

// SendNotification sends a push notification to a specific device
func (s *NotificationService) SendNotification(ctx context.Context, token string, title, body string, data map[string]string) error {
	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}
	// Send message
	_, err := s.client.Send(ctx, message)
	return err
}

// SendBatchNotifications sends notifications to multiple devices
func (s *NotificationService) SendBatchNotifications(ctx context.Context, tokens []string, title, body string, data map[string]string) []error {
	var errors []error
	for _, token := range tokens {
		if err := s.SendNotification(ctx, token, title, body, data); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}
