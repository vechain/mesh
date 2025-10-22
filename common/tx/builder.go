package tx

import (
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/math"
	meshcommon "github.com/vechain/mesh/common"
	meshcrypto "github.com/vechain/mesh/common/crypto"
	"github.com/vechain/mesh/common/vip180"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/api/transactions"
	"github.com/vechain/thor/v2/thor"
	thorTx "github.com/vechain/thor/v2/tx"
)

type TransactionBuilder struct {
	vip180Encoder *vip180.VIP180Encoder
	bytesHandler  *meshcrypto.BytesHandler
}

func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{
		vip180Encoder: vip180.NewVIP180Encoder(),
		bytesHandler:  meshcrypto.NewBytesHandler(),
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
	metadata := b.buildTransactionMetadata(tx.ChainTag, tx.BlockRef, tx.Expiration, tx.Gas, tx.Size, tx.GasPriceCoef, tx.MaxFeePerGas, tx.MaxPriorityFeePerGas)
	return b.buildMeshTransaction(tx.ID.String(), operations, metadata)
}

// BuildMeshTransactionFromTransactions builds a Mesh transaction directly from transactions.Transaction
func (b *TransactionBuilder) BuildMeshTransactionFromTransaction(tx *transactions.Transaction, operations []*types.Operation) *types.Transaction {
	metadata := b.buildTransactionMetadata(tx.ChainTag, tx.BlockRef, tx.Expiration, tx.Gas, tx.Size, tx.GasPriceCoef, tx.MaxFeePerGas, tx.MaxPriorityFeePerGas)
	return b.buildMeshTransaction(tx.ID.String(), operations, metadata)
}

// buildTransactionMetadata builds metadata for a transaction, detecting whether it's legacy or dynamic
func (b *TransactionBuilder) buildTransactionMetadata(
	chainTag byte,
	blockRef string,
	expiration uint32,
	gas uint64,
	size uint32,
	gasPriceCoef *uint8,
	maxFeePerGas, maxPriorityFeePerGas *math.HexOrDecimal256,
) map[string]any {
	metadata := map[string]any{
		"chainTag":   chainTag,
		"blockRef":   blockRef,
		"expiration": expiration,
		"gas":        gas,
		"size":       size,
	}

	// Detect transaction type and add appropriate fields
	if gasPriceCoef != nil {
		// Legacy transaction
		metadata["transactionType"] = meshcommon.TransactionTypeLegacy
		metadata["gasPriceCoef"] = *gasPriceCoef
	} else {
		// Dynamic transaction
		metadata["transactionType"] = meshcommon.TransactionTypeDynamic
		if maxFeePerGas != nil {
			metadata["maxFeePerGas"] = (*big.Int)(maxFeePerGas).String()
		}
		if maxPriorityFeePerGas != nil {
			metadata["maxPriorityFeePerGas"] = (*big.Int)(maxPriorityFeePerGas).String()
		}
	}

	return metadata
}

// BuildTransactionFromRequest builds a VeChain transaction from a construction request
func (b *TransactionBuilder) BuildTransactionFromRequest(request types.ConstructionPayloadsRequest, expiration uint32) (*thorTx.Transaction, error) {
	// Extract metadata
	metadata := request.Metadata
	blockRef, ok := metadata["blockRef"].(string)
	if !ok {
		return nil, fmt.Errorf("blockRef is required and must be a string")
	}

	chainTagFloat, ok := metadata["chainTag"].(float64)
	if !ok {
		return nil, fmt.Errorf("chainTag is required and must be a number")
	}

	gasFloat, ok := metadata["gas"].(float64)
	if !ok {
		return nil, fmt.Errorf("gas is required and must be a number")
	}

	transactionType, ok := metadata["transactionType"].(string)
	if !ok {
		return nil, fmt.Errorf("transactionType is required and must be a string")
	}

	nonce, ok := metadata["nonce"].(string)
	if !ok {
		return nil, fmt.Errorf("nonce is required and must be a string")
	}

	chainTag := int(chainTagFloat)
	gas := int64(gasFloat)

	// Parse nonce to uint64
	nonceStr := strings.TrimPrefix(nonce, "0x")
	nonceValue := new(big.Int)
	nonceValue, ok = nonceValue.SetString(nonceStr, 16)
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

	blockRefBytes, err := b.bytesHandler.DecodeHexStringWithPrefix(blockRef)
	if err != nil {
		return nil, fmt.Errorf("invalid blockRef: %w", err)
	}

	builder.BlockRef(thorTx.BlockRef(blockRefBytes))

	// Set expiration from configuration
	builder.Expiration(expiration)
	builder.Gas(uint64(gas))
	builder.Nonce(nonceValue.Uint64())

	if _, hasDelegator := metadata[meshcommon.DelegatorAccountMetadataKey]; hasDelegator {
		builder.Features(thorTx.DelegationFeature)
	}

	// Add clauses from operations
	if err := b.addClausesToBuilder(builder, request.Operations); err != nil {
		return nil, err
	}

	return builder.Build(), nil
}

// createTransactionBuilder creates a transaction builder based on type
func (b *TransactionBuilder) createTransactionBuilder(transactionType string, metadata map[string]any) (*thorTx.Builder, error) {
	if transactionType == meshcommon.TransactionTypeLegacy {
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

	maxFeeStr, ok := metadata["maxFeePerGas"].(string)
	if !ok {
		return nil, fmt.Errorf("maxFeePerGas is required and must be a string for dynamic transactions")
	}
	maxFee, ok := new(big.Int).SetString(maxFeeStr, 10)
	if !ok {
		return nil, fmt.Errorf("invalid maxFeePerGas: %s", maxFeeStr)
	}

	maxPriorityStr, ok := metadata["maxPriorityFeePerGas"].(string)
	if !ok {
		return nil, fmt.Errorf("maxPriorityFeePerGas is required and must be a string for dynamic transactions")
	}
	maxPriority, ok := new(big.Int).SetString(maxPriorityStr, 10)
	if !ok {
		return nil, fmt.Errorf("invalid maxPriorityFeePerGas: %s", maxPriorityStr)
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

	sort.Slice(sortedOps, func(i, j int) bool {
		return sortedOps[i].OperationIdentifier.Index < sortedOps[j].OperationIdentifier.Index
	})

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
						transferData, err := b.bytesHandler.DecodeHexStringWithPrefix(transferDataHex)
						if err != nil {
							return fmt.Errorf("failed to decode transfer data: %w", err)
						}

						// Create clause with contract call
						clause := thorTx.NewClause(&contractAddress)
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
