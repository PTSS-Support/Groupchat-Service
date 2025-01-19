package services

import (
	"Groupchat-Service/internal/util"
	"context"

	"Groupchat-Service/internal/database/repositories"
	"Groupchat-Service/internal/models"
)

type healthService struct {
	healthRepo repositories.HealthRepository
	logger     util.Logger
}

func NewHealthService(healthRepo repositories.HealthRepository, loggerFactory util.LoggerFactory) HealthService {
	return &healthService{
		healthRepo: healthRepo,
		logger:     loggerFactory.NewLogger("HealthService"),
	}
}

func (s *healthService) CheckHealth(ctx context.Context) (*models.HealthResponse, error) {
	// Get liveness check
	livenessResp, err := s.CheckLiveness(ctx)
	if err != nil {
		return nil, err
	}

	// Get readiness check
	readinessResp, err := s.CheckReadiness(ctx)
	if err != nil {
		return nil, err
	}

	// Combine the checks
	combinedResponse := &models.HealthResponse{
		Status: models.StatusUp,
		Checks: append(livenessResp.Checks, readinessResp.Checks...),
	}

	// If either check is down, mark overall status as down
	if livenessResp.Status == models.StatusDown || readinessResp.Status == models.StatusDown {
		combinedResponse.Status = models.StatusDown
	}

	return combinedResponse, nil
}

func (s *healthService) CheckReadiness(ctx context.Context) (*models.HealthResponse, error) {
	userServiceHealth, err := s.healthRepo.CheckHealth(ctx)

	if err != nil {
		return &models.HealthResponse{
			Status: models.StatusDown,
			Checks: []models.Check{
				{
					Name:   "User service health check",
					Status: models.StatusDown,
					Data: map[string]interface{}{
						"error": err.Error(),
					},
				},
			},
		}, nil
	}

	return &models.HealthResponse{
		Status: models.StatusUp,
		Checks: []models.Check{
			{
				Name:   "User service health check",
				Status: models.StatusUp,
				Data: map[string]interface{}{
					"response": userServiceHealth,
				},
			},
		},
	}, nil
}

func (s *healthService) CheckLiveness(ctx context.Context) (*models.HealthResponse, error) {
	return &models.HealthResponse{
		Status: models.StatusUp,
		Checks: []models.Check{},
	}, nil
}
