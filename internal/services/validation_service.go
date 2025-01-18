package services

import (
	"Groupchat-Service/internal/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strconv"
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

	if size, ok := queryParams["pageSize"]; ok {
		pageSize, err := strconv.Atoi(size)
		if err != nil || pageSize < MinPageSize || pageSize > MaxPageSize {
			return query, fmt.Errorf("page size must be between %d and %d", MinPageSize, MaxPageSize)
		}
		query.PageSize = pageSize
	}

	if cursor, ok := queryParams["cursor"]; ok {
		query.Cursor = &cursor
	}

	if direction, ok := queryParams["direction"]; ok {
		if direction != string(models.Next) && direction != string(models.Previous) {
			return query, fmt.Errorf("direction must be '%s' or '%s'", models.Next, models.Previous)
		}
		query.Direction = models.Direction(direction)
	}

	if search, ok := queryParams["search"]; ok {
		if utf8.RuneCountInString(search) > MaxSearchLength {
			return query, fmt.Errorf("search term too long: maximum %d characters", MaxSearchLength)
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

func (v *validationService) FetchUserName(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/users/me", v.userServiceURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch user profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to fetch user profile: user service returned non-200 status code")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Role      string `json:"role"`
		GroupID   string `json:"groupId"`
		LastSeen  string `json:"lastSeen"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return fmt.Sprintf("%s %s", result.FirstName, result.LastName), nil
}

func (v *validationService) FetchGroupMembers(ctx context.Context, groupID uuid.UUID) ([]models.UserSummary, error) {
	url := fmt.Sprintf("%s/groups/%s/members", v.userServiceURL, groupID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group members: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch group members: user service returned non-200 status code")
	}

	var members []models.UserSummary
	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return members, nil
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
