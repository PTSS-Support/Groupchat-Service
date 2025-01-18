package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"Groupchat-Service/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories and services
type MockMessageRepository struct {
	mock.Mock
}

type MockFCMTokenRepository struct {
	mock.Mock
}

type MockNotificationService struct {
	mock.Mock
}

type MockValidationService struct {
	mock.Mock
}

// Implement interface methods for mocks
func (m *MockMessageRepository) GetMessages(ctx context.Context, groupID uuid.UUID, query models.PaginationQuery) ([]models.Message, *models.PaginationResponse, error) {
	args := m.Called(ctx, groupID, query)
	messages := args.Get(0)
	if messages == nil {
		return nil, args.Get(1).(*models.PaginationResponse), args.Error(2)
	}
	return messages.([]models.Message), args.Get(1).(*models.PaginationResponse), args.Error(2)
}

func (m *MockMessageRepository) CreateMessage(ctx context.Context, groupID uuid.UUID, message *models.Message) error {
	args := m.Called(ctx, groupID, message)
	return args.Error(0)
}

func (m *MockMessageRepository) GetMessageByID(ctx context.Context, groupID uuid.UUID, messageID uuid.UUID) (*models.Message, error) {
	args := m.Called(ctx, groupID, messageID)
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *MockMessageRepository) ToggleMessagePin(ctx context.Context, groupID uuid.UUID, messageID uuid.UUID) (*models.Message, error) {
	args := m.Called(ctx, groupID, messageID)
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *MockMessageRepository) GetLastReadTime(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) (time.Time, error) {
	args := m.Called(ctx, groupID, userID)
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *MockMessageRepository) CountUnreadMessages(ctx context.Context, groupID uuid.UUID, lastReadTime time.Time) (int, error) {
	args := m.Called(ctx, groupID, lastReadTime)
	return args.Int(0), args.Error(1)
}

func (m *MockFCMTokenRepository) GetGroupMemberTokens(ctx context.Context, groupID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockFCMTokenRepository) SaveToken(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, token string) error {
	args := m.Called(ctx, groupID, userID, token)
	return args.Error(0)
}

func (m *MockFCMTokenRepository) DeleteToken(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, groupID, userID)
	return args.Error(0)
}

func (m *MockValidationService) ValidatePaginationQuery(queryParams map[string]string) (models.PaginationQuery, error) {
	args := m.Called(queryParams)
	return args.Get(0).(models.PaginationQuery), args.Error(1)
}

func (m *MockValidationService) ValidateUserContext(ctx context.Context) (uuid.UUID, string, error) {
	args := m.Called(ctx)
	return args.Get(0).(uuid.UUID), args.Get(1).(string), args.Error(2)
}

func (m *MockValidationService) ValidateGroupID(groupID string) (uuid.UUID, error) {
	args := m.Called(groupID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockValidationService) ValidateUserID(userID string) (uuid.UUID, error) {
	args := m.Called(userID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockValidationService) ValidateToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockNotificationService) SendGroupMessage(message Message, deviceTokens []string) (*BatchResponse, error) {
	args := m.Called(message, deviceTokens)
	return args.Get(0).(*BatchResponse), args.Error(1)
}

func (m *MockValidationService) FetchUserName(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockValidationService) FetchGroupMembers(ctx context.Context, groupID uuid.UUID) ([]models.UserSummary, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0).([]models.UserSummary), args.Error(1)
}

