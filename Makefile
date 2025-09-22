# VeChain Mesh API Makefile

.PHONY: help build test clean docker-build docker-up docker-down docker-logs docker-clean docker-solo-up docker-solo-down docker-solo-logs

# Default target
help:
	@echo "VeChain Mesh API - Available commands:"
	@echo ""
	@echo "Docker commands:"
	@echo "  docker-build     - Build the Docker image"
	@echo "  docker-up        - Start services in testnet mode"
	@echo "  docker-down      - Stop services"
	@echo "  docker-logs      - View service logs"
	@echo "  docker-clean     - Remove containers and images"
	@echo "  docker-solo-up   - Start services in solo mode"
	@echo "  docker-solo-down - Stop solo mode services"
	@echo "  docker-solo-logs - View solo mode logs"
	@echo ""
	@echo "Development commands:"
	@echo "  build - Build the Go binary"
	@echo "  test  - Run Go tests"
	@echo "  clean - Clean Go build artifacts and cache"
	@echo ""
	@echo "Utilities:"
	@echo "  help - Show this help message"

# Development commands
build:
	go build -o mesh-server .

test:
	go test ./...

clean:
	@echo "Cleaning Go build artifacts and cache..."
	go clean -cache
	go clean -modcache
	go clean -testcache
	rm -f mesh-server
	@echo "Clean completed!"

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-up-build:
	docker-compose up --build -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-clean:
	docker-compose down --rmi all --volumes --remove-orphans

docker-solo-up:
	NETWORK=solo docker-compose up --build -d

docker-solo-down:
	NETWORK=solo docker-compose down

docker-solo-logs:
	NETWORK=solo docker-compose logs -f
