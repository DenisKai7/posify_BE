package model

import "time"

type Discount struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	Code          string     `json:"code"`
	DiscountType  string     `json:"discount_type"`
	DiscountValue float64    `json:"discount_value"`
	MinPurchase   float64    `json:"min_purchase"`
	MaxDiscount   *float64   `json:"max_discount,omitempty"`
	IsActive      bool       `json:"is_active"`
	ValidFrom     time.Time  `json:"valid_from"`
	ValidUntil    *time.Time `json:"valid_until,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type CreateDiscountRequest struct {
	Code          string   `json:"code"`
	DiscountType  string   `json:"discount_type"`
	DiscountValue float64  `json:"discount_value"`
	MinPurchase   float64  `json:"min_purchase"`
	MaxDiscount   *float64 `json:"max_discount,omitempty"`
	ValidUntil    *string  `json:"valid_until,omitempty"`
}

type ValidateDiscountRequest struct {
	Code        string  `json:"code"`
	Subtotal    float64 `json:"subtotal"`
}

type ValidateDiscountResponse struct {
	Valid          bool    `json:"valid"`
	DiscountType   string  `json:"discount_type,omitempty"`
	DiscountValue  float64 `json:"discount_value,omitempty"`
	DiscountAmount float64 `json:"discount_amount,omitempty"`
	Message        string  `json:"message,omitempty"`
}
