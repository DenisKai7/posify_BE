package model

import "time"

type Customer struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Name       string    `json:"name"`
	Phone      string    `json:"phone,omitempty"`
	Email      string    `json:"email,omitempty"`
	Points     int       `json:"points"`
	Tier       string    `json:"tier"`
	TotalSpent float64   `json:"total_spent"`
	VisitCount int       `json:"visit_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateCustomerRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

type RedeemPointsRequest struct {
	CustomerID string `json:"customer_id"`
	Points     int    `json:"points"`
}

type AddPointsRequest struct {
	CustomerID string  `json:"customer_id"`
	Amount     float64 `json:"amount"` // transaction amount, auto-calculate points
	VisitCount int     `json:"visit_count"`
}
