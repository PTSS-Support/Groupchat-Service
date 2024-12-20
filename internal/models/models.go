// Package models defines the core data structures used throughout the FCM microservice.
// It includes models for messages, FCM tokens, and various request/response structures
// needed for the API endpoints.
package models

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Message represents a chat message within a group.
// It contains all the necessary information about the message,
// including metadata about the sender and the message state.
type Message struct {
	ID         uuid.UUID `json:"id" db:"id"`
	GroupID    uuid.UUID `json:"group_id" db:"group_id"`
	SenderID   uuid.UUID `json:"sender_id" db:"sender_id"`
	SenderName string    `json:"sender_name" db:"sender_name"`
	Content    string    `json:"content" db:"content"`
	SentAt     time.Time `json:"sent_at" db:"sent_at"`
	IsPinned   bool      `json:"is_pinned" db:"is_pinned"`
}

// FCMToken represents a Firebase Cloud Messaging token associated with a user's device.
// Each user might have multiple tokens if they use multiple devices.
type FCMToken struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	Token      string    `json:"token" db:"token"`
	Platform   Platform  `json:"platform" db:"platform"`
	DeviceName string    `json:"device_name" db:"device_name"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	LastUsedAt time.Time `json:"last_used_at" db:"last_used_at"`
	IsActive   bool      `json:"is_active" db:"is_active"`
}

// Platform represents the device platform type (iOS or Android)
type Platform string

const (
	PlatformIOS     Platform = "ios"
	PlatformAndroid Platform = "android"
)

// Validate checks if the platform value is valid
func (p Platform) Validate() error {
	switch p {
	case PlatformIOS, PlatformAndroid:
		return nil
	default:
		return fmt.Errorf("invalid platform: %s", p)
	}
}

// PaginationCursor represents the cursor used for pagination in the API.
// It encodes information about the last seen item to enable efficient pagination.
type PaginationCursor struct {
	LastID    uuid.UUID `json:"-"`
	Timestamp time.Time `json:"-"`
}

// Encode converts the cursor to a base64 string for use in API responses
func (c *PaginationCursor) Encode() string {
	data := fmt.Sprintf("%s:%d", c.LastID.String(), c.Timestamp.Unix())
	return base64.URLEncoding.EncodeToString([]byte(data))
}

// DecodeCursor creates a PaginationCursor from a base64 encoded string
func DecodeCursor(encoded string) (*PaginationCursor, error) {
	if encoded == "" {
		return nil, nil
	}

	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor encoding: %w", err)
	}

	var id uuid.UUID
	var timestamp int64
	_, err = fmt.Sscanf(string(data), "%s:%d", &id, &timestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor format: %w", err)
	}

	return &PaginationCursor{
		LastID:    id,
		Timestamp: time.Unix(timestamp, 0),
	}, nil
}

// Request/Response structures for API endpoints

// MessageCreateRequest represents the request body for creating a new message
type MessageCreateRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

// RegisterTokenRequest represents the request body for registering a new FCM token
type RegisterTokenRequest struct {
	Token      string   `json:"token" validate:"required"`
	Platform   Platform `json:"platform" validate:"required"`
	DeviceName string   `json:"device_name" validate:"required"`
}

// PaginatedMessagesResponse represents the response structure for message listing
type PaginatedMessagesResponse struct {
	Data       []Message `json:"data"`
	Pagination struct {
		NextCursor     string `json:"next_cursor,omitempty"`
		PreviousCursor string `json:"previous_cursor,omitempty"`
		HasMore        bool   `json:"has_more"`
	} `json:"pagination"`
}

// ValidationError represents a validation error in the API
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// APIError represents a standardized error response
type APIError struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Details []ValidationError `json:"details,omitempty"`
}

// MessageOptions represents query parameters for message retrieval
type MessageOptions struct {
	Limit     int    `json:"limit" validate:"required,min=1,max=100"`
	Cursor    string `json:"cursor"`
	Direction string `json:"direction" validate:"oneof=next previous"`
	Search    string `json:"search" validate:"omitempty,min=1,max=100"`
}

func (m *Message) Validate() error {
	if m.GroupID == uuid.Nil {
		return fmt.Errorf("group ID is required")
	}
	if m.SenderID == uuid.Nil {
		return fmt.Errorf("sender ID is required")
	}
	if m.Content == "" {
		return fmt.Errorf("content cannot be empty")
	}
	if len(m.Content) > 1000 {
		return fmt.Errorf("content exceeds maximum length of 1000 characters")
	}
	return nil
}

func (t *FCMToken) Validate() error {
	if t.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	if t.Token == "" {
		return fmt.Errorf("token cannot be empty")
	}
	if err := t.Platform.Validate(); err != nil {
		return err
	}
	return nil
}

// Helper methods for building responses

// NewPaginatedResponse creates a new paginated response with the given messages and cursors
func NewPaginatedResponse(messages []Message, nextCursor, prevCursor string, hasMore bool) *PaginatedMessagesResponse {
	return &PaginatedMessagesResponse{
		Data: messages,
		Pagination: struct {
			NextCursor     string `json:"next_cursor,omitempty"`
			PreviousCursor string `json:"previous_cursor,omitempty"`
			HasMore        bool   `json:"has_more"`
		}{
			NextCursor:     nextCursor,
			PreviousCursor: prevCursor,
			HasMore:        hasMore,
		},
	}
}

// NewAPIError creates a new API error response
func NewAPIError(code int, message string, details []ValidationError) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}
