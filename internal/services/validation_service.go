package services

import (
	"Groupchat-Service/internal/models"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	DefaultPageSize = 10
	MinPageSize     = 1
	MaxPageSize     = 50
	MaxSearchLength = 100
	MaxTokenLength  = 1024
)

type validationService struct {
	userServiceURL string
}

func NewValidationService(userServiceURL string) ValidationService {
	return &validationService{userServiceURL: userServiceURL}
}

func (v *validationService) ValidatePaginationQuery(queryParams map[string]string) (models.PaginationQuery, error) {
	query := models.PaginationQuery{
		PageSize:  DefaultPageSize,
		Direction: models.Next,
	}

	for key, value := range queryParams {
		lowerKey := strings.ToLower(key)
		switch lowerKey {
		case "pagesize":
			pageSize, err := strconv.Atoi(value)
			if err != nil || pageSize < MinPageSize || pageSize > MaxPageSize {
				return query, fmt.Errorf("page size must be between %d and %d", MinPageSize, MaxPageSize)
			}
			query.PageSize = pageSize
		case "cursor":
			query.Cursor = &value
		case "direction":
			lowerValue := strings.ToLower(value)
			if lowerValue != string(models.Next) && lowerValue != string(models.Previous) {
				return query, fmt.Errorf("direction must be '%s' or '%s'", models.Next, models.Previous)
			}
			query.Direction = models.Direction(lowerValue)
		case "search":
			if utf8.RuneCountInString(value) > MaxSearchLength {
				return query, fmt.Errorf("search term too long: maximum %d characters", MaxSearchLength)
			}
			query.Search = &value
		}
	}

	return query, nil
}

func (v *validationService) ValidateUserContext(ctx context.Context) (uuid.UUID, string, error) {
	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		return uuid.Nil, "", errors.New("user ID not found in context")
	}

	userName, ok := ctx.Value("userName").(string)
	if !ok {
		return uuid.Nil, "", errors.New("user name not found in context")
	}

	return userID, userName, nil
}

func (v *validationService) ValidateGroupID(groupID string) (uuid.UUID, error) {
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		return uuid.Nil, errors.New("invalid group ID")
	}
	return parsedGroupID, nil
}

func (v *validationService) ValidateUserID(userID string) (uuid.UUID, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID")
	}

	return parsedUserID, nil
}

func (v *validationService) ValidateToken(token string) error {
	if len(token) == 0 || len(token) > MaxTokenLength {
		return errors.New("invalid token length")
	}
	return nil
}

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Role      string    `json:"role"`
	GroupID   uuid.UUID `json:"groupId"`
	LastSeen  string    `json:"lastSeen"`
}
