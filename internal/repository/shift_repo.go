package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"posify-backend/internal/model"
)

type ShiftRepo struct {
	pool *pgxpool.Pool
}

func NewShiftRepo(pool *pgxpool.Pool) *ShiftRepo {
	return &ShiftRepo{pool: pool}
}

func (r *ShiftRepo) Start(ctx context.Context, tenantID, userID string, initialCash float64) (*model.Shift, error) {
	// Check if there's already an open shift
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM shifts WHERE tenant_id = $1 AND user_id = $2 AND status = 'OPEN')`,
		tenantID, userID,
	).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("check open shift: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("shift already open")
	}

	s := &model.Shift{}
	err = r.pool.QueryRow(ctx,
		`INSERT INTO shifts (tenant_id, user_id, initial_cash, status)
		 VALUES ($1, $2, $3, 'OPEN')
		 RETURNING id, tenant_id, user_id, initial_cash, total_sales, total_transactions, started_at, status`,
		tenantID, userID, initialCash,
	).Scan(&s.ID, &s.TenantID, &s.UserID, &s.InitialCash, &s.TotalSales, &s.TotalTransactions, &s.StartedAt, &s.Status)
	if err != nil {
		return nil, fmt.Errorf("insert shift: %w", err)
	}
	return s, nil
}

func (r *ShiftRepo) Close(ctx context.Context, tenantID, userID string, actualCash float64) (*model.Shift, error) {
	s := &model.Shift{}
	err := r.pool.QueryRow(ctx,
		`UPDATE shifts SET
			actual_cash = $3,
			expected_cash = initial_cash + total_sales,
			difference = $3 - (initial_cash + total_sales),
			ended_at = NOW(),
			status = 'CLOSED'
		 WHERE tenant_id = $1 AND user_id = $2 AND status = 'OPEN'
		 RETURNING id, tenant_id, user_id, initial_cash, actual_cash, expected_cash, difference,
		           total_sales, total_transactions, started_at, ended_at, status`,
		tenantID, userID, actualCash,
	).Scan(&s.ID, &s.TenantID, &s.UserID, &s.InitialCash, &s.ActualCash, &s.ExpectedCash,
		&s.Difference, &s.TotalSales, &s.TotalTransactions, &s.StartedAt, &s.EndedAt, &s.Status)
	if err != nil {
		return nil, fmt.Errorf("close shift: %w", err)
	}
	return s, nil
}

func (r *ShiftRepo) GetCurrent(ctx context.Context, tenantID, userID string) (*model.Shift, error) {
	s := &model.Shift{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, user_id, initial_cash, total_sales, total_transactions, started_at, status
		 FROM shifts WHERE tenant_id = $1 AND user_id = $2 AND status = 'OPEN'`,
		tenantID, userID,
	).Scan(&s.ID, &s.TenantID, &s.UserID, &s.InitialCash, &s.TotalSales, &s.TotalTransactions, &s.StartedAt, &s.Status)
	if err != nil {
		return nil, fmt.Errorf("no open shift: %w", err)
	}
	return s, nil
}

// AddToShiftSales increments sales counters on the open shift
func (r *ShiftRepo) AddToShiftSales(ctx context.Context, shiftID string, amount float64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE shifts SET total_sales = total_sales + $2, total_transactions = total_transactions + 1 WHERE id = $1 AND status = 'OPEN'`,
		shiftID, amount,
	)
	return err
}
