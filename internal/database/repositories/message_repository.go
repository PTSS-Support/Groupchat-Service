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

// LastReadEntity represents the structure for storing last read times
type LastReadEntity struct {
	PartitionKey string `json:"PartitionKey"` // GroupID
	RowKey       string `json:"RowKey"`       // UserID
	LastReadTime string `json:"LastReadTime"`
}

// entityMapper handles conversion between Message and MessageEntity
type entityMapper struct{}

func (m *entityMapper) toEntity(groupID uuid.UUID, message *models.Message) MessageEntity {
	return MessageEntity{
		PartitionKey: groupID.String(),
		RowKey:       message.ID.String(),
		SenderID:     message.SenderID.String(),
		SenderName:   message.SenderName,
		Content:      message.Content,
		SentAt:       message.SentAt.UTC().Format(time.RFC3339),
		IsPinned:     message.IsPinned,
	}
}

func (m *entityMapper) toMessage(rawEntity map[string]interface{}) (*models.Message, error) {
	messageID, err := uuid.Parse(rawEntity["RowKey"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to parse message ID: %w", err)
	}

	groupID, err := uuid.Parse(rawEntity["PartitionKey"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to parse group ID: %w", err)
	}

	senderID, err := uuid.Parse(rawEntity["SenderID"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to parse sender ID: %w", err)
	}

	sentAt, err := time.Parse(time.RFC3339, rawEntity["SentAt"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to parse sent time: %w", err)
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

// tableOperations handles common Azure Table operations
type tableOperations struct {
	table *aztables.Client
}

func (t *tableOperations) addEntity(ctx context.Context, entity interface{}) error {
	marshaled, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	_, err = t.table.AddEntity(ctx, marshaled, nil)
	if err != nil {
		return fmt.Errorf("failed to add entity: %w", err)
	}

	return nil
}

func (t *tableOperations) updateEntity(ctx context.Context, entity interface{}) error {
	marshaled, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	_, err = t.table.UpdateEntity(ctx, marshaled, nil)
	if err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}

	return nil
}

// queryBuilder handles building Azure Table queries
type queryBuilder struct{}

func (q *queryBuilder) buildMessageFilter(groupID uuid.UUID, query *models.PaginationQuery, cursorTime *time.Time) string {
	filter := fmt.Sprintf("PartitionKey eq '%s'", groupID.String())

	if cursorTime != nil {
		if query.Direction == "previous" {
			filter += fmt.Sprintf(" and SentAt gt '%s'", cursorTime.UTC().Format("2006-01-02T15:04:05Z"))
		} else {
			filter += fmt.Sprintf(" and SentAt lt '%s'", cursorTime.UTC().Format("2006-01-02T15:04:05Z"))
		}
	}

	return filter
}

func (q *queryBuilder) buildUnreadMessagesFilter(groupID uuid.UUID, lastReadTime time.Time) string {
	return fmt.Sprintf("PartitionKey eq '%s' and SentAt gt '%s'",
		groupID.String(),
		lastReadTime.UTC().Format(time.RFC3339))
}

func (m *entityMapper) parseLastReadTime(rawEntity map[string]interface{}) (time.Time, error) {
	lastReadTime, err := time.Parse(time.RFC3339, rawEntity["LastReadTime"].(string))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse last read time: %w", err)
	}
	return lastReadTime, nil
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
	mapper := &entityMapper{}
	ops := &tableOperations{table: r.table}

	entity := mapper.toEntity(groupID, message)
	return ops.addEntity(ctx, entity)
}

func (r *messageRepository) ToggleMessagePin(ctx context.Context, groupID uuid.UUID, messageID uuid.UUID) (*models.Message, error) {
	message, err := r.GetMessageByID(ctx, groupID, messageID)
	if err != nil {
		return nil, fmt.Errorf("error getting message by ID: %v", err)
	}

	message.IsPinned = !message.IsPinned

	mapper := &entityMapper{}
	ops := &tableOperations{table: r.table}

	entity := mapper.toEntity(groupID, message)
	if err := ops.updateEntity(ctx, entity); err != nil {
		return nil, err
	}

	return message, nil
}

func (r *messageRepository) GetMessageByID(ctx context.Context, groupID uuid.UUID, messageID uuid.UUID) (*models.Message, error) {
	entity, err := r.table.GetEntity(ctx, groupID.String(), messageID.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	var rawEntity map[string]interface{}
	if err := json.Unmarshal(entity.Value, &rawEntity); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entity: %w", err)
	}

	mapper := &entityMapper{}
	return mapper.toMessage(rawEntity)
}

func (r *messageRepository) GetMessages(ctx context.Context, groupID uuid.UUID, query models.PaginationQuery) ([]models.Message, *models.PaginationResponse, error) {
	var cursorTime *time.Time
	if query.Cursor != nil && *query.Cursor != "" {
		cursorUUID, err := uuid.Parse(*query.Cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid cursor format: %w", err)
		}

		cursorMsg, err := r.GetMessageByID(ctx, groupID, cursorUUID)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid cursor: %w", err)
		}
		t := cursorMsg.SentAt
		cursorTime = &t
	}

	qb := &queryBuilder{}
	filter := qb.buildMessageFilter(groupID, &query, cursorTime)

	messages, err := r.fetchMessages(ctx, filter, query.PageSize)
	if err != nil {
		return nil, nil, err
	}

	if query.Search != nil && *query.Search != "" {
		messages = r.filterMessagesByContent(messages, *query.Search)
	}

	return r.buildPaginatedResponse(messages, query)
}

func (r *messageRepository) fetchMessages(ctx context.Context, filter string, pageSize int) ([]models.Message, error) {
	size := int32(pageSize)
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
		Top:    &size,
	}

	pager := r.table.NewListEntitiesPager(options)
	mapper := &entityMapper{}
	var messages []models.Message

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list messages: %w", err)
		}

		for _, entity := range page.Entities {
			var rawEntity map[string]interface{}
			if err := json.Unmarshal(entity, &rawEntity); err != nil {
				return nil, fmt.Errorf("failed to unmarshal entity: %w", err)
			}

			message, err := mapper.toMessage(rawEntity)
			if err != nil {
				return nil, err
			}
			messages = append(messages, *message)
		}
	}

	return messages, nil
}

