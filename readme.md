# � Rinha de Backend 2025 - Fase 2 (Go)

Backend otimizado para alta performance na **Rinha de Backend 2025**, construído em Go puro com foco em throughput máximo e uso mínimo de recursos.

## 🎯 Objetivo

Processar o **máximo de requisições POST /payments** em **1 minuto** com:
- **CPU:** 1.5 cores
- **RAM:** 350MB
- Processamento assíncrono com fallback automático
- Circuit breaker para resiliência

## 🏗️ Arquitetura

### Componentes Principais

```
├── handlers/          # HTTP endpoints otimizados
│   └── payments.go    # Handler de payments com fila assíncrona
├── queue/             # Sistema de filas e processamento
│   ├── processor.go   # Circuit breaker e fallback automático
│   └── worker.go      # Pool de workers com batch processing
├── types/             # Estruturas de dados eficientes
│   └── payment.go     # Tipos e validações otimizadas
└── main.go           # Servidor HTTP com graceful shutdown
```

### Fluxo de Processamento

1. **Recepção** → POST /payments (resposta imediata 202 Accepted)
2. **Enfileiramento** → WorkerPool com buffer de 20k
3. **Processamento** → Batch processing com até 50 workers paralelos
4. **Fallback** → Circuit breaker automático entre processadores
5. **Monitoramento** → Health checks e estatísticas em tempo real

## 🚦 Endpoints

### `POST /payments`
```bash
curl -X POST http://localhost:8080/payments \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1000,
    "description": "Test payment",
    "type": "credit"
  }'
```

**Resposta:**
```json
{
  "id": "req_1752034000_1",
  "status": "accepted",
  "message": "Payment queued for processing"
}
```

### `GET /payments-summary`
```bash
curl http://localhost:8080/payments-summary
```

**Resposta:**
```json
{
  "total_payments": 1000,
  "default_success": 850,
  "fallback_success": 100,
  "total_errors": 50
}
```

### `GET /health`
```bash
curl http://localhost:8080/health
```

## ⚡ Otimizações de Performance

### 1. **Processamento Assíncrono**
- Resposta imediata (202 Accepted)
- WorkerPool com 4x CPUs workers
- Batch processing para reduzir overhead

### 2. **Circuit Breaker Inteligente**
- Fallback automático em 300ms
- Health checks a cada 10s
- Recuperação automática de processadores

### 3. **HTTP Otimizado**
- Timeouts agressivos (2s read/write)
- Connection pooling
- JSON streaming sem buffering

### 4. **Concorrência Segura**
- `sync/atomic` para estatísticas
- Channels não-bloqueantes
- Graceful shutdown

## 🐳 Docker

### Build & Run
```bash
# Build e iniciar
docker compose up --build

# Apenas iniciar (se já buildado)
docker compose up
```

### Configuração
```yaml
# docker-compose.yml
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DEFAULT_PROCESSOR_URL=http://processor-default:8080/process
      - FALLBACK_PROCESSOR_URL=http://processor-fallback:8080/process
    deploy:
      resources:
        limits:
          cpus: "1.5"
          memory: 350M
```

## 🛠️ Desenvolvimento Local

### Pré-requisitos
- Go 1.22+
- Docker & Docker Compose

### Executar Localmente
```bash
# Build
go build -o rinha .

# Run com processadores de teste
DEFAULT_PROCESSOR_URL=http://httpbin.org/status/200 \
FALLBACK_PROCESSOR_URL=http://httpbin.org/status/200 \
./rinha
```

### Teste de Carga
```bash
# Teste simples
for i in {1..100}; do
  curl -X POST http://localhost:8080/payments \
    -H "Content-Type: application/json" \
    -d '{"amount":'$i',"type":"credit"}' &
done

# Verificar estatísticas
curl http://localhost:8080/payments-summary
```

## 📊 Monitoramento

### Métricas Disponíveis
- **total_payments**: Total de payments recebidos
- **default_success**: Sucessos no processador padrão
- **fallback_success**: Sucessos no processador fallback
- **total_errors**: Erros de processamento

