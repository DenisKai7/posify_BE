package model

import "time"

type ReportSummary struct {
	RevenueToday    float64       `json:"revenue_today"`
	RevenueMonth    float64       `json:"revenue_month"`
	TxCountToday    int           `json:"tx_count_today"`
	TxCountMonth    int           `json:"tx_count_month"`
	ActiveProducts  int           `json:"active_products"`
	LowStockItems   []LowStockItem `json:"low_stock_items"`
}

type LowStockItem struct {
	ID    string `json:"id"`
	SKU   string `json:"sku"`
	Name  string `json:"name"`
	Stock int    `json:"stock"`
}

type SalesChartPoint struct {
	Date    time.Time `json:"date"`
	Revenue float64   `json:"revenue"`
	Count   int       `json:"count"`
}
