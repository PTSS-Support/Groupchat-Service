package services

import (
	fcmclient "Groupchat-Service/pkg"
	"context"
)

// NotificationService handles all push notification operations
type NotificationService struct {
	fcmClient *fcmclient.FCMClient
}

// NewNotificationService creates a new instance of NotificationService with the provided FCM client
func NewNotificationService(fcmClient *fcmclient.FCMClient) *NotificationService {
	return &NotificationService{
		fcmClient: fcmClient,
	}
}

// SendNotification sends a push notification to a specific device
func (s *NotificationService) SendNotification(ctx context.Context, token string, title, body string, data map[string]string) error {
	notification := &fcmclient.Notification{
		Title:    title,
		Body:     body,
		Data:     data,
		Priority: "high",
	}

	return s.fcmClient.SendNotification(ctx, token, notification)
}

// SendBatchNotifications sends notifications to multiple devices
func (s *NotificationService) SendBatchNotifications(ctx context.Context, tokens []string, title, body string, data map[string]string) []error {
	notification := &fcmclient.Notification{
		Title:    title,
		Body:     body,
		Data:     data,
		Priority: "high",
	}

	return s.fcmClient.SendBatchNotifications(ctx, tokens, notification)
}
