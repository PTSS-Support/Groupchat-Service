package repositories

import (
	"Groupchat-Service/internal/models"
	"Groupchat-Service/internal/util"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type healthRepository struct {
	baseURL    string
	httpClient *http.Client
	logger     util.Logger
}

func NewHealthRepository(baseURL string, loggerFactory util.LoggerFactory) HealthRepository {
	return &healthRepository{
		baseURL:    baseURL,
		httpClient: &http.Client{},
		logger:     loggerFactory.NewLogger("HealthRepository"),
	}
}

func (r *healthRepository) CheckHealth(ctx context.Context) (*models.HealthResponse, error) {
	log := r.logger.WithContext(ctx)
	healthURL := fmt.Sprintf("%s/q/health/ready", r.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		log.Error("Failed to create HTTP request", "error", err)
		return &models.HealthResponse{
			Status: models.StatusDown,
			Checks: []models.Check{{
				Name:   "Health check",
				Status: models.StatusDown,
				Data: map[string]interface{}{
					"error": err.Error(),
				},
			}},
		}, nil
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		log.Error("HTTP request failed", "error", err)
		return &models.HealthResponse{
			Status: models.StatusDown,
			Checks: []models.Check{{
				Name:   "Health check",
				Status: models.StatusDown,
				Data: map[string]interface{}{
					"error": err.Error(),
				},
			}},
		}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read response body", "error", err)
		return &models.HealthResponse{
			Status: models.StatusDown,
			Checks: []models.Check{{
				Name:   "Health check",
				Status: models.StatusDown,
				Data: map[string]interface{}{
					"error": fmt.Sprintf("failed to read response: %v", err),
				},
			}},
		}, nil
	}

	var healthResponse models.HealthResponse
	if err := json.Unmarshal(body, &healthResponse); err != nil {
		log.Error("Failed to parse response", "error", err)
		return &models.HealthResponse{
			Status: models.StatusDown,
			Checks: []models.Check{{
				Name:   "Health check",
				Status: models.StatusDown,
				Data: map[string]interface{}{
					"error": fmt.Sprintf("failed to parse response: %v", err),
				},
			}},
		}, nil
	}

	return &healthResponse, nil
}
