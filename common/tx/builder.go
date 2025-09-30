package tx

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	"github.com/vechain/mesh/common/vip180"
	"github.com/vechain/mesh/config"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/api/transactions"
	"github.com/vechain/thor/v2/thor"
	thorTx "github.com/vechain/thor/v2/tx"
)

type TransactionBuilder struct {
	vip180Encoder *vip180.VIP180Encoder
}

func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{
		vip180Encoder: vip180.NewVIP180Encoder(),
	}
}

// buildMeshTransaction is a helper function that builds a Mesh transaction
func (b *TransactionBuilder) buildMeshTransaction(hash string, operations []*types.Operation, metadata map[string]any) *types.Transaction {
	return &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{Hash: hash},
		Operations:            operations,
		Metadata:              metadata,
	}
}

// BuildMeshTransactionFromAPI builds a Mesh transaction directly from api.JSONEmbeddedTx
func (b *TransactionBuilder) BuildMeshTransactionFromAPI(tx *api.JSONEmbeddedTx, operations []*types.Operation) *types.Transaction {
	return b.buildMeshTransaction(tx.ID.String(), operations, map[string]any{
		"chainTag": tx.ChainTag, "blockRef": "0x" + fmt.Sprintf("%x", tx.BlockRef),
		"expiration": tx.Expiration, "gas": tx.Gas, "gasPriceCoef": tx.GasPriceCoef, "size": tx.Size,
	})
}

// BuildTransactionFromRequest builds a VeChain transaction from a construction request
func (b *TransactionBuilder) BuildTransactionFromRequest(request types.ConstructionPayloadsRequest, config *config.Config) (*thorTx.Transaction, error) {
	// Extract metadata
	metadata := request.Metadata
	blockRef := metadata["blockRef"].(string)
	chainTag := int(metadata["chainTag"].(float64))
	gas := int64(metadata["gas"].(float64))
	transactionType := metadata["transactionType"].(string)
	nonce := metadata["nonce"].(string)

	// Parse nonce to uint64
	nonceStr := strings.TrimPrefix(nonce, "0x")
	nonceValue := new(big.Int)
	nonceValue, ok := nonceValue.SetString(nonceStr, 16)
	if !ok {
		return nil, fmt.Errorf("invalid nonce: %s", nonceStr)
	}

	// Create transaction builder
	builder, err := b.createTransactionBuilder(transactionType, metadata)
	if err != nil {
		return nil, err
	}

	// Set common fields
	builder.ChainTag(byte(chainTag))

	blockRefBytes, err := hex.DecodeString(blockRef[2:])
	if err != nil {
		return nil, fmt.Errorf("invalid blockRef: %w", err)
	}

	builder.BlockRef(thorTx.BlockRef(blockRefBytes))

	// Set expiration from configuration
	expiration := config.GetExpiration()
	builder.Expiration(expiration)

	builder.Gas(uint64(gas))
	builder.Nonce(nonceValue.Uint64())

	// Add clauses from operations
	if err := b.addClausesToBuilder(builder, request.Operations); err != nil {
		return nil, err
	}

	return builder.Build(), nil
}

// createTransactionBuilder creates a transaction builder based on type
func (b *TransactionBuilder) createTransactionBuilder(transactionType string, metadata map[string]any) (*thorTx.Builder, error) {
	if transactionType == "legacy" {
		builder := thorTx.NewBuilder(thorTx.TypeLegacy)
		gasPriceCoefValue := metadata["gasPriceCoef"]
		var gasPriceCoef uint8
		switch v := gasPriceCoefValue.(type) {
		case float64:
			gasPriceCoef = uint8(v)
		case uint8:
			gasPriceCoef = v
		default:
			return nil, fmt.Errorf("invalid gasPriceCoef type: %T, value: %v", v, v)
		}
		builder.GasPriceCoef(gasPriceCoef)
		return builder, nil
	}

	// Dynamic fee transaction
	builder := thorTx.NewBuilder(thorTx.TypeDynamicFee)
	maxFee, ok := new(big.Int).SetString(metadata["maxFeePerGas"].(string), 10)
	if !ok {
		return nil, fmt.Errorf("invalid maxFeePerGas: %s", metadata["maxFeePerGas"])
	}
	maxPriority, ok := new(big.Int).SetString(metadata["maxPriorityFeePerGas"].(string), 10)
	if !ok {
		return nil, fmt.Errorf("invalid maxPriorityFeePerGas: %s", metadata["maxPriorityFeePerGas"])
	}
	builder.MaxFeePerGas(maxFee)
	builder.MaxPriorityFeePerGas(maxPriority)
	return builder, nil
}