func (r *messageRepository) filterMessagesByContent(messages []models.Message, search string) []models.Message {
	var filtered []models.Message
	searchLower := strings.ToLower(search)

	for _, msg := range messages {
		if strings.Contains(strings.ToLower(msg.Content), searchLower) {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// ptr returns a pointer to the string
func ptr(s string) *string {
	return &s
}

func (r *messageRepository) buildPaginatedResponse(messages []models.Message, query models.PaginationQuery) ([]models.Message, *models.PaginationResponse, error) {
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

func (r *messageRepository) CountUnreadMessages(ctx context.Context, groupID uuid.UUID, lastReadTime time.Time) (int, error) {
	qb := &queryBuilder{}
	filter := qb.buildUnreadMessagesFilter(groupID, lastReadTime)

	// Only select PartitionKey to minimize data transfer
	selectFields := "PartitionKey"
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
		Select: &selectFields,
	}

	return r.countFilteredEntities(ctx, options)
}

func (r *messageRepository) countFilteredEntities(ctx context.Context, options *aztables.ListEntitiesOptions) (int, error) {
	pager := r.table.NewListEntitiesPager(options)
	count := 0

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to list entities: %w", err)
		}
		count += len(page.Entities)
	}

	return count, nil
}

func (r *messageRepository) GetLastReadTime(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) (time.Time, error) {
	entity, err := r.table.GetEntity(ctx, groupID.String(), userID.String(), nil)
	if err != nil {
		if strings.Contains(err.Error(), "ResourceNotFound") {
			// Return the current time if the entity does not exist
			return time.Now().UTC(), nil
		}
		return time.Time{}, fmt.Errorf("failed to get last read time: %w", err)
	}

	var rawEntity map[string]interface{}
	if err := json.Unmarshal(entity.Value, &rawEntity); err != nil {
		return time.Time{}, fmt.Errorf("failed to unmarshal entity: %w", err)
	}

	mapper := &entityMapper{}
	return mapper.parseLastReadTime(rawEntity)
}
