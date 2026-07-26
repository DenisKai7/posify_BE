package model

import "time"

type StockAdjustment struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	ProductID string    `json:"product_id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`
	QtyChange int       `json:"qty_change"`
	Reason    string    `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	// Joined fields for display
	ProductName string `json:"product_name,omitempty"`
	SKU         string `json:"sku,omitempty"`
	UserName    string `json:"user_name,omitempty"`
}

type CreateStockAdjustmentRequest struct {
	ProductID string `json:"product_id"`
	Type      string `json:"type"` // IN, OUT, OPNAME, WASTE
	QtyChange int    `json:"qty_change"`
	Reason    string `json:"reason"`
}

type LowStockAlert struct {
	ProductID   string  `json:"product_id"`
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Stock       int     `json:"stock"`
	Threshold   int     `json:"threshold"`
	Price       float64 `json:"price"`
}
