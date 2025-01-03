package controllers

import (
	"Groupchat-Service/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type fcmTokenController struct {
	fcmTokenService   services.FCMTokenService
	validationService services.ValidationService
}

func NewFCMTokenController(service services.FCMTokenService, validationService services.ValidationService) FCMTokenController {
	return &fcmTokenController{fcmTokenService: service, validationService: validationService}
}

func (c *fcmTokenController) RegisterRoutes(router *gin.Engine) {
	router.POST("/groups/:groupId/users/:userId/tokens", c.SaveToken)
	router.DELETE("/groups/:groupId/users/:userId/tokens", c.DeleteToken)
}

func (c *fcmTokenController) SaveToken(ctx *gin.Context) {
	var request struct {
		Token string `json:"token" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	groupID, err := c.validationService.ValidateGroupID(ctx.Param("groupId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	userID, err := c.validationService.ValidateUserID(ctx.Param("userId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := c.fcmTokenService.SaveToken(ctx.Request.Context(), groupID, userID, request.Token); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Token saved successfully"})
}

func (c *fcmTokenController) DeleteToken(ctx *gin.Context) {
	groupID, err := c.validationService.ValidateGroupID(ctx.Param("groupId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	userID, err := c.validationService.ValidateUserID(ctx.Param("userId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := c.fcmTokenService.DeleteToken(ctx.Request.Context(), groupID, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Token deleted successfully"})
}
