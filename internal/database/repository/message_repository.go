package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"Groupchat-Service/internal/models"

	"github.com/google/uuid"
)

type MessageRepository interface {
	CreateMessage(ctx context.Context, message *models.Message) error
	GetGroupMessages(ctx context.Context, groupID uuid.UUID, opts MessageQueryOptions) ([]models.Message, string, error)
	PinMessage(ctx context.Context, messageID uuid.UUID, isPinned bool) error
}

type PostgresMessageRepository struct {
	db *sql.DB
}

type MessageQueryOptions struct {
	Limit     int
	Cursor    string
	Direction string
	Search    string
}

func NewMessageRepository(db *sql.DB) *PostgresMessageRepository {
	return &PostgresMessageRepository{db: db}
}

func (r *PostgresMessageRepository) CreateMessage(ctx context.Context, message *models.Message) error {
	query := `
        INSERT INTO group_messages 
        (group_id, sender_id, sender_name, content, sent_at) 
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `

	err := r.db.QueryRowContext(ctx, query,
		message.GroupID,
		message.SenderID,
		message.SenderName,
		message.Content,
		time.Now().UTC(),
	).Scan(&message.ID)

	return err
}

func (r *PostgresMessageRepository) GetGroupMessages(
	ctx context.Context,
	groupID uuid.UUID,
	opts MessageQueryOptions,
) ([]models.Message, string, error) {

	// Validate inputs
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 50 // Apply sensible default
	}
	if opts.Direction != "next" && opts.Direction != "previous" {
		opts.Direction = "next" // Default direction
	}

	// Build the query
	query := `
        SELECT id, sender_id, sender_name, content, sent_at, is_pinned 
        FROM group_messages 
        WHERE group_id = $1
    `
	params := []interface{}{groupID}
	paramIndex := 2 // Keep track of positional parameter placeholders

	// Add cursor-based pagination
	if opts.Cursor != "" {
		if opts.Direction == "next" {
			query += fmt.Sprintf(" AND sent_at < $%d", paramIndex)
		} else { // opts.Direction == "previous"
			query += fmt.Sprintf(" AND sent_at > $%d", paramIndex)
		}
		params = append(params, opts.Cursor)
		paramIndex++
	}

	// Add ordering and limit
	query += fmt.Sprintf(" ORDER BY sent_at %s LIMIT $%d", "DESC", paramIndex)
	params = append(params, opts.Limit)

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	// Parse results
	var messages []models.Message
	for rows.Next() {
		var message models.Message
		if err := rows.Scan(&message.ID, &message.SenderID, &message.SenderName, &message.Content, &message.SentAt, &message.IsPinned); err != nil {
			return nil, "", err
		}
		messages = append(messages, message)
	}

	// Check for rows errors
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	// Generate nextCursor (if messages exist)
	var nextCursor string
	if len(messages) > 0 {
		nextCursor = messages[len(messages)-1].SentAt.Format(time.RFC3339)
	}

	return messages, nextCursor, nil
}

func (r *PostgresMessageRepository) PinMessage(
	ctx context.Context,
	messageID uuid.UUID,
	isPinned bool,
) error {
	query := `
        UPDATE group_messages 
        SET is_pinned = $1 
        WHERE id = $2
    `

	_, err := r.db.ExecContext(ctx, query, isPinned, messageID)
	return err
}
