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

type mockMessageService struct {
	mock.Mock
}

func (m *mockMessageService) GetMessages(ctx context.Context, groupID uuid.UUID, query models.PaginationQuery) ([]models.MessageResponse, *models.PaginationResponse, error) {
	args := m.Called(ctx, groupID, query)
	return args.Get(0).([]models.MessageResponse), args.Get(1).(*models.PaginationResponse), args.Error(2)
}

func (m *mockMessageService) CreateMessage(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, userName string, create models.MessageCreate) (*models.Message, error) {
	args := m.Called(ctx, groupID, userID, userName, create)
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *mockMessageService) ToggleMessagePin(ctx context.Context, groupID uuid.UUID, messageID uuid.UUID) (*models.Message, error) {
	args := m.Called(ctx, groupID, messageID)
	return args.Get(0).(*models.Message), args.Error(1)
}

func setupMessageController() (*FCMMessageController, *mockMessageService, *mockValidationService) {
	mockMsgService := new(mockMessageService)
	mockValidation := new(mockValidationService)
	controller := NewMessageController(mockMsgService, mockValidation)
	return controller, mockMsgService, mockValidation
}

func TestGetMessages(t *testing.T) {
	t.Run("Successfully get messages with pagination", func(t *testing.T) {
		controller, mockMsgService, mockValidation := setupMessageController()
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		groupID := uuid.New()
		ctx.Set("groupID", groupID.String())

		ctx.Request = httptest.NewRequest("GET", "/?pageSize=10&direction=next", nil)

		expectedQuery := models.PaginationQuery{
			PageSize:  10,
			Direction: models.Next,
		}

		mockValidation.On("ValidatePaginationQuery", mock.Anything, mock.Anything).
			Return(expectedQuery, nil)

		messages := []models.MessageResponse{{ID: uuid.New()}}
		pagination := &models.PaginationResponse{HasNext: true}

		mockMsgService.On("GetMessages", mock.Anything, groupID, expectedQuery).
			Return(messages, pagination, nil)

		controller.GetMessages(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			return
		}
		assert.NotNil(t, response["data"])
		assert.NotNil(t, response["pagination"])
	})

	t.Run("Invalid pagination parameters", func(t *testing.T) {
		controller, _, mockValidation := setupMessageController()
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		groupID := uuid.New()
		ctx.Set("groupID", groupID.String())
		ctx.Request = httptest.NewRequest("GET", "/?pageSize=invalid", nil)

		mockValidation.On("ValidatePaginationQuery", mock.Anything, mock.Anything).
			Return(models.PaginationQuery{}, errors.New("invalid pagination"))

		controller.GetMessages(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCreateMessage(t *testing.T) {
	t.Run("Successfully create message", func(t *testing.T) {
		controller, mockMsgService, mockValidation := setupMessageController()
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		groupID := uuid.New()
		userID := uuid.New()
		ctx.Set("groupID", groupID.String())
		ctx.Set("userID", userID.String())

		createReq := models.MessageCreate{Content: "test message"}
		jsonBody, _ := json.Marshal(createReq)
		ctx.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		mockValidation.On("FetchUserName", mock.Anything).
			Return("testUser", nil)

		expectedMsg := &models.Message{ID: uuid.New(), Content: createReq.Content}
		mockMsgService.On("CreateMessage", mock.Anything, groupID, userID, "testUser", createReq).
			Return(expectedMsg, nil)

		controller.CreateMessage(ctx)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response models.Message
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			return
		}
		assert.Equal(t, expectedMsg.ID, response.ID)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		controller, _, _ := setupMessageController()
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		ctx.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte("invalid json")))
		ctx.Request.Header.Set("Content-Type", "application/json")

		controller.CreateMessage(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestToggleMessagePin(t *testing.T) {
	t.Run("Successfully toggle message pin", func(t *testing.T) {
		controller, mockMsgService, _ := setupMessageController()
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		groupID := uuid.New()
		messageID := uuid.New()
		ctx.Set("groupID", groupID.String())
		ctx.AddParam("messageId", messageID.String())

		ctx.Request = httptest.NewRequest("PUT", "/", nil)

		expectedMsg := &models.Message{ID: messageID, IsPinned: true}
		mockMsgService.On("ToggleMessagePin", mock.Anything, groupID, messageID).
			Return(expectedMsg, nil)

		controller.ToggleMessagePin(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.Message
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			return
		}
		assert.Equal(t, expectedMsg.ID, response.ID)
		assert.Equal(t, expectedMsg.IsPinned, response.IsPinned)
	})

	t.Run("Invalid message ID", func(t *testing.T) {
		controller, _, _ := setupMessageController()
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		ctx.AddParam("messageId", "invalid-uuid")
		ctx.Request = httptest.NewRequest("PUT", "/", nil)

		controller.ToggleMessagePin(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
