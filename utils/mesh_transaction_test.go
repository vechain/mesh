package utils

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/vechain/mesh/config"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/api/transactions"
	"github.com/vechain/thor/v2/thor"
	thorTx "github.com/vechain/thor/v2/tx"
)

// Test helper functions
func createTestVeChainTransaction() *thorTx.Transaction {
	builder := thorTx.NewBuilder(thorTx.TypeLegacy)
	builder.ChainTag(0x27)
	blockRef := thorTx.BlockRef([8]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef})
	builder.BlockRef(blockRef)
	builder.Expiration(720)
	builder.Gas(21000)
	builder.GasPriceCoef(0)
	builder.Nonce(0x1234567890abcdef)

	// Add a clause
	toAddr, _ := thor.ParseAddress("0x16277a1ff38678291c41d1820957c78bb5da59ce")
	value := new(big.Int)
	value.SetString("1000000000000000000", 10) // 1 VET

	thorClause := thorTx.NewClause(&toAddr)
	thorClause = thorClause.WithValue(value)
	thorClause = thorClause.WithData([]byte{})
	builder.Clause(thorClause)

	return builder.Build()
}

func createTestVeChainDynamicTransaction() *thorTx.Transaction {
	builder := thorTx.NewBuilder(thorTx.TypeDynamicFee)
	builder.ChainTag(0x27)
	blockRef := thorTx.BlockRef([8]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef})
	builder.BlockRef(blockRef)
	builder.Expiration(720)
	builder.Gas(21000)
	builder.MaxFeePerGas(big.NewInt(1000000000000000000))        // 1 VET
	builder.MaxPriorityFeePerGas(big.NewInt(100000000000000000)) // 0.1 VET
	builder.Nonce(0x1234567890abcdef)

	// Add a clause
	toAddr, _ := thor.ParseAddress("0x16277a1ff38678291c41d1820957c78bb5da59ce")
	value := new(big.Int)
	value.SetString("1000000000000000000", 10) // 1 VET

	thorClause := thorTx.NewClause(&toAddr)
	thorClause = thorClause.WithValue(value)
	thorClause = thorClause.WithData([]byte{})
	builder.Clause(thorClause)

	return builder.Build()
}

func createTestMeshTransaction() *MeshTransaction {
	vechainTx := createTestVeChainTransaction()
	return &MeshTransaction{
		Transaction: vechainTx,
		Origin:      []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
		Delegator:   []byte{},
		Signature:   []byte{0x03, 0xf8, 0xa1, 0xca, 0x4d, 0xd3, 0x9d, 0x99, 0xab, 0x54, 0x97, 0xf9, 0x4b, 0x8b, 0x79, 0x11, 0x34, 0x0c, 0xea, 0xc7, 0x18, 0x20, 0x19, 0xb7, 0xbe, 0x9d, 0x81, 0xf0, 0x43, 0xc7, 0x43, 0xf9, 0x5a, 0x69, 0x43, 0x1d, 0x71, 0x5a, 0xde, 0x0c, 0x9b, 0x74, 0x1f, 0x7c, 0x83, 0xd9, 0x57, 0x2a, 0xd8, 0x42, 0x71, 0xb4, 0xf2, 0xec, 0xb6, 0x2c, 0x8f, 0x49, 0xdd, 0xfa, 0x3e, 0x8c, 0x3a, 0xea, 0x01},
	}
}

func createTestConfig() *config.Config {
	return &config.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000",
	}
}

// Tests for NewMeshTransactionEncoder
func TestNewMeshTransactionEncoder(t *testing.T) {
	encoder := NewMeshTransactionEncoder()
	if encoder == nil {
		t.Errorf("NewMeshTransactionEncoder() returned nil")
	}
}

// Tests for EncodeUnsignedTransaction
func TestMeshTransactionEncoder_EncodeUnsignedTransaction_Legacy(t *testing.T) {
	encoder := NewMeshTransactionEncoder()
	vechainTx := createTestVeChainTransaction()
	origin := []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f}
	delegator := []byte{}

	encoded, err := encoder.EncodeUnsignedTransaction(vechainTx, origin, delegator)
	if err != nil {
		t.Errorf("EncodeUnsignedTransaction() error = %v", err)
	}
	if len(encoded) == 0 {
		t.Errorf("EncodeUnsignedTransaction() returned empty data")
	}
}

