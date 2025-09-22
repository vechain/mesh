# VeChain Mesh API Implementation

> **⚠️ Work in Progress (WIP)**: This repository is currently under active development. The current implementation is based on the reference implementation from [vechain/rosetta](https://github.com/vechain/rosetta) but is being reviewed and refactored to improve efficiency, code organization, and maintainability.
>
> **📋 TODO List:**
> - Implement mempool endpoints
> - Implement search endpoints
> - Implement events endpoints
> - Validate that all the middlewares are being applied as expected
> - Refactor Thor client within this repo so we use the existing types in [vechain/thor](https://github.com/vechain/thor) when possible
> - Add GitHub Actions (build, lint, test)
> - Add e2e tests like in vechain/rosetta
> - Add unit tests for coverage
> - Use mesh-cli to validate endpoints
>
> Expect more changes and improvements in upcoming releases.

A Coinbase Mesh API implementation for the VeChain blockchain, built in Go.

## Features

- ✅ HTTP server with Mesh API endpoints
- ✅ Support for VeChain networks (mainnet, testnet, and solo mode)
- ✅ Transaction construction endpoints
- ✅ Balance query and network status endpoints
- ✅ Modular architecture with separated concerns
- ✅ Solo mode for local development and testing

## Quick Start

### Prerequisites

- Docker and Docker Compose installed
- At least 4GB of available RAM
- At least 10GB of available disk space

### Docker Setup (Recommended)

```bash
# Clone the repository
git clone <repository-url>
cd mesh

# Start with default configuration (testnet)
make docker-up-build

# Start in solo mode (local development)
make docker-solo-up

# View logs
make docker-logs

# Stop services
make docker-down

# For solo mode logs
make docker-solo-logs
```

### Manual Setup

```bash
# Install dependencies
go mod tidy

# Build the server
go build -o mesh-server .

# Run with environment variables
MODE=online NETWORK=test ./mesh-server
```

## Configuration

### Environment Variables

- `MODE`: Server mode - `online` or `offline` (default: `online`)
- `NETWORK`: Network type - `main`, `test`, or `solo` (default: `test`)
- `PORT`: Server port (default: `8080`)

### Example Configurations

**Testnet (Online Mode):**
```bash
export MODE=online
export NETWORK=test
export PORT=8080
```

**Mainnet (Online Mode):**
```bash
export MODE=online
export NETWORK=main
export PORT=8080
```

**Solo Mode (Local Development):**
```bash
export MODE=online
export NETWORK=solo
export PORT=8080
```

## Usage

### Available Endpoints

- **Health Check**: `GET /health`
- **Network List**: `POST /network/list`
- **Network Status**: `POST /network/status`
- **Account Balance**: `POST /account/balance`
- **Construction Endpoints**: All Mesh API construction endpoints

### Example Requests

```bash
# Health check
curl http://localhost:8080/health

# Network list
curl -X POST http://localhost:8080/network/list \
  -H "Content-Type: application/json" \
  -d '{}'

# Account balance
curl -X POST http://localhost:8080/account/balance \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {"blockchain": "vechainthor", "network": "test"},
    "account_identifier": {"address": "0x1234567890123456789012345678901234567890"}
  }'
```

## Docker Services

**VeChain Thor Node:**
- **Image**: `vechain/thor` (official Docker image)
- **Ports**: 8669 (HTTP API), 11235 (P2P)
- **URL**: `http://localhost:8669`

**VeChain Mesh API:**
- **Port**: 8080
- **URL**: `http://localhost:8080`
- **Health Check**: `http://localhost:8080/health`

## Development

### Using Makefile (Recommended)

```bash
# Show available commands
make help

# Build and start in testnet mode
make docker-up-build

# Start in solo mode for local development
make docker-solo-up

# View logs
make docker-logs

# Stop services
make docker-down

# Clean up Docker resources
make docker-clean
```

### Manual Development

```bash
# Install dependencies
go mod tidy

# Run tests
go test ./...

# Build and run
go build -o mesh-server .
./mesh-server
```

## Troubleshooting

### Common Issues

1. **Services not starting**
   ```bash
   docker-compose logs
   docker-compose ps
   ```

2. **Port conflicts**
   ```bash
   lsof -i :8080
   lsof -i :8669
   ```

3. **Out of disk space**
   ```bash
   docker system prune -a
   ```

## Documentation

- [Mesh API Specification](https://github.com/coinbase/rosetta-sdk-go) - Official Mesh API documentation
- [VeChain Documentation](https://docs.vechain.org/) - VeChain blockchain documentation

## Contributing

Please follow the [contributing guidelines](CONTRIBUTING.md)

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## Credits

- [Coinbase Mesh SDK](https://github.com/coinbase/rosetta-sdk-go) - Base framework for Mesh API implementation
- [Gorilla Mux](https://github.com/gorilla/mux) - HTTP router and URL matcher
- VeChain developer community for blockchain insights