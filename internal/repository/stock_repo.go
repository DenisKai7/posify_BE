package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"posify-backend/internal/model"
)

type StockRepo struct {
	pool *pgxpool.Pool
}

func NewStockRepo(pool *pgxpool.Pool) *StockRepo {
	return &StockRepo{pool: pool}
}

func (r *StockRepo) Adjust(ctx context.Context, tenantID, userID string, req model.CreateStockAdjustmentRequest) (*model.StockAdjustment, error) {
	// Begin transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update product stock
	var newStock int
	err = tx.QueryRow(ctx,
		`UPDATE products SET stock = stock + $1, updated_at = NOW(), version = version + 1
		 WHERE id = $2 AND tenant_id = $3 AND deleted_at IS NULL
		 RETURNING stock`,
		req.QtyChange, req.ProductID, tenantID,
	).Scan(&newStock)
	if err != nil {
		return nil, fmt.Errorf("update stock: %w", err)
	}
	if newStock < 0 {
		return nil, fmt.Errorf("insufficient stock (would go negative)")
	}

	// Insert audit record
	sa := &model.StockAdjustment{}
	err = tx.QueryRow(ctx,
		`INSERT INTO stock_adjustments (tenant_id, product_id, user_id, type, qty_change, reason)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, tenant_id, product_id, user_id, type, qty_change, reason, created_at`,
		tenantID, req.ProductID, userID, req.Type, req.QtyChange, req.Reason,
	).Scan(&sa.ID, &sa.TenantID, &sa.ProductID, &sa.UserID, &sa.Type, &sa.QtyChange, &sa.Reason, &sa.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert adjustment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return sa, nil
}

func (r *StockRepo) LowStockAlerts(ctx context.Context, tenantID string, threshold int) ([]model.LowStockAlert, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, sku, name, stock, price FROM products
		 WHERE tenant_id = $1 AND is_active = TRUE AND deleted_at IS NULL AND stock <= $2
		 ORDER BY stock ASC`,
		tenantID, threshold,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.LowStockAlert
	for rows.Next() {
		var a model.LowStockAlert
		a.Threshold = threshold
		if err := rows.Scan(&a.ProductID, &a.SKU, &a.Name, &a.Stock, &a.Price); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	if items == nil {
		items = []model.LowStockAlert{}
	}
	return items, rows.Err()
}

func (r *StockRepo) History(ctx context.Context, tenantID, productID string, limit int) ([]model.StockAdjustment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT sa.id, sa.tenant_id, sa.product_id, sa.user_id, sa.type, sa.qty_change, sa.reason, sa.created_at,
		        p.name, p.sku, pr.full_name
		 FROM stock_adjustments sa
		 JOIN products p ON p.id = sa.product_id
		 JOIN profiles pr ON pr.id = sa.user_id
		 WHERE sa.tenant_id = $1 AND sa.product_id = $2
		 ORDER BY sa.created_at DESC
		 LIMIT $3`,
		tenantID, productID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.StockAdjustment
	for rows.Next() {
		var s model.StockAdjustment
		if err := rows.Scan(&s.ID, &s.TenantID, &s.ProductID, &s.UserID, &s.Type, &s.QtyChange, &s.Reason, &s.CreatedAt, &s.ProductName, &s.SKU, &s.UserName); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	if items == nil {
		items = []model.StockAdjustment{}
	}
	return items, rows.Err()
}

// LogTransactionDeduct records automatic stock deduction from a sale
func (r *StockRepo) LogTransactionDeduct(ctx context.Context, tenantID, userID, productID string, qty int) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO stock_adjustments (tenant_id, product_id, user_id, type, qty_change, reason)
		 VALUES ($1, $2, $3, 'OUT', $4, 'Sale deduction')`,
		tenantID, productID, userID, -qty,
	)
	return err
}
