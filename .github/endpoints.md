# Endpoints Coverage

This table shows the endpoint coverage of the Mesh API implementation for VeChain.

## Account

| Method | Endpoint           | Implemented | Description               | Mode    |
|--------|--------------------|--------------|---------------------------|---------|
| POST   | /account/balance   | ✅ Yes       | Get account balance       | online  |
| POST   | /account/coins     | ❌ No        | Get account coins         | -       |

## Block

| Method | Endpoint             | Implemented | Description               | Mode    |
|--------|----------------------|--------------|---------------------------|---------|
| POST   | /block               | ✅ Yes       | Get block information     | online  |
| POST   | /block/transaction   | ✅ Yes       | Get block transaction     | online  |

## Call

| Method | Endpoint | Implemented | Description | Mode |
|--------|----------|--------------|-------------|------|
| POST   | /call    | ❌ No        | Call contract method      | - |

## Construction

| Method | Endpoint                   | Implemented | Description                                       | Mode             |
|--------|----------------------------|--------------|---------------------------------------------------|------------------|
| POST   | /construction/combine      | ✅ Yes       | Create network transaction from signatures        | online & offline |
| POST   | /construction/derive       | ✅ Yes       | Derive AccountIdentifier from PublicKey           | online & offline |
| POST   | /construction/hash         | ✅ Yes       | Get hash of signed transaction                    | online & offline |
| POST   | /construction/metadata     | ✅ Yes       | Get metadata for transaction construction         | online           |
| POST   | /construction/parse        | ✅ Yes       | Parse transaction                                 | online & offline |
| POST   | /construction/payloads     | ✅ Yes       | Generate unsigned transaction and signing payloads | online & offline |
| POST   | /construction/preprocess   | ✅ Yes       | Create request for metadata                       | online & offline |
| POST   | /construction/submit       | ✅ Yes       | Submit signed transaction                         | online           |

## Events

| Method | Endpoint        | Implemented | Description                         | Mode    |
|--------|-----------------|--------------|-------------------------------------|---------|
| POST   | /events/blocks  | ❌ No        | Get range of block events | - |

## Mempool

| Method | Endpoint               | Implemented | Description               | Mode |
|--------|------------------------|--------------|---------------------------|------|
| POST   | /mempool               | ✅ Yes       | Get pending transactions  | online |
| POST   | /mempool/transaction   | ✅ Yes       | Get specific mempool transaction | online |

## Network

| Method | Endpoint           | Implemented | Description                   | Mode             |
|--------|--------------------|--------------|-------------------------------|------------------|
| POST   | /network/list      | ✅ Yes       | Get list of available networks | online & offline |
| POST   | /network/options   | ✅ Yes       | Get network options           | online & offline |
| POST   | /network/status    | ✅ Yes       | Get network status            | online           |

## Search

| Method | Endpoint               | Implemented | Description                         | Mode    |
|--------|------------------------|--------------|-------------------------------------|---------|
| POST   | /search/transactions   | ❌ No        | Search transactions       | -       |

## Health Check

| Method | Endpoint | Implemented | Description | Mode |
|--------|----------|--------------|-------------|------|
| GET    | /health  | ✅ Yes       | Check server status | - |

## Coverage Summary

- **Total standard endpoints**: 20
- **Implemented**: 15 (75%)
- **Not implemented**: 5 (25%)

### Pending endpoints to implement:
- `/account/coins`
- `/call`
- `/events/blocks`
- `/search/transactions`

### Notes:
- **Construction** endpoints are fully implemented and support both online and offline modes
- **Network** endpoints are fully implemented
- **Block** and **Mempool** endpoints are implemented for online mode
- **Events** and **Search** endpoints will be added in future versions
