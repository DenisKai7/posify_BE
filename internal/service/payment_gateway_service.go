package service

import (
	"context"

	"posify-backend/internal/model"
	"posify-backend/internal/repository"
)

type PaymentGatewayService struct {
	repo *repository.PaymentGatewayRepo
}

func NewPaymentGatewayService(repo *repository.PaymentGatewayRepo) *PaymentGatewayService {
	return &PaymentGatewayService{repo: repo}
}

func (s *PaymentGatewayService) Get(ctx context.Context, tenantID string) (*model.PaymentGateway, error) {
	pg, err := s.repo.GetByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	// Mask server key
	pg.ServerKey = repository.MaskServerKey(pg.ServerKey)
	return pg, nil
}

func (s *PaymentGatewayService) GetRaw(ctx context.Context, tenantID string) (*model.PaymentGateway, error) {
	return s.repo.GetByTenant(ctx, tenantID)
}

func (s *PaymentGatewayService) Update(ctx context.Context, tenantID string, req model.UpdatePaymentGatewayRequest) (*model.PaymentGateway, error) {
	pg, err := s.repo.Upsert(ctx, tenantID, req)
	if err != nil {
		return nil, err
	}
	pg.ServerKey = repository.MaskServerKey(pg.ServerKey)
	return pg, nil
}
