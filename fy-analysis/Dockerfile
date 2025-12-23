# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Instalar dependencias del sistema
RUN apk add --no-cache git ca-certificates

# Copiar go.mod y descargar dependencias
COPY go.mod ./
RUN go mod download && go mod tidy

# Copiar código fuente
COPY . .

# Compilar
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fy-analysis ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Instalar certificados CA para llamadas HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copiar binario desde builder
COPY --from=builder /app/fy-analysis .

# Usuario no root
RUN adduser -D -g '' appuser
USER appuser

# Puerto por defecto
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Ejecutar
CMD ["./fy-analysis"]
