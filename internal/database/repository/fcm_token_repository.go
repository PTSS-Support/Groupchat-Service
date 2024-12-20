package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"Groupchat-Service/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/google/uuid"
)

// AzureFCMTokenRepository implements FCMTokenRepository using Azure Table Storage
type AzureFCMTokenRepository struct {
	client *aztables.ServiceClient
	table  *aztables.Client
}

// NewAzureFCMTokenRepository creates a new FCM token repository
func NewAzureFCMTokenRepository(client *aztables.ServiceClient) (*AzureFCMTokenRepository, error) {
	table := client.NewClient(FCMTokensTable)

	_, err := table.CreateTable(context.Background(), nil)
	if err != nil {
		// Check if error is "TableAlreadyExists"
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.ErrorCode == "TableAlreadyExists" {
			// Table exists, which is fine
		} else {
			return nil, fmt.Errorf("failed to create tokens table: %w", err)
		}
	}

	return &AzureFCMTokenRepository{
		client: client,
		table:  table,
	}, nil
}

// FCMTokenEntity represents how we store FCM tokens in Azure Tables
type FCMTokenEntity struct {
	aztables.Entity           // Embedded Entity for PartitionKey and RowKey
	UserID          string    `json:"UserID"`
	Token           string    `json:"Token"`
	Platform        string    `json:"Platform"`
	DeviceName      string    `json:"DeviceName"`
	CreatedAt       time.Time `json:"CreatedAt"`
	LastUsedAt      time.Time `json:"LastUsedAt"`
	IsActive        bool      `json:"IsActive"`
}

func (r *AzureFCMTokenRepository) CreateFCMToken(ctx context.Context, token *models.FCMToken) error {
	// In Azure Tables, we'll use UserID as PartitionKey for efficient querying
	// and TokenID as RowKey for uniqueness
	entity := FCMTokenEntity{
		Entity: aztables.Entity{
			PartitionKey: token.UserID.String(),
			RowKey:       token.ID.String(),
		},
		UserID:     token.UserID.String(),
		Token:      token.Token,
		Platform:   string(token.Platform),
		DeviceName: token.DeviceName,
		CreatedAt:  token.CreatedAt,
		LastUsedAt: token.LastUsedAt,
		IsActive:   token.IsActive,
	}

	marshalled, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal token entity: %w", err)
	}

	_, err = r.table.AddEntity(ctx, marshalled, nil)
	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}

	return nil
}

func (r *AzureFCMTokenRepository) GetFCMTokensByUser(ctx context.Context, userID uuid.UUID) ([]models.FCMToken, error) {
	filter := fmt.Sprintf(
		"PartitionKey eq '%s' and IsActive eq true",
		userID.String(),
	)

	pager := r.table.NewListEntitiesPager(&aztables.ListEntitiesOptions{
		Filter: &filter,
	})

	var tokens []models.FCMToken
	for pager.More() {
		response, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list tokens: %w", err)
		}

		for _, entity := range response.Entities {
			var tokenEntity FCMTokenEntity
			if err := json.Unmarshal(entity, &tokenEntity); err != nil {
				return nil, fmt.Errorf("failed to unmarshal token: %w", err)
			}

			userID, _ := uuid.Parse(tokenEntity.UserID)
			tokenID, _ := uuid.Parse(tokenEntity.Entity.RowKey)

			token := models.FCMToken{
				ID:         tokenID,
				UserID:     userID,
				Token:      tokenEntity.Token,
				Platform:   models.Platform(tokenEntity.Platform),
				DeviceName: tokenEntity.DeviceName,
				CreatedAt:  tokenEntity.CreatedAt,
				LastUsedAt: tokenEntity.LastUsedAt,
				IsActive:   tokenEntity.IsActive,
			}
			tokens = append(tokens, token)
		}
	}

	return tokens, nil
}
