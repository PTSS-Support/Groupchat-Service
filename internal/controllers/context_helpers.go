package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getGroupIDFromContext extracts the group ID from the context
func getGroupIDFromContext(ctx *gin.Context) (uuid.UUID, error) {
	groupID, exists := ctx.Get("groupID")
	if !exists {
		return uuid.Nil, errors.New("group ID not found in context")
	}

	parsedGroupID, err := uuid.Parse(groupID.(string))
	if err != nil {
		return uuid.Nil, errors.New("invalid group ID")
	}

	return parsedGroupID, nil
}

// getUserIDFromContext extracts the user ID from the context
func getUserIDFromContext(ctx *gin.Context) (uuid.UUID, error) {
	userID, exists := ctx.Get("userID")
	if !exists {
		return uuid.Nil, errors.New("user ID not found in context")
	}

	parsedUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID")
	}

	return parsedUserID, nil
}

func respondWithError(ctx *gin.Context, code int, message string) {
	ctx.JSON(code, gin.H{"error": message})
}
