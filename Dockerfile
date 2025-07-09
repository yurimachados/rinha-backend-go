# Etapa 1 - Build
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o rinha .

# Etapa 2 - Runtime
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/rinha .

EXPOSE 8080
CMD ["./rinha"]
