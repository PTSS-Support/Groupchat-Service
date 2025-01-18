package services

import (
	"Groupchat-Service/internal/models"
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
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

func TestFetchUserName(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func() *httptest.Server
		expectName string
		expectErr  bool
	}{
		{
			name: "Success",
			setupMock: func() *httptest.Server {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"firstName": "John", "lastName": "Doe"}`))
				})
				return httptest.NewServer(handler)
			},
			expectName: "John Doe",
			expectErr:  false,
		},
		{
			name: "Non-200 status code",
			setupMock: func() *httptest.Server {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				})
				return httptest.NewServer(handler)
			},
			expectName: "",
			expectErr:  true,
		},
		{
			name: "Invalid response body",
			setupMock: func() *httptest.Server {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`invalid`))
				})
				return httptest.NewServer(handler)
			},
			expectName: "",
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupMock()
			defer server.Close()

			vs := NewValidationService(server.URL)
			name, err := vs.FetchUserName(context.Background())

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectName, name)
			}
		})
	}
}

func TestFetchGroupMembers(t *testing.T) {
	groupID := uuid.New()

	tests := []struct {
		name        string
		setupMock   func() *httptest.Server
		expectErr   bool
		expectCount int
	}{
		{
			name: "Success",
			setupMock: func() *httptest.Server {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`[{"id": "123e4567-e89b-12d3-a456-426614174000", "userName": "JohnDoe"}]`))
				})
				return httptest.NewServer(handler)
			},
			expectErr:   false,
			expectCount: 1,
		},
		{
			name: "Non-200 status code",
			setupMock: func() *httptest.Server {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				})
				return httptest.NewServer(handler)
			},
			expectErr:   true,
			expectCount: 0,
		},
		{
			name: "Invalid response body",
			setupMock: func() *httptest.Server {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`invalid`))
				})
				return httptest.NewServer(handler)
			},
			expectErr:   true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupMock()
			defer server.Close()

			vs := NewValidationService(server.URL)
			members, err := vs.FetchGroupMembers(context.Background(), groupID)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, members)
			} else {
				assert.NoError(t, err)
				assert.Len(t, members, tt.expectCount)
			}
		})
	}
}
