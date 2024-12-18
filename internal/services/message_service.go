package services

import (
	"Groupchat-Service/internal/database/repository"
)

type MessageService struct {
	messageRepo repository.MessageRepository
}
