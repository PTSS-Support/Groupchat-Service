package services

import (
	"Groupchat-Service/internal/models"
	"context"
	"github.com/google/uuid"
)

type MessageService interface {
	GetMessages(ctx context.Context, groupID uuid.UUID, query models.PaginationQuery) ([]models.MessageResponse, *models.PaginationResponse, error)
	CreateMessage(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, userName string, create models.MessageCreate) (*models.Message, error)
	ToggleMessagePin(ctx context.Context, groupID uuid.UUID, messageID uuid.UUID) (*models.Message, error)
}

type NotificationService interface {
	SendGroupMessage(message Message, deviceTokens []string) (*BatchResponse, error)
}

type ValidationService interface {
	ValidatePaginationQuery(queryParams map[string]string) (models.PaginationQuery, error)
	ValidateUserContext(ctx context.Context) (uuid.UUID, string, error)
	ValidateGroupID(groupID string) (uuid.UUID, error)
	ValidateUserID(userID string) (uuid.UUID, error)
	ValidateToken(token string) error
}

type FCMTokenService interface {
	SaveToken(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, token string) error
	DeleteToken(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) error
}

type HealthService interface {
	CheckHealth(ctx context.Context) (*models.HealthResponse, error)
	CheckReadiness(ctx context.Context) (*models.HealthResponse, error)
	CheckLiveness(ctx context.Context) (*models.HealthResponse, error)
}
