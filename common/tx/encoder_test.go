package tx

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/math"
	meshcommon "github.com/vechain/mesh/common"
	"github.com/vechain/mesh/config"
	meshtests "github.com/vechain/mesh/tests"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/api"
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
	toAddr, _ := thor.ParseAddress(meshtests.TestAddress1)
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
	toAddr, _ := thor.ParseAddress(meshtests.TestAddress1)
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
	vechainTx = vechainTx.WithSignature([]byte{0x03, 0xf8, 0xa1, 0xca, 0x4d, 0xd3, 0x9d, 0x99, 0xab, 0x54, 0x97, 0xf9, 0x4b, 0x8b, 0x79, 0x11, 0x34, 0x0c, 0xea, 0xc7, 0x18, 0x20, 0x19, 0xb7, 0xbe, 0x9d, 0x81, 0xf0, 0x43, 0xc7, 0x43, 0xf9, 0x5a, 0x69, 0x43, 0x1d, 0x71, 0x5a, 0xde, 0x0c, 0x9b, 0x74, 0x1f, 0x7c, 0x83, 0xd9, 0x57, 0x2a, 0xd8, 0x42, 0x71, 0xb4, 0xf2, 0xec, 0xb6, 0x2c, 0x8f, 0x49, 0xdd, 0xfa, 0x3e, 0x8c, 0x3a, 0xea, 0x01})
	return &MeshTransaction{
		Transaction: vechainTx,
		Origin:      []byte{99, 168, 194, 144, 235, 95, 211, 61, 237, 249, 155, 252, 216, 31, 111, 240, 115, 232, 135, 100},
		Delegator:   []byte{},
	}
}

func createTestConfig() *config.Config {
	return &config.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         meshcommon.OnlineMode,
		BaseGasPrice: "1000000000000000000",
	}
}

// Tests for NewMeshTransactionEncoder
func TestNewMeshTransactionEncoder(t *testing.T) {
	encoder := NewMeshTransactionEncoder(meshthor.NewMockVeChainClient())
	if encoder == nil {
		t.Errorf("NewMeshTransactionEncoder() returned nil")
	}
}

// Tests for EncodeUnsignedTransaction
func TestEncodeUnsignedTransaction_Legacy(t *testing.T) {
	encoder := NewMeshTransactionEncoder(meshthor.NewMockVeChainClient())
	vechainTx := createTestVeChainTransaction()
	origin := []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f}
	delegator := []byte{}

	encoded, err := encoder.EncodeTransaction(&MeshTransaction{
		Transaction: vechainTx,
		Origin:      origin,
		Delegator:   delegator,
	})
	if err != nil {
		t.Errorf("EncodeUnsignedTransaction() error = %v", err)
	}
	if len(encoded) == 0 {
		t.Errorf("EncodeUnsignedTransaction() returned empty data")
	}
}

func TestEncodeUnsignedTransaction_Dynamic(t *testing.T) {
	encoder := NewMeshTransactionEncoder(meshthor.NewMockVeChainClient())

	vechainTx := createTestVeChainDynamicTransaction()
	origin := []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f}
	delegator := []byte{}

	encoded, err := encoder.EncodeTransaction(&MeshTransaction{
		Transaction: vechainTx,
		Origin:      origin,
		Delegator:   delegator,
	})
	if err != nil {
		t.Errorf("EncodeUnsignedTransaction() error = %v", err)
	}
	if len(encoded) == 0 {
		t.Errorf("EncodeUnsignedTransaction() returned empty data")
	}
}

