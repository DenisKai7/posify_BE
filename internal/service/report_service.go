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
