package service

import (
	"context"

	"posify-backend/internal/model"
	"posify-backend/internal/repository"
)

type ReportService struct {
	repo *repository.ReportRepo
}

func NewReportService(repo *repository.ReportRepo) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) Summary(ctx context.Context, tenantID string) (*model.ReportSummary, error) {
	return s.repo.Summary(ctx, tenantID)
}

func (s *ReportService) SalesChart(ctx context.Context, tenantID string, days int) ([]model.SalesChartPoint, error) {
	return s.repo.SalesChart(ctx, tenantID, days)
}

func (s *ReportService) SalesSummary(ctx context.Context, tenantID, startDate, endDate string) (*model.SalesSummary, error) {
	return s.repo.SalesSummary(ctx, tenantID, startDate, endDate)
}

func (s *ReportService) TopProducts(ctx context.Context, tenantID string, limit int) ([]model.TopProduct, error) {
	return s.repo.TopProducts(ctx, tenantID, limit)
}

func (s *ReportService) PaymentMethods(ctx context.Context, tenantID string) ([]model.PaymentMethodBreakdown, error) {
	return s.repo.PaymentMethods(ctx, tenantID)
}

func (s *ReportService) SalesSummaryRows(ctx context.Context, tenantID, startDate, endDate string) ([]model.TopProduct, error) {
	return s.repo.SalesSummaryRows(ctx, tenantID, startDate, endDate)
}
