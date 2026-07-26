package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"posify-backend/internal/model"
)

type ReportRepo struct {
	pool *pgxpool.Pool
}

func NewReportRepo(pool *pgxpool.Pool) *ReportRepo {
	return &ReportRepo{pool: pool}
}

func (r *ReportRepo) Summary(ctx context.Context, tenantID string) (*model.ReportSummary, error) {
	s := &model.ReportSummary{}
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Revenue & count today
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		 FROM transactions
		 WHERE tenant_id = $1 AND payment_status = 'PAID' AND created_at >= $2`,
		tenantID, startOfDay,
	).Scan(&s.RevenueToday, &s.TxCountToday)
	if err != nil {
		return nil, fmt.Errorf("today stats: %w", err)
	}

	// Revenue & count this month
	err = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		 FROM transactions
		 WHERE tenant_id = $1 AND payment_status = 'PAID' AND created_at >= $2`,
		tenantID, startOfMonth,
	).Scan(&s.RevenueMonth, &s.TxCountMonth)
	if err != nil {
		return nil, fmt.Errorf("month stats: %w", err)
	}

	// Active products count
	err = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM products WHERE tenant_id = $1 AND is_active = TRUE AND deleted_at IS NULL`,
		tenantID,
	).Scan(&s.ActiveProducts)
	if err != nil {
		return nil, fmt.Errorf("product count: %w", err)
	}

	// Low stock items
	rows, err := r.pool.Query(ctx,
		`SELECT id, sku, name, stock FROM products
		 WHERE tenant_id = $1 AND is_active = TRUE AND deleted_at IS NULL AND stock < 10
		 ORDER BY stock ASC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("low stock: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item model.LowStockItem
		if err := rows.Scan(&item.ID, &item.SKU, &item.Name, &item.Stock); err != nil {
			return nil, fmt.Errorf("scan low stock: %w", err)
		}
		s.LowStockItems = append(s.LowStockItems, item)
	}
	if s.LowStockItems == nil {
		s.LowStockItems = []model.LowStockItem{}
	}
	return s, rows.Err()
}

func (r *ReportRepo) SalesChart(ctx context.Context, tenantID string, days int) ([]model.SalesChartPoint, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT d::date AS date,
			COALESCE(SUM(t.total_amount), 0) AS revenue,
			COUNT(t.id)::int AS count
		 FROM generate_series(CURRENT_DATE - $2::int, CURRENT_DATE, '1 day') d
		 LEFT JOIN transactions t
			ON t.tenant_id = $1
			AND t.payment_status = 'PAID'
			AND t.created_at::date = d::date
		 GROUP BY d::date
		 ORDER BY d::date`,
		tenantID, days-1,
	)
	if err != nil {
		return nil, fmt.Errorf("sales chart: %w", err)
	}
	defer rows.Close()

	var points []model.SalesChartPoint
	for rows.Next() {
		var p model.SalesChartPoint
		if err := rows.Scan(&p.Date, &p.Revenue, &p.Count); err != nil {
			return nil, fmt.Errorf("scan chart: %w", err)
		}
		points = append(points, p)
	}
	if points == nil {
		points = []model.SalesChartPoint{}
	}
	return points, rows.Err()
}

func (r *ReportRepo) SalesSummary(ctx context.Context, tenantID, startDate, endDate string) (*model.SalesSummary, error) {
	s := &model.SalesSummary{StartDate: startDate, EndDate: endDate}
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_amount), 0), COUNT(*), COALESCE(AVG(total_amount), 0)
		 FROM transactions
		 WHERE tenant_id = $1 AND payment_status = 'PAID'
		   AND created_at::date BETWEEN $2 AND $3`,
		tenantID, startDate, endDate,
	).Scan(&s.TotalRev, &s.TotalTx, &s.AvgTx)
	if err != nil {
		return nil, err
	}
	// Total items sold
	err = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(ti.quantity), 0)
		 FROM transaction_items ti
		 JOIN transactions t ON t.id = ti.transaction_id
		 WHERE t.tenant_id = $1 AND t.payment_status = 'PAID'
		   AND t.created_at::date BETWEEN $2 AND $3`,
		tenantID, startDate, endDate,
	).Scan(&s.TotalItems)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *ReportRepo) TopProducts(ctx context.Context, tenantID string, limit int) ([]model.TopProduct, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ti.product_id, ti.product_name, p.sku,
			SUM(ti.quantity)::int AS total_qty,
			SUM(ti.subtotal) AS total_revenue
		 FROM transaction_items ti
		 JOIN transactions t ON t.id = ti.transaction_id
		 JOIN products p ON p.id = ti.product_id
		 WHERE t.tenant_id = $1 AND t.payment_status = 'PAID'
		 GROUP BY ti.product_id, ti.product_name, p.sku
		 ORDER BY total_qty DESC
		 LIMIT $2`,
		tenantID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.TopProduct
	for rows.Next() {
		var tp model.TopProduct
		if err := rows.Scan(&tp.ProductID, &tp.ProductName, &tp.SKU, &tp.TotalQty, &tp.TotalRevenue); err != nil {
			return nil, err
		}
		items = append(items, tp)
	}
	if items == nil {
		items = []model.TopProduct{}
	}
	return items, rows.Err()
}

func (r *ReportRepo) PaymentMethods(ctx context.Context, tenantID string) ([]model.PaymentMethodBreakdown, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT payment_method, COUNT(*)::int, COALESCE(SUM(total_amount), 0)
		 FROM transactions
		 WHERE tenant_id = $1 AND payment_status = 'PAID'
		 GROUP BY payment_method
		 ORDER BY total DESC`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.PaymentMethodBreakdown
	for rows.Next() {
		var pm model.PaymentMethodBreakdown
		if err := rows.Scan(&pm.Method, &pm.Count, &pm.Total); err != nil {
			return nil, err
		}
		items = append(items, pm)
	}
	if items == nil {
		items = []model.PaymentMethodBreakdown{}
	}
	return items, rows.Err()
}

func (r *ReportRepo) SalesSummaryRows(ctx context.Context, tenantID, startDate, endDate string) ([]model.TopProduct, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ti.product_id, ti.product_name, p.sku,
			SUM(ti.quantity)::int AS total_qty,
			SUM(ti.subtotal) AS total_revenue
		 FROM transaction_items ti
		 JOIN transactions t ON t.id = ti.transaction_id
		 JOIN products p ON p.id = ti.product_id
		 WHERE t.tenant_id = $1 AND t.payment_status = 'PAID'
		   AND t.created_at::date BETWEEN $2 AND $3
		 GROUP BY ti.product_id, ti.product_name, p.sku
		 ORDER BY total_revenue DESC`,
		tenantID, startDate, endDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.TopProduct
	for rows.Next() {
		var tp model.TopProduct
		if err := rows.Scan(&tp.ProductID, &tp.ProductName, &tp.SKU, &tp.TotalQty, &tp.TotalRevenue); err != nil {
			return nil, err
		}
		items = append(items, tp)
	}
	if items == nil {
		items = []model.TopProduct{}
	}
	return items, rows.Err()
}
