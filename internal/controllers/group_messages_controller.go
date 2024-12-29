package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/google/uuid"

	"Groupchat-Service/internal/models"
	"Groupchat-Service/internal/services"
)

type FCMMessageController struct {
	messageService    services.MessageService
	validationService services.ValidationService
}

func NewMessageController(messageService services.MessageService, validationService services.ValidationService) *FCMMessageController {
	return &FCMMessageController{
		messageService:    messageService,
		validationService: validationService,
	}
}

func (c *FCMMessageController) RegisterRoutes(router *gin.Engine) {
	router.GET("/groups/:groupId/messages", c.GetMessages)
	router.POST("/groups/:groupId/messages", c.CreateMessage)
	router.PUT("/groups/:groupId/messages/:messageId/pin", c.ToggleMessagePin)
}

func (c *FCMMessageController) GetMessages(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	queryParams := map[string]string{
		"pageSize":  ctx.Query("pageSize"),
		"cursor":    ctx.Query("cursor"),
		"direction": ctx.Query("direction"),
		"search":    ctx.Query("search"),
	}

	query, err := c.validationService.ValidatePaginationQuery(ctx.Request.Context(), queryParams)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	messages, pagination, err := c.messageService.GetMessages(ctx, groupID, query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving messages"})
		return
	}

	response := models.PaginatedResponse{
		Data:       messages,
		Pagination: *pagination,
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *FCMMessageController) CreateMessage(ctx *gin.Context) {
	userID, userName, err := c.validationService.ValidateUserContext(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	groupID, err := c.validationService.ValidateGroupID(ctx.Param("groupId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var createReq models.MessageCreate
	if err := ctx.ShouldBindJSON(&createReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	message, err := c.messageService.CreateMessage(ctx.Request.Context(), groupID, userID, userName, createReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating message"})
		return
	}

	ctx.JSON(http.StatusCreated, message)
}

func (c *FCMMessageController) ToggleMessagePin(ctx *gin.Context) {
	messageID, err := uuid.Parse(ctx.Param("messageId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	message, err := c.messageService.ToggleMessagePin(ctx.Request.Context(), messageID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error toggling message pin"})
		return
	}

	ctx.JSON(http.StatusOK, message)
}
