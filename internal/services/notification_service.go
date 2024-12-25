package services

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// FCMNotificationService handles sending notifications using Firebase Cloud Messaging (FCM).
// It maintains an FCM client and context for handling communication with the FCM API.
type FCMNotificationService struct {
	client *messaging.Client
	ctx    context.Context
}

// NewNotificationService initializes a new NotificationService with a Firebase messaging client and context.
// Returns an instance of NotificationService or an error if the initialization fails.
func NewNotificationService(credentialFile string) (*FCMNotificationService, error) {
	ctx := context.Background()
	opt := option.WithCredentialsFile(credentialFile)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting messaging client: %v", err)
	}

	return &FCMNotificationService{
		client: client,
		ctx:    ctx,
	}, nil
}

// Message represents a structure for group messages, containing sender details, content, group ID, and timestamp.
type Message struct {
	SenderID   string `json:"senderId"`
	SenderName string `json:"senderName"`
	Content    string `json:"content"`
	GroupID    string `json:"groupId"`
	Timestamp  int64  `json:"timestamp"`
}

// BatchResponse represents the result of sending a batch of messages, including success and failure counts and invalid tokens.
type BatchResponse struct {
	SuccessCount  int
	FailureCount  int
	InvalidTokens []string
}

// SendGroupMessage sends a group notification with message content to the provided device tokens in batches.
// It returns a BatchResponse containing the count of successes, failures, and invalid tokens or an error if one occurs.
func (s *FCMNotificationService) SendGroupMessage(message Message, deviceTokens []string) (*BatchResponse, error) {
	// Create the notification payload
	notification := &messaging.Notification{
		Title: fmt.Sprintf("New message from %s", message.SenderName),
		Body:  message.Content,
	}

	// Add data payload for additional message details
	data := map[string]string{
		"groupId":    message.GroupID,
		"senderId":   message.SenderID,
		"senderName": message.SenderName,
		"timestamp":  fmt.Sprintf("%d", message.Timestamp),
		"type":       "group_message",
	}

	badgeNumber := 1

	batchSize := 500
	response := &BatchResponse{
		InvalidTokens: make([]string, 0),
	}

	// Process tokens in batches
	for i := 0; i < len(deviceTokens); i += batchSize {
		end := i + batchSize
		if end > len(deviceTokens) {
			end = len(deviceTokens)
		}

		batch := deviceTokens[i:end]
		batchMessage := &messaging.MulticastMessage{
			Tokens:       batch,
			Notification: notification,
			Data:         data,
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					ClickAction: "OPEN_GROUP_CHAT",
					ChannelID:   "support_group_messages",
				},
			},
			APNS: &messaging.APNSConfig{
				Headers: map[string]string{
					"apns-priority": "10",
				},
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Alert: &messaging.ApsAlert{
							Title: notification.Title,
							Body:  notification.Body,
						},
						Sound: "default",
						Badge: &badgeNumber,
					},
				},
			},
		}

		// Send the batch
		batchResponse, err := s.client.SendEachForMulticast(s.ctx, batchMessage)
		if err != nil {
			return response, fmt.Errorf("error sending batch: %v", err)
		}

		// Update response counts
		response.SuccessCount += batchResponse.SuccessCount
		response.FailureCount += batchResponse.FailureCount

		// Check for invalid tokens
		for idx, resp := range batchResponse.Responses {
			if !resp.Success {
				if resp.Error != nil && resp.Error.Error() == "registration-token-not-registered" {
					response.InvalidTokens = append(response.InvalidTokens, batch[idx])
					log.Printf("Invalid token found: %s", batch[idx])
				}
			}
		}
	}

	log.Printf("Message sending complete. Success: %d, Failure: %d, Invalid Tokens: %d",
		response.SuccessCount, response.FailureCount, len(response.InvalidTokens))

	return response, nil
}
