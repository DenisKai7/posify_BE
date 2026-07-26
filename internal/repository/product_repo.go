package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"posify-backend/internal/model"
)

type ProductRepo struct {
	pool *pgxpool.Pool
}

func NewProductRepo(pool *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{pool: pool}
}

func (r *ProductRepo) Create(ctx context.Context, tenantID string, req model.CreateProductRequest) (*model.Product, error) {
	p := &model.Product{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO products (tenant_id, sku, name, price, cost_price, stock, category)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, tenant_id, sku, name, price, cost_price, stock, category, version, is_active, created_at, updated_at`,
		tenantID, req.SKU, req.Name, req.Price, req.CostPrice, req.Stock, req.Category,
	).Scan(&p.ID, &p.TenantID, &p.SKU, &p.Name, &p.Price, &p.CostPrice, &p.Stock, &p.Category, &p.Version, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}
	return p, nil
}

func (r *ProductRepo) Update(ctx context.Context, tenantID, productID string, req model.UpdateProductRequest) (*model.Product, error) {
	p := &model.Product{}
	err := r.pool.QueryRow(ctx,
		`UPDATE products SET
			sku = COALESCE($3, sku),
			name = COALESCE($4, name),
			price = COALESCE($5, price),
			cost_price = COALESCE($6, cost_price),
			stock = COALESCE($7, stock),
			category = COALESCE($8, category),
			version = version + 1,
			updated_at = NOW()
		 WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		 RETURNING id, tenant_id, sku, name, price, cost_price, stock, category, version, is_active, created_at, updated_at`,
		productID, tenantID, req.SKU, req.Name, req.Price, req.CostPrice, req.Stock, req.Category,
	).Scan(&p.ID, &p.TenantID, &p.SKU, &p.Name, &p.Price, &p.CostPrice, &p.Stock, &p.Category, &p.Version, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update product: %w", err)
	}
	return p, nil
}

func (r *ProductRepo) SoftDelete(ctx context.Context, tenantID, productID string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE products SET deleted_at = NOW(), is_active = FALSE, updated_at = NOW() WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`,
		productID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

func (r *ProductRepo) ListByTenant(ctx context.Context, tenantID string) ([]model.Product, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, sku, name, price, cost_price, stock, COALESCE(category,''), version, is_active, created_at, updated_at
		 FROM products WHERE tenant_id = $1 AND deleted_at IS NULL ORDER BY name`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.TenantID, &p.SKU, &p.Name, &p.Price, &p.CostPrice, &p.Stock, &p.Category, &p.Version, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, p)
	}
	return products, rows.Err()
}
