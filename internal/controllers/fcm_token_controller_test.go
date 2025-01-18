package controllers

import (
	"Groupchat-Service/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock services
type mockFCMTokenService struct {
	mock.Mock
}

func (m *mockFCMTokenService) SaveToken(ctx context.Context, groupID, userID uuid.UUID, token string) error {
	args := m.Called(ctx, groupID, userID, token)
	return args.Error(0)
}

func (m *mockFCMTokenService) DeleteToken(ctx context.Context, groupID, userID uuid.UUID) error {
	args := m.Called(ctx, groupID, userID)
	return args.Error(0)
}

type mockValidationService struct {
	mock.Mock
}

func (m *mockValidationService) ValidatePaginationQuery(queryParams map[string]string) (models.PaginationQuery, error) {
	args := m.Called(queryParams)
	return args.Get(0).(models.PaginationQuery), args.Error(1)
}

func (m *mockValidationService) ValidateUserContext(ctx context.Context) (uuid.UUID, string, error) {
	args := m.Called(ctx)
	return args.Get(0).(uuid.UUID), args.String(1), args.Error(2)
}

func (m *mockValidationService) ValidateGroupID(groupID string) (uuid.UUID, error) {
	args := m.Called(groupID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockValidationService) ValidateUserID(userID string) (uuid.UUID, error) {
	args := m.Called(userID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockValidationService) ValidateToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *mockValidationService) FetchUserName(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *mockValidationService) FetchGroupMembers(ctx context.Context, groupID uuid.UUID) ([]models.UserSummary, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0).([]models.UserSummary), args.Error(1)
}

func setupTestController() (*fcmTokenController, *mockFCMTokenService, *mockValidationService) {
	mockFCMService := new(mockFCMTokenService)
	mockValidation := new(mockValidationService)
	controller := NewFCMTokenController(mockFCMService, mockValidation).(*fcmTokenController)
	return controller, mockFCMService, mockValidation
}

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	// Set up common context values
	groupID := uuid.New()
	userID := uuid.New()
	ctx.Set("groupID", groupID.String())
	ctx.Set("userID", userID.String())

	return ctx, w
}

func TestSaveToken(t *testing.T) {
	t.Run("Successfully saves token", func(t *testing.T) {
		// Set up controller and mocks
		controller, mockFCMService, mockValidation := setupTestController()
		ctx, w := setupTestContext()

		// Prepare request body
		token := "valid-fcm-token"
		body := gin.H{"token": token}
		jsonBody, _ := json.Marshal(body)
		ctx.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// Set up mock expectations
		mockValidation.On("ValidateToken", token).Return(nil)
		mockFCMService.On("SaveToken", mock.Anything, mock.Anything, mock.Anything, token).Return(nil)

		// Execute the handler
		controller.SaveToken(ctx)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Token saved successfully", response["message"])
		mockValidation.AssertExpectations(t)
		mockFCMService.AssertExpectations(t)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		controller, _, _ := setupTestController()
		ctx, w := setupTestContext()

		// Empty request body
		ctx.Request = httptest.NewRequest("POST", "/", nil)
		ctx.Request.Header.Set("Content-Type", "application/json")

		controller.SaveToken(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request body", response["error"])
	})

	t.Run("Token validation fails", func(t *testing.T) {
		controller, _, mockValidation := setupTestController()
		ctx, w := setupTestContext()

		// Prepare request with invalid token
		token := "invalid-token"
		body := gin.H{"token": token}
		jsonBody, _ := json.Marshal(body)
		ctx.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		mockValidation.On("ValidateToken", token).Return(errors.New("invalid token format"))

		controller.SaveToken(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "invalid token format", response["error"])
		mockValidation.AssertExpectations(t)
	})

	t.Run("Service fails to save token", func(t *testing.T) {
		controller, mockFCMService, mockValidation := setupTestController()
		ctx, w := setupTestContext()

		token := "valid-token"
		body := gin.H{"token": token}
		jsonBody, _ := json.Marshal(body)
		ctx.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		mockValidation.On("ValidateToken", token).Return(nil)
		mockFCMService.On("SaveToken", mock.Anything, mock.Anything, mock.Anything, token).
			Return(errors.New("database error"))

		controller.SaveToken(ctx)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to save token", response["error"])
		mockValidation.AssertExpectations(t)
		mockFCMService.AssertExpectations(t)
	})
}

func TestDeleteToken(t *testing.T) {
	t.Run("Successfully deletes token", func(t *testing.T) {
		controller, mockFCMService, _ := setupTestController()
		ctx, w := setupTestContext()

		ctx.Request = httptest.NewRequest("DELETE", "/", nil)

		mockFCMService.On("DeleteToken", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		controller.DeleteToken(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Token deleted successfully", response["message"])
		mockFCMService.AssertExpectations(t)
	})

	t.Run("Service fails to delete token", func(t *testing.T) {
		controller, mockFCMService, _ := setupTestController()
		ctx, w := setupTestContext()

		ctx.Request = httptest.NewRequest("DELETE", "/", nil)

		mockFCMService.On("DeleteToken", mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("database error"))

		controller.DeleteToken(ctx)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to delete token", response["error"])
		mockFCMService.AssertExpectations(t)
	})
}
