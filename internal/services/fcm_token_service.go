package services

import (
	"Groupchat-Service/internal/database/repositories"
	"context"
	"github.com/google/uuid"
)

type fcmTokenService struct {
	repo repositories.FCMTokenRepository
}

func NewFCMTokenService(repo repositories.FCMTokenRepository) FCMTokenService {
	return &fcmTokenService{repo: repo}
}

func (s *fcmTokenService) SaveToken(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, token string) error {
	return s.repo.SaveToken(ctx, groupID, userID, token)
}

func (s *fcmTokenService) DeleteToken(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) error {
	return s.repo.DeleteToken(ctx, groupID, userID)
}