// Tests for DecodeUnsignedTransaction
func TestDecodeUnsignedTransaction_ValidData(t *testing.T) {
	encoder := NewMeshTransactionEncoder(meshthor.NewMockVeChainClient())

	// First encode a transaction
	vechainTx := createTestVeChainTransaction()
	origin := []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f}
	delegator := []byte{}

	encoded, err := encoder.EncodeTransaction(&MeshTransaction{
		Transaction: vechainTx,
		Origin:      origin,
		Delegator:   delegator,
	})
	if err != nil {
		t.Fatalf("EncodeUnsignedTransaction() error = %v", err)
	}

	// Then decode it
	decoded, err := encoder.DecodeUnsignedTransaction(encoded)
	if err != nil {
		t.Errorf("DecodeUnsignedTransaction() error = %v", err)
	}
	if decoded.Transaction == nil {
		t.Errorf("DecodeUnsignedTransaction() returned nil Transaction")
	}

	// Validate that decoded transaction matches original
	if decoded.ChainTag() != vechainTx.ChainTag() {
		t.Errorf("DecodeUnsignedTransaction() ChainTag = %v, want %v", decoded.ChainTag(), vechainTx.ChainTag())
	}
	if decoded.Expiration() != vechainTx.Expiration() {
		t.Errorf("DecodeUnsignedTransaction() Expiration = %v, want %v", decoded.Expiration(), vechainTx.Expiration())
	}
	if decoded.Gas() != vechainTx.Gas() {
		t.Errorf("DecodeUnsignedTransaction() Gas = %v, want %v", decoded.Gas(), vechainTx.Gas())
	}
	if decoded.Nonce() != vechainTx.Nonce() {
		t.Errorf("DecodeUnsignedTransaction() Nonce = %v, want %v", decoded.Nonce(), vechainTx.Nonce())
	}
	if decoded.GasPriceCoef() != vechainTx.GasPriceCoef() {
		t.Errorf("DecodeUnsignedTransaction() GasPriceCoef = %v, want %v", decoded.GasPriceCoef(), vechainTx.GasPriceCoef())
	}

	// Validate origin and delegator
	if !bytes.Equal(decoded.Origin, origin) {
		t.Errorf("DecodeUnsignedTransaction() Origin = %v, want %v", decoded.Origin, origin)
	}
	if !bytes.Equal(decoded.Delegator, delegator) {
		t.Errorf("DecodeUnsignedTransaction() Delegator = %v, want %v", decoded.Delegator, delegator)
	}
}

func TestDecodeUnsignedTransaction_Dynamic(t *testing.T) {
	encoder := NewMeshTransactionEncoder(meshthor.NewMockVeChainClient())

	// First encode a dynamic transaction
	vechainTx := createTestVeChainDynamicTransaction()
	origin := []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f}
	delegator := []byte{}

	encoded, err := encoder.EncodeTransaction(&MeshTransaction{
		Transaction: vechainTx,
		Origin:      origin,
		Delegator:   delegator,
	})
	if err != nil {
		t.Fatalf("EncodeUnsignedTransaction() error = %v", err)
	}

	// Then decode it
	decoded, err := encoder.DecodeUnsignedTransaction(encoded)
	if err != nil {
		t.Errorf("DecodeUnsignedTransaction() error = %v", err)
	}
	if decoded.Transaction == nil {
		t.Errorf("DecodeUnsignedTransaction() returned nil Transaction")
	}

	// Validate that decoded transaction matches original
	if decoded.ChainTag() != vechainTx.ChainTag() {
		t.Errorf("DecodeUnsignedTransaction() ChainTag = %v, want %v", decoded.ChainTag(), vechainTx.ChainTag())
	}
	if decoded.Expiration() != vechainTx.Expiration() {
		t.Errorf("DecodeUnsignedTransaction() Expiration = %v, want %v", decoded.Expiration(), vechainTx.Expiration())
	}
	if decoded.Gas() != vechainTx.Gas() {
		t.Errorf("DecodeUnsignedTransaction() Gas = %v, want %v", decoded.Gas(), vechainTx.Gas())
	}
	if decoded.Nonce() != vechainTx.Nonce() {
		t.Errorf("DecodeUnsignedTransaction() Nonce = %v, want %v", decoded.Nonce(), vechainTx.Nonce())
	}

	// Validate dynamic fee specific fields
	if decoded.Transaction.MaxFeePerGas().Cmp(vechainTx.MaxFeePerGas()) != 0 {
		t.Errorf("DecodeUnsignedTransaction() MaxFeePerGas = %v, want %v", decoded.MaxFeePerGas(), vechainTx.MaxFeePerGas())
	}
	if decoded.Transaction.MaxPriorityFeePerGas().Cmp(vechainTx.MaxPriorityFeePerGas()) != 0 {
		t.Errorf("DecodeUnsignedTransaction() MaxPriorityFeePerGas = %v, want %v", decoded.MaxPriorityFeePerGas(), vechainTx.MaxPriorityFeePerGas())
	}

	// Validate origin and delegator
	if !bytes.Equal(decoded.Origin, origin) {
		t.Errorf("DecodeUnsignedTransaction() Origin = %v, want %v", decoded.Origin, origin)
	}
	if !bytes.Equal(decoded.Delegator, delegator) {
		t.Errorf("DecodeUnsignedTransaction() Delegator = %v, want %v", decoded.Delegator, delegator)
	}
}

