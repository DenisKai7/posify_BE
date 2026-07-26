package service

import (
	"context"

	"posify-backend/internal/model"
	"posify-backend/internal/repository"
)

type ProductService struct {
	repo *repository.ProductRepo
}

func NewProductService(repo *repository.ProductRepo) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) Create(ctx context.Context, tenantID string, req model.CreateProductRequest) (*model.Product, error) {
	return s.repo.Create(ctx, tenantID, req)
}

func (s *ProductService) Update(ctx context.Context, tenantID, productID string, req model.UpdateProductRequest) (*model.Product, error) {
	return s.repo.Update(ctx, tenantID, productID, req)
}

func (s *ProductService) Delete(ctx context.Context, tenantID, productID string) error {
	return s.repo.SoftDelete(ctx, tenantID, productID)
}

func (s *ProductService) List(ctx context.Context, tenantID string) ([]model.Product, error) {
	return s.repo.ListByTenant(ctx, tenantID)
}
