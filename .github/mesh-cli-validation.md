# Mesh CLI Validation

This document explains how to use Coinbase's `mesh-cli` tool to validate the VeChain Mesh API implementation.

## Overview

`mesh-cli` is Coinbase's official tool for validating Mesh/Rosetta API implementations. It verifies that endpoints comply with the specification and that data is consistent across different API calls.

## Quick Start

### Build mesh-cli (first time only)

```bash
make mesh-cli-build
```

### Validate Data API

```bash
# 1. Start server in solo mode (for quick testing)
make docker-solo-up

# 2. Run validation for solo network
make mesh-cli-check-data-solo

# Or for testnet (requires synced node)
make docker-up
make mesh-cli-check-data-test
```

### Validate Construction API

```bash
# 1. Start server in solo mode
make docker-solo-up

# 2. Run validation for solo network
make mesh-cli-check-construction-solo
```

## Available Commands

### General Commands

| Command | Description |
|---------|-------------|
| `make mesh-cli-build` | Build mesh-cli Docker image |
| `make mesh-cli-check-data ENV=<env>` | Validate Data API for specific environment |
| `make mesh-cli-check-construction ENV=<env>` | Validate Construction API for specific environment |
| `make mesh-cli-view-data` | View validation results in JSON format |

### Convenience Commands (Environment-Specific)

| Command | Description |
|---------|-------------|
| `make mesh-cli-check-data-solo` | Validate Data API on solo network |
| `make mesh-cli-check-construction-solo` | Validate Construction API on solo network |
| `make mesh-cli-check-data-test` | Validate Data API on testnet |
| `make mesh-cli-check-construction-test` | Validate Construction API on testnet |
| `make mesh-cli-check-data-main` | Validate Data API on mainnet |
| `make mesh-cli-check-construction-main` | Validate Construction API on mainnet |

## Validation Types

### 1. Data API Validation

Validates the following endpoints:
- **Network**: `/network/list`, `/network/status`, `/network/options`
- **Account**: `/account/balance`
- **Block**: `/block`, `/block/transaction`
- **Mempool**: `/mempool`, `/mempool/transaction`
- **Events**: `/events/blocks`
- **Search**: `/search/transactions`
- **Call**: `/call`

**Requirements:**
- Server running (`make docker-up` for testnet, `make docker-solo-up` for solo)
- Access to VeChain node (testnet, mainnet, or solo)

**Configuration files:**
- Solo network: `config/solo/mesh-cli-data.json`
- Testnet: `config/test/mesh-cli-data.json`
- Mainnet: `config/main/mesh-cli-data.json`

**Network values:**
- `test` - VeChain Testnet
- `main` - VeChain Mainnet  
- `solo` - Local solo node for development

### 2. Construction API Validation

Validates the following endpoints:
- `/construction/derive` - Derive address from public key
- `/construction/preprocess` - Prepare metadata request
- `/construction/metadata` - Get transaction metadata
- `/construction/payloads` - Generate unsigned transaction
- `/construction/parse` - Parse transaction
- `/construction/combine` - Combine transaction with signatures
- `/construction/hash` - Get transaction hash
- `/construction/submit` - Submit transaction to network

**Requirements:**
- Server running in solo mode (`make docker-solo-up`)
- Local Thor node with pre-funded accounts

**Configuration files:**
- Solo network: `config/solo/mesh-cli-construction.json`
- Testnet: `config/test/mesh-cli-construction.json`
- Mainnet: `config/main/mesh-cli-construction.json` (to be created)

## Configuration

Configuration files are organized by environment in separate directories:
- `config/solo/` - Local development configuration
- `config/test/` - Testnet configuration
- `config/main/` - Mainnet configuration

### Data API Configuration (e.g., `config/solo/mesh-cli-data.json`)

```json
{
  "network": {
    "blockchain": "vechainthor",
    "network": "test"
  },
  "online_url": "http://mesh:8080",
  "data_directory": "/data",
  "http_timeout": 300,
  "tip_delay": 5,
  "data": {
    "initial_balance_fetch_disabled": false,
    "historical_balance_disabled": true,
    "reconciliation_disabled": true,
    "inactive_discrepancy_search_disabled": true,
    "balance_tracking_disabled": true,
    "coin_tracking_disabled": true,
    "end_conditions": {
      "tip": true
    },
    "results_output_file": "/data/data_results.json"
  }
}
```

**Key parameters:**
- `network.blockchain`: Must be `"vechainthor"` (lowercase)
- `network.network`: Network identifier - `"test"` (testnet), `"main"` (mainnet), or `"solo"` (local development)
- `online_url`: URL to the Mesh API server (use `http://mesh:8080` when running in Docker)
- `http_timeout`: Request timeout in seconds (default: 300)
- `tip_delay`: Time in seconds to wait before considering a block as tip (default: 5)
- `data.reconciliation_disabled`: Set to `true` for faster validation (skips balance reconciliation)
- `data.end_conditions.tip`: Set to `true` to validate up to chain tip

### Construction API Configuration (e.g., `config/solo/mesh-cli-construction.json`)

```json
{
  "network": {
    "blockchain": "vechainthor",
    "network": "solo"
  },
  "online_url": "http://mesh:8080",
  "data_directory": "/data",
  "http_timeout": 300,
  "construction": {
    "offline_url": "http://mesh:8080",
    "max_offline_connections": 120,
    "stale_depth": 10,
    "broadcast_limit": 5,
    "workflows": [
      {
        "name": "transfer",
        "concurrency": 1,
        "scenarios": [
          {
            "name": "simple_transfer",
            "actions": [
              {
                "type": "set_variable",
                "input": "{\"symbol\":\"VET\", \"decimals\":18}",
                "output_path": "currency"
              },
              {
                "type": "set_variable",
                "input": "1000000000000000000",
                "output_path": "transfer_amount"
              }
            ]
          }
        ]
      }
    ],
    "end_conditions": {
      "create_account": 10,
      "transfer": 10
    },
    "results_output_file": "/data/construction_results.json"
  }
}
```

