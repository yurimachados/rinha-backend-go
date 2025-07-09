package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/yurimachados/rinha-backend-go/queue"
	"github.com/yurimachados/rinha-backend-go/types"
)

// PaymentHandler gerencia endpoints de payments
type PaymentHandler struct {
	processor      *queue.PaymentProcessor
	workerPool     *queue.WorkerPool
	requestCounter int64
}

// NewPaymentHandler cria um novo handler otimizado
func NewPaymentHandler(defaultURL, fallbackURL string) *PaymentHandler {
	processor := queue.NewPaymentProcessor(defaultURL, fallbackURL)
	workerPool := queue.NewWorkerPool(processor, 20000) // fila de 20k para alta carga

	handler := &PaymentHandler{
		processor:  processor,
		workerPool: workerPool,
	}

	// Iniciar pool de workers
	workerPool.Start()

	return handler
}

// PostPayments endpoint otimizado para receber payments
func (h *PaymentHandler) PostPayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON eficiente
	var payment types.PaymentRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // performance

	if err := decoder.Decode(&payment); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validação rápida
	if err := payment.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Enfileirar de forma não-bloqueante usando WorkerPool
	if h.workerPool.Submit(&payment) {
		// Sucesso - responder imediatamente
		requestID := atomic.AddInt64(&h.requestCounter, 1)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)

		response := map[string]interface{}{
			"id":      fmt.Sprintf("req_%d_%d", time.Now().Unix(), requestID),
			"status":  "accepted",
			"message": "Payment queued for processing",
		}

		json.NewEncoder(w).Encode(response)

	} else {
		// Fila cheia - rejeitar
		http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
	}
}

// GetPaymentsSummary endpoint para estatísticas
func (h *PaymentHandler) GetPaymentsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	summary := h.processor.GetSummary()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summary)
}

// StartHealthChecker inicia verificação de saúde dos processadores
func (h *PaymentHandler) StartHealthChecker() {
	ctx := context.Background()
	go h.processor.HealthChecker(ctx)
}

// Stop para o handler graciosamente
func (h *PaymentHandler) Stop() {
	h.workerPool.Stop()
}
