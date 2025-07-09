package queue

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/yurimachados/rinha-backend-go/types"
)

// WorkerPool gerencia um pool de workers para processamento assíncrono
type WorkerPool struct {
	processor    *PaymentProcessor
	workQueue    chan *types.PaymentRequest
	workerCount  int
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewWorkerPool cria um novo pool de workers otimizado
func NewWorkerPool(processor *PaymentProcessor, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Número de workers baseado no número de CPUs
	workerCount := runtime.NumCPU() * 4 // 4x o número de CPUs para I/O intensivo
	if workerCount > 100 {
		workerCount = 100 // limite máximo
	}
	
	return &WorkerPool{
		processor:   processor,
		workQueue:   make(chan *types.PaymentRequest, queueSize),
		workerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start inicia os workers do pool
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// Stop para os workers do pool graciosamente
func (wp *WorkerPool) Stop() {
	close(wp.workQueue)
	wp.cancel()
	wp.wg.Wait()
}

// Submit envia um payment para processamento
func (wp *WorkerPool) Submit(payment *types.PaymentRequest) bool {
	select {
	case wp.workQueue <- payment:
		return true
	default:
		return false // fila cheia
	}
}

// worker processa payments da fila
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	
	// Batch processing para eficiência
	batch := make([]*types.PaymentRequest, 0, 10)
	ticker := time.NewTicker(50 * time.Millisecond) // flush batch a cada 50ms
	defer ticker.Stop()
	
	for {
		select {
		case <-wp.ctx.Done():
			// Processar batch restante antes de sair
			wp.processBatch(batch)
			return
			
		case payment, ok := <-wp.workQueue:
			if !ok {
				// Canal fechado, processar batch restante
				wp.processBatch(batch)
				return
			}
			
			batch = append(batch, payment)
			
			// Processar batch quando estiver cheio
			if len(batch) >= 10 {
				wp.processBatch(batch)
				batch = batch[:0] // reset slice
			}
			
		case <-ticker.C:
			// Flush batch periodicamente
			if len(batch) > 0 {
				wp.processBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// processBatch processa um lote de payments de forma paralela
func (wp *WorkerPool) processBatch(batch []*types.PaymentRequest) {
	if len(batch) == 0 {
		return
	}
	
	// Processar até 5 payments em paralelo por batch
	semaphore := make(chan struct{}, 5)
	var batchWg sync.WaitGroup
	
	for _, payment := range batch {
		semaphore <- struct{}{}
		batchWg.Add(1)
		
		go func(p *types.PaymentRequest) {
			defer func() {
				<-semaphore
				batchWg.Done()
			}()
			
			wp.processor.ProcessPayment(p)
		}(payment)
	}
	
	batchWg.Wait()
}

// GetQueueSize retorna o tamanho atual da fila
func (wp *WorkerPool) GetQueueSize() int {
	return len(wp.workQueue)
}
