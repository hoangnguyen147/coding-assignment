package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type PaymentRequest struct {
	UserID        string  `json:"userID"`
	Amount        float64 `json:"amount"`
	TransactionID string  `json:"transactionID"`
}

type PaymentResponse struct {
	TraceID       string    `json:"traceID"`
	TransactionID string    `json:"transactionID"`
	UserID        string    `json:"userID"`
	Amount        float64   `json:"amount"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
	ProcessedAt   time.Time `json:"processedAt"`
}

type Transaction struct {
	TransactionID string
	UserID        string
	Amount        float64
	Status        string
	ProcessedAt   time.Time
}

type PaymentService struct {
	mu           sync.RWMutex
	transactions map[string]*Transaction
	balances     map[string]float64
}

func NewPaymentService() *PaymentService {
	return &PaymentService{
		transactions: make(map[string]*Transaction),
		balances:     make(map[string]float64),
	}
}

func (s *PaymentService) ProcessPayment(req PaymentRequest) (*PaymentResponse, error) {
	traceID := uuid.New().String()

	if req.TransactionID == "" {
		log.Printf("[%s] ERROR: transactionID is required", traceID)
		return nil, fmt.Errorf("transactionID is required")
	}
	if req.UserID == "" {
		log.Printf("[%s] ERROR: userID is required", traceID)
		return nil, fmt.Errorf("userID is required")
	}
	if req.Amount <= 0 {
		log.Printf("[%s] ERROR: amount must be greater than 0", traceID)
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if existingTxn, exists := s.transactions[req.TransactionID]; exists {
		log.Printf("[%s] IDEMPOTENT: Transaction %s already processed", traceID, req.TransactionID)
		return &PaymentResponse{
			TraceID:       traceID,
			TransactionID: existingTxn.TransactionID,
			UserID:        existingTxn.UserID,
			Amount:        existingTxn.Amount,
			Status:        existingTxn.Status,
			Message:       "Transaction already processed (idempotent response)",
			ProcessedAt:   existingTxn.ProcessedAt,
		}, nil
	}

	balance := s.balances[req.UserID]
	if balance < req.Amount {
		log.Printf("[%s] ERROR: insufficient funds for user %s: balance=%.2f, required=%.2f",
			traceID, req.UserID, balance, req.Amount)
		return nil, fmt.Errorf("insufficient funds: balance=%.2f, required=%.2f", balance, req.Amount)
	}

	s.balances[req.UserID] = balance - req.Amount

	txn := &Transaction{
		TransactionID: req.TransactionID,
		UserID:        req.UserID,
		Amount:        req.Amount,
		Status:        "success",
		ProcessedAt:   time.Now(),
	}
	s.transactions[req.TransactionID] = txn

	log.Printf("[%s] SUCCESS: Processed payment %s for user %s, amount %.2f",
		traceID, req.TransactionID, req.UserID, req.Amount)

	return &PaymentResponse{
		TraceID:       traceID,
		TransactionID: txn.TransactionID,
		UserID:        txn.UserID,
		Amount:        txn.Amount,
		Status:        txn.Status,
		Message:       "Payment processed successfully",
		ProcessedAt:   txn.ProcessedAt,
	}, nil
}

func (s *PaymentService) GetBalance(userID string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.balances[userID]
}

func (s *PaymentService) SetBalance(userID string, balance float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.balances[userID] = balance
}

func (s *PaymentService) GetTransaction(transactionID string) (*Transaction, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	txn, exists := s.transactions[transactionID]
	return txn, exists
}

func (s *PaymentService) HandlePayment(w http.ResponseWriter, r *http.Request) {
	traceID := uuid.New().String()

	if r.Method != http.MethodPost {
		log.Printf("[%s] ERROR: Method not allowed: %s", traceID, r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[%s] ERROR: Invalid request body: %v", traceID, err)
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("[%s] INFO: Received payment request for user %s, amount %.2f", traceID, req.UserID, req.Amount)

	resp, err := s.ProcessPayment(req)
	if err != nil {
		log.Printf("[%s] ERROR: Payment processing failed: %v", traceID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[%s] ERROR: Failed to encode response: %v", traceID, err)
	}
}

func main() {
	service := NewPaymentService()
	http.HandleFunc("/pay", service.HandlePayment)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
