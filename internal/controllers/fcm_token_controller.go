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
		respondWithError(ctx, http.StatusBadRequest, "Invalid request body")
		return
	}

	groupID, err := getGroupIDFromContext(ctx)
	if err != nil {
		respondWithError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		respondWithError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	if err := c.validationService.ValidateToken(request.Token); err != nil {
		respondWithError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.fcmTokenService.SaveToken(ctx.Request.Context(), groupID, userID, request.Token); err != nil {
		respondWithError(ctx, http.StatusInternalServerError, "Failed to save token")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Token saved successfully"})
}

func (c *fcmTokenController) DeleteToken(ctx *gin.Context) {
	groupID, err := getGroupIDFromContext(ctx)
	if err != nil {
		respondWithError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		respondWithError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	if err := c.fcmTokenService.DeleteToken(ctx.Request.Context(), groupID, userID); err != nil {
		respondWithError(ctx, http.StatusInternalServerError, "Failed to delete token")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Token deleted successfully"})
}
