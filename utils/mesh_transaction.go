package utils

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
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

// decodeMeshTransaction decodes a Mesh transaction format
func (e *MeshTransactionEncoder) decodeMeshTransaction(data []byte, signed bool) (*MeshTransaction, error) {
	if signed {
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

// DecodeUnsignedTransaction decodes an unsigned transaction from Mesh RLP format
func (e *MeshTransactionEncoder) DecodeUnsignedTransaction(data []byte) (*MeshTransaction, error) {
	meshTx, err := e.decodeMeshTransaction(data, false)
	if err != nil {
		return nil, err
	}

	return meshTx, nil
}

// DecodeSignedTransaction decodes a signed transaction from Mesh RLP format
func (e *MeshTransactionEncoder) DecodeSignedTransaction(data []byte) (*MeshTransaction, error) {
	meshTx, err := e.decodeMeshTransaction(data, true)
	if err != nil {
		return nil, err
	}

	return meshTx, nil
}

// parseTransactionOperationsFromClauses is a helper function that parses operations from clauses
func parseTransactionOperationsFromClauses(clauses []*api.JSONClause, originAddr string, gas uint64, status string) []*types.Operation {
	var operations []*types.Operation
	hasValueTransfer, hasContractInteraction, hasEnergyTransfer := false, false, gas > 0

	// Analyze clauses
	for _, clause := range clauses {
		valueBytes, _ := clause.Value.MarshalText()
		value := new(big.Int)
		value.SetString(string(valueBytes), 10)
		if value.Cmp(big.NewInt(0)) > 0 {
			hasValueTransfer = true
		}
		if len(clause.Data) > 0 || (clause.To != nil && !clause.To.IsZero()) {
			hasContractInteraction = true
		}
	}

	if !hasValueTransfer && !hasContractInteraction && !hasEnergyTransfer {
		return operations
	}

	operationIndex := 0
	for clauseIndex, clause := range clauses {
		valueBytes, _ := clause.Value.MarshalText()
		value := new(big.Int)
		value.SetString(string(valueBytes), 10)

		if value.Cmp(big.NewInt(0)) > 0 {
			valueStr := value.String()
			// Sender operation
			operations = append(operations, &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{Index: int64(operationIndex)},
				Type:                OperationTypeTransfer,
				Status:              StringPtr(status),
				Account:             &types.AccountIdentifier{Address: originAddr},
				Amount:              &types.Amount{Value: "-" + valueStr, Currency: VETCurrency},
				Metadata:            map[string]any{"clauseIndex": clauseIndex},
			})
			operationIndex++

			// Receiver operation
			if clause.To != nil && !clause.To.IsZero() {
				operations = append(operations, &types.Operation{
					OperationIdentifier: &types.OperationIdentifier{Index: int64(operationIndex)},
					Type:                OperationTypeTransfer,
					Status:              StringPtr(status),
					Account:             &types.AccountIdentifier{Address: clause.To.String()},
					Amount:              &types.Amount{Value: valueStr, Currency: VETCurrency},
					Metadata:            map[string]any{"clauseIndex": clauseIndex},
				})
				operationIndex++
			}
		}

		// Contract interaction operation
		if len(clause.Data) > 0 || (clause.To != nil && !clause.To.IsZero()) {
			toAddr := ""
			if clause.To != nil {
				toAddr = clause.To.String()
			}
			operations = append(operations, &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{Index: int64(operationIndex)},
				Type:                "ContractCall",
				Status:              StringPtr(status),
				Account:             &types.AccountIdentifier{Address: originAddr},
				Amount:              &types.Amount{Value: "0", Currency: VETCurrency},
				Metadata: map[string]any{
					"clauseIndex": clauseIndex,
					"to":          toAddr,
					"data":        "0x" + fmt.Sprintf("%x", clause.Data),
				},
			})
			operationIndex++
		}
	}

	// Energy transfer operation
	if hasEnergyTransfer {
		operations = append(operations, &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{Index: int64(operationIndex)},
			Type:                OperationTypeFee,
			Status:              StringPtr(status),
			Account:             &types.AccountIdentifier{Address: originAddr},
			Amount:              &types.Amount{Value: "-" + strconv.FormatUint(gas, 10), Currency: VTHOCurrency},
			Metadata:            map[string]any{"gas": strconv.FormatUint(gas, 10)},
		})
	}

	return operations
}

