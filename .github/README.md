# VeChain Mesh API Implementation

> **⚠️ Work in Progress (WIP)**: This repository is currently under active development. The current implementation is based on the reference implementation from [vechain/rosetta](https://github.com/vechain/rosetta) but is being reviewed and refactored to improve efficiency, code organization, and maintainability.
>
> **📋 TODO List:**
> - Implement call endpoint
> - Codecov integration (GHA)
> - Use mesh-cli to validate endpoints
> - Publish docker image on release
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

For a complete overview of endpoint coverage and implementation status, see [Endpoints Coverage](endpoints.md).

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

## Scripts

The `scripts/` directory contains utility scripts for development and testing.

### sign_payload

A command-line tool for signing VeChain transaction payloads using secp256k1 cryptography.

**Usage:**
```bash
# Build the script
cd scripts
go build -o sign_payload sign_payload.go

# Sign a payload
./sign_payload <private_key_hex> <payload_hex>
```

**Note:** The `payload_hex` should be the `hex_bytes` field from the `construction/payloads` response, which is a 32-byte hash ready for signing.

**Example construction/payloads response:**
```json
{
    "unsigned_transaction": "0xf85281f68800000005e6911c7481b4dad99416277a1ff38678291c41d1820957c78bb5da59ce8227108082bb808864d53d1260b9a69f94f077b491b355e64048ce21e3a6fc4751eeea77fa808609184e72a00080",
    "payloads": [
        {
            "address": "0xf077b491b355E64048cE21E3A6Fc4751eEeA77fa",
            "hex_bytes": "8d351a7849c1b8b22a6cb366dc54a6fa43599bc4a8304901e0f0df5af4e90251",
            "account_identifier": {
                "address": "0xf077b491b355E64048cE21E3A6Fc4751eEeA77fa"
            },
            "signature_type": "ecdsa_recovery"
        }
    ]
}
```

**Sample Output:**
```
85b9599a774600fd1791031f81c957b0dd9570610f34da4719ca266b3b8db92565513f67cea6e2f0daa7611d96935d6697eddccc241fcdf1399de65e7dc0423901
Derived address: 0xf077b491b355E64048cE21E3A6Fc4751eEeA77fa
```

**Features:**
- Outputs signature in format: r + s + v (where v is 0x00 or 0x01)
- Validates derived address for verification
- Handles hex strings with or without "0x" prefix

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

- [Mesh API Specification](https://docs.cdp.coinbase.com/mesh/mesh-api-spec/api-reference) - Official Mesh API documentation
- [VeChain Documentation](https://docs.vechain.org/) - VeChain blockchain documentation

## Contributing

Please follow the [contributing guidelines](CONTRIBUTING.md)

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## Credits

- [Coinbase Mesh SDK](https://github.com/coinbase/rosetta-sdk-go) - Base framework for Mesh API implementation