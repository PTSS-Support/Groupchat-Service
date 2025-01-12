package controllers

import (
	"Groupchat-Service/internal/models"
	"Groupchat-Service/internal/services"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type healthCheck func(ctx context.Context) (*models.HealthResponse, error)

type healthController struct {
	healthService services.HealthService
}

func NewHealthController(healthService services.HealthService) HealthController {
	return &healthController{
		healthService: healthService,
	}
}

func (c *healthController) RegisterRoutes(router *gin.Engine) {
	router.GET("/q/health", c.getHealth)
	router.GET("/q/health/live", c.getLiveness)
	router.GET("/q/health/ready", c.getReadiness)
}

func (c *healthController) getHealth(ctx *gin.Context) {
	c.handleHealthCheck(ctx, c.healthService.CheckHealth)
}

func (c *healthController) getLiveness(ctx *gin.Context) {
	c.handleHealthCheck(ctx, c.healthService.CheckLiveness)
}

func (c *healthController) getReadiness(ctx *gin.Context) {
	c.handleHealthCheck(ctx, c.healthService.CheckReadiness)
}

func (c *healthController) handleHealthCheck(ctx *gin.Context, check healthCheck) {
	response, err := check(ctx)
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, models.HealthResponse{
			Status: models.StatusDown,
			Checks: []models.Check{{
				Name:   "Health check error",
				Status: models.StatusDown,
				Data: map[string]interface{}{
					"error": err.Error(),
				},
			}},
		})
		return
	}

	statusCode := http.StatusOK
	if response.Status == models.StatusDown {
		statusCode = http.StatusServiceUnavailable
	}
	ctx.JSON(statusCode, response)
}
