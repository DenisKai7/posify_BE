package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/xuri/excelize/v2"

	"posify-backend/internal/middleware"
	"posify-backend/internal/service"
)

type ReportHandler struct {
	svc *service.ReportService
}

func NewReportHandler(svc *service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

// Summary handles GET /api/v1/reports/summary
func (h *ReportHandler) Summary(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	summary, err := h.svc.Summary(r.Context(), tenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// SalesChart handles GET /api/v1/reports/sales-chart?days=7
func (h *ReportHandler) SalesChart(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 && v <= 90 {
			days = v
		}
	}
	data, err := h.svc.SalesChart(r.Context(), tenantID, days)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"chart": data, "days": days})
}

// SalesSummary handles GET /api/v1/reports/sales-summary?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD
func (h *ReportHandler) SalesSummary(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	if startDate == "" || endDate == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "start_date and end_date required"})
		return
	}
	data, err := h.svc.SalesSummary(r.Context(), tenantID, startDate, endDate)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, data)
}

// TopProducts handles GET /api/v1/reports/top-products?limit=10
func (h *ReportHandler) TopProducts(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	data, err := h.svc.TopProducts(r.Context(), tenantID, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"products": data})
}

// PaymentMethods handles GET /api/v1/reports/payment-methods
func (h *ReportHandler) PaymentMethods(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	data, err := h.svc.PaymentMethods(r.Context(), tenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"methods": data})
}

// ExportExcel handles GET /api/v1/reports/export/excel?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD
func (h *ReportHandler) ExportExcel(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	if startDate == "" || endDate == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "start_date and end_date required"})
		return
	}

	// Fetch data
	summary, err := h.svc.SalesSummary(r.Context(), tenantID, startDate, endDate)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	products, err := h.svc.SalesSummaryRows(r.Context(), tenantID, startDate, endDate)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Build Excel
	f := excelize.NewFile()
	defer f.Close()

	// Summary sheet
	f.SetCellValue("Sheet1", "A1", "Laporan Penjualan")
	f.SetCellValue("Sheet1", "A2", fmt.Sprintf("Periode: %s s/d %s", startDate, endDate))
	f.SetCellValue("Sheet1", "A4", "Total Pendapatan")
	f.SetCellValue("Sheet1", "B4", summary.TotalRev)
	f.SetCellValue("Sheet1", "A5", "Total Transaksi")
	f.SetCellValue("Sheet1", "B5", summary.TotalTx)
	f.SetCellValue("Sheet1", "A6", "Rata-rata Transaksi")
	f.SetCellValue("Sheet1", "B6", summary.AvgTx)
	f.SetCellValue("Sheet1", "A7", "Total Item Terjual")
	f.SetCellValue("Sheet1", "B7", summary.TotalItems)

	// Products sheet
	f.NewSheet("Produk Terlaris")
	f.SetCellValue("Produk Terlaris", "A1", "SKU")
	f.SetCellValue("Produk Terlaris", "B1", "Nama Produk")
	f.SetCellValue("Produk Terlaris", "C1", "Qty Terjual")
	f.SetCellValue("Produk Terlaris", "D1", "Total Revenue")
	for i, p := range products {
		row := i + 2
		f.SetCellValue("Produk Terlaris", fmt.Sprintf("A%d", row), p.SKU)
		f.SetCellValue("Produk Terlaris", fmt.Sprintf("B%d", row), p.ProductName)
		f.SetCellValue("Produk Terlaris", fmt.Sprintf("C%d", row), p.TotalQty)
		f.SetCellValue("Produk Terlaris", fmt.Sprintf("D%d", row), p.TotalRevenue)
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=laporan_%s_%s.xlsx", startDate, endDate))
	if err := f.Write(w); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate excel"})
	}
}
