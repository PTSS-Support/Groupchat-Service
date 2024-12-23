package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"Groupchat-Service/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/google/uuid"
)

// TODO: ADD INTERFACES

// AzureMessageRepository implements MessageRepository using Azure Table Storage
type AzureMessageRepository struct {
	table *aztables.Client
}

// MessageEntity represents how we store messages in Azure Tables
type MessageEntity struct {
	aztables.Entity           // Embedded Entity for PartitionKey and RowKey
	GroupID         string    `json:"GroupID"`
	SenderID        string    `json:"SenderID"`
	SenderName      string    `json:"SenderName"`
	Content         string    `json:"Content"`
	SentAt          time.Time `json:"SentAt"`
	IsPinned        bool      `json:"IsPinned"`
}

func NewAzureMessageRepository(client *aztables.ServiceClient) (*AzureMessageRepository, error) {
	table := client.NewClient(MessagesTable)
	_, err := table.CreateTable(context.Background(), nil)
	if err != nil && err.Error() != "TableAlreadyExists" {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &AzureMessageRepository{table: table}, nil
}

func (r *AzureMessageRepository) CreateMessage(ctx context.Context, message *models.Message) error {
	entity := MessageEntity{
		Entity: aztables.Entity{
			PartitionKey: message.GroupID.String(),
			RowKey:       fmt.Sprintf("%d-%s", time.Now().UnixNano(), message.ID),
		},
		GroupID:    message.GroupID.String(),
		SenderID:   message.SenderID.String(),
		SenderName: message.SenderName,
		Content:    message.Content,
		SentAt:     message.SentAt,
		IsPinned:   message.IsPinned,
	}

	data, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	_, err = r.table.AddEntity(ctx, data, nil)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

func (r *AzureMessageRepository) GetGroupMessages(
	ctx context.Context,
	groupID uuid.UUID,
	opts MessageQueryOptions,
) ([]models.Message, string, error) {
	filter := fmt.Sprintf("PartitionKey eq '%s'", groupID)
	if opts.Cursor != "" {
		if opts.Direction == "next" {
			filter += " and RowKey lt '" + opts.Cursor + "'"
		} else {
			filter += " and RowKey gt '" + opts.Cursor + "'"
		}
	}

	pager := r.table.NewListEntitiesPager(&aztables.ListEntitiesOptions{Filter: &filter, Top: &int32(opts.Limit)})
	messages := []models.Message{}
	var lastKey string

	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, "", fmt.Errorf("failed to retrieve messages: %w", err)
		}

		for _, raw := range page.Entities {
			var entity MessageEntity
			if err := json.Unmarshal(raw, &entity); err != nil {
				continue
			}
			id, _ := uuid.Parse(entity.RowKey)
			messages = append(messages, models.Message{
				ID:         id,
				GroupID:    uuid.MustParse(entity.GroupID),
				SenderID:   uuid.MustParse(entity.SenderID),
				SenderName: entity.SenderName,
				Content:    entity.Content,
				SentAt:     entity.SentAt,
				IsPinned:   entity.IsPinned,
			})
			lastKey = entity.RowKey
		}
	}

	return messages, lastKey, nil
}
