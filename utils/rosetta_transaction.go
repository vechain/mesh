package utils

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/vechain/thor/v2/tx"
)

// RosettaTransaction represents a transaction in Rosetta format with additional fields
type RosettaTransaction struct {
	ChainTag   byte
	BlockRef   []byte
	Expiration uint32
	Clauses    []RosettaClause
	Gas        uint64
	Nonce      []byte
	Origin     []byte
	Delegator  []byte
	// Legacy fields
	GasPriceCoef *uint8
	// Dynamic fee fields
	MaxFeePerGas         *big.Int
	MaxPriorityFeePerGas *big.Int
	// Signature (for signed transactions)
	Signature []byte
}

// RosettaClause represents a clause in Rosetta format
type RosettaClause struct {
	To    []byte
	Value *big.Int
	Data  []byte
}

// RosettaTransactionEncoder handles encoding and decoding of Rosetta transactions
type RosettaTransactionEncoder struct{}

// NewRosettaTransactionEncoder creates a new Rosetta transaction encoder
func NewRosettaTransactionEncoder() *RosettaTransactionEncoder {
	return &RosettaTransactionEncoder{}
}

// EncodeUnsignedTransaction encodes an unsigned transaction using Rosetta RLP schema
// This reuses Thor's transaction structure but adapts it to Rosetta's RLP format
func (e *RosettaTransactionEncoder) EncodeUnsignedTransaction(vechainTx *tx.Transaction, origin, delegator []byte) ([]byte, error) {
	// Extract transaction fields using Thor's native methods
	clauses := make([]RosettaClause, len(vechainTx.Clauses()))
	for i, clause := range vechainTx.Clauses() {
		clauses[i] = RosettaClause{
			To:    clause.To().Bytes(),
			Value: clause.Value(),
			Data:  clause.Data(),
		}
	}

	// Convert nonce to bytes (Thor returns uint64, we need []byte)
	nonceBytes := make([]byte, 8)
	nonce := vechainTx.Nonce()
	for i := 7; i >= 0; i-- {
		nonceBytes[i] = byte(nonce)
		nonce >>= 8
	}

	// Convert BlockRef to bytes (Thor returns [8]byte, we need []byte)
	blockRef := vechainTx.BlockRef()
	blockRefBytes := make([]byte, 8)
	copy(blockRefBytes, blockRef[:])

	// Create Rosetta transaction structure
	rosettaTx := RosettaTransaction{
		ChainTag:   vechainTx.ChainTag(),
		BlockRef:   blockRefBytes,
		Expiration: vechainTx.Expiration(),
		Clauses:    clauses,
		Gas:        vechainTx.Gas(),
		Nonce:      nonceBytes,
		Origin:     origin,
		Delegator:  delegator,
	}

	// Add type-specific fields
	if vechainTx.Type() == tx.TypeLegacy {
		gasPriceCoef := vechainTx.GasPriceCoef()
		rosettaTx.GasPriceCoef = &gasPriceCoef
	} else {
		rosettaTx.MaxFeePerGas = vechainTx.MaxFeePerGas()
		rosettaTx.MaxPriorityFeePerGas = vechainTx.MaxPriorityFeePerGas()
	}

	return rlp.EncodeToBytes(rosettaTx)
}

// DecodeUnsignedTransaction decodes an unsigned transaction from Rosetta RLP format
// It automatically detects whether it's legacy or dynamic fee based on the presence of fields
func (e *RosettaTransactionEncoder) DecodeUnsignedTransaction(data []byte) (*RosettaTransaction, error) {
	var rosettaTx RosettaTransaction
	if err := rlp.DecodeBytes(data, &rosettaTx); err != nil {
		return nil, err
	}
	return &rosettaTx, nil
}

// DecodeSignedTransaction decodes a signed transaction from Rosetta RLP format
func (e *RosettaTransactionEncoder) DecodeSignedTransaction(data []byte) (*RosettaTransaction, error) {
	var rosettaTx RosettaTransaction
	if err := rlp.DecodeBytes(data, &rosettaTx); err != nil {
		return nil, err
	}
	return &rosettaTx, nil
}

// EncodeSignedTransaction encodes a signed Rosetta transaction
func (e *RosettaTransactionEncoder) EncodeSignedTransaction(rosettaTx *RosettaTransaction) ([]byte, error) {
	return rlp.EncodeToBytes(rosettaTx)
}

// GetTransactionType determines if a RosettaTransaction is legacy or dynamic fee
func (e *RosettaTransactionEncoder) GetTransactionType(rosettaTx *RosettaTransaction) string {
	if rosettaTx.GasPriceCoef != nil {
		return "legacy"
	}
	return "dynamic"
}

// ConvertToHexString converts a RosettaTransaction to hex string format
func (e *RosettaTransactionEncoder) ConvertToHexString(rosettaTx *RosettaTransaction) (string, error) {
	data, err := rlp.EncodeToBytes(rosettaTx)
	if err != nil {
		return "", fmt.Errorf("failed to encode Rosetta transaction: %w", err)
	}
	return "0x" + hex.EncodeToString(data), nil
}

// ConvertFromHexString converts a hex string to RosettaTransaction
func (e *RosettaTransactionEncoder) ConvertFromHexString(hexStr string) (*RosettaTransaction, error) {
	// Remove 0x prefix if present
	if len(hexStr) > 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}

	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %w", err)
	}

	return e.DecodeUnsignedTransaction(data)
}