// parseTransactionOperationsFromTransactionsClauses is a helper function that parses operations from transactions.Clauses
func parseTransactionOperationsFromTransactionsClauses(clauses api.Clauses, originAddr string, gas uint64, status string) []*types.Operation {
	var operations []*types.Operation
	hasValueTransfer, hasContractInteraction, hasEnergyTransfer := false, false, gas > 0

	// Analyze clauses
	for _, clause := range clauses {
		valueBytes, _ := clause.Value.MarshalText()
		value := new(big.Int)
		value.SetString(string(valueBytes), 10)
		if value.Cmp(big.NewInt(0)) > 0 {
			hasValueTransfer = true
		}
		if len(clause.Data) > 0 || (clause.To != nil && !clause.To.IsZero()) {
			hasContractInteraction = true
		}
	}

	if !hasValueTransfer && !hasContractInteraction && !hasEnergyTransfer {
		return operations
	}

	operationIndex := 0
	for clauseIndex, clause := range clauses {
		valueBytes, _ := clause.Value.MarshalText()
		value := new(big.Int)
		value.SetString(string(valueBytes), 10)

		if value.Cmp(big.NewInt(0)) > 0 {
			valueStr := value.String()
			// Sender operation
			operations = append(operations, &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{Index: int64(operationIndex)},
				Type:                OperationTypeTransfer,
				Status:              StringPtr(status),
				Account:             &types.AccountIdentifier{Address: originAddr},
				Amount:              &types.Amount{Value: "-" + valueStr, Currency: VETCurrency},
				Metadata:            map[string]any{"clauseIndex": clauseIndex},
			})
			operationIndex++

			// Receiver operation
			if clause.To != nil && !clause.To.IsZero() {
				operations = append(operations, &types.Operation{
					OperationIdentifier: &types.OperationIdentifier{Index: int64(operationIndex)},
					Type:                OperationTypeTransfer,
					Status:              StringPtr(status),
					Account:             &types.AccountIdentifier{Address: clause.To.String()},
					Amount:              &types.Amount{Value: valueStr, Currency: VETCurrency},
					Metadata:            map[string]any{"clauseIndex": clauseIndex},
				})
				operationIndex++
			}
		}

		// Contract interaction operation
		if len(clause.Data) > 0 || (clause.To != nil && !clause.To.IsZero()) {
			toAddr := ""
			if clause.To != nil {
				toAddr = clause.To.String()
			}
			operations = append(operations, &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{Index: int64(operationIndex)},
				Type:                "ContractCall",
				Status:              StringPtr(status),
				Account:             &types.AccountIdentifier{Address: originAddr},
				Amount:              &types.Amount{Value: "0", Currency: VETCurrency},
				Metadata: map[string]any{
					"clauseIndex": clauseIndex,
					"to":          toAddr,
					"data":        "0x" + fmt.Sprintf("%x", clause.Data),
				},
			})
			operationIndex++
		}
	}

	// Energy transfer operation
	if hasEnergyTransfer {
		operations = append(operations, &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{Index: int64(operationIndex)},
			Type:                OperationTypeFee,
			Status:              StringPtr(status),
			Account:             &types.AccountIdentifier{Address: originAddr},
			Amount:              &types.Amount{Value: "-" + strconv.FormatUint(gas, 10), Currency: VTHOCurrency},
			Metadata:            map[string]any{"gas": strconv.FormatUint(gas, 10)},
		})
	}

	return operations
}

