package repository

import (
	"Groupchat-Service/internal/models"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

// FCMTokenRepository is an interface for operations related to FCM tokens.
type FCMTokenRepository interface {
	// CreateFCMToken inserts a new FCM token into the database.
	CreateFCMToken(ctx context.Context, token *models.FCMToken) error

	// GetFCMTokensByUser retrieves all active FCM tokens for a specific user.
	GetFCMTokensByUser(ctx context.Context, userID uuid.UUID) ([]models.FCMToken, error)

	// GetGroupMemberTokens retrieves active FCM tokens for all group members.
	GetGroupMemberTokens(ctx context.Context, groupID uuid.UUID) ([]string, error)

	// DeleteFCMToken deactivates a specific FCM token in the database.
	DeleteFCMToken(ctx context.Context, tokenID uuid.UUID) error
}

// PostgresFCMTokenRepository is the Postgres implementation of the FCMTokenRepository.
type PostgresFCMTokenRepository struct {
	db *sql.DB
}

// NewFCMTokenRepository creates a new instance of PostgresFCMTokenRepository.
func NewFCMTokenRepository(db *sql.DB) *PostgresFCMTokenRepository {
	return &PostgresFCMTokenRepository{
		db: db,
	}
}

// CreateFCMToken inserts a new FCM token into the database.
func (r *PostgresFCMTokenRepository) CreateFCMToken(ctx context.Context, token *models.FCMToken) error {
	query := `
		INSERT INTO fcm_tokens (id, user_id, token, platform, device_name, created_at, last_used_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.UserID,
		token.Token,
		token.Platform,
		token.DeviceName,
		token.CreatedAt,
		token.LastUsedAt,
		token.IsActive,
	)
	return err
}

// GetFCMTokensByUser retrieves all active FCM tokens for a specific user.
func (r *PostgresFCMTokenRepository) GetFCMTokensByUser(ctx context.Context, userID uuid.UUID) ([]models.FCMToken, error) {
	query := `
		SELECT id, user_id, token, platform, device_name, created_at, last_used_at, is_active
		FROM fcm_tokens
		WHERE user_id = $1 AND is_active = TRUE
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []models.FCMToken
	for rows.Next() {
		var token models.FCMToken
		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.Token,
			&token.Platform,
			&token.DeviceName,
			&token.CreatedAt,
			&token.LastUsedAt,
			&token.IsActive,
		)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

// GetGroupMemberTokens retrieves active FCM tokens for all group members by group ID.
func (r *PostgresFCMTokenRepository) GetGroupMemberTokens(ctx context.Context, groupID uuid.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT t.token
		FROM fcm_tokens t
		INNER JOIN group_members gm ON t.user_id = gm.user_id
		WHERE gm.group_id = $1 AND t.is_active = TRUE
	`
	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

// DeleteFCMToken deactivates a specific FCM token in the database by setting is_active to FALSE.
func (r *PostgresFCMTokenRepository) DeleteFCMToken(ctx context.Context, tokenID uuid.UUID) error {
	query := `
		UPDATE fcm_tokens
		SET is_active = FALSE
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no rows affected, token ID might not exist")
	}
	return nil
}
