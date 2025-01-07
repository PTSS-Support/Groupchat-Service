package models

import (
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FCMToken struct {
	PartitionKey string                 `json:"PartitionKey"` // GroupID
	RowKey       string                 `json:"RowKey"`       // UserID
	Token        string                 `json:"Token"`
	IsActive     bool                   `json:"IsActive"`
	Timestamp    *timestamppb.Timestamp `json:"Timestamp"`
}