func TestMeshTransactionEncoder_EncodeUnsignedTransaction_Dynamic(t *testing.T) {
	encoder := NewMeshTransactionEncoder()

	vechainTx := createTestVeChainDynamicTransaction()
	origin := []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f}
	delegator := []byte{}

	encoded, err := encoder.EncodeUnsignedTransaction(vechainTx, origin, delegator)
	if err != nil {
		t.Errorf("EncodeUnsignedTransaction() error = %v", err)
	}
	if len(encoded) == 0 {
		t.Errorf("EncodeUnsignedTransaction() returned empty data")
	}
}

// Tests for DecodeUnsignedTransaction
func TestMeshTransactionEncoder_DecodeUnsignedTransaction_ValidData(t *testing.T) {
	encoder := NewMeshTransactionEncoder()

	// First encode a transaction
	vechainTx := createTestVeChainTransaction()
	origin := []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f}
	delegator := []byte{}

	encoded, err := encoder.EncodeUnsignedTransaction(vechainTx, origin, delegator)
	if err != nil {
		t.Fatalf("EncodeUnsignedTransaction() error = %v", err)
	}

	// Then decode it
	decoded, err := encoder.DecodeUnsignedTransaction(encoded)
	if err != nil {
		t.Errorf("DecodeUnsignedTransaction() error = %v", err)
	}
	if decoded == nil {
		t.Errorf("DecodeUnsignedTransaction() returned nil")
	}
	if decoded.Transaction == nil {
		t.Errorf("DecodeUnsignedTransaction() returned nil Transaction")
	}
}

func TestMeshTransactionEncoder_DecodeUnsignedTransaction_Dynamic(t *testing.T) {
	encoder := NewMeshTransactionEncoder()

	// First encode a dynamic transaction
	vechainTx := createTestVeChainDynamicTransaction()
	origin := []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f}
	delegator := []byte{}

	encoded, err := encoder.EncodeUnsignedTransaction(vechainTx, origin, delegator)
	if err != nil {
		t.Fatalf("EncodeUnsignedTransaction() error = %v", err)
	}

	// Then decode it
	decoded, err := encoder.DecodeUnsignedTransaction(encoded)
	if err != nil {
		t.Errorf("DecodeUnsignedTransaction() error = %v", err)
	}
	if decoded == nil {
		t.Errorf("DecodeUnsignedTransaction() returned nil")
	}
	if decoded.Transaction == nil {
		t.Errorf("DecodeUnsignedTransaction() returned nil Transaction")
	}
}

func TestMeshTransactionEncoder_DecodeUnsignedTransaction_InvalidData(t *testing.T) {
	encoder := NewMeshTransactionEncoder()

	invalidData := []byte{0x01, 0x02, 0x03} // Invalid RLP data

	_, err := encoder.DecodeUnsignedTransaction(invalidData)
	if err == nil {
		t.Errorf("DecodeUnsignedTransaction() should return error for invalid data")
	}
}

// Tests for DecodeSignedTransaction
func TestMeshTransactionEncoder_DecodeSignedTransaction_ValidData(t *testing.T) {
	encoder := NewMeshTransactionEncoder()

	// First encode a signed transaction
	meshTx := createTestMeshTransaction()
	encoded, err := encoder.EncodeSignedTransaction(meshTx)
	if err != nil {
		t.Fatalf("EncodeSignedTransaction() error = %v", err)
	}

	// Then decode it
	decoded, err := encoder.DecodeSignedTransaction(encoded)
	if err != nil {
		t.Errorf("DecodeSignedTransaction() error = %v", err)
	}
	if decoded == nil {
		t.Errorf("DecodeSignedTransaction() returned nil")
	}
	if decoded.Transaction == nil {
		t.Errorf("DecodeSignedTransaction() returned nil Transaction")
	}
}

func TestMeshTransactionEncoder_DecodeSignedTransaction_Dynamic(t *testing.T) {
	encoder := NewMeshTransactionEncoder()

	// Create a dynamic mesh transaction
	vechainTx := createTestVeChainDynamicTransaction()
	meshTx := &MeshTransaction{
		Transaction: vechainTx,
		Origin:      []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
		Delegator:   []byte{},
		Signature:   []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f, 0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0x5b, 0x5c, 0x5d, 0x5e, 0x5f, 0x60, 0x61, 0x62, 0x63, 0x64, 0x65},
	}

	encoded, err := encoder.EncodeSignedTransaction(meshTx)
	if err != nil {
		t.Fatalf("EncodeSignedTransaction() error = %v", err)
	}

	// Then decode it
	decoded, err := encoder.DecodeSignedTransaction(encoded)
	if err != nil {
		t.Errorf("DecodeSignedTransaction() error = %v", err)
	}
	if decoded == nil {
		t.Errorf("DecodeSignedTransaction() returned nil")
	}
	if decoded.Transaction == nil {
		t.Errorf("DecodeSignedTransaction() returned nil Transaction")
	}
}

