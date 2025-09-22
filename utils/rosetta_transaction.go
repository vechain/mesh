package utils

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/tx"
)

// RosettaTransaction represents a transaction with Rosetta-specific fields
type RosettaTransaction struct {
	*tx.Transaction
	Origin    []byte
	Delegator []byte
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
	// Create Rosetta RLP structure based on transaction type
	if vechainTx.Type() == tx.TypeLegacy {
		return e.encodeUnsignedLegacyTransaction(vechainTx, origin, delegator)
	} else {
		return e.encodeUnsignedDynamicTransaction(vechainTx, origin, delegator)
	}
}

// DecodeUnsignedTransaction decodes an unsigned transaction from Rosetta RLP format
func (e *RosettaTransactionEncoder) DecodeUnsignedTransaction(data []byte) (*RosettaTransaction, error) {
	// Try to decode as legacy transaction first (9 fields)
	if rosettaTx, err := e.decodeUnsignedLegacyTransaction(data); err == nil {
		return rosettaTx, nil
	}

	// Try to decode as dynamic fee transaction (10 fields)
	if rosettaTx, err := e.decodeUnsignedDynamicTransaction(data); err == nil {
		return rosettaTx, nil
	}

	return nil, fmt.Errorf("failed to decode as either legacy or dynamic fee transaction")
}

// DecodeSignedTransaction decodes a signed transaction from Rosetta RLP format
func (e *RosettaTransactionEncoder) DecodeSignedTransaction(data []byte) (*RosettaTransaction, error) {
	// Try to decode as signed legacy transaction first (10 fields)
	if rosettaTx, err := e.decodeSignedLegacyTransaction(data); err == nil {
		return rosettaTx, nil
	}

	// Try to decode as signed dynamic fee transaction (11 fields)
	if rosettaTx, err := e.decodeSignedDynamicTransaction(data); err == nil {
		return rosettaTx, nil
	}

	return nil, fmt.Errorf("failed to decode as either signed legacy or signed dynamic fee transaction")
}

// EncodeSignedTransaction encodes a signed Rosetta transaction
func (e *RosettaTransactionEncoder) EncodeSignedTransaction(rosettaTx *RosettaTransaction) ([]byte, error) {
	// Create Rosetta RLP structure based on transaction type
	if rosettaTx.Transaction.Type() == tx.TypeLegacy {
		return e.encodeSignedLegacyTransaction(rosettaTx)
	} else {
		return e.encodeSignedDynamicTransaction(rosettaTx)
	}
}

// encodeUnsignedLegacyTransaction encodes a legacy transaction using Rosetta RLP schema
func (e *RosettaTransactionEncoder) encodeUnsignedLegacyTransaction(vechainTx *tx.Transaction, origin, delegator []byte) ([]byte, error) {
	// Create Rosetta legacy transaction RLP structure (9 fields)
	blockRef := vechainTx.BlockRef()
	rosettaTx := []any{
		vechainTx.ChainTag(),
		blockRef[:],
		vechainTx.Expiration(),
		e.convertClausesToRosetta(vechainTx.Clauses()),
		vechainTx.Gas(),
		e.convertNonceToBytes(vechainTx.Nonce()),
		origin,
		delegator,
		vechainTx.GasPriceCoef(),
	}

	return rlp.EncodeToBytes(rosettaTx)
}

// encodeUnsignedDynamicTransaction encodes a dynamic fee transaction using Rosetta RLP schema
func (e *RosettaTransactionEncoder) encodeUnsignedDynamicTransaction(vechainTx *tx.Transaction, origin, delegator []byte) ([]byte, error) {
	// Create Rosetta dynamic fee transaction RLP structure (10 fields)
	blockRef := vechainTx.BlockRef()
	rosettaTx := []any{
		vechainTx.ChainTag(),
		blockRef[:],
		vechainTx.Expiration(),
		e.convertClausesToRosetta(vechainTx.Clauses()),
		vechainTx.Gas(),
		e.convertNonceToBytes(vechainTx.Nonce()),
		origin,
		delegator,
		vechainTx.MaxFeePerGas(),
		vechainTx.MaxPriorityFeePerGas(),
	}

	return rlp.EncodeToBytes(rosettaTx)
}

