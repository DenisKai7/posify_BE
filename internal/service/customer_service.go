package service

import (
	"context"

	"posify-backend/internal/model"
	"posify-backend/internal/repository"
)

type CustomerService struct {
	repo *repository.CustomerRepo
}

func NewCustomerService(repo *repository.CustomerRepo) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) Create(ctx context.Context, tenantID string, req model.CreateCustomerRequest) (*model.Customer, error) {
	return s.repo.Create(ctx, tenantID, req)
}

func (s *CustomerService) GetByID(ctx context.Context, tenantID, customerID string) (*model.Customer, error) {
	return s.repo.GetByID(ctx, tenantID, customerID)
}

func (s *CustomerService) Search(ctx context.Context, tenantID, query string) ([]model.Customer, error) {
	return s.repo.Search(ctx, tenantID, query)
}

func (s *CustomerService) AddPoints(ctx context.Context, tenantID, customerID string, points int, amount float64) error {
	return s.repo.AddPoints(ctx, tenantID, customerID, points, amount)
}

func (s *CustomerService) RedeemPoints(ctx context.Context, tenantID, customerID string, points int) error {
	return s.repo.RedeemPoints(ctx, tenantID, customerID, points)
}

func (s *CustomerService) List(ctx context.Context, tenantID string) ([]model.Customer, error) {
	return s.repo.List(ctx, tenantID)
}