func TestGetMessages(t *testing.T) {
	ctx := context.Background()
	groupID := uuid.New()
	query := models.PaginationQuery{
		PageSize:  10,
		Direction: models.Next,
	}

	tests := []struct {
		name           string
		setupMocks     func(*MockMessageRepository, *MockValidationService)
		expectedErr    error
		expectedCount  int
		expectedSender string
	}{
		{
			name: "Success",
			setupMocks: func(mr *MockMessageRepository, vs *MockValidationService) {
				messages := []models.Message{
					{
						ID:       uuid.New(),
						GroupID:  groupID,
						SenderID: uuid.New(),
						Content:  "Test message",
						SentAt:   time.Now(),
					},
				}
				pagination := &models.PaginationResponse{HasNext: false}
				mr.On("GetMessages", ctx, groupID, query).Return(messages, pagination, nil)

				members := []models.UserSummary{{ID: messages[0].SenderID, UserName: "TestUser"}}
				vs.On("FetchGroupMembers", ctx, groupID).Return(members, nil)
			},
			expectedCount:  1,
			expectedSender: "TestUser",
		},
		{
			name: "Repository Error",
			setupMocks: func(mr *MockMessageRepository, vs *MockValidationService) {
				mr.On("GetMessages", ctx, groupID, query).Return(nil, &models.PaginationResponse{}, errors.New("db error"))
			},
			expectedErr: errors.New("error getting messages from repository: db error"),
		},
		{
			name: "Validation Service Error",
			setupMocks: func(mr *MockMessageRepository, vs *MockValidationService) {
				messages := []models.Message{{}}
				pagination := &models.PaginationResponse{}
				mr.On("GetMessages", ctx, groupID, query).Return(messages, pagination, nil)
				vs.On("FetchGroupMembers", ctx, groupID).Return([]models.UserSummary{}, errors.New("validation error"))
			},
			expectedErr: errors.New("error fetching group members: validation error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMsgRepo := new(MockMessageRepository)
			mockFCMRepo := new(MockFCMTokenRepository)
			mockNotifService := new(MockNotificationService)
			mockValidService := new(MockValidationService)

			tt.setupMocks(mockMsgRepo, mockValidService)

			service := NewMessageService(mockMsgRepo, mockFCMRepo, mockNotifService, mockValidService)
			messages, _, err := service.GetMessages(ctx, groupID, query)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Len(t, messages, tt.expectedCount)
			if tt.expectedCount > 0 {
				assert.Equal(t, tt.expectedSender, messages[0].SenderName)
			}

			mockMsgRepo.AssertExpectations(t)
			mockValidService.AssertExpectations(t)
		})
	}
}

func TestCreateMessage(t *testing.T) {
	ctx := context.Background()
	groupID := uuid.New()
	userID := uuid.New()
	userName := "TestUser"
	create := models.MessageCreate{Content: "Test message"}

	tests := []struct {
		name       string
		setupMocks func(*MockMessageRepository, *MockFCMTokenRepository, *MockNotificationService, chan bool)
		wantErr    bool
		expectSend bool
	}{
		{
			name: "Success",
			setupMocks: func(mockMsgRepo *MockMessageRepository, mockFCMRepo *MockFCMTokenRepository, mockNotifService *MockNotificationService, ch chan bool) {
				mockMsgRepo.On("CreateMessage", mock.Anything, groupID, mock.Anything).Return(nil)
				mockFCMRepo.On("GetGroupMemberTokens", mock.Anything, groupID).Return([]string{"token1", "token2"}, nil)
				mockNotifService.On("SendGroupMessage", mock.Anything, []string{"token1", "token2"}).Return(&BatchResponse{}, nil).Run(func(args mock.Arguments) {
					ch <- true
				})
			},
			wantErr:    false,
			expectSend: true,
		},
		{
			name: "Repository Error",
			setupMocks: func(mockMsgRepo *MockMessageRepository, mockFCMRepo *MockFCMTokenRepository, mockNotifService *MockNotificationService, ch chan bool) {
				mockMsgRepo.On("CreateMessage", mock.Anything, groupID, mock.Anything).Return(errors.New("repository error"))
			},
			wantErr:    true,
			expectSend: false,
		},
		{
			name: "FCM Token Error - Continues Successfully",
			setupMocks: func(mockMsgRepo *MockMessageRepository, mockFCMRepo *MockFCMTokenRepository, mockNotifService *MockNotificationService, ch chan bool) {
				mockMsgRepo.On("CreateMessage", mock.Anything, groupID, mock.Anything).Return(nil)
				mockFCMRepo.On("GetGroupMemberTokens", mock.Anything, groupID).Return([]string{}, errors.New("FCM token error"))
			},
			wantErr:    false,
			expectSend: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMsgRepo := new(MockMessageRepository)
			mockFCMRepo := new(MockFCMTokenRepository)
			mockNotifService := new(MockNotificationService)
			ch := make(chan bool, 1)

			tt.setupMocks(mockMsgRepo, mockFCMRepo, mockNotifService, ch)

			service := NewMessageService(mockMsgRepo, mockFCMRepo, mockNotifService, nil)
			message, err := service.CreateMessage(ctx, groupID, userID, userName, create)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, message)
				if tt.expectSend {
					select {
					case <-ch:
						// SendGroupMessage was called
					case <-time.After(1 * time.Second):
						t.Fatal("SendGroupMessage was not called")
					}
				}
			}

			mockMsgRepo.AssertExpectations(t)
			mockFCMRepo.AssertExpectations(t)
			mockNotifService.AssertExpectations(t)
		})
	}
}