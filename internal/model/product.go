package model

import "time"

type Product struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	SKU       string    `json:"sku"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	CostPrice float64   `json:"cost_price"`
	Stock     int       `json:"stock"`
	Category  string    `json:"category"`
	Version   int       `json:"version"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	SKU       string  `json:"sku"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	CostPrice float64 `json:"cost_price"`
	Stock     int     `json:"stock"`
	Category  string  `json:"category"`
}

type UpdateProductRequest struct {
	SKU       *string  `json:"sku,omitempty"`
	Name      *string  `json:"name,omitempty"`
	Price     *float64 `json:"price,omitempty"`
	CostPrice *float64 `json:"cost_price,omitempty"`
	Stock     *int     `json:"stock,omitempty"`
	Category  *string  `json:"category,omitempty"`
}
