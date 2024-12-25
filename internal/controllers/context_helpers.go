package controllers

import (
	"context"
	"errors"
	"github.com/google/uuid"
)

// getGroupIDFromContext extracts the group ID from the context
func getGroupIDFromContext(ctx context.Context) (uuid.UUID, error) {
	groupID, ok := ctx.Value("groupID").(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("group ID not found in context")
	}
	return groupID, nil
}

// getUserFromContext extracts the user ID and user name from the context
func getUserFromContext(ctx context.Context) (uuid.UUID, string, error) {
	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		return uuid.Nil, "", errors.New("user ID not found in context")
	}
	userName, ok := ctx.Value("userName").(string)
	if !ok {
		return uuid.Nil, "", errors.New("user name not found in context")
	}
	return userID, userName, nil
}
