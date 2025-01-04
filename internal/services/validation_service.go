package services

import (
	"Groupchat-Service/internal/models"
	"context"
	"errors"
	"github.com/google/uuid"
	"strconv"
)

const (
	DefaultPageSize = 10
	MinPageSize     = 1
	MaxPageSize     = 50
	MaxSearchLength = 100
	MaxTokenLength  = 1024
)

type validationService struct{}

func NewValidationService() ValidationService {
	return &validationService{}
}

func (v *validationService) ValidatePaginationQuery(ctx context.Context, queryParams map[string]string) (models.PaginationQuery, error) {
	query := models.PaginationQuery{
		PageSize:  DefaultPageSize,
		Direction: models.Next,
	}

	if size, ok := queryParams["pageSize"]; ok {
		pageSize, err := strconv.Atoi(size)
		if err != nil || pageSize < MinPageSize || pageSize > MaxPageSize {
			return query, errors.New("page size must be between 1 and 50")
		}
		query.PageSize = pageSize
	}

	if cursor, ok := queryParams["cursor"]; ok {
		query.Cursor = &cursor
	}

	if direction, ok := queryParams["direction"]; ok {
		if direction != "next" && direction != "previous" {
			return query, errors.New("direction must be 'next' or 'previous'")
		}
		query.Direction = models.Direction(direction)
	}

	if search, ok := queryParams["search"]; ok {
		if len(search) > MaxSearchLength {
			return query, errors.New("search term too long: maximum 100 characters")
		}
		query.Search = &search
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
