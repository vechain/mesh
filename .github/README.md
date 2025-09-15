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

- [Project](#project)
  - [Introduction](#introduction)
  - [Table of Contents](#table-of-contents)
  - [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Installation](#installation)
    - [Configuration](#configuration)
    - [Usage](#usage)
    - [Documentation](#documentation)
    - [Contributing](#contributing)
    - [Roadmap](#roadmap)
    - [Changelog](#changelog)
    - [License](#license)
    - [Credits](#credits)

## Getting Started

### Prerequisites

- Go 1.24.5 or higher
- Git for version control
- Basic understanding of blockchain concepts

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd mesh
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the server:
```bash
go build -o mesh-server .
```

### Configuration

The VeChain Mesh API can be configured using environment variables and a JSON configuration file.

#### Environment Variables

**Required Variables:**
- `MODE`: Server mode - `online` or `offline` (default: `online`)
- `NETWORK`: Network type - `main`, `test`, or `custom` (default: `test`)
- `NODEURL`: VeChain node API URL (required for online mode)

**Optional Variables:**
- `PORT`: Server port (default: `8000`)

#### Configuration File

The base configuration is loaded from `config/config.json`. Environment variables override the JSON configuration.

#### Example Configurations

**Mainnet (Online Mode):**
```bash
export MODE=online
export NETWORK=main
export NODEURL=https://mainnet.veblocks.net
export PORT=8000
```

**Testnet (Online Mode):**
```bash
export MODE=online
export NETWORK=test
export NODEURL=https://testnet.veblocks.net
export PORT=8000
```

**Offline Mode:**
```bash
export MODE=offline
export NETWORK=test
export NODEURL=
export PORT=8000
```

#### Chain Tags

- Mainnet: `0x4a` (74)
- Testnet: `0x27` (39)

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
curl http://localhost:8000/health

# Network list
curl -X POST http://localhost:8000/network/list \
  -H "Content-Type: application/json" \
  -d '{}'

# Account balance
curl -X POST http://localhost:8000/account/balance \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {"blockchain": "vechainthor", "network": "test"},
    "account_identifier": {"address": "0x1234567890123456789012345678901234567890"}
  }'
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
