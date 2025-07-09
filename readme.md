# 🐐 Rinha de Backend 2025 — Golang Edition

Este projeto é uma implementação otimizada para a competição **Rinha de Backend 2025**, com foco total em throughput, fallback inteligente e respeito aos limites de CPU e memória definidos pela organização.

---

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