// decodeUnsignedLegacyTransaction decodes a legacy transaction from Rosetta RLP format
func (e *RosettaTransactionEncoder) decodeUnsignedLegacyTransaction(data []byte) (*RosettaTransaction, error) {
	var fields []any
	if err := rlp.DecodeBytes(data, &fields); err != nil {
		return nil, err
	}

	// Legacy transaction should have 9 fields
	if len(fields) != 9 {
		return nil, fmt.Errorf("invalid legacy transaction: expected 9 fields, got %d", len(fields))
	}

	// Extract fields according to Rosetta schema
	chainTagBytes := fields[0].([]byte)
	chainTag := chainTagBytes[0]
	blockRef := fields[1].([]byte)
	expirationBytes := fields[2].([]byte)
	expiration := e.convertBytesToUint32(expirationBytes)
	clauses := fields[3].([]any)
	gasBytes := fields[4].([]byte)
	gas := e.convertBytesToUint64(gasBytes)
	nonce := fields[5].([]byte)
	origin := fields[6].([]byte)
	delegator := fields[7].([]byte)
	gasPriceCoefBytes := fields[8].([]byte)
	gasPriceCoef := gasPriceCoefBytes[0]

	// Build Thor transaction using native builder
	builder := tx.NewBuilder(tx.TypeLegacy)
	builder.ChainTag(chainTag)

	// Convert blockRef to 8-byte array
	var blockRefArray [8]byte
	copy(blockRefArray[:], blockRef)
	builder.BlockRef(tx.BlockRef(blockRefArray))

	builder.Expiration(expiration)
	builder.Gas(gas)
	builder.Nonce(e.convertBytesToNonce(nonce))
	builder.GasPriceCoef(gasPriceCoef)

	// Add clauses using Thor's native methods
	for _, clauseData := range clauses {
		clause := clauseData.([]any)
		to := clause[0].([]byte)
		valueBytes := clause[1].([]byte)
		value := new(big.Int).SetBytes(valueBytes)
		data := clause[2].([]byte)

		toAddr := thor.BytesToAddress(to)
		thorClause := tx.NewClause(&toAddr)
		thorClause = thorClause.WithValue(value)
		thorClause = thorClause.WithData(data)
		builder.Clause(thorClause)
	}

	vechainTx := builder.Build()

	return &RosettaTransaction{
		Transaction: vechainTx,
		Origin:      origin,
		Delegator:   delegator,
	}, nil
}

// decodeUnsignedDynamicTransaction decodes a dynamic fee transaction from Rosetta RLP format
func (e *RosettaTransactionEncoder) decodeUnsignedDynamicTransaction(data []byte) (*RosettaTransaction, error) {
	var fields []any
	if err := rlp.DecodeBytes(data, &fields); err != nil {
		return nil, err
	}

	// Dynamic fee transaction should have 10 fields
	if len(fields) != 10 {
		return nil, fmt.Errorf("invalid dynamic fee transaction: expected 10 fields, got %d", len(fields))
	}

	// Extract fields according to Rosetta schema
	chainTagBytes := fields[0].([]byte)
	chainTag := chainTagBytes[0]
	blockRef := fields[1].([]byte)
	expirationBytes := fields[2].([]byte)
	expiration := e.convertBytesToUint32(expirationBytes)
	clauses := fields[3].([]any)
	gasBytes := fields[4].([]byte)
	gas := e.convertBytesToUint64(gasBytes)
	nonce := fields[5].([]byte)
	origin := fields[6].([]byte)
	delegator := fields[7].([]byte)
	maxFeePerGasBytes := fields[8].([]byte)
	maxFeePerGas := new(big.Int).SetBytes(maxFeePerGasBytes)
	maxPriorityFeePerGasBytes := fields[9].([]byte)
	maxPriorityFeePerGas := new(big.Int).SetBytes(maxPriorityFeePerGasBytes)

	// Build Thor transaction using native builder
	builder := tx.NewBuilder(tx.TypeDynamicFee)
	builder.ChainTag(chainTag)

	// Convert blockRef to 8-byte array
	var blockRefArray [8]byte
	copy(blockRefArray[:], blockRef)
	builder.BlockRef(tx.BlockRef(blockRefArray))

	builder.Expiration(expiration)
	builder.Gas(gas)
	builder.Nonce(e.convertBytesToNonce(nonce))
	builder.MaxFeePerGas(maxFeePerGas)
	builder.MaxPriorityFeePerGas(maxPriorityFeePerGas)

	// Add clauses using Thor's native methods
	for _, clauseData := range clauses {
		clause := clauseData.([]any)
		to := clause[0].([]byte)
		valueBytes := clause[1].([]byte)
		value := new(big.Int).SetBytes(valueBytes)
		data := clause[2].([]byte)

		toAddr := thor.BytesToAddress(to)
		thorClause := tx.NewClause(&toAddr)
		thorClause = thorClause.WithValue(value)
		thorClause = thorClause.WithData(data)
		builder.Clause(thorClause)
	}

	vechainTx := builder.Build()

	return &RosettaTransaction{
		Transaction: vechainTx,
		Origin:      origin,
		Delegator:   delegator,
	}, nil
}

