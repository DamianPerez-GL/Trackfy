# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Instalar dependencias del sistema
RUN apk add --no-cache git ca-certificates

# Copiar go.mod primero
COPY go.mod ./

# Descargar dependencias y generar go.sum
RUN go mod download

# Copiar código fuente
COPY . .

# Regenerar go.sum por si acaso y compilar
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o api-gateway ./cmd/main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Instalar certificados CA
RUN apk --no-cache add ca-certificates tzdata wget

# Copiar binario
COPY --from=builder /app/api-gateway .

# Puerto
EXPOSE 8080

# Ejecutar
CMD ["./api-gateway"]
