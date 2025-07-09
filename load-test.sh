#!/bin/bash

# 🚀 Script de Teste de Carga - Rinha de Backend 2025
# Simula alta carga de requisições POST /payments

set -e

HOST=${1:-"localhost:8080"}
REQUESTS=${2:-1000}
CONCURRENCY=${3:-50}

echo "🚀 Iniciando teste de carga..."
echo "📊 Host: $HOST"
echo "📈 Requests: $REQUESTS"
echo "⚡ Concurrency: $CONCURRENCY"
echo ""

# Verificar se o servidor está rodando
echo "🔍 Verificando saúde do servidor..."
if ! curl -s "$HOST/health" > /dev/null; then
    echo "❌ Servidor não está respondendo em $HOST"
    exit 1
fi
echo "✅ Servidor OK"
echo ""

# Função para enviar um payment
send_payment() {
    local id=$1
    curl -s -X POST "http://$HOST/payments" \
        -H "Content-Type: application/json" \
        -d "{\"amount\": $((RANDOM % 10000 + 100)), \"description\": \"Load test payment $id\", \"type\": \"credit\"}" \
        > /dev/null
}

# Função para executar batch de requests
run_batch() {
    local start=$1
    local end=$2
    
    for i in $(seq $start $end); do
        send_payment $i &
    done
    wait
}

# Calcular batches
BATCH_SIZE=$CONCURRENCY
BATCHES=$((REQUESTS / BATCH_SIZE))
REMAINDER=$((REQUESTS % BATCH_SIZE))

echo "🚀 Executando $BATCHES batches de $BATCH_SIZE requests..."

# Medir tempo de execução
START_TIME=$(date +%s.%N)

# Executar batches
for batch in $(seq 1 $BATCHES); do
    start_req=$(( (batch - 1) * BATCH_SIZE + 1 ))
    end_req=$(( batch * BATCH_SIZE ))
    
    echo "📦 Batch $batch/$BATCHES (requests $start_req-$end_req)..."
    run_batch $start_req $end_req
done

# Executar requests restantes
if [ $REMAINDER -gt 0 ]; then
    start_req=$(( BATCHES * BATCH_SIZE + 1 ))
    end_req=$REQUESTS
    echo "📦 Batch final (requests $start_req-$end_req)..."
    run_batch $start_req $end_req
fi

END_TIME=$(date +%s.%N)
DURATION=$(echo "$END_TIME - $START_TIME" | bc -l)

echo ""
echo "✅ Teste concluído!"
echo "⏱️  Duração: ${DURATION}s"
echo "📊 Throughput: $(echo "scale=2; $REQUESTS / $DURATION" | bc -l) req/s"
echo ""

# Obter estatísticas finais
echo "📈 Estatísticas finais:"
curl -s "http://$HOST/payments-summary" | jq '.' || curl -s "http://$HOST/payments-summary"

echo ""
echo "🎯 Teste de carga finalizado!"