// buildMeshTransaction is a helper function that builds a Mesh transaction
func buildMeshTransaction(hash string, operations []*types.Operation, metadata map[string]any) *types.Transaction {
	return &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{Hash: hash},
		Operations:            operations,
		Metadata:              metadata,
	}
}

// ParseTransactionOperationsFromAPI parses operations directly from api.JSONEmbeddedTx
func ParseTransactionOperationsFromAPI(tx *api.JSONEmbeddedTx) []*types.Operation {
	return parseTransactionOperationsFromClauses(tx.Clauses, tx.Origin.String(), tx.Gas, "Success")
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
	var vechainTx *thorTx.Transaction
	var meshTx *MeshTransaction
	var err error

	if signed {
		// For signed transactions, try to decode as Mesh transaction first
		meshTx, err = encoder.DecodeSignedTransaction(txBytes)
		if err != nil {
			// Fallback to native Thor decoding
			var nativeTx thorTx.Transaction
			stream := rlp.NewStream(bytes.NewReader(txBytes), 0)
			if err := nativeTx.DecodeRLP(stream); err != nil {
				return nil, nil, nil, fmt.Errorf("failed to decode transaction: %w", err)
			}
			vechainTx = &nativeTx
		} else {
			vechainTx = meshTx.Transaction
		}
	} else {
		// For unsigned transactions, decode as Mesh transaction
		meshTx, err = encoder.DecodeUnsignedTransaction(txBytes)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to decode unsigned transaction: %w", err)
		}
		vechainTx = meshTx.Transaction
	}

	// Parse operations and signers
	operations, signers := parseTransactionSignersAndOperations(vechainTx, meshTx)

	return meshTx, operations, signers, nil
}

// parseTransactionSignersAndOperations parses signers and operations from a transaction
func parseTransactionSignersAndOperations(vechainTx *thorTx.Transaction, meshTx *MeshTransaction) ([]*types.Operation, []*types.AccountIdentifier) {
	var operations []*types.Operation
	var signers []*types.AccountIdentifier

	// Add origin signer
	var originAddr thor.Address
	var delegatorAddr *thor.Address

	if meshTx != nil {
		// Use Mesh transaction fields
		originAddr = thor.BytesToAddress(meshTx.Origin)
		if len(meshTx.Delegator) > 0 {
			delegator := thor.BytesToAddress(meshTx.Delegator)
			delegatorAddr = &delegator
		}
	} else {
		// Use native Thor methods
		originAddr, _ = vechainTx.Origin()
		delegator, err := vechainTx.Delegator()
		if err == nil && delegator != nil {
			delegatorAddr = delegator
		}
	}

	signers = append(signers, &types.AccountIdentifier{
		Address: originAddr.String(),
	})

	// Add delegator signer if present
	if delegatorAddr != nil {
		signers = append(signers, &types.AccountIdentifier{
			Address: delegatorAddr.String(),
		})
	}

	// Parse clauses as operations
	clauses := vechainTx.Clauses()
	operationIndex := 0
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

// ParseTransactionOperationsFromTransactions parses operations directly from transactions.Transaction
func ParseTransactionOperationsFromTransactions(tx *transactions.Transaction) []*types.Operation {
	return parseTransactionOperationsFromTransactionsClauses(tx.Clauses, tx.Origin.String(), tx.Gas, "Pending")
}

// BuildMeshTransactionFromTransactions builds a Mesh transaction directly from transactions.Transaction
func BuildMeshTransactionFromTransactions(tx *transactions.Transaction, operations []*types.Operation) *types.Transaction {
	return buildMeshTransaction(tx.ID.String(), operations, map[string]any{
		"chainTag": tx.ChainTag, "blockRef": "0x" + fmt.Sprintf("%x", tx.BlockRef),
		"expiration": tx.Expiration, "gas": tx.Gas, "gasPriceCoef": tx.GasPriceCoef, "size": tx.Size,
	})
}
