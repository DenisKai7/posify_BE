package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"posify-backend/internal/model"
)

type PaymentRepo struct {
	pool *pgxpool.Pool
}

func NewPaymentRepo(pool *pgxpool.Pool) *PaymentRepo {
	return &PaymentRepo{pool: pool}
}

func (r *PaymentRepo) CreateQRIS(ctx context.Context, tenantID string, req model.GenerateQRISRequest) (*model.QRISPayment, error) {
	// Generate mock QR string (in production, call Midtrans/Xendit API)
	qrString := fmt.Sprintf("00020101021226580011ID.CO.QRIS.WWW0215%s0303UME51440014ID.CO.QRIS.WWW0215%s0303UME520459995303360540%.0f5802ID5909POSIFY6007Jakarta", req.OrderID, req.OrderID, req.Amount)
	qrURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=%s", qrString)

	p := &model.QRISPayment{}
	expiresAt := time.Now().Add(5 * time.Minute)

	err := r.pool.QueryRow(ctx,
		`INSERT INTO qris_payments (tenant_id, order_id, amount, qr_string, qr_url, status, expires_at)
		 VALUES ($1, $2, $3, $4, $5, 'PENDING', $6)
		 RETURNING id, tenant_id, order_id, amount, qr_string, qr_url, status, expires_at, created_at`,
		tenantID, req.OrderID, req.Amount, qrString, qrURL, expiresAt,
	).Scan(&p.ID, &p.TenantID, &p.OrderID, &p.Amount, &p.QRString, &p.QRURL, &p.Status, &p.ExpiresAt, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create qris: %w", err)
	}
	return p, nil
}

func (r *PaymentRepo) GetStatus(ctx context.Context, orderID string) (*model.QRISStatusResponse, error) {
	s := &model.QRISStatusResponse{}
	err := r.pool.QueryRow(ctx,
		`SELECT order_id, status, COALESCE(payment_method, ''), paid_at FROM qris_payments WHERE order_id = $1`,
		orderID,
	).Scan(&s.OrderID, &s.Status, &s.PaymentMethod, &s.PaidAt)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}
	return s, nil
}

func (r *PaymentRepo) MarkPaid(ctx context.Context, orderID, paymentMethod string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE qris_payments SET status = 'PAID', payment_method = $2, paid_at = NOW()
		 WHERE order_id = $1 AND status = 'PENDING'`,
		orderID, paymentMethod,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("payment not found or already processed")
	}
	return nil
}

func (r *PaymentRepo) MarkExpired(ctx context.Context, orderID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE qris_payments SET status = 'EXPIRED' WHERE order_id = $1 AND status = 'PENDING' AND expires_at < NOW()`,
		orderID,
	)
	return err
}
