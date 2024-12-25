package repositories

import (
	"Groupchat-Service/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/google/uuid"
)

type FcmTokenRepository struct {
	table *aztables.Client
}

func NewFCMTokenRepository(client *aztables.ServiceClient) (*FcmTokenRepository, error) {
	table := client.NewClient(FCMTokensTable)
	_, err := table.CreateTable(context.Background(), nil)
	if err != nil && err.Error() != "TableAlreadyExists" {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &FcmTokenRepository{table: table}, nil
}

func (r *FcmTokenRepository) GetGroupMemberTokens(ctx context.Context, groupID uuid.UUID) ([]string, error) {
	filter := fmt.Sprintf("PartitionKey eq '%s' and IsActive eq true", groupID.String())
	pager := r.table.NewListEntitiesPager(&aztables.ListEntitiesOptions{
		Filter: &filter,
	})

	var tokens []string

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list group member tokens: %w", err)
		}

		for _, entity := range page.Entities {
			var tokenEntity models.FCMToken
			if err := json.Unmarshal(entity, &tokenEntity); err != nil {
				return nil, err
			}
			tokens = append(tokens, tokenEntity.Token)
		}
	}

	return tokens, nil
}

func (r *FcmTokenRepository) SaveToken(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, token string) error {
	entity := models.FCMToken{
		PartitionKey: groupID.String(),
		RowKey:       userID.String(),
		Token:        token,
		IsActive:     true,
	}

	marshaled, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	_, err = r.table.AddEntity(ctx, marshaled, nil)
	if err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

func (r *FcmTokenRepository) DeleteToken(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) error {
	// TODO implement

	return nil
}
