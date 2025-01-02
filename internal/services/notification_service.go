package services

import (
	"Groupchat-Service/internal/database/repositories"
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type FCMNotificationService struct {
	client      *messaging.Client
	ctx         context.Context
	messageRepo repositories.MessageRepository
}

func NewNotificationService(credentialFile string, messageRepo repositories.MessageRepository) (*FCMNotificationService, error) {
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
		client:      client,
		ctx:         ctx,
		messageRepo: messageRepo,
	}, nil
}

type Message struct {
	SenderID   string `json:"senderId"`
	SenderName string `json:"senderName"`
	Content    string `json:"content"`
	GroupID    string `json:"groupId"`
	Timestamp  int64  `json:"timestamp"`
}

type BatchResponse struct {
	SuccessCount  int
	FailureCount  int
	InvalidTokens []string
}

func (s *FCMNotificationService) SendGroupMessage(message Message, deviceTokens []string) (*BatchResponse, error) {
	notification := &messaging.Notification{
		Title: fmt.Sprintf("New message from %s", message.SenderName),
		Body:  message.Content,
	}

	data := map[string]string{
		"groupId":    message.GroupID,
		"senderId":   message.SenderID,
		"senderName": message.SenderName,
		"timestamp":  fmt.Sprintf("%d", message.Timestamp),
		"type":       "group_message",
	}

	batchSize := 500
	response := &BatchResponse{
		InvalidTokens: make([]string, 0),
	}

	for i := 0; i < len(deviceTokens); i += batchSize {
		end := i + batchSize
		if end > len(deviceTokens) {
			end = len(deviceTokens)
		}

		batch := deviceTokens[i:end]

		// Get the last read time for the user
		lastReadTime, err := s.messageRepo.GetLastReadTime(s.ctx, uuid.MustParse(message.GroupID), uuid.MustParse(message.SenderID))
		if err != nil {
			return response, fmt.Errorf("error getting last read time: %v", err)
		}

		badgeNumber, err := s.messageRepo.CountUnreadMessages(s.ctx, uuid.MustParse(message.GroupID), lastReadTime)
		if err != nil {
			return response, fmt.Errorf("error counting unread messages: %v", err)
		}

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

		batchResponse, err := s.client.SendEachForMulticast(s.ctx, batchMessage)
		if err != nil {
			return response, fmt.Errorf("error sending batch: %v", err)
		}

		response.SuccessCount += batchResponse.SuccessCount
		response.FailureCount += batchResponse.FailureCount

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