func TestMeshTransactionEncoder_DecodeSignedTransaction_InvalidData(t *testing.T) {
	encoder := NewMeshTransactionEncoder()

	invalidData := []byte{0x01, 0x02, 0x03} // Invalid RLP data

	_, err := encoder.DecodeSignedTransaction(invalidData)
	if err == nil {
		t.Errorf("DecodeSignedTransaction() should return error for invalid data")
	}
}

// Tests for EncodeSignedTransaction
func TestMeshTransactionEncoder_EncodeSignedTransaction_Legacy(t *testing.T) {
	encoder := NewMeshTransactionEncoder()
	meshTx := createTestMeshTransaction()

	encoded, err := encoder.EncodeSignedTransaction(meshTx)
	if err != nil {
		t.Errorf("EncodeSignedTransaction() error = %v", err)
	}
	if len(encoded) == 0 {
		t.Errorf("EncodeSignedTransaction() returned empty data")
	}
}

// Tests for convertClausesToMesh
func TestMeshTransactionEncoder_convertClausesToMesh(t *testing.T) {
	encoder := NewMeshTransactionEncoder()

	// Create test clauses
	toAddr, _ := thor.ParseAddress("0x16277a1ff38678291c41d1820957c78bb5da59ce")
	value := new(big.Int)
	value.SetString("1000000000000000000", 10)

	clause := thorTx.NewClause(&toAddr)
	clause = clause.WithValue(value)
	clause = clause.WithData([]byte{0x01, 0x02, 0x03})

	clauses := []*thorTx.Clause{clause}

	meshClauses := encoder.convertClausesToMesh(clauses)
	if len(meshClauses) != 1 {
		t.Errorf("convertClausesToMesh() returned %d clauses, want 1", len(meshClauses))
	}

	clauseArray := meshClauses[0].([]any)
	if len(clauseArray) != 3 {
		t.Errorf("convertClausesToMesh() clause array length = %d, want 3", len(clauseArray))
	}

	// Check that the first element (to address) matches
	toBytes := clauseArray[0].([]byte)
	expectedToBytes := toAddr.Bytes()
	if !bytes.Equal(toBytes, expectedToBytes) {
		t.Errorf("convertClausesToMesh() clause 'to' bytes = %v, want %v", toBytes, expectedToBytes)
	}
}

// Tests for convertNonceToBytes
func TestMeshTransactionEncoder_convertNonceToBytes(t *testing.T) {
	encoder := NewMeshTransactionEncoder()

	nonce := uint64(0x1234567890abcdef)
	bytes := encoder.convertNonceToBytes(nonce)

	if len(bytes) == 0 {
		t.Errorf("convertNonceToBytes() returned empty bytes")
	}
}

// Tests for convertBytesToNonce
func TestMeshTransactionEncoder_convertBytesToNonce(t *testing.T) {
	encoder := NewMeshTransactionEncoder()

	nonce := uint64(0x1234567890abcdef)
	bytes := encoder.convertNonceToBytes(nonce)

	convertedNonce := encoder.convertBytesToNonce(bytes)
	if convertedNonce != nonce {
		t.Errorf("convertBytesToNonce() = %v, want %v", convertedNonce, nonce)
	}
}

// Tests for convertBytesToUint64
func TestMeshTransactionEncoder_convertBytesToUint64(t *testing.T) {
	encoder := NewMeshTransactionEncoder()

	bytes := []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef}
	result := encoder.convertBytesToUint64(bytes)

	expected := uint64(0x1234567890abcdef)
	if result != expected {
		t.Errorf("convertBytesToUint64() = %v, want %v", result, expected)
	}
}

// Tests for convertBytesToUint32
func TestMeshTransactionEncoder_convertBytesToUint32(t *testing.T) {
	encoder := NewMeshTransactionEncoder()

	bytes := []byte{0x12, 0x34, 0x56, 0x78}
	result := encoder.convertBytesToUint32(bytes)

	expected := uint32(0x12345678)
	if result != expected {
		t.Errorf("convertBytesToUint32() = %v, want %v", result, expected)
	}
}

