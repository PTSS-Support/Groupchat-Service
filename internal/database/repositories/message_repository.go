package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"strings"
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
	if err != nil {
		if strings.Contains(err.Error(), "TableAlreadyExists") {
			return &messageRepository{table: table}, nil
		}
		return nil, fmt.Errorf("failed to create/verify table: %w", err)
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

func (r *messageRepository) ToggleMessagePin(ctx context.Context, groupID uuid.UUID, messageID uuid.UUID) (*models.Message, error) {
	message, err := r.GetMessageByID(ctx, groupID, messageID)
	if err != nil {
		return nil, fmt.Errorf("error getting message by ID: %v", err)
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

	message.IsPinned = !message.IsPinned
	return message, nil
}

func (r *messageRepository) GetMessageByID(ctx context.Context, groupID uuid.UUID, messageID uuid.UUID) (*models.Message, error) {
	partitionKey := groupID.String()
	rowKey := messageID.String()

	entity, err := r.table.GetEntity(ctx, partitionKey, rowKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	var rawEntity map[string]interface{}
	if err := json.Unmarshal(entity.Value, &rawEntity); err != nil {
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

func (r *messageRepository) GetMessages(ctx context.Context, groupID uuid.UUID, query models.PaginationQuery) ([]models.Message, *models.PaginationResponse, error) {
	// Filter messages by group ID
	filter := fmt.Sprintf("PartitionKey eq '%s'", groupID.String())

	if query.Cursor != nil && *query.Cursor != "" {
		cursorUUID, err := uuid.Parse(*query.Cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid cursor format: %w", err)
		}

		cursorMsg, err := r.GetMessageByID(ctx, groupID, cursorUUID)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid cursor: %w", err)
		}

		// Azure Tables expects ISO8601 format directly
		cursorTime := cursorMsg.SentAt.UTC().Format("2006-01-02T15:04:05Z")
		if query.Direction == "previous" {
			filter += fmt.Sprintf(" and SentAt gt '%s'", cursorTime)
		} else {
			filter += fmt.Sprintf(" and SentAt lt '%s'", cursorTime)
		}
	}

	pageSize := int32(query.PageSize)
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
		Top:    &pageSize,
	}

	pager := r.table.NewListEntitiesPager(options)

	var messages []models.Message
	var filteredMessages []models.Message

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

			sentAt, err := time.Parse(time.RFC3339Nano, rawEntity["SentAt"].(string))
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

			content, _ := rawEntity["Content"].(string)
			message := models.Message{
				ID:         messageID,
				GroupID:    groupID,
				SenderID:   senderID,
				SenderName: rawEntity["SenderName"].(string),
				Content:    rawEntity["Content"].(string),
				SentAt:     sentAt,
				IsPinned:   rawEntity["IsPinned"].(bool),
			}

			// If search is provided, filter in memory because Azure Tables does not support full-text search
			if query.Search != nil && *query.Search != "" {
				if strings.Contains(
					strings.ToLower(content),
					strings.ToLower(*query.Search),
				) {
					filteredMessages = append(filteredMessages, message)
				}
			} else {
				messages = append(messages, message)
			}
		}
	}

	if query.Search != nil && *query.Search != "" {
		messages = filteredMessages
	}

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

func (r *messageRepository) CountUnreadMessages(ctx context.Context, groupID uuid.UUID, lastReadTime time.Time) (int, error) {
	filter := fmt.Sprintf("PartitionKey eq '%s' and SentAt gt '%s'", groupID.String(), lastReadTime.UTC().Format(time.RFC3339))
	selectFields := "PartitionKey"
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
		Select: &selectFields,
	}

	unreadCount := 0
	pager := r.table.NewListEntitiesPager(options)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to list unread messages: %w", err)
		}
		unreadCount += len(page.Entities)
	}

	return unreadCount, nil
}

func (r *messageRepository) GetLastReadTime(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) (time.Time, error) {
	partitionKey := groupID.String()
	rowKey := userID.String()

	entity, err := r.table.GetEntity(ctx, partitionKey, rowKey, nil)
	if err != nil {
		if strings.Contains(err.Error(), "ResourceNotFound") {
			// Return the current time if the entity does not exist. For example: If the user has not read any messages
			return time.Now().UTC(), nil
		}
		return time.Time{}, fmt.Errorf("failed to get last read time: %w", err)
	}

	var rawEntity map[string]interface{}
	if err := json.Unmarshal(entity.Value, &rawEntity); err != nil {
		return time.Time{}, fmt.Errorf("failed to unmarshal entity: %w", err)
	}

	lastReadTime, err := time.Parse(time.RFC3339, rawEntity["LastReadTime"].(string))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse last read time: %w", err)
	}

	return lastReadTime, nil
}
