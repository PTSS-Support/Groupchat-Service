package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"time"

	"Groupchat-Service/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/google/uuid"
)

// AzureMessageRepository implements MessageRepository using Azure Table Storage
type AzureMessageRepository struct {
	client *aztables.ServiceClient
	table  *aztables.Client
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

// NewAzureMessageRepository creates a new message repository
func NewAzureMessageRepository(client *aztables.ServiceClient) (*AzureMessageRepository, error) {
	table := client.NewClient(MessagesTable)

	// Check if table exists and create it if it doesn't
	_, err := table.CreateTable(context.Background(), nil)
	if err != nil {
		// Check if error is "TableAlreadyExists"
		if azerr, ok := err.(*azcore.ResponseError); ok && azerr.ErrorCode == "TableAlreadyExists" {
			// Table exists, which is fine
		} else {
			return nil, fmt.Errorf("failed to create messages table: %w", err)
		}
	}

	return &AzureMessageRepository{
		client: client,
		table:  table,
	}, nil
}

func (r *AzureMessageRepository) CreateMessage(ctx context.Context, message *models.Message) error {
	// Use GroupID as PartitionKey for efficient querying of messages in a group
	// Use a timestamp-based RowKey for natural ordering
	// Format: {timestamp}-{messageID} to ensure uniqueness
	timestampPrefix := fmt.Sprintf("%d", time.Now().UnixNano())

	entity := MessageEntity{
		Entity: aztables.Entity{
			PartitionKey: message.GroupID.String(),
			RowKey:       fmt.Sprintf("%s-%s", timestampPrefix, message.ID.String()),
		},
		GroupID:    message.GroupID.String(),
		SenderID:   message.SenderID.String(),
		SenderName: message.SenderName,
		Content:    message.Content,
		SentAt:     message.SentAt,
		IsPinned:   message.IsPinned,
	}

	marshalled, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal message entity: %w", err)
	}

	_, err = r.table.AddEntity(ctx, marshalled, nil)
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
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 50 // Default limit
	}

	// Build filter for the specific group
	filter := fmt.Sprintf("PartitionKey eq '%s'", groupID.String())

	// Add cursor-based filtering if provided
	if opts.Cursor != "" {
		if opts.Direction == "next" {
			filter += fmt.Sprintf(" and RowKey lt '%s'", opts.Cursor)
		} else {
			filter += fmt.Sprintf(" and RowKey gt '%s'", opts.Cursor)
		}
	}

	// Configure paging options
	pageSize := int32(opts.Limit) // Convert to int32 as required by Azure SDK
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
		Top:    &pageSize,
	}

	pager := r.table.NewListEntitiesPager(options)

	var messages []models.Message
	var lastRowKey string

	for pager.More() && len(messages) < opts.Limit {
		response, err := pager.NextPage(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("failed to list messages: %w", err)
		}

		for _, entity := range response.Entities {
			var messageEntity MessageEntity
			if err := json.Unmarshal(entity, &messageEntity); err != nil {
				return nil, "", fmt.Errorf("failed to unmarshal message: %w", err)
			}

			groupID, _ := uuid.Parse(messageEntity.GroupID)
			senderID, _ := uuid.Parse(messageEntity.SenderID)
			messageID, _ := uuid.Parse(messageEntity.Entity.RowKey)

			message := models.Message{
				ID:         messageID,
				GroupID:    groupID,
				SenderID:   senderID,
				SenderName: messageEntity.SenderName,
				Content:    messageEntity.Content,
				SentAt:     messageEntity.SentAt,
				IsPinned:   messageEntity.IsPinned,
			}
			messages = append(messages, message)
			lastRowKey = messageEntity.RowKey
		}
	}

	return messages, lastRowKey, nil
}