func TestDecodeUnsignedTransaction_InvalidData(t *testing.T) {
	encoder := NewMeshTransactionEncoder(meshthor.NewMockVeChainClient())

	invalidData := []byte{0x01, 0x02, 0x03} // Invalid RLP data

	_, err := encoder.DecodeUnsignedTransaction(invalidData)
	if err == nil {
		t.Errorf("DecodeUnsignedTransaction() should return error for invalid data")
	}
}

// Tests for DecodeSignedTransaction
func TestDecodeSignedTransaction_ValidData(t *testing.T) {
	encoder := NewMeshTransactionEncoder(meshthor.NewMockVeChainClient())

	// First encode a signed transaction
	meshTx := createTestMeshTransaction()
	encoded, err := encoder.EncodeTransaction(meshTx)
	if err != nil {
		t.Fatalf("EncodeSignedTransaction() error = %v", err)
	}

	// Then decode it
	decoded, err := encoder.DecodeSignedTransaction(encoded)
	if err != nil {
		t.Errorf("DecodeSignedTransaction() error = %v", err)
	}
	if decoded.Transaction == nil {
		t.Errorf("DecodeSignedTransaction() returned nil Transaction")
	}

	// Validate that decoded transaction matches original
	if decoded.ChainTag() != meshTx.ChainTag() {
		t.Errorf("DecodeSignedTransaction() ChainTag = %v, want %v", decoded.ChainTag(), meshTx.ChainTag())
	}
	if decoded.Expiration() != meshTx.Expiration() {
		t.Errorf("DecodeSignedTransaction() Expiration = %v, want %v", decoded.Expiration(), meshTx.Expiration())
	}
	if decoded.Gas() != meshTx.Gas() {
		t.Errorf("DecodeSignedTransaction() Gas = %v, want %v", decoded.Gas(), meshTx.Gas())
	}
	if decoded.Nonce() != meshTx.Nonce() {
		t.Errorf("DecodeSignedTransaction() Nonce = %v, want %v", decoded.Nonce(), meshTx.Nonce())
	}
	if decoded.GasPriceCoef() != meshTx.GasPriceCoef() {
		t.Errorf("DecodeSignedTransaction() GasPriceCoef = %v, want %v", decoded.GasPriceCoef(), meshTx.GasPriceCoef())
	}

	// Validate origin, delegator, and signature
	if !bytes.Equal(decoded.Origin, meshTx.Origin) {
		t.Errorf("DecodeSignedTransaction() Origin = %v, want %v", decoded.Origin, meshTx.Origin)
	}
	if !bytes.Equal(decoded.Delegator, meshTx.Delegator) {
		t.Errorf("DecodeSignedTransaction() Delegator = %v, want %v", decoded.Delegator, meshTx.Delegator)
	}
	if !bytes.Equal(decoded.Signature(), meshTx.Signature()) {
		t.Errorf("DecodeSignedTransaction() Signature = %+v, want %+v", decoded.Signature(), meshTx.Signature())
	}
}

