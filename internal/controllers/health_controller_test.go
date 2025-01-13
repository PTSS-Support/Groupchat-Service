package controllers

import (
	"Groupchat-Service/internal/models"
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http/httptest"
	"testing"
)

type mockHealthService struct {
	mock.Mock
}

func (m *mockHealthService) CheckHealth(ctx context.Context) (*models.HealthResponse, error) {
	args := m.Called(ctx)
	if resp := args.Get(0); resp != nil {
		return resp.(*models.HealthResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockHealthService) CheckLiveness(ctx context.Context) (*models.HealthResponse, error) {
	args := m.Called(ctx)
	if resp := args.Get(0); resp != nil {
		return resp.(*models.HealthResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockHealthService) CheckReadiness(ctx context.Context) (*models.HealthResponse, error) {
	args := m.Called(ctx)
	if resp := args.Get(0); resp != nil {
		return resp.(*models.HealthResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestHealthController(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mockHealthService)
		checkFunc      func(*healthController, *gin.Context)
		expectedStatus int
		expectedState  models.HealthStatus
	}{
		{
			name: "Health check success",
			setupMock: func(m *mockHealthService) {
				m.On("CheckHealth", mock.Anything).Return(&models.HealthResponse{
					Status: models.StatusUp,
					Checks: []models.Check{{Name: "test", Status: models.StatusUp}},
				}, nil)
			},
			checkFunc: func(c *healthController, ctx *gin.Context) {
				c.getHealth(ctx)
			},
			expectedStatus: 200,
			expectedState:  models.StatusUp,
		},
		{
			name: "Liveness check failure",
			setupMock: func(m *mockHealthService) {
				m.On("CheckLiveness", mock.Anything).Return(nil, errors.New("liveness check failed"))
			},
			checkFunc: func(c *healthController, ctx *gin.Context) {
				c.getLiveness(ctx)
			},
			expectedStatus: 503,
			expectedState:  models.StatusDown,
		},
		{
			name: "Readiness check degraded",
			setupMock: func(m *mockHealthService) {
				m.On("CheckReadiness", mock.Anything).Return(&models.HealthResponse{
					Status: models.StatusDown,
					Checks: []models.Check{{Name: "database", Status: models.StatusDown}},
				}, nil)
			},
			checkFunc: func(c *healthController, ctx *gin.Context) {
				c.getReadiness(ctx)
			},
			expectedStatus: 503,
			expectedState:  models.StatusDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			mockService := new(mockHealthService)
			tt.setupMock(mockService)

			controller := NewHealthController(mockService).(*healthController)
			tt.checkFunc(controller, ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			var response models.HealthResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedState, response.Status)
			mockService.AssertExpectations(t)
		})
	}
}
