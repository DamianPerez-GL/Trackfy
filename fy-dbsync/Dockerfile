# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Instalar dependencias del sistema
RUN apk add --no-cache git ca-certificates

# Copiar go.mod y go.sum y descargar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fy-dbsync ./cmd/syncer

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Instalar certificados CA para llamadas HTTPS y wget para healthcheck
RUN apk --no-cache add ca-certificates tzdata wget

# Crear directorio para datos
RUN mkdir -p /data

# Copiar binario desde builder
COPY --from=builder /app/fy-dbsync .

# Usuario no root con acceso a /data
RUN adduser -D -g '' appuser && chown -R appuser:appuser /data
USER appuser

# Puerto para health check
EXPOSE 9091

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9091/health || exit 1

# Ejecutar
CMD ["./fy-dbsync"]
