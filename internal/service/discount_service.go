package service

import (
	"context"

	"posify-backend/internal/model"
	"posify-backend/internal/repository"
)

type DiscountService struct {
	repo *repository.DiscountRepo
}

func NewDiscountService(repo *repository.DiscountRepo) *DiscountService {
	return &DiscountService{repo: repo}
}

func (s *DiscountService) Create(ctx context.Context, tenantID string, req model.CreateDiscountRequest) (*model.Discount, error) {
	return s.repo.Create(ctx, tenantID, req)
}

func (s *DiscountService) Validate(ctx context.Context, tenantID, code string, subtotal float64) (*model.ValidateDiscountResponse, error) {
	return s.repo.Validate(ctx, tenantID, code, subtotal)
}

func (s *DiscountService) List(ctx context.Context, tenantID string) ([]model.Discount, error) {
	return s.repo.List(ctx, tenantID)
}
