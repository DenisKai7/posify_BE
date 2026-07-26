package repository

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"posify-backend/internal/model"
)

type DiscountRepo struct {
	pool *pgxpool.Pool
}

func NewDiscountRepo(pool *pgxpool.Pool) *DiscountRepo {
	return &DiscountRepo{pool: pool}
}

func (r *DiscountRepo) Create(ctx context.Context, tenantID string, req model.CreateDiscountRequest) (*model.Discount, error) {
	d := &model.Discount{}
	var validUntil interface{}
	if req.ValidUntil != nil {
		validUntil = *req.ValidUntil
	}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO discounts (tenant_id, code, discount_type, discount_value, min_purchase, max_discount, valid_until)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, tenant_id, code, discount_type, discount_value, min_purchase, max_discount, is_active, valid_from, valid_until, created_at`,
		tenantID, req.Code, req.DiscountType, req.DiscountValue, req.MinPurchase, req.MaxDiscount, validUntil,
	).Scan(&d.ID, &d.TenantID, &d.Code, &d.DiscountType, &d.DiscountValue, &d.MinPurchase, &d.MaxDiscount, &d.IsActive, &d.ValidFrom, &d.ValidUntil, &d.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create discount: %w", err)
	}
	return d, nil
}

func (r *DiscountRepo) Validate(ctx context.Context, tenantID, code string, subtotal float64) (*model.ValidateDiscountResponse, error) {
	d := &model.Discount{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, code, discount_type, discount_value, min_purchase, max_discount, is_active, valid_until
		 FROM discounts
		 WHERE tenant_id = $1 AND UPPER(code) = UPPER($2)`,
		tenantID, code,
	).Scan(&d.ID, &d.Code, &d.DiscountType, &d.DiscountValue, &d.MinPurchase, &d.MaxDiscount, &d.IsActive, &d.ValidUntil)
	if err != nil {
		return &model.ValidateDiscountResponse{Valid: false, Message: "Kode diskon tidak ditemukan"}, nil
	}

	if !d.IsActive {
		return &model.ValidateDiscountResponse{Valid: false, Message: "Diskon tidak aktif"}, nil
	}
	if d.ValidUntil != nil && d.ValidUntil.Before(time.Now()) {
		return &model.ValidateDiscountResponse{Valid: false, Message: "Diskon sudah kedaluwarsa"}, nil
	}
	if subtotal < d.MinPurchase {
		return &model.ValidateDiscountResponse{Valid: false, Message: fmt.Sprintf("Minimal pembelian %s", fmt.Sprintf("%.0f", d.MinPurchase))}, nil
	}

	var amount float64
	if d.DiscountType == "PERCENT" {
		amount = subtotal * d.DiscountValue / 100
		if d.MaxDiscount != nil && amount > *d.MaxDiscount {
			amount = *d.MaxDiscount
		}
	} else {
		amount = math.Min(d.DiscountValue, subtotal)
	}

	return &model.ValidateDiscountResponse{
		Valid:          true,
		DiscountType:   d.DiscountType,
		DiscountValue:  d.DiscountValue,
		DiscountAmount: amount,
	}, nil
}

func (r *DiscountRepo) List(ctx context.Context, tenantID string) ([]model.Discount, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, code, discount_type, discount_value, min_purchase, max_discount, is_active, valid_from, valid_until, created_at
		 FROM discounts WHERE tenant_id = $1 ORDER BY created_at DESC`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.Discount
	for rows.Next() {
		var d model.Discount
		if err := rows.Scan(&d.ID, &d.TenantID, &d.Code, &d.DiscountType, &d.DiscountValue, &d.MinPurchase, &d.MaxDiscount, &d.IsActive, &d.ValidFrom, &d.ValidUntil, &d.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	if items == nil {
		items = []model.Discount{}
	}
	return items, rows.Err()
}
