package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type MessageController interface {
	RegisterRoutes(router *gin.Engine)
	GetMessages(w http.ResponseWriter, r *http.Request)
	CreateMessage(w http.ResponseWriter, r *http.Request)
	ToggleMessagePin(w http.ResponseWriter, r *http.Request)
}

type FCMTokenController interface {
	RegisterRoutes(router *gin.Engine)
	SaveToken(ctx *gin.Context)
	DeleteToken(ctx *gin.Context)
}

type HealthController interface {
	RegisterRoutes(router *gin.Engine)
	getHealth(ctx *gin.Context)
	getLiveness(ctx *gin.Context)
	getReadiness(ctx *gin.Context)
	handleHealthCheck(ctx *gin.Context, check healthCheck)
}
