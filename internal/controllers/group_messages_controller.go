package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"Groupchat-Service/internal/models"
	"Groupchat-Service/internal/services"
)

type FCMMessageController struct {
	messageService services.MessageService
}

func NewMessageController(messageService services.MessageService) *FCMMessageController {
	return &FCMMessageController{
		messageService: messageService,
	}
}

func (c *FCMMessageController) GetMessages(ctx *gin.Context) {
	// Extract group ID from URL
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	// Parse pagination query
	query := models.PaginationQuery{
		PageSize:  10, // Default page size
		Direction: models.Next,
	}

	if size := ctx.Query("pageSize"); size != "" {
		pageSize, err := strconv.Atoi(size)
		if err != nil || pageSize < 1 || pageSize > 50 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "page size must be between 1 and 50"})
			return
		}
		query.PageSize = pageSize
	}

	if cursor := ctx.Query("cursor"); cursor != "" {
		query.Cursor = &cursor
	}

	if direction := ctx.Query("direction"); direction != "" {
		if direction != "next" && direction != "previous" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "direction must be 'next' or 'previous'"})
			return
		}
		query.Direction = models.Direction(direction)
	}

	if search := ctx.Query("search"); search != "" {
		if len(search) > 100 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "search term too long: maximum 100 characters"})
			return
		}
		query.Search = &search
	}

	// Get messages from service
	messages, pagination, err := c.messageService.GetMessages(ctx, groupID, query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving messages"})
		return
	}

	// Format response
	response := models.PaginatedResponse{
		Data:       messages,
		Pagination: *pagination,
	}

	ctx.JSON(http.StatusOK, response)
}

// CreateMessage handles POST /groups/messages
func (c *FCMMessageController) CreateMessage(ctx *gin.Context) {
	// Get user info from context
	userID, userName, err := getUserFromContext(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user context"})
		return
	}

	groupID, err := getGroupIDFromContext(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid group context"})
		return
	}

	// Parse request body
	var createReq models.MessageCreate
	if err := ctx.ShouldBindJSON(&createReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Call service to create message
	message, err := c.messageService.CreateMessage(ctx.Request.Context(), groupID, userID, userName, createReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating message"})
		return
	}

	ctx.JSON(http.StatusCreated, message)
}

// ToggleMessagePin handles PUT /groups/messages/{messageId}/pin
func (c *FCMMessageController) ToggleMessagePin(ctx *gin.Context) {
	// Extract message ID from URL
	messageID, err := uuid.Parse(ctx.Param("messageId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	// Call service to toggle pin
	message, err := c.messageService.ToggleMessagePin(ctx.Request.Context(), messageID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error toggling message pin"})
		return
	}

	ctx.JSON(http.StatusOK, message)
}