// Tests for ParseTransactionOperationsFromAPI
func TestParseTransactionOperationsFromAPI(t *testing.T) {
	// Create test transaction
	tx := &api.JSONEmbeddedTx{
		ID: func() thor.Bytes32 {
			hash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
			return hash
		}(),
		Clauses: []*api.JSONClause{
			{
				To: func() *thor.Address {
					addr, _ := thor.ParseAddress("0x16277a1ff38678291c41d1820957c78bb5da59ce")
					return &addr
				}(),
				Value: func() math.HexOrDecimal256 {
					val, _ := new(big.Int).SetString("0xde0b6b3a7640000", 0)
					return math.HexOrDecimal256(*val)
				}(),
				Data: "0x",
			},
		},
		Origin: func() thor.Address {
			addr, _ := thor.ParseAddress("0xf077b491b355e64048ce21e3a6fc4751eeea77fa")
			return addr
		}(),
	}

	operations := ParseTransactionOperationsFromAPI(tx)
	if len(operations) == 0 {
		t.Errorf("ParseTransactionOperationsFromAPI() returned no operations")
	}

	// Check first operation
	op := operations[0]
	if op.Type != OperationTypeTransfer && op.Type != "ContractCall" {
		t.Errorf("ParseTransactionOperationsFromAPI() operation type = %v, want %v or ContractCall", op.Type, OperationTypeTransfer)
	}
}

// Tests for BuildMeshTransactionFromAPI
func TestBuildMeshTransactionFromAPI(t *testing.T) {
	// Create test transaction
	tx := &api.JSONEmbeddedTx{
		ID: func() thor.Bytes32 {
			hash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
			return hash
		}(),
		Clauses: []*api.JSONClause{
			{
				To: func() *thor.Address {
					addr, _ := thor.ParseAddress("0x16277a1ff38678291c41d1820957c78bb5da59ce")
					return &addr
				}(),
				Value: func() math.HexOrDecimal256 {
					val, _ := new(big.Int).SetString("0xde0b6b3a7640000", 0)
					return math.HexOrDecimal256(*val)
				}(),
				Data: "0x",
			},
		},
		Origin: func() thor.Address {
			addr, _ := thor.ParseAddress("0xf077b491b355e64048ce21e3a6fc4751eeea77fa")
			return addr
		}(),
	}

	operations := ParseTransactionOperationsFromAPI(tx)
	meshTx := BuildMeshTransactionFromAPI(tx, operations)

	if meshTx == nil {
		t.Errorf("BuildMeshTransactionFromAPI() returned nil")
	}
	if meshTx.TransactionIdentifier == nil {
		t.Errorf("BuildMeshTransactionFromAPI() returned nil TransactionIdentifier")
	}
}

// Tests for ParseTransactionFromBytes
func TestParseTransactionFromBytes(t *testing.T) {
	encoder := NewMeshTransactionEncoder()

	// Create and encode a transaction
	vechainTx := createTestVeChainTransaction()
	origin := []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f}
	delegator := []byte{}

	encoded, err := encoder.EncodeUnsignedTransaction(vechainTx, origin, delegator)
	if err != nil {
		t.Fatalf("EncodeUnsignedTransaction() error = %v", err)
	}

	// Parse unsigned transaction
	meshTx, operations, signers, err := ParseTransactionFromBytes(encoded, false, encoder)
	if err != nil {
		t.Errorf("ParseTransactionFromBytes() error = %v", err)
	}
	if meshTx == nil {
		t.Errorf("ParseTransactionFromBytes() returned nil meshTx")
	}
	if operations == nil {
		t.Errorf("ParseTransactionFromBytes() returned nil operations")
	}
	if signers == nil {
		t.Errorf("ParseTransactionFromBytes() returned nil signers")
	}
}

// Tests for BuildTransactionFromRequest
func TestBuildTransactionFromRequest(t *testing.T) {
	config := createTestConfig()

	request := types.ConstructionPayloadsRequest{
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: VETCurrency,
				},
			},
		},
		Metadata: map[string]any{
			"transactionType": "legacy",
			"blockRef":        "0x0000000000000000",
			"chainTag":        float64(1),
			"gas":             float64(21000),
			"nonce":           "0x1",
			"gasPriceCoef":    uint8(128),
		},
	}

	tx, err := BuildTransactionFromRequest(request, config)
	if err != nil {
		t.Errorf("BuildTransactionFromRequest() error = %v", err)
	}
	if tx == nil {
		t.Errorf("BuildTransactionFromRequest() returned nil transaction")
	}
}

