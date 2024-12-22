package services

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type NotificationService struct {
	FCMServerKey string
}

func NewNotificationService(fcmServerKey string) *NotificationService {
	return &NotificationService{
		FCMServerKey: fcmServerKey,
	}
}

func (n *NotificationService) SendBatchNotifications(
	ctx context.Context,
	tokens []string,
	title string,
	body string,
	data map[string]string,
) {
	if len(tokens) == 0 {
		log.Println("No tokens provided for the notification")
		return
	}

	message := map[string]interface{}{
		"registration_ids": tokens,
		"notification": map[string]string{
			"title": title,
			"body":  body,
		},
		"data": data,
	}

	n.sendNotification(ctx, message)
}

func (n *NotificationService) sendNotification(ctx context.Context, message map[string]interface{}) {
	body, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal notification body: %v\n", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Failed to create notification request: %v\n", err)
		return
	}

	req.Header.Set("Authorization", "key="+n.FCMServerKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send notification: %v\n", err)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to send notification, status: %d\n", resp.StatusCode)
	}
}
