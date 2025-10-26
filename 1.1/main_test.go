package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProcessPaymentSuccess(t *testing.T) {
	service := NewPaymentService()
	service.SetBalance("user123", 1000.00)

	req := PaymentRequest{
		UserID:        "user123",
		Amount:        -100.00,
		TransactionID: "txn-001",
	}

	resp, err := service.ProcessPayment(req)
	if err != nil {
		t.Fatalf("ProcessPayment failed: %v", err)
	}

	if resp.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", resp.Status)
	}
	if resp.TransactionID != "txn-001" {
		t.Errorf("Expected transactionID 'txn-001', got '%s'", resp.TransactionID)
	}
	if resp.Amount != -100.00 {
		t.Errorf("Expected amount -100.00, got %.2f", resp.Amount)
	}
	if resp.TraceID == "" {
		t.Error("TraceID should not be empty")
	}

	balance := service.GetBalance("user123")
	if balance != 900.00 {
		t.Errorf("Expected balance 900.00 (1000 - 100), got %.2f", balance)
	}
}

func TestProcessPaymentIdempotency(t *testing.T) {
	service := NewPaymentService()
	service.SetBalance("user123", 1000.00)

	req := PaymentRequest{
		UserID:        "user123",
		Amount:        -100.00,
		TransactionID: "txn-001",
	}

	resp1, err := service.ProcessPayment(req)
	if err != nil {
		t.Fatalf("First payment failed: %v", err)
	}

	resp2, err := service.ProcessPayment(req)
	if err != nil {
		t.Fatalf("Second payment failed: %v", err)
	}

	if resp1.TransactionID != resp2.TransactionID {
		t.Errorf("Transaction IDs should match: %s != %s", resp1.TransactionID, resp2.TransactionID)
	}

	balance := service.GetBalance("user123")
	if balance != 900.00 {
		t.Errorf("Balance should be 900.00 (1000 - 100) after idempotent request, got %.2f", balance)
	}
}

func TestProcessPaymentInsufficientFunds(t *testing.T) {
	service := NewPaymentService()
	service.SetBalance("user123", 50.00)

	req := PaymentRequest{
		UserID:        "user123",
		Amount:        -100.00,
		TransactionID: "txn-001",
	}

	_, err := service.ProcessPayment(req)
	if err == nil {
		t.Error("Expected error for insufficient funds")
	}

	balance := service.GetBalance("user123")
	if balance != 50.00 {
		t.Errorf("Balance should remain 50.00, got %.2f", balance)
	}
}

func TestProcessPaymentValidation(t *testing.T) {
	service := NewPaymentService()
	service.SetBalance("user123", 1000.00)

	tests := []struct {
		name string
		req  PaymentRequest
	}{
		{
			name: "Missing TransactionID",
			req: PaymentRequest{
				UserID: "user123",
				Amount: 100.00,
			},
		},
		{
			name: "Missing UserID",
			req: PaymentRequest{
				Amount:        100.00,
				TransactionID: "txn-001",
			},
		},
		{
			name: "Zero Amount",
			req: PaymentRequest{
				UserID:        "user123",
				Amount:        0,
				TransactionID: "txn-001",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ProcessPayment(tt.req)
			if err == nil {
				t.Errorf("Expected error for %s", tt.name)
			}
		})
	}
}

func TestProcessPaymentPositiveAmount(t *testing.T) {
	service := NewPaymentService()
	service.SetBalance("user123", 100.00)

	req := PaymentRequest{
		UserID:        "user123",
		Amount:        50.00,
		TransactionID: "txn-add-001",
	}

	resp, err := service.ProcessPayment(req)
	if err != nil {
		t.Fatalf("ProcessPayment failed: %v", err)
	}

	if resp.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", resp.Status)
	}

	balance := service.GetBalance("user123")
	expectedBalance := 150.00
	if balance != expectedBalance {
		t.Errorf("Expected balance %.2f (100 + 50), got %.2f", expectedBalance, balance)
	}
}

func TestProcessPaymentNegativeAmount(t *testing.T) {
	service := NewPaymentService()
	service.SetBalance("user123", 100.00)

	req := PaymentRequest{
		UserID:        "user123",
		Amount:        -30.00,
		TransactionID: "txn-deduct-001",
	}

	resp, err := service.ProcessPayment(req)
	if err != nil {
		t.Fatalf("ProcessPayment failed: %v", err)
	}

	if resp.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", resp.Status)
	}

	balance := service.GetBalance("user123")
	expectedBalance := 70.00
	if balance != expectedBalance {
		t.Errorf("Expected balance %.2f (100 - 30), got %.2f", expectedBalance, balance)
	}
}

func TestProcessPaymentNegativeAmountInsufficientFunds(t *testing.T) {
	service := NewPaymentService()
	service.SetBalance("user123", 50.00)

	req := PaymentRequest{
		UserID:        "user123",
		Amount:        -100.00,
		TransactionID: "txn-deduct-fail",
	}

	_, err := service.ProcessPayment(req)
	if err == nil {
		t.Error("Expected error for insufficient funds with negative amount")
	}

	balance := service.GetBalance("user123")
	if balance != 50.00 {
		t.Errorf("Balance should remain 50.00, got %.2f", balance)
	}
}

