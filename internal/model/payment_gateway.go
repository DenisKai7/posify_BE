package model

import "time"

type PaymentGateway struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Provider    string    `json:"provider"`
	MerchantID  string    `json:"merchant_id"`
	ClientKey   string    `json:"client_key"`
	ServerKey   string    `json:"server_key,omitempty"` // masked in response
	IsProduction bool    `json:"is_production"`
	IsActive    bool      `json:"is_active"`
	WebhookURL  string    `json:"webhook_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdatePaymentGatewayRequest struct {
	MerchantID   *string `json:"merchant_id,omitempty"`
	ClientKey    *string `json:"client_key,omitempty"`
	ServerKey    *string `json:"server_key,omitempty"`
	IsProduction *bool   `json:"is_production,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

type MembershipTier struct {
	ID                 string    `json:"id"`
	TenantID           string    `json:"tenant_id"`
	Name               string    `json:"name"`
	MinSpend           float64   `json:"min_spend"`
	MultiplierPoints   float64   `json:"multiplier_points"`
	DiscountPercentage float64   `json:"discount_percentage"`
	SortOrder          int       `json:"sort_order"`
	CreatedAt          time.Time `json:"created_at"`
}

type UpdateTierRequest struct {
	Name               *string  `json:"name,omitempty"`
	MinSpend           *float64 `json:"min_spend,omitempty"`
	MultiplierPoints   *float64 `json:"multiplier_points,omitempty"`
	DiscountPercentage *float64 `json:"discount_percentage,omitempty"`
}
