package controllers

import (
	"Groupchat-Service/internal/middleware"
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
	router.GET("/groups/messages",
		middleware.RequireRoles(models.RolePatient, models.RolePrimaryCaregiver, models.RoleFamilyMember),
		c.GetMessages)

	router.POST("/groups/messages",
		middleware.RequireRoles(models.RolePatient, models.RolePrimaryCaregiver, models.RoleFamilyMember),
		c.CreateMessage)

	router.PUT("/groups/messages/:messageId/pin",
		middleware.RequireRoles(models.RolePatient, models.RolePrimaryCaregiver, models.RoleFamilyMember),
		c.ToggleMessagePin)
}

func (c *FCMMessageController) GetMessages(ctx *gin.Context) {
	groupID, err := getGroupIDFromContext(ctx)
	if err != nil {
		respondWithError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	queryParams := map[string]string{
		"pageSize":  ctx.Query("pageSize"),
		"cursor":    ctx.Query("cursor"),
		"direction": ctx.Query("direction"),
		"search":    ctx.Query("search"),
	}

	query, err := c.validationService.ValidatePaginationQuery(queryParams)
	if err != nil {
		respondWithError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	messages, pagination, err := c.messageService.GetMessages(ctx.Request.Context(), groupID, query)
	if err != nil {
		respondWithError(ctx, http.StatusInternalServerError, "Error getting messages")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":       messages,
		"pagination": pagination,
	})
}

func (c *FCMMessageController) CreateMessage(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		respondWithError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	groupID, err := getGroupIDFromContext(ctx)
	if err != nil {
		respondWithError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	userName, err := c.validationService.FetchUserName(ctx.Request.Context())
	if err != nil {
		respondWithError(ctx, http.StatusInternalServerError, "Error fetching user name")
		return
	}

	var createReq models.MessageCreate
	if err := ctx.ShouldBindJSON(&createReq); err != nil {
		respondWithError(ctx, http.StatusBadRequest, "Invalid request body")
		return
	}

	message, err := c.messageService.CreateMessage(ctx.Request.Context(), groupID, userID, userName, createReq)
	if err != nil {
		respondWithError(ctx, http.StatusInternalServerError, "Error creating message")
		return
	}

	ctx.JSON(http.StatusCreated, message)
}

func (c *FCMMessageController) ToggleMessagePin(ctx *gin.Context) {
	groupID, err := getGroupIDFromContext(ctx)
	if err != nil {
		respondWithError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	messageID, err := uuid.Parse(ctx.Param("messageId"))
	if err != nil {
		respondWithError(ctx, http.StatusBadRequest, "Invalid message ID")
		return
	}

	message, err := c.messageService.ToggleMessagePin(ctx.Request.Context(), groupID, messageID)
	if err != nil {
		respondWithError(ctx, http.StatusInternalServerError, "Error toggling message pin")
		return
	}

	ctx.JSON(http.StatusOK, message)
}
