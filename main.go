package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yurimachados/rinha-backend-go/handlers"
)

func main() {
	// URLs dos processadores (podem vir de variáveis de ambiente)
	defaultURL := getEnv("DEFAULT_PROCESSOR_URL", "http://processor-default:8080/process")
	fallbackURL := getEnv("FALLBACK_PROCESSOR_URL", "http://processor-fallback:8080/process")
	
	// Criar handler otimizado
	paymentHandler := handlers.NewPaymentHandler(defaultURL, fallbackURL)
	
	// Iniciar health checker
	paymentHandler.StartHealthChecker()
	
	// Configurar rotas otimizadas
	mux := http.NewServeMux()
	
	// Health check simples
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})
	
	// Endpoint principal para payments
	mux.HandleFunc("/payments", paymentHandler.PostPayments)
	
	// Endpoint para estatísticas
	mux.HandleFunc("/payments-summary", paymentHandler.GetPaymentsSummary)
	
	// Servidor HTTP otimizado
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  2 * time.Second,  // timeout agressivo
		WriteTimeout: 2 * time.Second,
		IdleTimeout:  10 * time.Second,
	}
	
	// Graceful shutdown
	// Capturar sinais do sistema
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Iniciar servidor em goroutine
	go func() {
		log.Printf("🚀 Rinha Backend Server rodando na porta 8080")
		log.Printf("📊 Default Processor: %s", defaultURL)
		log.Printf("🔄 Fallback Processor: %s", fallbackURL)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar servidor: %v", err)
		}
	}()
	
	// Aguardar sinal de shutdown
	<-sigChan
	log.Println("🛑 Iniciando graceful shutdown...")
	
	// Timeout para shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Erro durante shutdown: %v", err)
	} else {
		log.Println("✅ Servidor finalizado graciosamente")
	}
}

// getEnv retorna variável de ambiente ou valor padrão
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