// encodeSignedLegacyTransaction encodes a signed legacy transaction
func (e *RosettaTransactionEncoder) encodeSignedLegacyTransaction(rosettaTx *RosettaTransaction) ([]byte, error) {
	// Create Rosetta signed legacy transaction RLP structure (10 fields)
	blockRef := rosettaTx.BlockRef()
	rosettaTxRLP := []any{
		rosettaTx.ChainTag(),
		blockRef[:],
		rosettaTx.Expiration(),
		e.convertClausesToRosetta(rosettaTx.Clauses()),
		rosettaTx.Gas(),
		e.convertNonceToBytes(rosettaTx.Nonce()),
		rosettaTx.Origin,
		rosettaTx.Delegator,
		rosettaTx.GasPriceCoef(),
		rosettaTx.Signature,
	}

	return rlp.EncodeToBytes(rosettaTxRLP)
}

// encodeSignedDynamicTransaction encodes a signed dynamic fee transaction
func (e *RosettaTransactionEncoder) encodeSignedDynamicTransaction(rosettaTx *RosettaTransaction) ([]byte, error) {
	// Create Rosetta signed dynamic fee transaction RLP structure (11 fields)
	blockRef := rosettaTx.BlockRef()
	rosettaTxRLP := []any{
		rosettaTx.ChainTag(),
		blockRef[:],
		rosettaTx.Expiration(),
		e.convertClausesToRosetta(rosettaTx.Clauses()),
		rosettaTx.Gas(),
		e.convertNonceToBytes(rosettaTx.Nonce()),
		rosettaTx.Origin,
		rosettaTx.Delegator,
		rosettaTx.MaxFeePerGas(),
		rosettaTx.MaxPriorityFeePerGas(),
		rosettaTx.Signature,
	}

	return rlp.EncodeToBytes(rosettaTxRLP)
}

// decodeSignedLegacyTransaction decodes a signed legacy transaction from Rosetta RLP format
func (e *RosettaTransactionEncoder) decodeSignedLegacyTransaction(data []byte) (*RosettaTransaction, error) {
	var fields []any
	if err := rlp.DecodeBytes(data, &fields); err != nil {
		return nil, err
	}

	// Signed legacy transaction should have 10 fields
	if len(fields) != 10 {
		return nil, fmt.Errorf("invalid signed legacy transaction: expected 10 fields, got %d", len(fields))
	}

	// Extract fields according to Rosetta schema
	chainTagBytes := fields[0].([]byte)
	chainTag := chainTagBytes[0]
	blockRef := fields[1].([]byte)
	expirationBytes := fields[2].([]byte)
	expiration := e.convertBytesToUint32(expirationBytes)
	clauses := fields[3].([]any)
	gasBytes := fields[4].([]byte)
	gas := e.convertBytesToUint64(gasBytes)
	nonce := fields[5].([]byte)
	origin := fields[6].([]byte)
	delegator := fields[7].([]byte)
	gasPriceCoefBytes := fields[8].([]byte)
	gasPriceCoef := gasPriceCoefBytes[0]
	signature := fields[9].([]byte)

	// Build Thor transaction using native builder
	builder := tx.NewBuilder(tx.TypeLegacy)
	builder.ChainTag(chainTag)

	// Convert blockRef to 8-byte array
	var blockRefArray [8]byte
	copy(blockRefArray[:], blockRef)
	builder.BlockRef(tx.BlockRef(blockRefArray))

	builder.Expiration(expiration)
	builder.Gas(gas)
	builder.Nonce(e.convertBytesToNonce(nonce))
	builder.GasPriceCoef(gasPriceCoef)

	// Add clauses using Thor's native methods
	for _, clauseData := range clauses {
		clause := clauseData.([]any)
		to := clause[0].([]byte)
		valueBytes := clause[1].([]byte)
		value := new(big.Int).SetBytes(valueBytes)
		data := clause[2].([]byte)

		toAddr := thor.BytesToAddress(to)
		thorClause := tx.NewClause(&toAddr)
		thorClause = thorClause.WithValue(value)
		thorClause = thorClause.WithData(data)
		builder.Clause(thorClause)
	}

	vechainTx := builder.Build()

	return &RosettaTransaction{
		Transaction: vechainTx,
		Origin:      origin,
		Delegator:   delegator,
		Signature:   signature,
	}, nil
}

