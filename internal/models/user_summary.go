package models

import "github.com/google/uuid"

type UserSummary struct {
	ID       uuid.UUID `json:"id"`
	UserName string    `json:"userName"`
}