func TestDecodeSignedTransaction_Dynamic(t *testing.T) {
	encoder := NewMeshTransactionEncoder(meshthor.NewMockVeChainClient())

	// Create a dynamic mesh transaction
	vechainTx := createTestVeChainDynamicTransaction()
	// Use a valid signature generated for dynamic transactions
	vechainTx = vechainTx.WithSignature([]byte{
		0x1b, 0xc4, 0xaf, 0xf0, 0xc0, 0xd4, 0x25, 0xec, 0xd1, 0xa9, 0x31, 0xad, 0x43, 0x51, 0x56, 0xa4,
		0x8b, 0x3e, 0x74, 0xb9, 0xa7, 0x6b, 0x79, 0xfc, 0xc6, 0xe8, 0x66, 0x33, 0x7a, 0x73, 0xf0, 0x5a,
		0x7e, 0x05, 0x77, 0x3a, 0x5f, 0xcb, 0xd4, 0xcf, 0x12, 0x51, 0xcf, 0x02, 0x6e, 0x70, 0xc5, 0xcc,
		0x5b, 0x35, 0x24, 0x86, 0x64, 0x46, 0xde, 0x93, 0xa7, 0xd4, 0x98, 0x97, 0xc0, 0xba, 0xc5, 0x79,
		0x00,
	})
	meshTx := &MeshTransaction{
		Transaction: vechainTx,
		Origin:      []byte{248, 243, 89, 245, 8, 156, 213, 223, 233, 220, 251, 65, 123, 213, 249, 106, 75, 44, 231, 64},
		Delegator:   []byte{},
	}

	encoded, err := encoder.EncodeTransaction(meshTx)
	if err != nil {
		t.Fatalf("EncodeSignedTransaction() error = %v", err)
	}

	// Then decode it
	decoded, err := encoder.DecodeSignedTransaction(encoded)
	if err != nil {
		t.Errorf("DecodeSignedTransaction() error = %v", err)
	}
	if decoded.Transaction == nil {
		t.Errorf("DecodeSignedTransaction() returned nil Transaction")
	}

	// Validate that decoded transaction matches original
	if decoded.ChainTag() != meshTx.ChainTag() {
		t.Errorf("DecodeSignedTransaction() ChainTag = %v, want %v", decoded.ChainTag(), meshTx.ChainTag())
	}
	if decoded.Expiration() != meshTx.Expiration() {
		t.Errorf("DecodeSignedTransaction() Expiration = %v, want %v", decoded.Expiration(), meshTx.Expiration())
	}
	if decoded.Gas() != meshTx.Gas() {
		t.Errorf("DecodeSignedTransaction() Gas = %v, want %v", decoded.Gas(), meshTx.Gas())
	}
	if decoded.Nonce() != meshTx.Nonce() {
		t.Errorf("DecodeSignedTransaction() Nonce = %v, want %v", decoded.Nonce(), meshTx.Nonce())
	}

	// Validate dynamic fee specific fields
	if decoded.Transaction.MaxFeePerGas().Cmp(meshTx.MaxFeePerGas()) != 0 {
		t.Errorf("DecodeSignedTransaction() MaxFeePerGas = %v, want %v", decoded.MaxFeePerGas(), meshTx.MaxFeePerGas())
	}
	if decoded.Transaction.MaxPriorityFeePerGas().Cmp(meshTx.MaxPriorityFeePerGas()) != 0 {
		t.Errorf("DecodeSignedTransaction() MaxPriorityFeePerGas = %v, want %v", decoded.MaxPriorityFeePerGas(), meshTx.MaxPriorityFeePerGas())
	}

	// Validate origin, delegator, and signature
	if !bytes.Equal(decoded.Origin, meshTx.Origin) {
		t.Errorf("DecodeSignedTransaction() Origin = %v, want %v", decoded.Origin, meshTx.Origin)
	}
	if !bytes.Equal(decoded.Delegator, meshTx.Delegator) {
		t.Errorf("DecodeSignedTransaction() Delegator = %v, want %v", decoded.Delegator, meshTx.Delegator)
	}
	if !bytes.Equal(decoded.Signature(), meshTx.Signature()) {
		t.Errorf("DecodeSignedTransaction() Signature = %+v, want %+v", decoded.Signature(), meshTx.Signature())
	}
}

