package controllers

import (
	"net/http"
)

type MessageController interface {
	GetMessages(w http.ResponseWriter, r *http.Request)
	CreateMessage(w http.ResponseWriter, r *http.Request)
	ToggleMessagePin(w http.ResponseWriter, r *http.Request)
}
