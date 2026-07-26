package service

import (
	"context"

	"posify-backend/internal/model"
	"posify-backend/internal/repository"
)

type PaymentService struct {
	repo *repository.PaymentRepo
}

func NewPaymentService(repo *repository.PaymentRepo) *PaymentService {
	return &PaymentService{repo: repo}
}

func (s *PaymentService) GenerateQRIS(ctx context.Context, tenantID string, req model.GenerateQRISRequest) (*model.GenerateQRISResponse, error) {
	p, err := s.repo.CreateQRIS(ctx, tenantID, req)
	if err != nil {
		return nil, err
	}
	return &model.GenerateQRISResponse{
		OrderID:   p.OrderID,
		QRString:  p.QRString,
		QRURL:     p.QRURL,
		Amount:    p.Amount,
		ExpiresAt: p.ExpiresAt,
	}, nil
}

func (s *PaymentService) GetStatus(ctx context.Context, orderID string) (*model.QRISStatusResponse, error) {
	return s.repo.GetStatus(ctx, orderID)
}

func (s *PaymentService) HandleWebhook(ctx context.Context, payload model.WebhookPayload) error {
	// ponytail: In production, verify signature here
	if payload.Status == "PAID" {
		return s.repo.MarkPaid(ctx, payload.OrderID, payload.PaymentMethod)
	}
	return nil
}
