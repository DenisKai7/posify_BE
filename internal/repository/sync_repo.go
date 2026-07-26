package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"posify-backend/internal/model"
)

type SyncRepo struct {
	pool *pgxpool.Pool
}

func NewSyncRepo(pool *pgxpool.Pool) *SyncRepo {
	return &SyncRepo{pool: pool}
}

// SyncTransaction inserts a single offline transaction idempotently.
// Returns true if inserted (new), false if duplicate (skipped).
func (r *SyncRepo) SyncTransaction(ctx context.Context, tx pgx.Tx, tenantID string, ot model.OfflineTransaction) (bool, error) {
	// Idempotency check
	var exists bool
	err := tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM transactions WHERE offline_transaction_id = $1)`,
		ot.OfflineTransactionID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("idempotency check: %w", err)
	}
	if exists {
		return false, nil // duplicate, skip
	}

	createdAt, err := time.Parse(time.RFC3339, ot.CreatedAt)
	if err != nil {
		createdAt = time.Now()
	}

	// Insert transaction
	var txID string
	err = tx.QueryRow(ctx,
		`INSERT INTO transactions (tenant_id, cashier_id, offline_transaction_id, total_amount, pay_amount, change_amount, payment_method, payment_status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'PAID', $8)
		 RETURNING id`,
		tenantID, ot.CashierID, ot.OfflineTransactionID,
		ot.TotalAmount, ot.PayAmount, ot.ChangeAmount,
		ot.PaymentMethod, createdAt,
	).Scan(&txID)
	if err != nil {
		return false, fmt.Errorf("insert transaction: %w", err)
	}

	// Insert items & decrement stock
	for _, item := range ot.Items {
		_, err = tx.Exec(ctx,
			`INSERT INTO transaction_items (transaction_id, product_id, product_name, price_at_sale, quantity, subtotal)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			txID, item.ProductID, item.ProductName, item.PriceAtSale, item.Quantity, item.Subtotal,
		)
		if err != nil {
			return false, fmt.Errorf("insert item: %w", err)
		}

		_, err = tx.Exec(ctx,
			`UPDATE products SET stock = stock - $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`,
			item.Quantity, item.ProductID, tenantID,
		)
		if err != nil {
			return false, fmt.Errorf("update stock: %w", err)
		}
	}

	return true, nil
}

// GetProducts returns active products for a tenant, optionally filtered by last sync time (delta sync / LWW)
func (r *SyncRepo) GetProducts(ctx context.Context, tenantID string, since *time.Time) ([]model.SyncProduct, error) {
	var query string
	var args []any

	if since != nil {
		query = `SELECT id, sku, name, price, stock, version, updated_at
			FROM products
			WHERE tenant_id = $1 AND is_active = TRUE AND deleted_at IS NULL AND updated_at > $2
			ORDER BY name`
		args = []any{tenantID, *since}
	} else {
		query = `SELECT id, sku, name, price, stock, version, updated_at
			FROM products
			WHERE tenant_id = $1 AND is_active = TRUE AND deleted_at IS NULL
			ORDER BY name`
		args = []any{tenantID}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	var products []model.SyncProduct
	for rows.Next() {
		var p model.SyncProduct
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Price, &p.Stock, &p.Version, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

// InsertSyncLog records a push batch for audit
func (r *SyncRepo) InsertSyncLog(ctx context.Context, sl model.SyncLog) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sync_logs (tenant_id, user_id, batch_size, synced_count, duplicate_count, error_count, errors)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		sl.TenantID, sl.UserID, sl.BatchSize, sl.SyncedCount, sl.DuplicateCount, sl.ErrorCount, sl.Errors,
	)
	return err
}
