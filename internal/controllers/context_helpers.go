package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getGroupIDFromContext extracts the group ID from the context
func getGroupIDFromContext(ctx *gin.Context) (uuid.UUID, error) {
	groupID, ok := ctx.Get("groupID")
	if !ok {
		return uuid.Nil, errors.New("group ID not found in context")
	}

	parsedGroupID, err := uuid.Parse(fmt.Sprintf("%v", groupID))
	if err != nil {
		return uuid.Nil, errors.New("invalid group ID")
	}

	return parsedGroupID, nil
}

// getUserFromContext extracts the user ID from the context
func getUserIDFromContext(ctx *gin.Context) (uuid.UUID, error) {
	userID, ok := ctx.Get("userID")
	if !ok {
		return uuid.Nil, errors.New("user ID not found in context")
	}

	parsedUserID, err := uuid.Parse(fmt.Sprintf("%v", userID))
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID")
	}

	return parsedUserID, nil
}

func respondWithError(ctx *gin.Context, code int, message string) {
	ctx.JSON(code, gin.H{"error": message})
}
