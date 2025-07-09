package queue

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yurimachados/rinha-backend-go/types"
)

// ProcessorStatus representa o status de um processador
type ProcessorStatus struct {
	IsHealthy      int64 // usar atomic para thread-safety
	FailureCount   int64
	LastCheckTime  int64
	ResponseTimeMs int64
}

// PaymentProcessor gerencia o processamento de payments
type PaymentProcessor struct {
	defaultURL   string
	fallbackURL  string
	client       *http.Client
	defaultStatus  *ProcessorStatus
	fallbackStatus *ProcessorStatus
	
	// Estatísticas atômicas
	totalPayments   int64
	defaultSuccess  int64
	fallbackSuccess int64
	totalErrors     int64
}

// NewPaymentProcessor cria um novo processador otimizado
func NewPaymentProcessor(defaultURL, fallbackURL string) *PaymentProcessor {
	return &PaymentProcessor{
		defaultURL:   defaultURL,
		fallbackURL:  fallbackURL,
		client: &http.Client{
			Timeout: 300 * time.Millisecond, // timeout agressivo
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		defaultStatus: &ProcessorStatus{
			IsHealthy: 1, // inicializar como saudável
		},
		fallbackStatus: &ProcessorStatus{
			IsHealthy: 1,
		},
	}
}

// ProcessPayment processa um payment com fallback automático
func (p *PaymentProcessor) ProcessPayment(payment *types.PaymentRequest) *types.ProcessorResult {
	atomic.AddInt64(&p.totalPayments, 1)
	
	// Tentar processador padrão primeiro se estiver saudável
	if atomic.LoadInt64(&p.defaultStatus.IsHealthy) == 1 {
		result := p.sendToProcessor(p.defaultURL, "default", payment, p.defaultStatus)
		if result.Success {
			atomic.AddInt64(&p.defaultSuccess, 1)
			return result
		}
	}
	
	// Fallback para processador secundário
	if atomic.LoadInt64(&p.fallbackStatus.IsHealthy) == 1 {
		result := p.sendToProcessor(p.fallbackURL, "fallback", payment, p.fallbackStatus)
		if result.Success {
			atomic.AddInt64(&p.fallbackSuccess, 1)
			return result
		}
	}
	
	// Ambos falharam
	atomic.AddInt64(&p.totalErrors, 1)
	return &types.ProcessorResult{
		Success:     false,
		ProcessorID: "none",
		Error:       fmt.Errorf("all processors unavailable"),
	}
}

// sendToProcessor envia para um processador específico
func (p *PaymentProcessor) sendToProcessor(url, processorID string, payment *types.PaymentRequest, status *ProcessorStatus) *types.ProcessorResult {
	start := time.Now()
	
	payloadBytes, err := payment.ToJSON()
	if err != nil {
		p.markUnhealthy(status)
		return &types.ProcessorResult{
			Success:     false,
			ProcessorID: processorID,
			Error:       err,
		}
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		p.markUnhealthy(status)
		return &types.ProcessorResult{
			Success:     false,
			ProcessorID: processorID,
			Error:       err,
		}
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		p.markUnhealthy(status)
		return &types.ProcessorResult{
			Success:     false,
			ProcessorID: processorID,
			Error:       err,
		}
	}
	defer resp.Body.Close()
	
	responseTime := time.Since(start).Milliseconds()
	atomic.StoreInt64(&status.ResponseTimeMs, responseTime)
	
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		p.markHealthy(status)
		return &types.ProcessorResult{
			Success:     true,
			ProcessorID: processorID,
		}
	}
	
	// Status de erro ou timeout
	if resp.StatusCode == 429 || resp.StatusCode >= 500 {
		p.markUnhealthy(status)
	}
	
	return &types.ProcessorResult{
		Success:     false,
		ProcessorID: processorID,
		Error:       fmt.Errorf("HTTP %d", resp.StatusCode),
	}
}

// markHealthy marca processador como saudável
func (p *PaymentProcessor) markHealthy(status *ProcessorStatus) {
	atomic.StoreInt64(&status.IsHealthy, 1)
	atomic.StoreInt64(&status.FailureCount, 0)
	atomic.StoreInt64(&status.LastCheckTime, time.Now().Unix())
}

// markUnhealthy marca processador como não saudável
func (p *PaymentProcessor) markUnhealthy(status *ProcessorStatus) {
	failures := atomic.AddInt64(&status.FailureCount, 1)
	if failures >= 3 { // circuit breaker após 3 falhas
		atomic.StoreInt64(&status.IsHealthy, 0)
	}
	atomic.StoreInt64(&status.LastCheckTime, time.Now().Unix())
}

// GetSummary retorna estatísticas de processamento
func (p *PaymentProcessor) GetSummary() *types.PaymentSummary {
	return &types.PaymentSummary{
		TotalPayments:   atomic.LoadInt64(&p.totalPayments),
		DefaultSuccess:  atomic.LoadInt64(&p.defaultSuccess),
		FallbackSuccess: atomic.LoadInt64(&p.fallbackSuccess),
		TotalErrors:     atomic.LoadInt64(&p.totalErrors),
	}
}

// HealthChecker executa verificações periódicas de saúde
func (p *PaymentProcessor) HealthChecker(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.checkProcessorHealth()
		}
	}
}

// checkProcessorHealth verifica saúde dos processadores
func (p *PaymentProcessor) checkProcessorHealth() {
	var wg sync.WaitGroup
	
	// Verificar default
	wg.Add(1)
	go func() {
		defer wg.Done()
		if atomic.LoadInt64(&p.defaultStatus.IsHealthy) == 0 {
			if p.pingProcessor(p.defaultURL) {
				p.markHealthy(p.defaultStatus)
			}
		}
	}()
	
	// Verificar fallback
	wg.Add(1)
	go func() {
		defer wg.Done()
		if atomic.LoadInt64(&p.fallbackStatus.IsHealthy) == 0 {
			if p.pingProcessor(p.fallbackURL) {
				p.markHealthy(p.fallbackStatus)
			}
		}
	}()
	
	wg.Wait()
}

// pingProcessor faz um ping simples no processador
func (p *PaymentProcessor) pingProcessor(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", url+"/health", nil)
	if err != nil {
		return false
	}
	
	resp, err := p.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == 200
}