**Key parameters:**
- `network.blockchain`: Must be `"vechainthor"` (lowercase)
- `network.network`: Network identifier - `"test"`, `"main"`, or `"solo"` (recommended: `"solo"` for faster testing)
- `construction.offline_url`: URL to the Mesh API server for offline operations
- `construction.end_conditions.transfer`: Number of transfer transactions to test
- `construction.end_conditions.create_account`: Number of account creations to test
- `construction.workflows`: Defines test scenarios for transaction construction (see full config file for complete example)

## Understanding Results

### Success ✅

When validation passes, you'll see:

```
✅ All checks passed!
✅ Data API validation successful
✅ Construction API validation successful
```

The validation results are saved to:
- Data API: `mesh-cli-data/data_results.json`
- Construction API: `mesh-cli-data/construction_results.json`

### Common Errors ❌

**Balance mismatch**
```
ERROR: Balance does not match expected value
```
- **Cause**: Balance returned by `/account/balance` doesn't match transaction history
- **Fix**: Check balance calculation logic in `AccountService`

**Invalid signature**
```
ERROR: Invalid signature in combined transaction
```
- **Cause**: Signature generation or combination is incorrect
- **Fix**: Verify `/construction/payloads` and `/construction/combine` implementations

**Endpoint timeout**
```
ERROR: Request timeout after 300s
```
- **Cause**: Endpoint taking too long to respond
- **Fix**: 
  - Increase `http_timeout` in configuration
  - Optimize endpoint performance
  - Reduce validation range (lower `end_index`)

**Invalid response format**
```
ERROR: Response does not match expected schema
```
- **Cause**: Response doesn't comply with Mesh API specification
- **Fix**: Review endpoint response format against Mesh API spec

**Block not found**
```
ERROR: Block at index X not found
```
- **Cause**: Node hasn't synced to requested block yet
- **Fix**:
  - Wait for node to sync
  - Reduce `end_index` to a lower block
  - Verify node connectivity

## Troubleshooting

### Server Not Responding

```bash
# Check if server is running
docker ps | grep vechain-mesh

# Check server health
curl http://localhost:8080/health

# View server logs
make docker-logs
```

### Connection Refused

If mesh-cli can't connect to the server:

1. **Check both containers are in same network:**
   ```bash
   docker network inspect mesh_vechain-network
   ```

2. **Verify mesh container is accessible:**
   ```bash
   docker run --rm --network mesh_vechain-network \
     vechain-mesh-cli:latest \
     sh -c "curl -f http://mesh:8080/health"
   ```

3. **If running mesh-cli outside Docker:**
   - Change `online_url` to `http://localhost:8080`
   - Ensure server is listening on `0.0.0.0` not just `127.0.0.1`

### Timeout Issues

If validation times out frequently:

1. **Limit validation to fewer blocks:**
   ```json
   {
     "data": {
       "end_conditions": {
         "index": 10
       }
     }
   }
   ```
   
   Or use `tip_delay` to stop sooner:
   ```json
   {
     "tip_delay": 1,
     "data": {
       "end_conditions": {
         "tip": true
       }
     }
   }
   ```

2. **Increase timeout:**
   ```json
   {
     "http_timeout": 600,  // 10 minutes
     "retry_elapsed_time": 1200
   }
   ```

3. **Check system resources:**
   ```bash
   docker stats vechain-mesh
   ```

### Solo Mode Issues

**Error: Insufficient balance**
```
ERROR: Account has insufficient balance for transfer
```
- **Cause**: Test account ran out of VET
- **Fix**: Solo mode should have pre-funded accounts. Check Thor solo configuration.

**Error: Transaction reverted**
```
ERROR: Transaction execution reverted
```
- **Cause**: Invalid transaction parameters or contract logic error
- **Fix**: Check transaction construction logic and gas parameters

## Advanced Usage

### Manual Docker Execution

Run mesh-cli commands directly:

```bash
# Data API validation (solo network)
docker run --rm \
  --network mesh_vechain-network \
  -v $(pwd)/config/solo:/config:ro \
  -v $(pwd)/mesh-cli-data:/data \
  vechain-mesh-cli:latest \
  check:data --configuration-file /config/mesh-cli-data.json

# Construction API validation (solo network)
docker run --rm \
  --network mesh_vechain-network \
  -v $(pwd)/config/solo:/config:ro \
  -v $(pwd)/mesh-cli-data:/data \
  vechain-mesh-cli:latest \
  check:construction --configuration-file /config/mesh-cli-construction.json

# View help
docker run --rm vechain-mesh-cli:latest --help

# Check version
docker run --rm vechain-mesh-cli:latest version
```

### Custom Configuration

To create a configuration for a new environment:

1. Create a new directory under `config/` (e.g., `config/customnet/`)
2. Copy and modify the configuration files from `config/solo/`
3. Update the `network.network` field to match your environment
4. Run with the ENV parameter:
   ```bash
   make mesh-cli-check-data ENV=customnet
   ```

Or run directly with Docker:
   ```bash
   docker run --rm \
     --network mesh_vechain-network \
     -v $(pwd)/config/customnet:/config:ro \
     -v $(pwd)/mesh-cli-data:/data \
     vechain-mesh-cli:latest \
     check:data --configuration-file /config/mesh-cli-data.json
   ```

### Viewing Detailed Logs

View results file directly:

```bash
# View results using jq
cat mesh-cli-data/data_results.json | jq '.'

# Or use the make command
make mesh-cli-view-data
```