// Tests for createTransactionBuilder
func TestCreateTransactionBuilder_Legacy(t *testing.T) {
	metadata := map[string]any{
		"transactionType": "legacy",
		"blockRef":        "0x0000000000000000",
		"chainTag":        float64(1),
		"gas":             float64(21000),
		"nonce":           "0x1",
		"gasPriceCoef":    uint8(128),
	}

	builder, err := createTransactionBuilder("legacy", metadata)
	if err != nil {
		t.Errorf("createTransactionBuilder() error = %v", err)
	}
	if builder == nil {
		t.Errorf("createTransactionBuilder() returned nil builder")
	}
}

func TestCreateTransactionBuilder_Dynamic(t *testing.T) {
	metadata := map[string]any{
		"transactionType":      "dynamic",
		"blockRef":             "0x0000000000000000",
		"chainTag":             float64(1),
		"gas":                  float64(21000),
		"nonce":                "0x1",
		"maxFeePerGas":         "1000000000000000000",
		"maxPriorityFeePerGas": "1000000000000000000",
	}

	builder, err := createTransactionBuilder("dynamic", metadata)
	if err != nil {
		t.Errorf("createTransactionBuilder() error = %v", err)
	}
	if builder == nil {
		t.Errorf("createTransactionBuilder() returned nil builder")
	}
}

// Tests for addClausesToBuilder
func TestAddClausesToBuilder(t *testing.T) {
	builder := thorTx.NewBuilder(thorTx.TypeLegacy)

	operations := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			Type:                OperationTypeTransfer,
			Account: &types.AccountIdentifier{
				Address: "0x16277a1ff38678291c41d1820957c78bb5da59ce",
			},
			Amount: &types.Amount{
				Value:    "1000000000000000000",
				Currency: VETCurrency,
			},
		},
	}

	err := addClausesToBuilder(builder, operations)
	if err != nil {
		t.Errorf("addClausesToBuilder() error = %v", err)
	}
}

// Tests for ParseTransactionOperationsFromTransactions
func TestParseTransactionOperationsFromTransactions(t *testing.T) {
	// Create test transaction
	tx := &transactions.Transaction{
		ID: func() thor.Bytes32 {
			hash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
			return hash
		}(),
		Clauses: api.Clauses{
			{
				To: func() *thor.Address {
					addr, _ := thor.ParseAddress("0x16277a1ff38678291c41d1820957c78bb5da59ce")
					return &addr
				}(),
				Value: func() *math.HexOrDecimal256 {
					val, _ := new(big.Int).SetString("0xde0b6b3a7640000", 0)
					hexVal := math.HexOrDecimal256(*val)
					return &hexVal
				}(),
				Data: "0x",
			},
		},
		Origin: func() thor.Address {
			addr, _ := thor.ParseAddress("0xf077b491b355e64048ce21e3a6fc4751eeea77fa")
			return addr
		}(),
	}

	operations := ParseTransactionOperationsFromTransactions(tx)
	if len(operations) == 0 {
		t.Errorf("ParseTransactionOperationsFromTransactions() returned no operations")
	}

	// Check first operation
	op := operations[0]
	if op.Type != OperationTypeTransfer && op.Type != "ContractCall" {
		t.Errorf("ParseTransactionOperationsFromTransactions() operation type = %v, want %v or ContractCall", op.Type, OperationTypeTransfer)
	}
}

// Tests for BuildMeshTransactionFromTransactions
func TestBuildMeshTransactionFromTransactions(t *testing.T) {
	// Create test transaction
	tx := &transactions.Transaction{
		ID: func() thor.Bytes32 {
			hash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
			return hash
		}(),
		Clauses: api.Clauses{
			{
				To: func() *thor.Address {
					addr, _ := thor.ParseAddress("0x16277a1ff38678291c41d1820957c78bb5da59ce")
					return &addr
				}(),
				Value: func() *math.HexOrDecimal256 {
					val, _ := new(big.Int).SetString("0xde0b6b3a7640000", 0)
					hexVal := math.HexOrDecimal256(*val)
					return &hexVal
				}(),
				Data: "0x",
			},
		},
		Origin: func() thor.Address {
			addr, _ := thor.ParseAddress("0xf077b491b355e64048ce21e3a6fc4751eeea77fa")
			return addr
		}(),
	}

	operations := ParseTransactionOperationsFromTransactions(tx)
	meshTx := BuildMeshTransactionFromTransactions(tx, operations)

	if meshTx == nil {
		t.Errorf("BuildMeshTransactionFromTransactions() returned nil")
	}
	if meshTx.TransactionIdentifier == nil {
		t.Errorf("BuildMeshTransactionFromTransactions() returned nil TransactionIdentifier")
	}
}
