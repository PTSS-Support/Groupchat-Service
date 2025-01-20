package repositories

import (
	"Groupchat-Service/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"
)

type FcmTokenRepository struct {
	table *aztables.Client
}

func NewFCMTokenRepository(client *aztables.ServiceClient) (*FcmTokenRepository, error) {
	table := client.NewClient(FCMTokensTable)

	_, err := table.CreateTable(context.Background(), nil)
	if err != nil {
		if strings.Contains(err.Error(), "TableAlreadyExists") {
			return &FcmTokenRepository{table: table}, nil
		}
		return nil, fmt.Errorf("failed to create/verify table: %w", err)
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
			var temp struct {
				PartitionKey string `json:"PartitionKey"`
				RowKey       string `json:"RowKey"`
				Token        string `json:"Token"`
				IsActive     bool   `json:"IsActive"`
				Timestamp    string `json:"Timestamp"`
			}
			if err := json.Unmarshal(entity, &temp); err != nil {
				return nil, err
			}

			// Convert string timestamp to timestamppb.Timestamp
			timestamp, err := time.Parse(time.RFC3339, temp.Timestamp)
			if err != nil {
				return nil, fmt.Errorf("failed to parse timestamp: %w", err)
			}

			tokenEntity := models.FCMToken{
				PartitionKey: temp.PartitionKey,
				RowKey:       temp.RowKey,
				Token:        temp.Token,
				IsActive:     temp.IsActive,
				Timestamp:    timestamppb.New(timestamp),
			}

			tokens = append(tokens, tokenEntity.Token)
		}
	}

	return tokens, nil
}

type AzureTableEntity struct {
	PartitionKey string `json:"PartitionKey"`
	RowKey       string `json:"RowKey"`
	Token        string `json:"Token"`
	IsActive     bool   `json:"IsActive"`
	Timestamp    string `json:"Timestamp"`
}

func (r *FcmTokenRepository) SaveToken(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, token string) error {
	entity := AzureTableEntity{
		PartitionKey: groupID.String(),
		RowKey:       userID.String(),
		Token:        token,
		IsActive:     true,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}

	marshaled, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	_, err = r.table.UpsertEntity(ctx, marshaled, &aztables.UpsertEntityOptions{
		UpdateMode: aztables.UpdateModeReplace,
	})
	if err != nil {
		return fmt.Errorf("failed to save token (upsert): %w", err)
	}

	return nil
}

func (r *FcmTokenRepository) DeleteToken(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) error {
	partitionKey := groupID.String()
	rowKey := userID.String()

	_, err := r.table.DeleteEntity(ctx, partitionKey, rowKey, nil)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	return nil
}
