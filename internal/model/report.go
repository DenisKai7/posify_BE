package model

import "time"

type ReportSummary struct {
	RevenueToday   float64        `json:"revenue_today"`
	RevenueMonth   float64        `json:"revenue_month"`
	TxCountToday   int            `json:"tx_count_today"`
	TxCountMonth   int            `json:"tx_count_month"`
	ActiveProducts int            `json:"active_products"`
	LowStockItems  []LowStockItem `json:"low_stock_items"`
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

type SalesSummary struct {
	StartDate  string  `json:"start_date"`
	EndDate    string  `json:"end_date"`
	TotalRev   float64 `json:"total_revenue"`
	TotalTx    int     `json:"total_transactions"`
	AvgTx      float64 `json:"avg_transaction"`
	TotalItems int     `json:"total_items_sold"`
}

type TopProduct struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	TotalQty    int     `json:"total_qty"`
	TotalRevenue float64 `json:"total_revenue"`
}

type PaymentMethodBreakdown struct {
	Method  string  `json:"method"`
	Count   int     `json:"count"`
	Total   float64 `json:"total"`
}
