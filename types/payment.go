package types

import (
	"encoding/json"
	"errors"
)

// PaymentRequest representa o payload de entrada
type PaymentRequest struct {
	Amount      int    `json:"amount"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"`
}

// PaymentResponse representa a resposta do processamento
type PaymentResponse struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	ProcessedBy string `json:"processed_by"`
}

// ProcessorResult representa o resultado do processamento
type ProcessorResult struct {
	Success     bool   `json:"success"`
	ProcessorID string `json:"processor_id"`
	Error       error  `json:"error,omitempty"`
}

// PaymentSummary representa o resumo de payments
type PaymentSummary struct {
	TotalPayments    int64 `json:"total_payments"`
	DefaultSuccess   int64 `json:"default_success"`
	FallbackSuccess  int64 `json:"fallback_success"`
	TotalErrors      int64 `json:"total_errors"`
}

// Validate valida o payload de payment
func (p *PaymentRequest) Validate() error {
	if p.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	if p.Type == "" {
		return errors.New("type is required")
	}
	if len(p.Description) > 255 {
		return errors.New("description too long")
	}
	return nil
}

// ToJSON converte para JSON de forma eficiente
func (p *PaymentRequest) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}
