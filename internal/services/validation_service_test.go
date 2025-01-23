package services

import (
	"Groupchat-Service/internal/models"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func ptr(s string) *string {
	return &s
}

func TestValidatePaginationQuery(t *testing.T) {
	vs := NewValidationService("")

	tests := []struct {
		name        string
		queryParams map[string]string
		expected    models.PaginationQuery
		expectErr   bool
	}{
		{
			name: "Valid query",
			queryParams: map[string]string{
				"pageSize":  "20",
				"cursor":    "some-cursor",
				"direction": "next",
				"search":    "test",
			},
			expected: models.PaginationQuery{
				PageSize:  20,
				Cursor:    ptr("some-cursor"),
				Direction: models.Next,
				Search:    ptr("test"),
			},
			expectErr: false,
		},
		{
			name: "Invalid page size",
			queryParams: map[string]string{
				"pageSize": "100",
			},
			expected:  models.PaginationQuery{},
			expectErr: true,
		},
		{
			name: "Invalid direction",
			queryParams: map[string]string{
				"direction": "up",
			},
			expected:  models.PaginationQuery{},
			expectErr: true,
		},
		{
			name: "Search term too long",
			queryParams: map[string]string{
				"search": strings.Repeat("a", MaxSearchLength+1),
			},
			expected:  models.PaginationQuery{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := vs.ValidatePaginationQuery(tt.queryParams)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, query)
			}
		})
	}
}
