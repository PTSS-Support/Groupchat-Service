package models

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type Direction string

const (
	Next     Direction = "next"
	Previous Direction = "previous"
)

type Message struct {
	ID         uuid.UUID `json:"id"`
	GroupID    uuid.UUID `json:"groupId"`
	SenderID   uuid.UUID `json:"senderId"`
	SenderName string    `json:"senderName"`
	Content    string    `json:"content"`
	SentAt     time.Time `json:"sentAt"`
	IsPinned   bool      `json:"isPinned"`
}

type MessageCreate struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

type PaginationQuery struct {
	Cursor    *string   `json:"cursor,omitempty"`
	PageSize  int       `json:"pageSize"`
	Direction Direction `json:"direction"`
	Search    *string   `json:"search,omitempty"`
}

type PaginatedResponse struct {
	Data       interface{}        `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

type PaginationResponse struct {
	NextCursor     *string `json:"nextCursor,omitempty"`
	PreviousCursor *string `json:"previousCursor,omitempty"`
	HasNext        bool    `json:"hasNext"`
	HasPrevious    bool    `json:"hasPrevious"`
}

type FCMToken struct {
	PartitionKey string                 `json:"PartitionKey"` // GroupID
	RowKey       string                 `json:"RowKey"`       // UserID
	Token        string                 `json:"Token"`
	IsActive     bool                   `json:"IsActive"`
	Timestamp    *timestamppb.Timestamp `json:"Timestamp"`
}
