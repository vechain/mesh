# VeChain Mesh API Implementation

A Coinbase Mesh API implementation for the VeChain blockchain, built in Go.

## Introduction

This project implements a server that exposes the Mesh API for interacting with the VeChain blockchain. The Mesh API is an open standard that allows exchanges and other applications to interact with different blockchains in a uniform way.

## Features

- ✅ HTTP server with Mesh API endpoints
- ✅ Support for VeChain networks (mainnet and testnet)
- ✅ Transaction construction endpoints
- ✅ Balance query and network status endpoints
- ✅ Modular architecture with separated concerns

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Docker Setup (Recommended)](#docker-setup-recommended)
  - [Manual Setup](#manual-setup)
- [Configuration](#configuration)
- [Usage](#usage)
- [Docker Management](#docker-management)
- [Troubleshooting](#troubleshooting)
- [Development](#development)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)
- [Credits](#credits)

## Getting Started

### Prerequisites

**For Manual Setup:**
- Go 1.21 or higher
- Git for version control
- Basic understanding of blockchain concepts

**For Docker Setup (Recommended):**
- Docker and Docker Compose installed
- At least 4GB of available RAM
- At least 10GB of available disk space

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd mesh
```

### Docker Setup (Recommended)

For a complete setup with VeChain Thor node and Mesh API:

#### Quick Start
```bash
# Start with default configuration (testnet)
./docker-start.sh

# Start with mainnet
./docker-start.sh --network main

# Start in offline mode
./docker-start.sh --mode offline
```

#### Manual Docker Compose
```bash
# Build and start services
docker-compose up --build -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

#### Individual Docker Builds
```bash
# Build Mesh API only
docker build -f Dockerfile.mesh -t vechain-mesh .

# Build Thor node only
docker build -f Dockerfile.thor -t vechain-thor .
```

### Manual Setup

If you prefer to run without Docker:

1. Install dependencies:
```bash
go mod tidy
```

2. Build the server:
```bash
go build -o mesh-server .
```

3. Run the server (see [Usage](#usage) section for configuration)

### Configuration

The VeChain Mesh API can be configured using environment variables and a JSON configuration file.

#### Environment Variables

**Required Variables:**
- `MODE`: Server mode - `online` or `offline` (default: `online`)
- `NETWORK`: Network type - `main`, `test`, or `custom` (default: `test`)
- `NODEURL`: VeChain node API URL (required for online mode)

**Optional Variables:**
- `PORT`: Server port (default: `8080`)

#### Configuration File

The base configuration is loaded from `config/config.json`. Environment variables override the JSON configuration.

#### Docker Configuration

When using Docker, the Mesh API service is configured via environment variables in `docker-compose.yml`:

```yaml
environment:
  - MODE=online              # online | offline
  - NETWORK=test             # main | test | custom
  - NODEURL=http://thor:8669 # Thor node URL (internal Docker network)
  - PORT=8080                # Mesh API port
```

**Thor Node Configuration:**
```yaml
services:
  thor:
    image: vechain/thor
    command: --network test --api-addr 0.0.0.0:8669 --p2p-port 11235  # testnet
    # or for mainnet:
    # command: --network main --api-addr 0.0.0.0:8669 --p2p-port 11235
```

#### Example Configurations

**Mainnet (Online Mode):**
```bash
export MODE=online
export NETWORK=main
export NODEURL=https://mainnet.veblocks.net
export PORT=8080
```

**Testnet (Online Mode):**
```bash
export MODE=online
export NETWORK=test
export NODEURL=https://testnet.veblocks.net
export PORT=8080
```

**Offline Mode:**
```bash
export MODE=offline
export NETWORK=test
export NODEURL=
export PORT=8080
```

#### Chain Tags

- Mainnet: `0x4a` (74)
- Testnet: `0x27` (39)

#### Docker Services

When using Docker Compose, the following services are available:

**VeChain Thor Node:**
- **Image**: `vechain/thor` (official Docker image)
- **Version**: Latest stable release
- **Ports**: 
  - 8669 (HTTP API - TCP)
  - 11235 (P2P - TCP/UDP)
- **URL**: `http://localhost:8669`
- **Purpose**: VeChain blockchain node for transaction processing

**VeChain Mesh API:**
- **Port**: 8080
- **URL**: `http://localhost:8080`
- **Purpose**: Rosetta API server for blockchain interaction
- **Health Check**: `http://localhost:8080/health`

### Usage

#### Run the server

**Using default configuration:**
```bash
./mesh-server
```

**Using environment variables:**
```bash
MODE=online NETWORK=test NODEURL=https://testnet.veblocks.net ./mesh-server
```

**Or set environment variables and run:**
```bash
export MODE=online
export NETWORK=test
export NODEURL=https://testnet.veblocks.net
./mesh-server
```

#### Available endpoints

- **Health Check**: `GET /health`
- **Network List**: `POST /network/list`
- **Network Status**: `POST /network/status`
- **Account Balance**: `POST /account/balance`
- **Construction Endpoints**: All Mesh API construction endpoints

#### Example requests

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

### Docker Management

#### Service Management

```bash
# View service status
docker-compose ps

# View logs for all services
docker-compose logs -f

# View logs for specific service
docker-compose logs -f mesh
docker-compose logs -f thor

# Restart specific service
docker-compose restart mesh
docker-compose restart thor

# Stop all services
docker-compose down

# Stop and remove volumes (WARNING: This will delete blockchain data)
docker-compose down -v
```

#### Individual Service Control

```bash
# Start only Thor node
docker-compose up thor -d

# Start only Mesh API (requires Thor to be running)
docker-compose up mesh -d
```

#### Data Persistence

The Thor node data is stored in a Docker volume named `thor-data`. This includes:
- Blockchain database
- Node configuration
- Peer information

**Backup data:**
```bash
docker run --rm -v thor-data:/data -v $(pwd):/backup alpine tar czf /backup/thor-data-backup.tar.gz -C /data .
```

**Restore data:**
```bash
docker run --rm -v thor-data:/data -v $(pwd):/backup alpine tar xzf /backup/thor-data-backup.tar.gz -C /data
```

#### Health Checks

```bash
# Check Thor node health
curl http://localhost:11235/blocks/best

# Check Mesh API health
curl http://localhost:8080/health

# Check service health status
docker-compose ps
```

### Troubleshooting

#### Common Issues

1. **Services not starting**
   ```bash
   # Check logs
   docker-compose logs
   
   # Check service status
   docker-compose ps
   ```

2. **Mesh API can't connect to Thor**
   ```bash
   # Check if Thor is healthy
   curl http://localhost:11235/blocks/best
   
   # Check Thor logs
   docker-compose logs thor
   ```

3. **Port conflicts**
   ```bash
   # Check what's using the ports
   lsof -i :8080
   lsof -i :11235
   lsof -i :11235/udp
   lsof -i :11236
   lsof -i :11237
   ```

4. **Out of disk space**
   ```bash
   # Clean up Docker system
   docker system prune -a
   
   # Check volume sizes
   docker system df
   ```

#### Debugging

```bash
# Access Mesh API container shell
docker-compose exec mesh sh

# Access Thor node container shell
docker-compose exec thor sh

# Check container resource usage
docker stats
```

### Development

#### Building Images Manually

```bash
# Build Mesh API image
docker build -f Dockerfile.mesh -t vechain-mesh .

# Build Thor node image
docker build -f Dockerfile.thor -t vechain-thor .

# Run Mesh API container
docker run -p 8080:8080 \
  -e MODE=online \
  -e NETWORK=test \
  -e NODEURL=http://host.docker.internal:11235 \
  vechain-mesh

# Run Thor node container
docker run -p 11235:11235 -p 11235:11235/udp -p 11236:11236 -p 11237:11237 \
  -v thor-data:/app/data \
  vechain-thor
```

### Documentation

- [Mesh API Specification](https://github.com/coinbase/rosetta-sdk-go) - Official Mesh API documentation
- [VeChain Documentation](https://docs.vechain.org/) - VeChain blockchain documentation

### Contributing

Please follow the [contributing guidelines](CONTRIBUTING.md)

#### Development Guidelines

- Follow Go best practices and conventions
- Ensure all tests pass before submitting
- Update documentation for new features
- Use meaningful commit messages
- Keep pull requests focused and atomic

#### Development Setup

**Using Docker (Recommended):**
```bash
# Start development environment
./docker-start.sh

# View logs during development
docker-compose logs -f mesh
```

**Manual Development:**
```bash
# Install dependencies
go mod tidy

# Run tests
go test ./...

# Build and run
go build -o mesh-server .
./mesh-server
```

#### Reporting Issues

- Use the GitHub issue tracker
- Provide detailed reproduction steps
- Include system information and logs
- Check existing issues before creating new ones

### Changelog

#### v0.1.0 (Current)
- Initial implementation of VeChain Mesh API server
- Modular architecture with separated concerns
- All basic Mesh API endpoints implemented
- Graceful shutdown and structured logging
- Comprehensive documentation

### License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

### Credits

#### Core Dependencies
- [Coinbase Rosetta SDK](https://github.com/coinbase/rosetta-sdk-go) - Base framework for Mesh API implementation
- [Gorilla Mux](https://github.com/gorilla/mux) - HTTP router and URL matcher

#### Community
- VeChain developer community for blockchain insights
- Open source contributors and maintainers
- Mesh API specification contributors

#### Special Thanks
- Coinbase for creating the Mesh API standard
- Go community for excellent tooling and libraries
- All contributors and users of this project
