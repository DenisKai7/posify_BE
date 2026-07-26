package service

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	pool *pgxpool.Pool
	cron *cron.Cron
}

func NewScheduler(pool *pgxpool.Pool) *Scheduler {
	return &Scheduler{
		pool: pool,
		cron: cron.New(cron.WithLocation(time.UTC)),
	}
}

func (s *Scheduler) Start() {
	// 1. Daily EOD Sales Summary at 23:59 UTC
	s.cron.AddFunc("59 23 * * *", s.dailySummaryJob)

	// 2. Expired QRIS cleanup every minute
	s.cron.AddFunc("* * * * *", s.cleanupExpiredQRIS)

	// 3. Low stock check every 5 minutes
	s.cron.AddFunc("*/5 * * * *", s.lowStockAlertJob)

	s.cron.Start()
	log.Println("✓ Scheduler started (daily summary, QRIS cleanup, low stock alerts)")
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("Scheduler stopped")
}

// dailySummaryJob aggregates sales data for each tenant
func (s *Scheduler) dailySummaryJob() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")

	// Get all tenants
	rows, err := s.pool.Query(ctx, `SELECT id FROM tenants`)
	if err != nil {
		log.Printf("scheduler: daily summary: list tenants: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tenantID string
		rows.Scan(&tenantID)

		_, err := s.pool.Exec(ctx, `
			INSERT INTO daily_summaries (tenant_id, summary_date, total_revenue, total_transactions, total_items_sold, cash_revenue, qris_revenue, top_product_name, top_product_qty)
			SELECT
				$1, $2::date,
				COALESCE(SUM(t.total_amount), 0),
				COUNT(t.id)::int,
				COALESCE(SUM(items.qty), 0)::int,
				COALESCE(SUM(CASE WHEN t.payment_method = 'CASH' THEN t.total_amount ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN t.payment_method = 'QRIS' THEN t.total_amount ELSE 0 END), 0),
				COALESCE((SELECT ti.product_name FROM transaction_items ti JOIN transactions t2 ON t2.id = ti.transaction_id WHERE t2.tenant_id = $1 AND t2.created_at::date = $2::date AND t2.payment_status = 'PAID' GROUP BY ti.product_name ORDER BY SUM(ti.quantity) DESC LIMIT 1), ''),
				COALESCE((SELECT SUM(ti.quantity)::int FROM transaction_items ti JOIN transactions t2 ON t2.id = ti.transaction_id WHERE t2.tenant_id = $1 AND t2.created_at::date = $2::date AND t2.payment_status = 'PAID' GROUP BY ti.product_name ORDER BY SUM(ti.quantity) DESC LIMIT 1), 0)
			FROM transactions t
			LEFT JOIN LATERAL (SELECT SUM(quantity) AS qty FROM transaction_items WHERE transaction_id = t.id) items ON true
			WHERE t.tenant_id = $1 AND t.created_at::date = $2::date AND t.payment_status = 'PAID'
			ON CONFLICT (tenant_id, summary_date) DO UPDATE SET
				total_revenue = EXCLUDED.total_revenue,
				total_transactions = EXCLUDED.total_transactions,
				total_items_sold = EXCLUDED.total_items_sold,
				cash_revenue = EXCLUDED.cash_revenue,
				qris_revenue = EXCLUDED.qris_revenue,
				top_product_name = EXCLUDED.top_product_name,
				top_product_qty = EXCLUDED.top_product_qty,
				created_at = NOW()`,
			tenantID, yesterday)
		if err != nil {
			log.Printf("scheduler: daily summary tenant %s: %v", tenantID, err)
		}
	}
	log.Printf("scheduler: daily summary generated for %s", yesterday)
}

// cleanupExpiredQRIS marks stale QRIS payments as EXPIRED
func (s *Scheduler) cleanupExpiredQRIS() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tag, err := s.pool.Exec(ctx,
		`UPDATE qris_payments SET status = 'EXPIRED' WHERE status = 'PENDING' AND expires_at < NOW()`)
	if err != nil {
		log.Printf("scheduler: QRIS cleanup: %v", err)
		return
	}
	if tag.RowsAffected() > 0 {
		log.Printf("scheduler: expired %d QRIS payments", tag.RowsAffected())
	}
}

// lowStockAlertJob logs low stock warnings (email integration placeholder)
func (s *Scheduler) lowStockAlertJob() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := s.pool.Query(ctx,
		`SELECT t.name, p.name, p.sku, p.stock
		 FROM products p
		 JOIN tenants t ON t.id = p.tenant_id
		 WHERE p.is_active = TRUE AND p.deleted_at IS NULL AND p.stock <= 5
		 ORDER BY p.stock ASC LIMIT 20`)
	if err != nil {
		return
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var tenantName, productName, sku string
		var stock int
		rows.Scan(&tenantName, &productName, &sku, &stock)
		if count == 0 {
			log.Println("scheduler: LOW STOCK ALERTS:")
		}
		log.Printf("  [%s] %s (%s) — %d left", tenantName, productName, sku, stock)
		count++
	}
	// ponytail: send email/webhook notification here when integrated
}
