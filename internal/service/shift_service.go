package service

import (
	"context"

	"posify-backend/internal/model"
	"posify-backend/internal/repository"
)

type ShiftService struct {
	repo *repository.ShiftRepo
}

func NewShiftService(repo *repository.ShiftRepo) *ShiftService {
	return &ShiftService{repo: repo}
}

func (s *ShiftService) Start(ctx context.Context, tenantID, userID string, initialCash float64) (*model.Shift, error) {
	return s.repo.Start(ctx, tenantID, userID, initialCash)
}

func (s *ShiftService) Close(ctx context.Context, tenantID, userID string, actualCash float64) (*model.Shift, error) {
	return s.repo.Close(ctx, tenantID, userID, actualCash)
}

func (s *ShiftService) GetCurrent(ctx context.Context, tenantID, userID string) (*model.Shift, error) {
	return s.repo.GetCurrent(ctx, tenantID, userID)
}
