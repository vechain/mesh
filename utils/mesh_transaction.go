package utils

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/vechain/mesh/config"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/api/transactions"
	"github.com/vechain/thor/v2/thor"
	thorTx "github.com/vechain/thor/v2/tx"
)

// MeshTransaction represents a transaction with Mesh-specific fields
type MeshTransaction struct {
	*thorTx.Transaction
	// Origin and delegator are required as separate fields because they get encoded when there is no signature
	Origin    []byte
	Delegator []byte
}

// MeshTransactionEncoder handles encoding and decoding of Mesh transactions
type MeshTransactionEncoder struct{}

// NewMeshTransactionEncoder creates a new Mesh transaction encoder
func NewMeshTransactionEncoder() *MeshTransactionEncoder {
	return &MeshTransactionEncoder{}
}

// EncodeTransaction encodes a transaction using Mesh RLP schema
func (e *MeshTransactionEncoder) EncodeTransaction(meshTx *MeshTransaction) ([]byte, error) {
	// Use native Thor encoding and add Mesh-specific fields
	thorBytes, err := meshTx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to encode Thor transaction: %w", err)
	}

	if len(meshTx.Signature()) > 0 {
		return thorBytes, nil
	}

	// Create Mesh structure if not signed: [thorTransaction, origin, delegator]
	meshTxRLP := []any{
		thorBytes,
		meshTx.Origin,
		meshTx.Delegator,
	}

	return rlp.EncodeToBytes(meshTxRLP)
}

// DecodeUnsignedTransaction decodes an unsigned transaction from Mesh RLP format
func (e *MeshTransactionEncoder) DecodeUnsignedTransaction(data []byte) (*MeshTransaction, error) {
	// For unsigned transactions, decode as RLP list: [thorTransaction, origin, delegator]
	var fields []any
	if err := rlp.DecodeBytes(data, &fields); err != nil {
		return nil, err
	}

	if len(fields) != 3 {
		return nil, fmt.Errorf("invalid Mesh transaction: expected 3 fields, got %d", len(fields))
	}

	// Decode Thor transaction from bytes
	thorBytes := fields[0].([]byte)
	var thorTx thorTx.Transaction
	if err := thorTx.UnmarshalBinary(thorBytes); err != nil {
		return nil, fmt.Errorf("failed to decode Thor transaction: %w", err)
	}

	origin := fields[1].([]byte)
	delegator := fields[2].([]byte)

	return &MeshTransaction{
		Transaction: &thorTx,
		Origin:      origin,
		Delegator:   delegator,
	}, nil
}

// DecodeSignedTransaction decodes a signed transaction from Mesh RLP format
func (e *MeshTransactionEncoder) DecodeSignedTransaction(data []byte) (*MeshTransaction, error) {
	// For signed transactions, decode directly as Thor transaction
	var thorTx thorTx.Transaction
	if err := thorTx.UnmarshalBinary(data); err != nil {
		return nil, fmt.Errorf("failed to decode Thor transaction: %w", err)
	}

	originAddr, err := thorTx.Origin()
	if err != nil {
		return nil, fmt.Errorf("failed to get origin: %w", err)
	}
	delegatorAddr, err := thorTx.Delegator()
	if err != nil {
		return nil, fmt.Errorf("failed to get delegator: %w", err)
	}

	var delegatorBytes []byte
	if delegatorAddr != nil {
		delegatorBytes = delegatorAddr.Bytes()
	}

	return &MeshTransaction{
		Transaction: &thorTx,
		Origin:      originAddr.Bytes(),
		Delegator:   delegatorBytes,
	}, nil
}

// buildMeshTransaction is a helper function that builds a Mesh transaction
func buildMeshTransaction(hash string, operations []*types.Operation, metadata map[string]any) *types.Transaction {
	return &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{Hash: hash},
		Operations:            operations,
		Metadata:              metadata,
	}
}

// BuildMeshTransactionFromAPI builds a Mesh transaction directly from api.JSONEmbeddedTx
func BuildMeshTransactionFromAPI(tx *api.JSONEmbeddedTx, operations []*types.Operation) *types.Transaction {
	return buildMeshTransaction(tx.ID.String(), operations, map[string]any{
		"chainTag": tx.ChainTag, "blockRef": "0x" + fmt.Sprintf("%x", tx.BlockRef),
		"expiration": tx.Expiration, "gas": tx.Gas, "gasPriceCoef": tx.GasPriceCoef, "size": tx.Size,
	})
}

