package service

import (
	"context"

	"posify-backend/internal/model"
	"posify-backend/internal/repository"
)

type TierService struct {
	repo *repository.TierRepo
}

func NewTierService(repo *repository.TierRepo) *TierService {
	return &TierService{repo: repo}
}

func (s *TierService) List(ctx context.Context, tenantID string) ([]model.MembershipTier, error) {
	return s.repo.List(ctx, tenantID)
}

func (s *TierService) Update(ctx context.Context, tenantID, tierID string, req model.UpdateTierRequest) (*model.MembershipTier, error) {
	return s.repo.Update(ctx, tenantID, tierID, req)
}

func (s *TierService) AutoUpgrade(ctx context.Context, tenantID, customerID string) error {
	return s.repo.AutoUpgradeTier(ctx, tenantID, customerID)
}
