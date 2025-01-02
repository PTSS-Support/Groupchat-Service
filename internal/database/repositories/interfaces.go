package repositories

import (
	"Groupchat-Service/internal/models"
	"context"
	"github.com/google/uuid"
	"time"
)

type MessageRepository interface {
	GetMessages(ctx context.Context, groupID uuid.UUID, query models.PaginationQuery) ([]models.Message, *models.PaginationResponse, error)
	CreateMessage(ctx context.Context, groupID uuid.UUID, message *models.Message) error
	ToggleMessagePin(ctx context.Context, groupID uuid.UUID, messageID uuid.UUID) (*models.Message, error)
	GetMessageByID(ctx context.Context, groupID uuid.UUID, messageID uuid.UUID) (*models.Message, error)
	GetLastReadTime(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) (time.Time, error)
	CountUnreadMessages(ctx context.Context, groupID uuid.UUID, lastReadTime time.Time) (int, error)
}

type FCMTokenRepository interface {
	GetGroupMemberTokens(ctx context.Context, groupID uuid.UUID) ([]string, error)
	SaveToken(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, token string) error
	DeleteToken(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) error
}
