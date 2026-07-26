package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"posify-backend/internal/model"
)

type PaymentGatewayRepo struct {
	pool *pgxpool.Pool
}

func NewPaymentGatewayRepo(pool *pgxpool.Pool) *PaymentGatewayRepo {
	return &PaymentGatewayRepo{pool: pool}
}

func (r *PaymentGatewayRepo) GetByTenant(ctx context.Context, tenantID string) (*model.PaymentGateway, error) {
	pg := &model.PaymentGateway{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, provider, COALESCE(merchant_id,''), COALESCE(client_key,''),
		 COALESCE(server_key,''), is_production, is_active, COALESCE(webhook_url,''), created_at, updated_at
		 FROM payment_gateways WHERE tenant_id = $1 AND provider = 'MIDTRANS'`,
		tenantID,
	).Scan(&pg.ID, &pg.TenantID, &pg.Provider, &pg.MerchantID, &pg.ClientKey,
		&pg.ServerKey, &pg.IsProduction, &pg.IsActive, &pg.WebhookURL, &pg.CreatedAt, &pg.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("payment gateway not configured: %w", err)
	}
	return pg, nil
}

func (r *PaymentGatewayRepo) Upsert(ctx context.Context, tenantID string, req model.UpdatePaymentGatewayRequest) (*model.PaymentGateway, error) {
	// First try to get existing
	existing, _ := r.GetByTenant(ctx, tenantID)

	if existing == nil {
		// Insert new
		pg := &model.PaymentGateway{}
		err := r.pool.QueryRow(ctx,
			`INSERT INTO payment_gateways (tenant_id, provider, merchant_id, client_key, server_key, is_production, is_active, webhook_url)
			 VALUES ($1, 'MIDTRANS', $2, $3, $4, $5, $6, $7)
			 RETURNING id, tenant_id, provider, merchant_id, client_key, server_key, is_production, is_active, webhook_url, created_at, updated_at`,
			tenantID,
			derefStr(req.MerchantID), derefStr(req.ClientKey), derefStr(req.ServerKey),
			derefBool(req.IsProduction, false), derefBool(req.IsActive, true),
			fmt.Sprintf("/api/v1/payments/midtrans/webhook"),
		).Scan(&pg.ID, &pg.TenantID, &pg.Provider, &pg.MerchantID, &pg.ClientKey,
			&pg.ServerKey, &pg.IsProduction, &pg.IsActive, &pg.WebhookURL, &pg.CreatedAt, &pg.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("insert payment gateway: %w", err)
		}
		return pg, nil
	}

	// Update existing
	pg := &model.PaymentGateway{}
	err := r.pool.QueryRow(ctx,
		`UPDATE payment_gateways SET
			merchant_id = COALESCE($2, merchant_id),
			client_key = COALESCE($3, client_key),
			server_key = COALESCE($4, server_key),
			is_production = COALESCE($5, is_production),
			is_active = COALESCE($6, is_active),
			updated_at = NOW()
		 WHERE tenant_id = $1 AND provider = 'MIDTRANS'
		 RETURNING id, tenant_id, provider, merchant_id, client_key, server_key, is_production, is_active, webhook_url, created_at, updated_at`,
		tenantID, req.MerchantID, req.ClientKey, req.ServerKey, req.IsProduction, req.IsActive,
	).Scan(&pg.ID, &pg.TenantID, &pg.Provider, &pg.MerchantID, &pg.ClientKey,
		&pg.ServerKey, &pg.IsProduction, &pg.IsActive, &pg.WebhookURL, &pg.CreatedAt, &pg.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update payment gateway: %w", err)
	}
	return pg, nil
}

// MaskServerKey hides the server key for safe API response
func MaskServerKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefBool(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}
