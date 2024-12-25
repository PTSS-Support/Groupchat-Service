package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"time"

	"Groupchat-Service/internal/models"
	"github.com/google/uuid"
)

type messageRepository struct {
	table *aztables.Client
}

// MessageEntity represents the structure for Azure Table Storage
type MessageEntity struct {
	PartitionKey string `json:"PartitionKey"`
	RowKey       string `json:"RowKey"`
	SenderID     string `json:"SenderID"`
	SenderName   string `json:"SenderName"`
	Content      string `json:"Content"`
	SentAt       string `json:"SentAt"`
	IsPinned     bool   `json:"IsPinned"`
}

func NewMessageRepository(client *aztables.ServiceClient) (MessageRepository, error) {
	table := client.NewClient(MessagesTable)
	_, err := table.CreateTable(context.Background(), nil)
	if err != nil && err.Error() != "TableAlreadyExists" {
		return nil, fmt.Errorf("failed to create messages table: %w", err)
	}
	return &messageRepository{table: table}, nil
}

func (r *messageRepository) CreateMessage(ctx context.Context, groupID uuid.UUID, message *models.Message) error {
	entity := MessageEntity{
		PartitionKey: groupID.String(),
		RowKey:       message.ID.String(),
		SenderID:     message.SenderID.String(),
		SenderName:   message.SenderName,
		Content:      message.Content,
		SentAt:       message.SentAt.UTC().Format(time.RFC3339),
		IsPinned:     message.IsPinned,
	}

	marshaled, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	_, err = r.table.AddEntity(ctx, marshaled, nil)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

func (r *messageRepository) ToggleMessagePin(ctx context.Context, messageID uuid.UUID) (*models.Message, error) {
	// First get the message to get its current state and partition key
	message, err := r.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	entity := MessageEntity{
		PartitionKey: message.GroupID.String(),
		RowKey:       messageID.String(),
		SenderID:     message.SenderID.String(),
		SenderName:   message.SenderName,
		Content:      message.Content,
		SentAt:       message.SentAt.UTC().Format(time.RFC3339),
		IsPinned:     !message.IsPinned,
	}

	marshaled, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %w", err)
	}

	_, err = r.table.UpdateEntity(ctx, marshaled, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	// Update the message object with new pin status
	message.IsPinned = !message.IsPinned
	return message, nil
}

func (r *messageRepository) GetMessageByID(ctx context.Context, messageID uuid.UUID) (*models.Message, error) {
	filter := fmt.Sprintf("RowKey eq '%s'", messageID.String())

	pager := r.table.NewListEntitiesPager(&aztables.ListEntitiesOptions{
		Filter: &filter,
	})

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get message: %w", err)
		}

		// Should only be one result
		for _, entity := range page.Entities {
			var rawEntity map[string]interface{}
			if err := json.Unmarshal(entity, &rawEntity); err != nil {
				return nil, fmt.Errorf("failed to unmarshal entity: %w", err)
			}

			sentAt, err := time.Parse(time.RFC3339, rawEntity["SentAt"].(string))
			if err != nil {
				return nil, fmt.Errorf("failed to parse sent time: %w", err)
			}

			senderID, err := uuid.Parse(rawEntity["SenderID"].(string))
			if err != nil {
				return nil, fmt.Errorf("failed to parse sender ID: %w", err)
			}

			groupID, err := uuid.Parse(rawEntity["PartitionKey"].(string))
			if err != nil {
				return nil, fmt.Errorf("failed to parse group ID: %w", err)
			}

			return &models.Message{
				ID:         messageID,
				GroupID:    groupID,
				SenderID:   senderID,
				SenderName: rawEntity["SenderName"].(string),
				Content:    rawEntity["Content"].(string),
				SentAt:     sentAt,
				IsPinned:   rawEntity["IsPinned"].(bool),
			}, nil
		}
	}

	return nil, fmt.Errorf("message not found")
}

func (r *messageRepository) GetMessages(ctx context.Context, groupID uuid.UUID, query models.PaginationQuery) ([]models.Message, *models.PaginationResponse, error) {
	filter := fmt.Sprintf("PartitionKey eq '%s'", groupID.String())

	// Add search filter if provided
	if query.Search != nil && *query.Search != "" {
		filter += fmt.Sprintf(" and Content ne '' and Content containing '%s'", *query.Search)
	}

	pageSize := int32(query.PageSize)
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
		Top:    &pageSize,
	}

	// Add cursor-based filtering if cursor is provided
	if query.Cursor != nil {
		cursorMsg, err := r.GetMessageByID(ctx, uuid.MustParse(*query.Cursor))
		if err != nil {
			return nil, nil, fmt.Errorf("invalid cursor: %w", err)
		}

		cursorTime := cursorMsg.SentAt.UTC().Format(time.RFC3339)
		if query.Direction == "previous" {
			filter += fmt.Sprintf(" and SentAt gt '%s'", cursorTime)
		} else {
			filter += fmt.Sprintf(" and SentAt lt '%s'", cursorTime)
		}
		options.Filter = &filter
	}

	pager := r.table.NewListEntitiesPager(options)

	var messages []models.Message
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list messages: %w", err)
		}

		for _, entity := range page.Entities {
			var rawEntity map[string]interface{}
			if err := json.Unmarshal(entity, &rawEntity); err != nil {
				return nil, nil, fmt.Errorf("failed to unmarshal entity: %w", err)
			}

			sentAt, err := time.Parse(time.RFC3339, rawEntity["SentAt"].(string))
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse sent time: %w", err)
			}

			messageID, err := uuid.Parse(rawEntity["RowKey"].(string))
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse message ID: %w", err)
			}

			senderID, err := uuid.Parse(rawEntity["SenderID"].(string))
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse sender ID: %w", err)
			}

			message := models.Message{
				ID:         messageID,
				GroupID:    groupID,
				SenderID:   senderID,
				SenderName: rawEntity["SenderName"].(string),
				Content:    rawEntity["Content"].(string),
				SentAt:     sentAt,
				IsPinned:   rawEntity["IsPinned"].(bool),
			}
			messages = append(messages, message)
		}
	}

	// Handle pagination response
	pagination := &models.PaginationResponse{}
	if len(messages) > 0 {
		if len(messages) > query.PageSize {
			messages = messages[:query.PageSize]
			pagination.HasNext = true
			pagination.NextCursor = ptr(messages[len(messages)-1].ID.String())
		}

		if query.Cursor != nil {
			pagination.HasPrevious = true
			pagination.PreviousCursor = ptr(messages[0].ID.String())
		}
	}

	return messages, pagination, nil
}

// ptr returns a pointer to the string
func ptr(s string) *string {
	return &s
}
