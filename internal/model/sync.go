package model

import "time"

// SyncPushRequest is the incoming batch from the frontend outbox
type SyncPushRequest struct {
	Transactions []OfflineTransaction `json:"transactions"`
}

type OfflineTransaction struct {
	OfflineTransactionID string                `json:"offlineTransactionId"`
	CashierID            string                `json:"cashierId"`
	TotalAmount          float64               `json:"totalAmount"`
	PayAmount            float64               `json:"payAmount"`
	ChangeAmount         float64               `json:"changeAmount"`
	PaymentMethod        string                `json:"paymentMethod"`
	CreatedAt            string                `json:"createdAt"`
	Items                []OfflineTransactionItem `json:"items"`
}

type OfflineTransactionItem struct {
	ProductID   string  `json:"productId"`
	ProductName string  `json:"productName"`
	PriceAtSale float64 `json:"priceAtSale"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}

type SyncPushResponse struct {
	SyncedIDs []string `json:"synced_ids"`
	Errors    []string `json:"errors,omitempty"`
}

// SyncPullResponse returns current product catalog for the tenant
type SyncPullResponse struct {
	Products []SyncProduct `json:"products"`
}

type SyncProduct struct {
	ID        string    `json:"id"`
	SKU       string    `json:"sku"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Stock     int       `json:"stock"`
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// SyncLog records each push batch for audit
type SyncLog struct {
	TenantID       string
	UserID         string
	BatchSize      int
	SyncedCount    int
	DuplicateCount int
	ErrorCount     int
	Errors         []string
}