func TestDecodeSignedTransaction_InvalidData(t *testing.T) {
	encoder := NewMeshTransactionEncoder(meshthor.NewMockVeChainClient())

	invalidData := []byte{0x01, 0x02, 0x03} // Invalid RLP data

	_, err := encoder.DecodeSignedTransaction(invalidData)
	if err == nil {
		t.Errorf("DecodeSignedTransaction() should return error for invalid data")
	}
}

// Tests for EncodeSignedTransaction
func TestEncodeSignedTransaction_Legacy(t *testing.T) {
	encoder := NewMeshTransactionEncoder(meshthor.NewMockVeChainClient())
	meshTx := createTestMeshTransaction()

	encoded, err := encoder.EncodeTransaction(meshTx)
	if err != nil {
		t.Errorf("EncodeSignedTransaction() error = %v", err)
	}
	if len(encoded) == 0 {
		t.Errorf("EncodeSignedTransaction() returned empty data")
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
					addr, _ := thor.ParseAddress(meshtests.TestAddress1)
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
			addr, _ := thor.ParseAddress(meshtests.FirstSoloAddress)
			return addr
		}(),
	}

	encoder := NewMeshTransactionEncoder(meshthor.NewMockVeChainClient())
	operations := encoder.ParseTransactionOperationsFromAPI(tx)
	if len(operations) == 0 {
		t.Errorf("ParseTransactionOperationsFromAPI() returned no operations")
	}

	// Check first operation
	op := operations[0]
	if op.Type != meshcommon.OperationTypeTransfer && op.Type != "ContractCall" {
		t.Errorf("ParseTransactionOperationsFromAPI() operation type = %v, want %v or ContractCall", op.Type, meshcommon.OperationTypeTransfer)
	}
}

// Tests for ParseTransactionFromBytes
func TestParseTransactionFromBytes(t *testing.T) {
	encoder := NewMeshTransactionEncoder(meshthor.NewMockVeChainClient())

	// Create and encode a transaction
	vechainTx := createTestVeChainTransaction()

	encoded, err := encoder.EncodeTransaction(&MeshTransaction{
		Transaction: vechainTx,
	})
	if err != nil {
		t.Fatalf("EncodeUnsignedTransaction() error = %v", err)
	}

	// Parse unsigned transaction
	meshTx, operations, signers, err := encoder.ParseTransactionFromBytes(encoded, false)
	if err != nil {
		t.Errorf("ParseTransactionFromBytes() error = %v", err)
	}
	if meshTx == nil {
		t.Errorf("ParseTransactionFromBytes() returned nil meshTx")
	}
	if operations == nil {
		t.Errorf("ParseTransactionFromBytes() returned nil operations")
	}
	if signers != nil {
		t.Errorf("ParseTransactionFromBytes() returned signers")
	}
}