// ParseTransactionFromBytes parses a transaction from bytes and returns operations and signers
func ParseTransactionFromBytes(txBytes []byte, signed bool, encoder *MeshTransactionEncoder) (*MeshTransaction, []*types.Operation, []*types.AccountIdentifier, error) {
	var meshTx *MeshTransaction
	var err error

	if signed {
		// For signed transactions, try to decode as Mesh transaction first
		meshTx, err = encoder.DecodeSignedTransaction(txBytes)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to decode signed transaction: %w", err)
		}
	} else {
		// For unsigned transactions, decode as Mesh transaction
		meshTx, err = encoder.DecodeUnsignedTransaction(txBytes)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to decode unsigned transaction: %w", err)
		}
	}

	// Parse operations and signers
	operations, signers := parseTransactionSignersAndOperations(meshTx)

	return meshTx, operations, signers, nil
}

// parseTransactionSignersAndOperations parses signers and operations from a transaction
func parseTransactionSignersAndOperations(meshTx *MeshTransaction) ([]*types.Operation, []*types.AccountIdentifier) {
	originAddr := thor.BytesToAddress(meshTx.Origin)
	var delegatorAddr *thor.Address
	if len(meshTx.Delegator) > 0 {
		delegator := thor.BytesToAddress(meshTx.Delegator)
		delegatorAddr = &delegator
	}

	signers := []*types.AccountIdentifier{
		{Address: originAddr.String()},
	}

	// Add delegator signer if present
	if delegatorAddr != nil {
		signers = append(signers, &types.AccountIdentifier{
			Address: delegatorAddr.String(),
		})
	}

	// Parse clauses as operations
	clauses := meshTx.Clauses()
	operationIndex := 0
	operations := make([]*types.Operation, 0, len(clauses))
	for _, clause := range clauses {
		to := clause.To()
		if to != nil {
			// Sender operation (negative amount)
			senderOp := &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: int64(operationIndex),
				},
				Type: OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: originAddr.String(),
				},
				Amount: &types.Amount{
					Value:    "-" + clause.Value().String(),
					Currency: VETCurrency,
				},
			}
			operations = append(operations, senderOp)
			operationIndex++

			// Receiver operation (positive amount)
			receiverOp := &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: int64(operationIndex),
				},
				Type: OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: to.String(),
				},
				Amount: &types.Amount{
					Value:    clause.Value().String(),
					Currency: VETCurrency,
				},
			}
			operations = append(operations, receiverOp)
			operationIndex++
		}
	}

	return operations, signers
}

// BuildTransactionFromRequest builds a VeChain transaction from a construction request
func BuildTransactionFromRequest(request types.ConstructionPayloadsRequest, config *config.Config) (*thorTx.Transaction, error) {
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
	builder, err := createTransactionBuilder(transactionType, metadata)
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
	if err := addClausesToBuilder(builder, request.Operations); err != nil {
		return nil, err
	}

	return builder.Build(), nil
}

// createTransactionBuilder creates a transaction builder based on type
func createTransactionBuilder(transactionType string, metadata map[string]any) (*thorTx.Builder, error) {
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
func addClausesToBuilder(builder *thorTx.Builder, operations []*types.Operation) error {
	for _, op := range operations {
		if op.Type == OperationTypeTransfer {
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

			clause := thorTx.NewClause(&toAddr)
			clause = clause.WithValue(value)
			builder.Clause(clause)
		}
		// FeeDelegation operations are handled in the signing process
	}
	return nil
}

// ParseTransactionOperationsFromAPI parses operations directly from api.JSONEmbeddedTx
func ParseTransactionOperationsFromAPI(tx *api.JSONEmbeddedTx) []*types.Operation {
	status := OperationStatusSucceeded
	if tx.Reverted {
		status = OperationStatusReverted
	}
	return parseTransactionOperationsFromClauses(tx.Clauses, tx.Origin.String(), tx.Gas, &status)
}

// BuildMeshTransactionFromTransactions builds a Mesh transaction directly from transactions.Transaction
func BuildMeshTransactionFromTransaction(tx *transactions.Transaction, operations []*types.Operation) *types.Transaction {
	return buildMeshTransaction(tx.ID.String(), operations, map[string]any{
		"chainTag": tx.ChainTag, "blockRef": "0x" + fmt.Sprintf("%x", tx.BlockRef),
		"expiration": tx.Expiration, "gas": tx.Gas, "gasPriceCoef": tx.GasPriceCoef, "size": tx.Size,
	})
}
