package handlers

import (
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
	processor     *queue.PaymentProcessor
	paymentQueue  chan *types.PaymentRequest
	requestCounter int64
}

// NewPaymentHandler cria um novo handler otimizado
func NewPaymentHandler(defaultURL, fallbackURL string) *PaymentHandler {
	processor := queue.NewPaymentProcessor(defaultURL, fallbackURL)

	handler := &PaymentHandler{
		processor:    processor,
		paymentQueue: make(chan *types.PaymentRequest, 10000), // buffer grande para alta carga
	}

	// Iniciar workers assíncronos
	for i := 0; i < 50; i++ { // 50 workers paralelos
		go handler.worker()
	}

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

	// Enfileirar de forma não-bloqueante
	select {
	case h.paymentQueue <- &payment:
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

	default:
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

// worker processa payments da fila de forma assíncrona
func (h *PaymentHandler) worker() {
	for payment := range h.paymentQueue {
		// Processar sem bloquear outros workers
		h.processor.ProcessPayment(payment)

		// Micro-sleep para evitar CPU-bound excessivo
		time.Sleep(100 * time.Microsecond)
	}
}

// StartHealthChecker inicia verificação de saúde dos processadores
func (h *PaymentHandler) StartHealthChecker() {
	go h.processor.HealthChecker(nil)
}
