package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"posify-backend/internal/model"
)

type TierRepo struct {
	pool *pgxpool.Pool
}

func NewTierRepo(pool *pgxpool.Pool) *TierRepo {
	return &TierRepo{pool: pool}
}

func (r *TierRepo) List(ctx context.Context, tenantID string) ([]model.MembershipTier, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, min_spend, multiplier_points, discount_percentage, sort_order, created_at
		 FROM membership_tiers WHERE tenant_id = $1 ORDER BY sort_order`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.MembershipTier
	for rows.Next() {
		var t model.MembershipTier
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Name, &t.MinSpend, &t.MultiplierPoints, &t.DiscountPercentage, &t.SortOrder, &t.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	if items == nil {
		items = []model.MembershipTier{}
	}
	return items, rows.Err()
}

func (r *TierRepo) Update(ctx context.Context, tenantID, tierID string, req model.UpdateTierRequest) (*model.MembershipTier, error) {
	t := &model.MembershipTier{}
	err := r.pool.QueryRow(ctx,
		`UPDATE membership_tiers SET
			name = COALESCE($3, name),
			min_spend = COALESCE($4, min_spend),
			multiplier_points = COALESCE($5, multiplier_points),
			discount_percentage = COALESCE($6, discount_percentage)
		 WHERE id = $1 AND tenant_id = $2
		 RETURNING id, tenant_id, name, min_spend, multiplier_points, discount_percentage, sort_order, created_at`,
		tierID, tenantID, req.Name, req.MinSpend, req.MultiplierPoints, req.DiscountPercentage,
	).Scan(&t.ID, &t.TenantID, &t.Name, &t.MinSpend, &t.MultiplierPoints, &t.DiscountPercentage, &t.SortOrder, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("update tier: %w", err)
	}
	return t, nil
}

// AutoUpgradeTier checks and upgrades customer tier based on total_spent
func (r *TierRepo) AutoUpgradeTier(ctx context.Context, tenantID, customerID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE customers c SET tier_id = (
			SELECT t.id FROM membership_tiers t
			WHERE t.tenant_id = c.tenant_id AND t.min_spend <= c.total_spent
			ORDER BY t.min_spend DESC LIMIT 1
		 ), tier = (
			SELECT t.name FROM membership_tiers t
			WHERE t.tenant_id = c.tenant_id AND t.min_spend <= c.total_spent
			ORDER BY t.min_spend DESC LIMIT 1
		 ), updated_at = NOW()
		 WHERE c.id = $1 AND c.tenant_id = $2`,
		customerID, tenantID,
	)
	return err
}