func TestGetBalanceSuccess(t *testing.T) {
	service := NewPaymentService()
	service.SetBalance("user123", 500.00)

	balance := service.GetBalance("user123")
	if balance != 500.00 {
		t.Errorf("Expected balance 500.00, got %.2f", balance)
	}
}

func TestGetBalanceNonExistent(t *testing.T) {
	service := NewPaymentService()

	balance := service.GetBalance("nonexistent")
	if balance != 0 {
		t.Errorf("Expected balance 0 for nonexistent user, got %.2f", balance)
	}
}

func TestSetBalanceSuccess(t *testing.T) {
	service := NewPaymentService()

	service.SetBalance("user123", 1000.00)
	balance := service.GetBalance("user123")
	if balance != 1000.00 {
		t.Errorf("Expected balance 1000.00, got %.2f", balance)
	}

	service.SetBalance("user123", 500.00)
	balance = service.GetBalance("user123")
	if balance != 500.00 {
		t.Errorf("Expected balance 500.00 after update, got %.2f", balance)
	}
}

func TestSetBalanceMultipleUsers(t *testing.T) {
	service := NewPaymentService()

	service.SetBalance("user1", 100.00)
	service.SetBalance("user2", 200.00)
	service.SetBalance("user3", 300.00)

	if service.GetBalance("user1") != 100.00 {
		t.Error("user1 balance mismatch")
	}
	if service.GetBalance("user2") != 200.00 {
		t.Error("user2 balance mismatch")
	}
	if service.GetBalance("user3") != 300.00 {
		t.Error("user3 balance mismatch")
	}
}

func TestGetTransactionSuccess(t *testing.T) {
	service := NewPaymentService()
	service.SetBalance("user123", 1000.00)

	req := PaymentRequest{
		UserID:        "user123",
		Amount:        100.00,
		TransactionID: "txn-001",
	}

	_, err := service.ProcessPayment(req)
	if err != nil {
		t.Fatalf("ProcessPayment failed: %v", err)
	}

	txn, exists := service.GetTransaction("txn-001")
	if !exists {
		t.Error("Transaction should exist")
	}
	if txn.TransactionID != "txn-001" {
		t.Errorf("Expected transactionID 'txn-001', got '%s'", txn.TransactionID)
	}
	if txn.UserID != "user123" {
		t.Errorf("Expected userID 'user123', got '%s'", txn.UserID)
	}
	if txn.Amount != 100.00 {
		t.Errorf("Expected amount 100.00, got %.2f", txn.Amount)
	}
	if txn.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", txn.Status)
	}
}

func TestGetTransactionNonExistent(t *testing.T) {
	service := NewPaymentService()

	txn, exists := service.GetTransaction("nonexistent")
	if exists {
		t.Error("Transaction should not exist")
	}
	if txn != nil {
		t.Error("Transaction should be nil")
	}
}

// Those tests below from this line should be integration test, not unit test
// But I added it to present the integration test in reality
func TestHandlePaymentSuccess(t *testing.T) {
	service := NewPaymentService()
	service.SetBalance("user123", 1000.00)

	reqBody := PaymentRequest{
		UserID:        "user123",
		Amount:        -100.00,
		TransactionID: "txn-001",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/pay", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	service.HandlePayment(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp PaymentResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", resp.Status)
	}
	if resp.TraceID == "" {
		t.Error("TraceID should not be empty")
	}
}

func TestHandlePaymentInvalidMethod(t *testing.T) {
	service := NewPaymentService()

	req := httptest.NewRequest(http.MethodGet, "/pay", nil)
	w := httptest.NewRecorder()

	service.HandlePayment(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandlePaymentInvalidJSON(t *testing.T) {
	service := NewPaymentService()

	req := httptest.NewRequest(http.MethodPost, "/pay", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	service.HandlePayment(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandlePaymentIdempotency(t *testing.T) {
	service := NewPaymentService()
	service.SetBalance("user123", 1000.00)

	reqBody := PaymentRequest{
		UserID:        "user123",
		Amount:        -100.00,
		TransactionID: "txn-001",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req1 := httptest.NewRequest(http.MethodPost, "/pay", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()

	service.HandlePayment(w1, req1)

	var resp1 PaymentResponse
	json.NewDecoder(w1.Body).Decode(&resp1)

	req2 := httptest.NewRequest(http.MethodPost, "/pay", bytes.NewBuffer(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()

	service.HandlePayment(w2, req2)

	var resp2 PaymentResponse
	json.NewDecoder(w2.Body).Decode(&resp2)

	if resp1.TransactionID != resp2.TransactionID {
		t.Error("Idempotent requests should return same transaction ID")
	}

	balance := service.GetBalance("user123")
	if balance != 900.00 {
		t.Errorf("Expected balance 900.00 (1000 - 100), got %.2f", balance)
	}
}
