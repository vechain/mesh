# Mesh CLI Validation

This document explains how to use Coinbase's `mesh-cli` tool to validate the VeChain Mesh API implementation.

## Overview

`mesh-cli` is Coinbase's official tool for validating Mesh/Rosetta API implementations. It verifies that endpoints comply with the specification and that data is consistent across different API calls.

## Quick Start

### Build mesh-cli (first time only)

```bash
make mesh-cli-build
```

### Validate APIs

```bash
# Data API validation (automatically starts/stops services)
make mesh-cli-check-data-solo

# Construction API validation (automatically starts/stops services)
make mesh-cli-check-construction-solo
```

**That's it!** The commands handle everything automatically: starting services, waiting for readiness, running validation, and cleaning up.

## Available Commands

| Command | Description |
|---------|-------------|
| `make mesh-cli-build` | Build mesh-cli Docker image (first time only) |
| `make mesh-cli-check-data-solo` | Validate Data API on solo network |
| `make mesh-cli-check-construction-solo` | Validate Construction API on solo network |
| `make mesh-cli-check-data ENV=<env>` | Validate Data API for specific environment |
| `make mesh-cli-check-construction ENV=<env>` | Validate Construction API for specific environment |

**Supported environments:** `solo`, `test`, `main`

## What Gets Validated

### Data API
Validates network, account, block, mempool, events, search, and call endpoints to ensure they comply with the Mesh API specification.

### Construction API  
Validates the complete transaction construction flow: derive → preprocess → metadata → payloads → parse → combine → hash → submit.

## Configuration

Configuration files are located in environment-specific directories:
- `config/solo/` - Local development (recommended for testing)
- `config/test/` - Testnet configuration  
- `config/main/` - Mainnet configuration

The configuration files are already set up and ready to use. No manual configuration needed for basic validation.

## Understanding Results

### Success ✅
When validation passes, you'll see:
```
✅ Data API validation passed!
✅ Construction API validation passed!
```

### Failure ❌
If validation fails, check the error message and refer to the troubleshooting section below.

## Troubleshooting

### Common Issues

**Server not responding**
```bash
# Check if server is running
docker ps | grep vechain-mesh

# Check server health  
curl http://localhost:8080/health
```

**Connection refused**
- The commands automatically handle Docker networking
- If issues persist, try rebuilding: `make mesh-cli-build`

**Timeout errors**
- Construction API validation can take several minutes
- This is normal behavior

**Insufficient balance (Construction API)**
- Solo mode has pre-funded accounts
- If this error occurs, check Thor solo configuration

## Advanced Usage

### Custom Environments

To validate against a different environment:

```bash
# For testnet
make mesh-cli-check-data ENV=test
make mesh-cli-check-construction ENV=test

# For mainnet  
make mesh-cli-check-data ENV=main
```

### Manual Docker Execution

If you need to run mesh-cli directly:

```bash
# Data API validation
docker run --rm \
  --network mesh_vechain-network \
  -v $(pwd)/config/solo:/config:ro \
  -v $(pwd)/mesh-cli-data:/data \
  vechain-mesh-cli:latest \
  check:data --configuration-file /config/mesh-cli-data.json
```

### Results

Validation results are saved to:
- Data API: `mesh-cli-data/data_results.json`
- Construction API: `mesh-cli-data/construction_results.json`
