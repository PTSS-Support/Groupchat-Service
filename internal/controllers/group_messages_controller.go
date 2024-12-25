package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"Groupchat-Service/internal/models"
	"Groupchat-Service/internal/services"
)

// FCMMessageController handles HTTP requests for messages
type FCMMessageController struct {
	messageService services.MessageService
}

func NewMessageController(messageService services.MessageService) *FCMMessageController {
	return &FCMMessageController{
		messageService: messageService,
	}
}

// GetMessages handles GET /groups/messages
func (c *FCMMessageController) GetMessages(w http.ResponseWriter, r *http.Request) {
	// Extract group ID from context (set by auth middleware)
	groupID, err := getGroupIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid group context")
		return
	}

	// Parse pagination query
	query := models.PaginationQuery{
		PageSize:  10, // Default page size
		Direction: "next",
	}

	if size := r.URL.Query().Get("pageSize"); size != "" {
		pageSize, err := strconv.Atoi(size)
		if err != nil || pageSize < 1 || pageSize > 50 {
			respondWithError(w, http.StatusBadRequest, "page size must be between 1 and 50")
			return
		}
		query.PageSize = pageSize
	}

	if cursor := r.URL.Query().Get("cursor"); cursor != "" {
		query.Cursor = &cursor
	}

	if direction := r.URL.Query().Get("direction"); direction != "" {
		if direction != "next" && direction != "previous" {
			respondWithError(w, http.StatusBadRequest, "direction must be 'next' or 'previous'")
			return
		}
		query.Direction = direction
	}

	if search := r.URL.Query().Get("search"); search != "" {
		if len(search) > 100 {
			respondWithError(w, http.StatusBadRequest, "search term too long: maximum 100 characters")
			return
		}
		query.Search = &search
	}

	// Get messages from service
	messages, pagination, err := c.messageService.GetMessages(r.Context(), groupID, query)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving messages")
		return
	}

	// Format response
	response := models.PaginatedResponse{
		Data:       messages,
		Pagination: *pagination,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// CreateMessage handles POST /groups/messages
func (c *FCMMessageController) CreateMessage(w http.ResponseWriter, r *http.Request) {
	// Get user info from context
	userID, userName, err := getUserFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid user context")
		return
	}

	groupID, err := getGroupIDFromContext(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid group context")
		return
	}

	// Parse request body
	var createReq models.MessageCreate
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Call service to create message
	message, err := c.messageService.CreateMessage(r.Context(), groupID, userID, userName, createReq)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating message")
		return
	}

	respondWithJSON(w, http.StatusCreated, message)
}

// ToggleMessagePin handles PUT /groups/messages/{messageId}/pin
func (c *FCMMessageController) ToggleMessagePin(w http.ResponseWriter, r *http.Request) {
	// Extract message ID from URL
	messageID, err := uuid.Parse(mux.Vars(r)["messageId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	// Call service to toggle pin
	message, err := c.messageService.ToggleMessagePin(r.Context(), messageID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error toggling message pin")
		return
	}

	respondWithJSON(w, http.StatusOK, message)
}

// respondWithError sends a JSON error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
