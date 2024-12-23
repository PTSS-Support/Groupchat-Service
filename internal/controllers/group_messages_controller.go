package controllers

import (
	"Groupchat-Service/internal/database/repositories"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"Groupchat-Service/internal/models"
	"Groupchat-Service/internal/services"
)

// GroupMessagesController handles HTTP requests for group messages
type GroupMessagesController struct {
	messageService *services.MessageService
}

// NewGroupMessagesController creates a new instance of GroupMessagesController
func NewGroupMessagesController(messageService *services.MessageService) *GroupMessagesController {
	return &GroupMessagesController{
		messageService: messageService,
	}
}

// RegisterRoutes registers all routes for group messages
func (c *GroupMessagesController) RegisterRoutes(router *mux.Router) {
	// Group message endpoints
	router.HandleFunc("/api/v1/groups/{groupId}/messages", c.GetGroupMessages).Methods("GET")
	router.HandleFunc("/api/v1/groups/{groupId}/messages", c.CreateGroupMessage).Methods("POST")
	router.HandleFunc("/api/v1/groups/{groupId}/messages/{messageId}/pin", c.ToggleMessagePin).Methods("PUT") // TODO: add methods for this
}

// GetGroupMessages handles the GET request for retrieving group messages
func (c *GroupMessagesController) GetGroupMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID, err := uuid.Parse(vars["groupId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 {
		limit = 20 // Default limit
	}

	cursor := query.Get("cursor")
	direction := query.Get("direction")
	search := query.Get("search")

	opts := repositories.MessageQueryOptions{
		Limit:     limit,
		Cursor:    cursor,
		Direction: direction,
		Search:    search,
	}

	messages, nextCursor, err := c.messageService.GetGroupMessages(r.Context(), groupID, opts)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve messages")
		return
	}

	response := models.NewPaginatedResponse(messages, nextCursor, "", len(messages) == limit)
	respondWithJSON(w, http.StatusOK, response)
}

// CreateGroupMessage handles the POST request for creating a new message
func (c *GroupMessagesController) CreateGroupMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID, err := uuid.Parse(vars["groupId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid group ID format")
		return
	}

	// Parse request body
	var messageCreate models.MessageCreate
	if err := json.NewDecoder(r.Body).Decode(&messageCreate); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Create message object
	message := &models.Message{
		GroupID:    groupID,
		SenderID:   messageCreate.SenderID,
		SenderName: messageCreate.SenderName,
		Content:    messageCreate.Content,
	}

	// Use service to create message
	err = c.messageService.SendGroupMessage(r.Context(), message)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create message")
		return
	}

	respondWithJSON(w, http.StatusCreated, message)
}

// ToggleMessagePin handles the PUT request for toggling a message's pin status
func (c *GroupMessagesController) ToggleMessagePin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageID, err := uuid.Parse(vars["messageId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid message ID format")
		return
	}

	// Toggle pin status using service
	message, err := c.messageService.ToggleMessagePin(r.Context(), messageID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to toggle message pin status")
		return
	}

	respondWithJSON(w, http.StatusOK, message)
}

// Helper functions for HTTP responses
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