// decodeSignedDynamicTransaction decodes a signed dynamic fee transaction from Rosetta RLP format
func (e *RosettaTransactionEncoder) decodeSignedDynamicTransaction(data []byte) (*RosettaTransaction, error) {
	var fields []any
	if err := rlp.DecodeBytes(data, &fields); err != nil {
		return nil, err
	}

	// Signed dynamic fee transaction should have 11 fields
	if len(fields) != 11 {
		return nil, fmt.Errorf("invalid signed dynamic fee transaction: expected 11 fields, got %d", len(fields))
	}

	// Extract fields according to Rosetta schema
	chainTagBytes := fields[0].([]byte)
	chainTag := chainTagBytes[0]
	blockRef := fields[1].([]byte)
	expirationBytes := fields[2].([]byte)
	expiration := e.convertBytesToUint32(expirationBytes)
	clauses := fields[3].([]any)
	gasBytes := fields[4].([]byte)
	gas := e.convertBytesToUint64(gasBytes)
	nonce := fields[5].([]byte)
	origin := fields[6].([]byte)
	delegator := fields[7].([]byte)
	maxFeePerGasBytes := fields[8].([]byte)
	maxFeePerGas := new(big.Int).SetBytes(maxFeePerGasBytes)
	maxPriorityFeePerGasBytes := fields[9].([]byte)
	var maxPriorityFeePerGas *big.Int
	if len(maxPriorityFeePerGasBytes) > 0 {
		maxPriorityFeePerGas = new(big.Int).SetBytes(maxPriorityFeePerGasBytes)
	} else {
		maxPriorityFeePerGas = big.NewInt(0)
	}
	signature := fields[10].([]byte)

	// Build Thor transaction using native builder
	builder := tx.NewBuilder(tx.TypeDynamicFee)
	builder.ChainTag(chainTag)

	// Convert blockRef to 8-byte array
	var blockRefArray [8]byte
	copy(blockRefArray[:], blockRef)
	builder.BlockRef(tx.BlockRef(blockRefArray))

	builder.Expiration(expiration)
	builder.Gas(gas)
	builder.Nonce(e.convertBytesToNonce(nonce))
	builder.MaxFeePerGas(maxFeePerGas)
	builder.MaxPriorityFeePerGas(maxPriorityFeePerGas)

	// Add clauses using Thor's native methods
	for _, clauseData := range clauses {
		clause := clauseData.([]any)
		to := clause[0].([]byte)
		valueBytes := clause[1].([]byte)
		value := new(big.Int).SetBytes(valueBytes)
		data := clause[2].([]byte)

		toAddr := thor.BytesToAddress(to)
		thorClause := tx.NewClause(&toAddr)
		thorClause = thorClause.WithValue(value)
		thorClause = thorClause.WithData(data)
		builder.Clause(thorClause)
	}

	vechainTx := builder.Build()

	return &RosettaTransaction{
		Transaction: vechainTx,
		Origin:      origin,
		Delegator:   delegator,
		Signature:   signature,
	}, nil
}

// Helper methods
func (e *RosettaTransactionEncoder) convertClausesToRosetta(clauses []*tx.Clause) []any {
	rosettaClauses := make([]any, len(clauses))
	for i, clause := range clauses {
		rosettaClauses[i] = []any{
			clause.To().Bytes(),
			clause.Value(),
			clause.Data(),
		}
	}
	return rosettaClauses
}

func (e *RosettaTransactionEncoder) convertNonceToBytes(nonce uint64) []byte {
	nonceBytes := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		nonceBytes[i] = byte(nonce)
		nonce >>= 8
	}
	return nonceBytes
}

func (e *RosettaTransactionEncoder) convertBytesToNonce(nonceBytes []byte) uint64 {
	var nonce uint64
	for i := range 8 {
		nonce = (nonce << 8) | uint64(nonceBytes[i])
	}
	return nonce
}

func (e *RosettaTransactionEncoder) convertBytesToUint64(bytes []byte) uint64 {
	var result uint64
	for i := range bytes {
		result = (result << 8) | uint64(bytes[i])
	}
	return result
}

func (e *RosettaTransactionEncoder) convertBytesToUint32(bytes []byte) uint32 {
	var result uint32
	for i := range bytes {
		result = (result << 8) | uint32(bytes[i])
	}
	return result
}