### Logs Estruturados
```
2025/07/09 01:06:05 🚀 Rinha Backend Server rodando na porta 8080
2025/07/09 01:06:05 📊 Default Processor: http://processor-default:8080/process
2025/07/09 01:06:05 🔄 Fallback Processor: http://processor-fallback:8080/process
```

## 🎯 Estratégia para a Rinha

### Configuração Recomendada
1. **Workers**: 4x número de CPUs (aprox. 24 workers)
2. **Buffer**: 20.000 payments em memória
3. **Batch Size**: 10 payments por lote
4. **Timeout**: 250ms por processador
5. **Circuit Breaker**: 3 falhas consecutivas

### Expectativa de Performance
- **Throughput**: 5.000+ req/s
- **Latência**: <5ms (resposta HTTP)
- **Memória**: <200MB em uso constante
- **CPU**: 80-90% utilização otimizada

## 🔧 Variáveis de Ambiente

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `DEFAULT_PROCESSOR_URL` | `http://processor-default:8080/process` | URL do processador padrão |
| `FALLBACK_PROCESSOR_URL` | `http://processor-fallback:8080/process` | URL do processador fallback |

## 📝 Notas Técnicas

- **Zero Dependências Externas**: Apenas Go stdlib + net/http
- **Memory Pool**: Reutilização de objetos para reduzir GC
- **Lock-Free**: Operações atômicas para alta concorrência
- **Graceful Shutdown**: Finalização segura sem perda de dados

---

**Desenvolvido para dominar a Rinha de Backend 2025** 🏆

## ⚙️ Tecnologias Utilizadas

- **Go 1.22** — `net/http`, goroutines, channels
- **Docker + Docker Compose** — Multi-stage build
- **Nginx (futuro)** — Proxy reverso leve
- **k6** — Testes de carga

---

## 📋 Requisitos da Rinha

- Processar o **maior número de pagamentos possível** em 1 minuto
- Utilizar preferencialmente o endpoint `default` para POSTs
- Utilizar `fallback` apenas se necessário (ex: indisponibilidade ou erro)
- Limite de recursos:
  - **CPU**: 1.5
  - **Memória**: 350MB
- Deve expor:
  - `POST /payments`
  - `GET /payments-summary`
  - `GET /health`

---

## 🧭 Roadmap do Projeto

### ✅ Etapa 1 – Setup básico

- [X]  Inicialização com Go 1.22
- [X]  Endpoint `/health`
- [X]  Dockerfile com multi-stage build
- [X]  docker-compose com limites de CPU e RAM

### 🚧 Etapa 2 – Fila e processamento assíncrono

- [ ]  Channel buffered com tipo `Payment`
- [ ]  Workers em goroutines processando pagamentos
- [ ]  Enfileiramento no `POST /payments`

### 🚧 Etapa 3 – Circuit Breaker

- [ ]  Health-check assíncrono dos processadores
- [ ]  Alternância entre default/fallback com base no estado
- [ ]  Tolerância a falhas (429, timeouts, 5xx)

### 🚧 Etapa 4 – Métricas e summary

- [ ]  `GET /payments-summary` com dados agregados
- [ ]  Monitoramento em tempo real da utilização dos processadores

### 🚧 Etapa 5 – Otimização fina

- [ ]  Benchmark com `k6`
- [ ]  Tuning de goroutines, timeouts, fila
- [ ]  Redução de alocações e footprint de memória

---

## 🏁 Como Rodar

```bash
docker-compose up --build
curl http://localhost:8080/health
```


## Estrutura

``````go

rinha-backend-go/
├── main.go
├── handlers/
│   └── payments.go
├── queue/
│   └── processor.go
├── types/
│   └── payment.go
├── Dockerfile
├── docker-compose.yml
└── README.md
``````



## 🧠 Estratégia de Vitória

* Fila assíncrona com `chan Payment` para evitar bloqueios
* Circuit breaker com estado leve (cache em memória)
* Requisições imediatas ao default se saudável, fallback só quando necessário
* Resposta rápida (`202 Accepted`) e processamento em segundo plano
