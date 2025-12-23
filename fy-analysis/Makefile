.PHONY: build run test clean docker-build docker-run

# Variables
BINARY_NAME=fy-analysis
MAIN_PATH=./cmd/server

# Build
build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# Run locally
run:
	go run $(MAIN_PATH)/main.go

# Run with hot reload (requiere air: go install github.com/cosmtrek/air@latest)
dev:
	air

# Tests
test:
	go test -v ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Linting
lint:
	golangci-lint run

# Clean
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Docker
docker-build:
	docker build -t fy-analysis:latest .

docker-run:
	docker run -p 8080:8080 --name fy-analysis fy-analysis:latest

docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down

# Mod
mod-tidy:
	go mod tidy

mod-download:
	go mod download
