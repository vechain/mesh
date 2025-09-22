package utils

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/vechain/thor/v2/tx"
)

// RosettaTransaction represents a transaction in Rosetta format with additional fields
type RosettaTransaction struct {
	// Reuse Thor's native transaction structure
	*tx.Transaction
	// Rosetta-specific fields
	Origin    []byte
	Delegator []byte
	// Signature (for signed transactions)
	Signature []byte
}

// RosettaTransactionEncoder handles encoding and decoding of Rosetta transactions
type RosettaTransactionEncoder struct{}

// NewRosettaTransactionEncoder creates a new Rosetta transaction encoder
func NewRosettaTransactionEncoder() *RosettaTransactionEncoder {
	return &RosettaTransactionEncoder{}
}

// EncodeUnsignedTransaction encodes an unsigned transaction using Rosetta RLP schema
func (e *RosettaTransactionEncoder) EncodeUnsignedTransaction(vechainTx *tx.Transaction, origin, delegator []byte) ([]byte, error) {
	rosettaTx := &RosettaTransaction{
		Transaction: vechainTx,
		Origin:      origin,
		Delegator:   delegator,
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
