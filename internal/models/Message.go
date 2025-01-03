package models

import (
	"github.com/google/uuid"
	"time"
)

type Message struct {
	ID         uuid.UUID `json:"id"`
	GroupID    uuid.UUID `json:"groupId"`
	SenderID   uuid.UUID `json:"senderId"`
	SenderName string    `json:"senderName"`
	Content    string    `json:"content"`
	SentAt     time.Time `json:"sentAt"`
	IsPinned   bool      `json:"isPinned"`
}

type MessageCreate struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
}
