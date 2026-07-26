package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"posify-backend/internal/model"
)

type CustomerRepo struct {
	pool *pgxpool.Pool
}

func NewCustomerRepo(pool *pgxpool.Pool) *CustomerRepo {
	return &CustomerRepo{pool: pool}
}

func (r *CustomerRepo) Create(ctx context.Context, tenantID string, req model.CreateCustomerRequest) (*model.Customer, error) {
	c := &model.Customer{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO customers (tenant_id, name, phone, email)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, tenant_id, name, phone, email, points, tier, total_spent, visit_count, created_at, updated_at`,
		tenantID, req.Name, req.Phone, req.Email,
	).Scan(&c.ID, &c.TenantID, &c.Name, &c.Phone, &c.Email, &c.Points, &c.Tier, &c.TotalSpent, &c.VisitCount, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create customer: %w", err)
	}
	return c, nil
}

func (r *CustomerRepo) GetByID(ctx context.Context, tenantID, customerID string) (*model.Customer, error) {
	c := &model.Customer{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, phone, email, points, tier, total_spent, visit_count, created_at, updated_at
		 FROM customers WHERE tenant_id = $1 AND id = $2`,
		tenantID, customerID,
	).Scan(&c.ID, &c.TenantID, &c.Name, &c.Phone, &c.Email, &c.Points, &c.Tier, &c.TotalSpent, &c.VisitCount, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}
	return c, nil
}

func (r *CustomerRepo) Search(ctx context.Context, tenantID, query string) ([]model.Customer, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, phone, email, points, tier, total_spent, visit_count, created_at, updated_at
		 FROM customers
		 WHERE tenant_id = $1 AND (phone ILIKE '%' || $2 || '%' OR name ILIKE '%' || $2 || '%')
		 ORDER BY name LIMIT 20`,
		tenantID, query,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.Customer
	for rows.Next() {
		var c model.Customer
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Name, &c.Phone, &c.Email, &c.Points, &c.Tier, &c.TotalSpent, &c.VisitCount, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	if items == nil {
		items = []model.Customer{}
	}
	return items, rows.Err()
}

func (r *CustomerRepo) AddPoints(ctx context.Context, tenantID, customerID string, points int, amount float64) error {
	// 1 point per Rp 10,000 spent
	earned := int(amount / 10000)
	if earned < 1 {
		earned = 0
	}

	_, err := r.pool.Exec(ctx,
		`UPDATE customers SET
			points = points + $3,
			total_spent = total_spent + $4,
			visit_count = visit_count + 1,
			tier = CASE
				WHEN total_spent + $4 >= 10000000 THEN 'PLATINUM'
				WHEN total_spent + $4 >= 5000000 THEN 'GOLD'
				ELSE 'SILVER'
			END,
			updated_at = NOW()
		 WHERE tenant_id = $1 AND id = $2`,
		tenantID, customerID, earned, amount,
	)
	return err
}

func (r *CustomerRepo) RedeemPoints(ctx context.Context, tenantID, customerID string, points int) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE customers SET points = points - $3, updated_at = NOW()
		 WHERE tenant_id = $1 AND id = $2 AND points >= $3`,
		tenantID, customerID, points,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient points")
	}
	return nil
}

func (r *CustomerRepo) List(ctx context.Context, tenantID string) ([]model.Customer, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, phone, email, points, tier, total_spent, visit_count, created_at, updated_at
		 FROM customers WHERE tenant_id = $1 ORDER BY name`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.Customer
	for rows.Next() {
		var c model.Customer
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Name, &c.Phone, &c.Email, &c.Points, &c.Tier, &c.TotalSpent, &c.VisitCount, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	if items == nil {
		items = []model.Customer{}
	}
	return items, rows.Err()
}
