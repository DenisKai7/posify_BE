package service

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"posify-backend/internal/model"
	"posify-backend/internal/repository"
)

type MidtransService struct {
	pool    *pgxpool.Pool
	pgRepo  *repository.PaymentGatewayRepo
	payRepo *repository.PaymentRepo
}

func NewMidtransService(pool *pgxpool.Pool, pgRepo *repository.PaymentGatewayRepo, payRepo *repository.PaymentRepo) *MidtransService {
	return &MidtransService{pool: pool, pgRepo: pgRepo, payRepo: payRepo}
}

type midtransChargeRequest struct {
	PaymentType string            `json:"payment_type"`
	Transaction midtransTransData `json:"transaction_details"`
	Customer    *midtransCustomer `json:"customer_details,omitempty"`
}

type midtransTransData struct {
	OrderID  string  `json:"order_id"`
	GrossAmt float64 `json:"gross_amount"`
}

type midtransCustomer struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone,omitempty"`
}

type midtransQRISResponse struct {
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	TransactionID     string `json:"transaction_id"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	PaymentType       string `json:"payment_type"`
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	Actions           []struct {
		Name   string `json:"name"`
		Method string `json:"method"`
		URL    string `json:"url"`
	} `json:"actions"`
	QRString string `json:"qr_string"`
}

type midtransStatusResponse struct {
	StatusCode        string `json:"status_code"`
	TransactionStatus string `json:"transaction_status"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
}

// ChargeQRIS creates a QRIS charge via Midtrans API
func (s *MidtransService) ChargeQRIS(ctx context.Context, tenantID, orderID string, amount float64, customerName string) (*model.GenerateQRISResponse, error) {
	pg, err := s.pgRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("payment gateway not configured: %w", err)
	}
	if !pg.IsActive {
		return nil, fmt.Errorf("payment gateway is disabled")
	}

	baseURL := "https://api.sandbox.midtrans.com"
	if pg.IsProduction {
		baseURL = "https://api.midtrans.com"
	}

	reqBody := midtransChargeRequest{
		PaymentType: "qris",
		Transaction: midtransTransData{OrderID: orderID, GrossAmt: amount},
	}
	if customerName != "" {
		reqBody.Customer = &midtransCustomer{FullName: customerName}
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequestWithContext(ctx, "POST", baseURL+"/v2/charge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(pg.ServerKey, "")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("midtrans request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var qrResp midtransQRISResponse
	if err := json.Unmarshal(respBody, &qrResp); err != nil {
		return nil, fmt.Errorf("parse midtrans response: %w", err)
	}

	if qrResp.StatusCode != "201" {
		return nil, fmt.Errorf("midtrans error: %s - %s", qrResp.StatusCode, qrResp.StatusMessage)
	}

	// Extract QR URL from actions
	qrURL := ""
	for _, a := range qrResp.Actions {
		if a.Name == "generate-qr-code" {
			qrURL = a.URL
			break
		}
	}

	// Store in local DB
	_, err = s.payRepo.CreateQRISFromMidtrans(ctx, tenantID, orderID, amount, qrResp.QRString, qrURL, qrResp.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("store payment: %w", err)
	}

	return &model.GenerateQRISResponse{
		OrderID:  orderID,
		QRString: qrResp.QRString,
		QRURL:    qrURL,
		Amount:   amount,
	}, nil
}

// HandleWebhook processes Midtrans webhook notification
func (s *MidtransService) HandleWebhook(ctx context.Context, tenantID string, payload map[string]any) error {
	pg, err := s.pgRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("gateway not configured")
	}

	orderID, _ := payload["order_id"].(string)
	statusCode, _ := payload["status_code"].(string)
	grossAmount, _ := payload["gross_amount"].(string)
	signatureKey, _ := payload["signature_key"].(string)
	transactionStatus, _ := payload["transaction_status"].(string)

	// Validate signature: SHA512(order_id + status_code + gross_amount + server_key)
	expected := sha512hex(orderID + statusCode + grossAmount + pg.ServerKey)
	if signatureKey != expected {
		return fmt.Errorf("invalid signature")
	}

	// Map status
	newStatus := "PENDING"
	switch transactionStatus {
	case "settlement":
		newStatus = "PAID"
	case "expire", "cancel", "deny":
		newStatus = "EXPIRED"
	case "failure":
		newStatus = "FAILED"
	}

	return s.payRepo.UpdateStatus(ctx, orderID, newStatus)
}

// CheckStatus polls Midtrans for transaction status
func (s *MidtransService) CheckStatus(ctx context.Context, tenantID, orderID string) (string, error) {
	pg, err := s.pgRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return "", err
	}

	baseURL := "https://api.sandbox.midtrans.com"
	if pg.IsProduction {
		baseURL = "https://api.midtrans.com"
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", baseURL+"/v2/"+orderID+"/status", nil)
	req.SetBasicAuth(pg.ServerKey, "")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var statusResp midtransStatusResponse
	json.NewDecoder(resp.Body).Decode(&statusResp)

	switch statusResp.TransactionStatus {
	case "settlement":
		return "PAID", nil
	case "expire", "cancel", "deny":
		return "EXPIRED", nil
	case "failure":
		return "FAILED", nil
	default:
		return "PENDING", nil
	}
}

func sha512hex(input string) string {
	h := sha512.Sum512([]byte(input))
	return hex.EncodeToString(h[:])
}

// Base64Encode for basic auth
func Base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
