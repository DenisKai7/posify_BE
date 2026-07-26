package model

import "time"

type QRISPayment struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	OrderID       string     `json:"order_id"`
	Amount        float64    `json:"amount"`
	QRString      string     `json:"qr_string"`
	QRURL         string     `json:"qr_url"`
	Status        string     `json:"status"` // PENDING, PAID, EXPIRED, FAILED
	PaymentMethod string     `json:"payment_method,omitempty"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	ExpiresAt     time.Time  `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type GenerateQRISRequest struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
}

type GenerateQRISResponse struct {
	OrderID   string `json:"order_id"`
	QRString  string `json:"qr_string"`
	QRURL     string `json:"qr_url"`
	Amount    float64 `json:"amount"`
	ExpiresAt time.Time `json:"expires_at"`
}

type QRISStatusResponse struct {
	OrderID       string     `json:"order_id"`
	Status        string     `json:"status"`
	PaymentMethod string     `json:"payment_method,omitempty"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
}

type WebhookPayload struct {
	OrderID       string  `json:"order_id"`
	Status        string  `json:"status"`
	PaymentMethod string  `json:"payment_method"`
	Amount        float64 `json:"amount"`
	Signature     string  `json:"signature"`
}
