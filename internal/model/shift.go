package model

import "time"

type Shift struct {
	ID                string     `json:"id"`
	TenantID          string     `json:"tenant_id"`
	UserID            string     `json:"user_id"`
	InitialCash       float64    `json:"initial_cash"`
	ActualCash        *float64   `json:"actual_cash,omitempty"`
	ExpectedCash      *float64   `json:"expected_cash,omitempty"`
	Difference        *float64   `json:"difference,omitempty"`
	TotalSales        float64    `json:"total_sales"`
	TotalTransactions int        `json:"total_transactions"`
	StartedAt         time.Time  `json:"started_at"`
	EndedAt           *time.Time `json:"ended_at,omitempty"`
	Status            string     `json:"status"`
}

type StartShiftRequest struct {
	InitialCash float64 `json:"initial_cash"`
}

type CloseShiftRequest struct {
	ActualCash float64 `json:"actual_cash"`
}
