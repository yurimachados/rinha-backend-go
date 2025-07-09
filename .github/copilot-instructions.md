# 🤖 GitHub Copilot Instructions – Rinha de Backend 2025 (Golang)

Este projeto é altamente otimizado para a competição **Rinha de Backend 2025**. Todas as sugestões do Copilot devem seguir as diretrizes abaixo para garantir performance máxima, simplicidade e uso mínimo de recursos.

---

## 🧠 Contexto do Desafio

- O objetivo é processar o **máximo de requisições POST /payments** em **1 minuto**.
- Toda requisição deve preferencialmente ser enviada ao endpoint `default`.
- Se o `default` estiver indisponível (timeout, 429, erro), enviar para o `fallback`.
- Requisições devem ser processadas **de forma assíncrona e paralela**, sem bloquear o recebimento.
- O backend precisa rodar dentro das seguintes limitações:
  - **CPU:** 1.5 cores
  - **RAM:** 350MB

---

## 🎯 Instruções para o Copilot

### ✅ Sobre o código

- Sempre usar **Go puro**, com `net/http` nativo.
- **Evite** bibliotecas/frameworks externos (ex: Fiber, Gin, Echo).
- Utilize **channels e goroutines** para fila assíncrona.
- Use **handlers simples** e **estruturas de dados eficientes**.
- Preferir `sync/atomic` ou `chan` para concorrência leve.
- Mantenha o código limpo, direto e com **zero alocação desnecessária**.
- Evite `interface{}` ou tipos genéricos desnecessários — use tipos explícitos.

### ✅ Sobre endpoints

- `POST /payments`: deve validar, enfileirar e responder `202 Accepted`.
- `GET /payments-summary`: retornar dados agregados (default vs fallback).
- `GET /health`: simples `200 OK` para health-check.

### ✅ Sobre arquitetura

- Estrutura mínima de pastas:
  - `handlers/` → endpoints HTTP
  - `queue/` → lógica de fila e workers
  - `circuitbreaker/` → lógica de fallback e health-checks
  - `types/` → definições de structs e payloads

### ✅ Sobre fallback e resiliência

- Toda chamada HTTP para processadores deve:
  - Ter timeout (ex: 300ms)
  - Ter retry controlado (no máximo 1)
  - Atualizar o estado de saúde (circuit breaker leve)
- Processadores devem ser escolhidos com base no status atualizado (memória)

### ✅ Sobre desempenho

- Evitar `log.Println` desnecessários em alta frequência
- Se precisar logar, use logs leves com tempo de execução
- Evite parsing JSON em massa ou structs aninhadas demais

---

## 🚫 Anti-patterns proibidos

- ❌ Uso de frameworks HTTP externos (Gin, Echo)
- ❌ Middleware pesado ou routers complexos
- ❌ Mutex excessivo ou lock global
- ❌ Uso de ORM ou banco de dados
- ❌ Códigos com múltiplos layers de abstração desnecessária

---

## 📌 Observação

Este projeto não é sobre arquitetura corporativa. É sobre **eficiência brutal** e **simplicidade funcional**. Copilot deve focar em entregar código **curto, legível e extremamente performático** para cenários de estresse com milhares de requisições por segundo.

---

> Qualquer sugestão que não esteja alinhada com esses objetivos deve ser descartada ou reescrita com foco em throughput.
>