func TestParseTransactionSignersAndOperations(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	encoder := NewMeshTransactionEncoder(mockClient)

	tests := []struct {
		name            string
		meshTx          *MeshTransaction
		expectedOps     int
		expectedSigners int
	}{
		{
			name: "simple VET transfer",
			meshTx: &MeshTransaction{
				Transaction: func() *thorTx.Transaction {
					builder := thorTx.NewBuilder(thorTx.TypeLegacy)
					builder.ChainTag(0x27)
					blockRef := thorTx.BlockRef([8]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef})
					builder.BlockRef(blockRef)
					builder.Expiration(720)
					builder.Gas(21000)
					builder.GasPriceCoef(0)
					builder.Nonce(0x1234567890abcdef)

					// Add a clause
					toAddr, _ := thor.ParseAddress(meshtests.TestAddress1)
					value := new(big.Int)
					value.SetString("1000000000000000000", 10) // 1 VET

					thorClause := thorTx.NewClause(&toAddr)
					thorClause = thorClause.WithValue(value)
					thorClause = thorClause.WithData([]byte{})
					builder.Clause(thorClause)

					return builder.Build()
				}(),
				Origin: func() []byte {
					addr, _ := thor.ParseAddress(meshtests.FirstSoloAddress)
					return addr.Bytes()
				}(),
				Delegator: []byte{},
			},
			expectedOps:     3,
			expectedSigners: 1,
		},
		{
			name: "VET transfer with delegator",
			meshTx: &MeshTransaction{
				Transaction: func() *thorTx.Transaction {
					builder := thorTx.NewBuilder(thorTx.TypeLegacy)
					builder.ChainTag(0x27)
					blockRef := thorTx.BlockRef([8]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef})
					builder.BlockRef(blockRef)
					builder.Expiration(720)
					builder.Gas(21000)
					builder.GasPriceCoef(0)
					builder.Nonce(0x1234567890abcdef)

					// Add a clause
					toAddr, _ := thor.ParseAddress("0x1234567890123456789012345678901234567890")
					value := new(big.Int)
					value.SetString("500000000000000000", 10) // 0.5 VET

					thorClause := thorTx.NewClause(&toAddr)
					thorClause = thorClause.WithValue(value)
					thorClause = thorClause.WithData([]byte{})
					builder.Clause(thorClause)

					return builder.Build()
				}(),
				Origin: func() []byte {
					addr, _ := thor.ParseAddress(meshtests.FirstSoloAddress)
					return addr.Bytes()
				}(),
				Delegator: func() []byte {
					addr, _ := thor.ParseAddress(meshtests.TestAddress1)
					return addr.Bytes()
				}(),
			},
			expectedOps:     3,
			expectedSigners: 2,
		},
		{
			name: "empty clauses",
			meshTx: &MeshTransaction{
				Transaction: func() *thorTx.Transaction {
					builder := thorTx.NewBuilder(thorTx.TypeLegacy)
					builder.ChainTag(0x27)
					blockRef := thorTx.BlockRef([8]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef})
					builder.BlockRef(blockRef)
					builder.Expiration(720)
					builder.Gas(0)
					builder.GasPriceCoef(0)
					builder.Nonce(0x1234567890abcdef)

					return builder.Build()
				}(),
				Origin: func() []byte {
					addr, _ := thor.ParseAddress(meshtests.FirstSoloAddress)
					return addr.Bytes()
				}(),
				Delegator: []byte{},
			},
			expectedOps:     0,
			expectedSigners: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operations, signers := encoder.parseTransactionSignersAndOperations(tt.meshTx, true)

			if len(operations) != tt.expectedOps {
				t.Errorf("parseTransactionSignersAndOperations() operations length = %v, want %v", len(operations), tt.expectedOps)
			}

			if len(signers) != tt.expectedSigners {
				t.Errorf("parseTransactionSignersAndOperations() signers length = %v, want %v", len(signers), tt.expectedSigners)
			}

			// Verify origin is always the first signer
			if len(signers) > 0 {
				originAddr := thor.BytesToAddress(tt.meshTx.Origin)
				if signers[0].Address != originAddr.String() {
					t.Errorf("parseTransactionSignersAndOperations() first signer = %v, want %v", signers[0].Address, originAddr.String())
				}
			}

			// Verify delegator is second signer if present
			if len(tt.meshTx.Delegator) > 0 && len(signers) > 1 {
				delegatorAddr := thor.BytesToAddress(tt.meshTx.Delegator)
				if signers[1].Address != delegatorAddr.String() {
					t.Errorf("parseTransactionSignersAndOperations() second signer = %v, want %v", signers[1].Address, delegatorAddr.String())
				}
			}
		})
	}
}
