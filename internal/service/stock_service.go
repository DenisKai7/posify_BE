package service

import (
	"context"

	"posify-backend/internal/model"
	"posify-backend/internal/repository"
)

type StockService struct {
	repo *repository.StockRepo
}

func NewStockService(repo *repository.StockRepo) *StockService {
	return &StockService{repo: repo}
}

func (s *StockService) Adjust(ctx context.Context, tenantID, userID string, req model.CreateStockAdjustmentRequest) (*model.StockAdjustment, error) {
	return s.repo.Adjust(ctx, tenantID, userID, req)
}

func (s *StockService) LowStockAlerts(ctx context.Context, tenantID string, threshold int) ([]model.LowStockAlert, error) {
	return s.repo.LowStockAlerts(ctx, tenantID, threshold)
}

func (s *StockService) History(ctx context.Context, tenantID, productID string, limit int) ([]model.StockAdjustment, error) {
	return s.repo.History(ctx, tenantID, productID, limit)
}