// addClausesToBuilder adds clauses to the transaction builder
func (b *TransactionBuilder) addClausesToBuilder(builder *thorTx.Builder, operations []*types.Operation) error {
	// Sort operations by index to maintain order
	sortedOps := make([]*types.Operation, len(operations))
	copy(sortedOps, operations)

	// Simple sort by operation index
	for i := 0; i < len(sortedOps)-1; i++ {
		for j := i + 1; j < len(sortedOps); j++ {
			if sortedOps[i].OperationIdentifier.Index > sortedOps[j].OperationIdentifier.Index {
				sortedOps[i], sortedOps[j] = sortedOps[j], sortedOps[i]
			}
		}
	}

	for _, op := range sortedOps {
		if op.Type == meshcommon.OperationTypeTransfer {
			// Only process Transfer operations with positive values (recipients)
			value := new(big.Int)
			value, ok := value.SetString(op.Amount.Value, 10)
			if !ok {
				return fmt.Errorf("invalid amount: %s", op.Amount.Value)
			}

			// Skip negative values (these represent the sender/origin, not a clause)
			if value.Sign() < 0 {
				continue
			}

			toAddr, err := thor.ParseAddress(op.Account.Address)
			if err != nil {
				return fmt.Errorf("invalid address: %w", err)
			}

			// Check if this is a VIP180 token transfer
			if op.Amount.Currency.Metadata != nil {
				if contractAddr, exists := op.Amount.Currency.Metadata["contractAddress"]; exists {
					if addr, ok := contractAddr.(string); ok {
						// This is a VIP180 token transfer
						contractAddress, err := thor.ParseAddress(addr)
						if err != nil {
							return fmt.Errorf("invalid contract address: %w", err)
						}

						// Encode VIP180 transfer call data
						transferDataHex, err := b.vip180Encoder.EncodeVIP180TransferCallData(op.Account.Address, op.Amount.Value)
						if err != nil {
							return fmt.Errorf("failed to encode VIP180 transfer: %w", err)
						}

						// Convert hex string to bytes
						transferData, err := hex.DecodeString(strings.TrimPrefix(transferDataHex, "0x"))
						if err != nil {
							return fmt.Errorf("failed to decode transfer data: %w", err)
						}

						// Create clause with contract call
						clause := thorTx.NewClause(&contractAddress)
						clause = clause.WithValue(big.NewInt(0)) // VIP180 transfers have 0 value
						clause = clause.WithData(transferData)
						builder.Clause(clause)
						continue
					}
				}
			}

			// Regular VET transfer
			clause := thorTx.NewClause(&toAddr)
			clause = clause.WithValue(value)
			builder.Clause(clause)
		}
		// FeeDelegation operations are handled in the signing process
	}
	return nil
}

// BuildMeshTransactionFromTransactions builds a Mesh transaction directly from transactions.Transaction
func (b *TransactionBuilder) BuildMeshTransactionFromTransaction(tx *transactions.Transaction, operations []*types.Operation) *types.Transaction {
	return b.buildMeshTransaction(tx.ID.String(), operations, map[string]any{
		"chainTag": tx.ChainTag, "blockRef": "0x" + fmt.Sprintf("%x", tx.BlockRef),
		"expiration": tx.Expiration, "gas": tx.Gas, "gasPriceCoef": tx.GasPriceCoef, "size": tx.Size,
	})
}
