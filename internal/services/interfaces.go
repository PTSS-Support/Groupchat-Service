package services

import (
	"Groupchat-Service/internal/models"
	"context"
	"github.com/google/uuid"
)

type MessageService interface {
	GetMessages(ctx context.Context, groupID uuid.UUID, query models.PaginationQuery) ([]models.Message, *models.PaginationResponse, error)
	CreateMessage(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, userName string, create models.MessageCreate) (*models.Message, error)
	ToggleMessagePin(ctx context.Context, messageID uuid.UUID) (*models.Message, error)
}

type NotificationService interface {
	SendGroupMessage(message Message, deviceTokens []string) (*BatchResponse, error)
}
