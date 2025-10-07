package operations

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/math"
	meshcommon "github.com/vechain/mesh/common"
	meshtests "github.com/vechain/mesh/tests"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/tx"
)

// Test helper functions for clause parser tests
var testStatus = meshcommon.OperationStatusSucceeded

func createTestJSONClause(to *thor.Address, value *big.Int, data string) *api.JSONClause {
	hexValue := math.HexOrDecimal256(*value)
	return &api.JSONClause{
		To:    to,
		Value: hexValue,
		Data:  data,
	}
}

func createTestClause(to *thor.Address, value *big.Int, data string) *api.Clause {
	hexValue := math.HexOrDecimal256(*value)
	return &api.Clause{
		To:    to,
		Value: &hexValue,
		Data:  data,
	}
}

func createTestAddress(addr string) *thor.Address {
	address, _ := thor.ParseAddress(addr)
	return &address
}

// Test analyzeClauses
func TestMeshTransactionEncoder_analyzeClauses(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	tests := []struct {
		name                        string
		clauseData                  []ClauseData
		gas                         uint64
		expectedValueTransfer       bool
		expectedContractInteraction bool
		expectedEnergyTransfer      bool
	}{
		{
			name:                        "empty clauses",
			clauseData:                  []ClauseData{},
			gas:                         0,
			expectedValueTransfer:       false,
			expectedContractInteraction: false,
			expectedEnergyTransfer:      false,
		},
		{
			name: "value transfer only",
			clauseData: []ClauseData{
				JSONClauseAdapter{Clause: createTestJSONClause(
					createTestAddress(meshtests.TestAddress1),
					big.NewInt(1000000000000000000),
					"0x",
				)},
			},
			gas:                         0,
			expectedValueTransfer:       true,
			expectedContractInteraction: true,
			expectedEnergyTransfer:      false,
		},
		{
			name: "contract interaction only",
			clauseData: []ClauseData{
				JSONClauseAdapter{Clause: createTestJSONClause(
					createTestAddress(meshtests.TestAddress1),
					big.NewInt(0),
					"0x1234",
				)},
			},
			gas:                         0,
			expectedValueTransfer:       false,
			expectedContractInteraction: true,
			expectedEnergyTransfer:      false,
		},
		{
			name:                        "energy transfer only",
			clauseData:                  []ClauseData{},
			gas:                         21000,
			expectedValueTransfer:       false,
			expectedContractInteraction: false,
			expectedEnergyTransfer:      true,
		},
		{
			name: "all types",
			clauseData: []ClauseData{
				JSONClauseAdapter{Clause: createTestJSONClause(
					createTestAddress(meshtests.TestAddress1),
					big.NewInt(1000000000000000000),
					"0x1234",
				)},
			},
			gas:                         21000,
			expectedValueTransfer:       true,
			expectedContractInteraction: true,
			expectedEnergyTransfer:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasValueTransfer, hasContractInteraction, hasEnergyTransfer := parser.analyzeClauses(tt.clauseData, tt.gas)

			if hasValueTransfer != tt.expectedValueTransfer {
				t.Errorf("analyzeClauses() hasValueTransfer = %v, want %v", hasValueTransfer, tt.expectedValueTransfer)
			}
			if hasContractInteraction != tt.expectedContractInteraction {
				t.Errorf("analyzeClauses() hasContractInteraction = %v, want %v", hasContractInteraction, tt.expectedContractInteraction)
			}
			if hasEnergyTransfer != tt.expectedEnergyTransfer {
				t.Errorf("analyzeClauses() hasEnergyTransfer = %v, want %v", hasEnergyTransfer, tt.expectedEnergyTransfer)
			}
		})
	}
}

// Test getClauseValue
func TestMeshTransactionEncoder_getClauseValue(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	tests := []struct {
		name     string
		clause   ClauseData
		expected *big.Int
	}{
		{
			name: "zero value",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				createTestAddress(meshtests.TestAddress1),
				big.NewInt(0),
				"0x",
			)},
			expected: big.NewInt(0),
		},
		{
			name: "positive value",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				createTestAddress(meshtests.TestAddress1),
				big.NewInt(1000000000000000000),
				"0x",
			)},
			expected: big.NewInt(1000000000000000000),
		},
		{
			name: "large value",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				createTestAddress(meshtests.TestAddress1),
				func() *big.Int { val, _ := big.NewInt(0).SetString("1000000000000000000000000", 10); return val }(),
				"0x",
			)},
			expected: func() *big.Int { val, _ := big.NewInt(0).SetString("1000000000000000000000000", 10); return val }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.getClauseValue(tt.clause)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("getClauseValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test isVIP180Transfer
func TestMeshTransactionEncoder_isVIP180Transfer(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	tests := []struct {
		name     string
		clause   ClauseData
		value    *big.Int
		expected bool
	}{
		{
			name: "VIP180 transfer",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				createTestAddress(meshtests.TestAddress1),
				big.NewInt(0),
				"0xa9059cbb00000000000000000000000016277a1ff38678291c41d1820957c78bb5da59ce0000000000000000000000000000000000000000000000000de0b6b3a7640000",
			)},
			value:    big.NewInt(0),
			expected: true,
		},
		{
			name: "non-VIP180 with value",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				createTestAddress(meshtests.TestAddress1),
				big.NewInt(1000000000000000000),
				"0x1234",
			)},
			value:    big.NewInt(1000000000000000000),
			expected: false,
		},
		{
			name: "no data",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				createTestAddress(meshtests.TestAddress1),
				big.NewInt(0),
				"0x",
			)},
			value:    big.NewInt(0),
			expected: false,
		},
		{
			name: "nil to address",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				nil,
				big.NewInt(0),
				"0xa9059cbb00000000000000000000000016277a1ff38678291c41d1820957c78bb5da59ce0000000000000000000000000000000000000000000000000de0b6b3a7640000",
			)},
			value:    big.NewInt(0),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isVIP180Transfer(tt.clause, tt.value)
			if result != tt.expected {
				t.Errorf("isVIP180Transfer() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test hasContractInteraction
func TestMeshTransactionEncoder_hasContractInteraction(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	tests := []struct {
		name     string
		clause   ClauseData
		expected bool
	}{
		{
			name: "has data",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				createTestAddress(meshtests.TestAddress1),
				big.NewInt(0),
				"0x1234",
			)},
			expected: true,
		},
		{
			name: "has to address",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				createTestAddress(meshtests.TestAddress1),
				big.NewInt(0),
				"0x",
			)},
			expected: true,
		},
		{
			name: "no data and nil to",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				nil,
				big.NewInt(0),
				"",
			)},
			expected: false,
		},
		{
			name: "zero to address",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				createTestAddress("0x0000000000000000000000000000000000000000"),
				big.NewInt(0),
				"",
			)},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.hasContractInteraction(tt.clause)
			if result != tt.expected {
				t.Errorf("hasContractInteraction() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test createTransferOperation
func TestMeshTransactionEncoder_createTransferOperation(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	tests := []struct {
		name         string
		index        int
		networkIndex *int64
		address      string
		amount       string
		currency     *types.Currency
		clauseIndex  int
		status       *string
	}{
		{
			name:         "basic transfer",
			index:        0,
			networkIndex: nil,
			address:      meshtests.TestAddress1,
			amount:       "1000000000000000000",
			currency:     meshcommon.VETCurrency,
			clauseIndex:  0,
			status:       &testStatus,
		},
		{
			name:         "with network index",
			index:        1,
			networkIndex: func() *int64 { i := int64(0); return &i }(),
			address:      meshtests.FirstSoloAddress,
			amount:       "-500000000000000000",
			currency:     meshcommon.VTHOCurrency,
			clauseIndex:  1,
			status:       &testStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation := parser.createTransferOperation(tt.index, tt.networkIndex, tt.address, tt.amount, tt.currency, tt.clauseIndex, tt.status)

			if operation.OperationIdentifier.Index != int64(tt.index) {
				t.Errorf("createTransferOperation() index = %v, want %v", operation.OperationIdentifier.Index, tt.index)
			}
			if operation.OperationIdentifier.NetworkIndex != tt.networkIndex {
				t.Errorf("createTransferOperation() networkIndex = %v, want %v", operation.OperationIdentifier.NetworkIndex, tt.networkIndex)
			}
			if operation.Type != meshcommon.OperationTypeTransfer {
				t.Errorf("createTransferOperation() type = %v, want %v", operation.Type, meshcommon.OperationTypeTransfer)
			}
			if operation.Status != tt.status {
				t.Errorf("createTransferOperation() status = %v, want %v", operation.Status, tt.status)
			}
			if operation.Account.Address != tt.address {
				t.Errorf("createTransferOperation() address = %v, want %v", operation.Account.Address, tt.address)
			}
			if operation.Amount.Value != tt.amount {
				t.Errorf("createTransferOperation() amount = %v, want %v", operation.Amount.Value, tt.amount)
			}
			if operation.Amount.Currency != tt.currency {
				t.Errorf("createTransferOperation() currency = %v, want %v", operation.Amount.Currency, tt.currency)
			}
			if operation.Metadata["clauseIndex"] != tt.clauseIndex {
				t.Errorf("createTransferOperation() clauseIndex = %v, want %v", operation.Metadata["clauseIndex"], tt.clauseIndex)
			}
		})
	}
}

// Test createContractInteractionOperation
func TestMeshTransactionEncoder_createContractInteractionOperation(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	tests := []struct {
		name           string
		clause         ClauseData
		clauseIndex    int
		operationIndex int
		originAddr     string
		status         *string
	}{
		{
			name: "with to address and data",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				createTestAddress(meshtests.TestAddress1),
				big.NewInt(0),
				"0x1234",
			)},
			clauseIndex:    0,
			operationIndex: 0,
			originAddr:     meshtests.FirstSoloAddress,
			status:         &testStatus,
		},
		{
			name: "nil to address",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				nil,
				big.NewInt(0),
				"0x5678",
			)},
			clauseIndex:    1,
			operationIndex: 1,
			originAddr:     meshtests.FirstSoloAddress,
			status:         &testStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation := parser.createContractInteractionOperation(tt.clause, tt.clauseIndex, tt.operationIndex, tt.originAddr, tt.status)

			if operation.OperationIdentifier.Index != int64(tt.operationIndex) {
				t.Errorf("createContractInteractionOperation() index = %v, want %v", operation.OperationIdentifier.Index, tt.operationIndex)
			}
			if operation.Type != meshcommon.OperationTypeContractCall {
				t.Errorf("createContractInteractionOperation() type = %v, want %v", operation.Type, meshcommon.OperationTypeContractCall)
			}
			if operation.Status != tt.status {
				t.Errorf("createContractInteractionOperation() status = %v, want %v", operation.Status, tt.status)
			}
			if operation.Account.Address != tt.originAddr {
				t.Errorf("createContractInteractionOperation() address = %v, want %v", operation.Account.Address, tt.originAddr)
			}
			if operation.Amount.Value != "0" {
				t.Errorf("createContractInteractionOperation() amount = %v, want 0", operation.Amount.Value)
			}
			if operation.Amount.Currency != meshcommon.VETCurrency {
				t.Errorf("createContractInteractionOperation() currency = %v, want %v", operation.Amount.Currency, meshcommon.VETCurrency)
			}
			if operation.Metadata["clauseIndex"] != tt.clauseIndex {
				t.Errorf("createContractInteractionOperation() clauseIndex = %v, want %v", operation.Metadata["clauseIndex"], tt.clauseIndex)
			}
		})
	}
}

// Test createEnergyTransferOperation
func TestMeshTransactionEncoder_createEnergyTransferOperation(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	tests := []struct {
		name           string
		operationIndex int
		originAddr     string
		gas            uint64
		status         *string
	}{
		{
			name:           "basic energy transfer",
			operationIndex: 0,
			originAddr:     meshtests.FirstSoloAddress,
			gas:            21000,
			status:         &testStatus,
		},
		{
			name:           "zero gas",
			operationIndex: 1,
			originAddr:     meshtests.TestAddress1,
			gas:            0,
			status:         &testStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation := parser.createEnergyTransferOperation(tt.operationIndex, tt.originAddr, tt.gas, tt.status)

			if operation.OperationIdentifier.Index != int64(tt.operationIndex) {
				t.Errorf("createEnergyTransferOperation() index = %v, want %v", operation.OperationIdentifier.Index, tt.operationIndex)
			}
			if operation.Type != meshcommon.OperationTypeFee {
				t.Errorf("createEnergyTransferOperation() type = %v, want %v", operation.Type, meshcommon.OperationTypeFee)
			}
			if operation.Status != tt.status {
				t.Errorf("createEnergyTransferOperation() status = %v, want %v", operation.Status, tt.status)
			}
			if operation.Account.Address != tt.originAddr {
				t.Errorf("createEnergyTransferOperation() address = %v, want %v", operation.Account.Address, tt.originAddr)
			}
			expectedValue := "-" + strconv.FormatUint(tt.gas, 10)
			if operation.Amount.Value != expectedValue {
				t.Errorf("createEnergyTransferOperation() amount = %v, want %v", operation.Amount.Value, expectedValue)
			}
			if operation.Amount.Currency != meshcommon.VTHOCurrency {
				t.Errorf("createEnergyTransferOperation() currency = %v, want %v", operation.Amount.Currency, meshcommon.VTHOCurrency)
			}
			if operation.Metadata["gas"] != strconv.FormatUint(tt.gas, 10) {
				t.Errorf("createEnergyTransferOperation() gas = %v, want %v", operation.Metadata["gas"], tt.gas)
			}
		})
	}
}

// Test parseVETTransfer
func TestMeshTransactionEncoder_parseVETTransfer(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	tests := []struct {
		name              string
		clause            ClauseData
		clauseIndex       int
		operationIndex    int
		originAddr        string
		value             *big.Int
		status            *string
		expectedOps       int
		expectedNextIndex int
	}{
		{
			name: "transfer with recipient",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				createTestAddress(meshtests.TestAddress1),
				big.NewInt(1000000000000000000),
				"0x",
			)},
			clauseIndex:       0,
			operationIndex:    0,
			originAddr:        meshtests.FirstSoloAddress,
			value:             big.NewInt(1000000000000000000),
			status:            &testStatus,
			expectedOps:       2,
			expectedNextIndex: 2,
		},
		{
			name: "transfer without recipient",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				nil,
				big.NewInt(500000000000000000),
				"0x",
			)},
			clauseIndex:       1,
			operationIndex:    0,
			originAddr:        meshtests.FirstSoloAddress,
			value:             big.NewInt(500000000000000000),
			status:            &testStatus,
			expectedOps:       1,
			expectedNextIndex: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operations, nextIndex := parser.parseVETTransfer(tt.clause, tt.clauseIndex, tt.operationIndex, tt.originAddr, tt.value, tt.status)

			if len(operations) != tt.expectedOps {
				t.Errorf("parseVETTransfer() operations length = %v, want %v", len(operations), tt.expectedOps)
			}
			if nextIndex != tt.expectedNextIndex {
				t.Errorf("parseVETTransfer() nextIndex = %v, want %v", nextIndex, tt.expectedNextIndex)
			}

			// Check first operation (sender)
			if len(operations) > 0 {
				senderOp := operations[0]
				if senderOp.Amount.Value != "-"+tt.value.String() {
					t.Errorf("parseVETTransfer() sender amount = %v, want %v", senderOp.Amount.Value, "-"+tt.value.String())
				}
				if senderOp.Account.Address != tt.originAddr {
					t.Errorf("parseVETTransfer() sender address = %v, want %v", senderOp.Account.Address, tt.originAddr)
				}
			}

			// Check second operation (receiver) if exists
			if len(operations) > 1 {
				receiverOp := operations[1]
				if receiverOp.Amount.Value != tt.value.String() {
					t.Errorf("parseVETTransfer() receiver amount = %v, want %v", receiverOp.Amount.Value, tt.value.String())
				}
				if tt.clause.GetTo() != nil && receiverOp.Account.Address != tt.clause.GetTo().String() {
					t.Errorf("parseVETTransfer() receiver address = %v, want %v", receiverOp.Account.Address, tt.clause.GetTo().String())
				}
			}
		})
	}
}

// Test parseTransactionOperationsFromClauses
func TestMeshTransactionEncoder_parseTransactionOperationsFromClauses(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	tests := []struct {
		name        string
		clauses     []*api.JSONClause
		originAddr  string
		gas         uint64
		status      *string
		expectedOps int
	}{
		{
			name:        "empty clauses",
			clauses:     []*api.JSONClause{},
			originAddr:  meshtests.FirstSoloAddress,
			gas:         0,
			status:      &testStatus,
			expectedOps: 0,
		},
		{
			name: "single VET transfer",
			clauses: []*api.JSONClause{
				createTestJSONClause(
					createTestAddress(meshtests.TestAddress1),
					big.NewInt(1000000000000000000),
					"0x",
				),
			},
			originAddr:  meshtests.FirstSoloAddress,
			gas:         0,
			status:      &testStatus,
			expectedOps: 3,
		},
		{
			name: "VET transfer with gas",
			clauses: []*api.JSONClause{
				createTestJSONClause(
					createTestAddress(meshtests.TestAddress1),
					big.NewInt(1000000000000000000),
					"0x",
				),
			},
			originAddr:  meshtests.FirstSoloAddress,
			gas:         21000,
			status:      &testStatus,
			expectedOps: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operations := parser.ParseTransactionOperationsFromJSONClauses(tt.clauses, tt.originAddr, tt.gas, tt.status)

			if len(operations) != tt.expectedOps {
				t.Errorf("parseTransactionOperationsFromClauses() operations length = %v, want %v", len(operations), tt.expectedOps)
			}
		})
	}
}

// Test ParseTransactionOperationsFromTransactionClauses
func TestMeshTransactionEncoder_ParseTransactionOperationsFromTransactionClauses(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	tests := []struct {
		name        string
		clauses     api.Clauses
		originAddr  string
		gas         uint64
		status      *string
		expectedOps int
	}{
		{
			name:        "empty clauses",
			clauses:     api.Clauses{},
			originAddr:  meshtests.FirstSoloAddress,
			gas:         0,
			status:      &testStatus,
			expectedOps: 0,
		},
		{
			name: "single VET transfer",
			clauses: api.Clauses{
				createTestClause(
					createTestAddress(meshtests.TestAddress1),
					big.NewInt(1000000000000000000),
					"0x",
				),
			},
			originAddr:  meshtests.FirstSoloAddress,
			gas:         0,
			status:      &testStatus,
			expectedOps: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operations := parser.ParseOperationsFromAPIClauses(tt.clauses, tt.originAddr, tt.gas, tt.status)

			if len(operations) != tt.expectedOps {
				t.Errorf("ParseTransactionOperationsFromTransactionClauses() operations length = %v, want %v", len(operations), tt.expectedOps)
			}
		})
	}
}

func TestClauseParser_ParseClausesFromOptions(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	tests := []struct {
		name          string
		clausesRaw    any
		expectError   bool
		expectedCount int
		validateFunc  func(*testing.T, []*tx.Clause)
	}{
		{
			name:          "empty clauses array",
			clausesRaw:    []any{},
			expectError:   false,
			expectedCount: 0,
		},
		{
			name: "single valid clause with all fields",
			clausesRaw: []any{
				map[string]any{
					"to":    meshtests.TestAddress1,
					"value": "1000000000000000000",
					"data":  "0x",
				},
			},
			expectError:   false,
			expectedCount: 1,
			validateFunc: func(t *testing.T, clauses []*tx.Clause) {
				if clauses[0].To() == nil {
					t.Error("Expected non-nil 'to' address")
				}
				if clauses[0].Value().Cmp(big.NewInt(1000000000000000000)) != 0 {
					t.Errorf("Expected value 1000000000000000000, got %v", clauses[0].Value())
				}
				if len(clauses[0].Data()) != 0 {
					t.Errorf("Expected empty data, got %d bytes", len(clauses[0].Data()))
				}
			},
		},
		{
			name: "clause with hex data",
			clausesRaw: []any{
				map[string]any{
					"to":    meshtests.TestAddress1,
					"value": "0",
					"data":  "0xa9059cbb0000000000000000000000001234567890123456789012345678901234567890000000000000000000000000000000000000000000000000000000000000000a",
				},
			},
			expectError:   false,
			expectedCount: 1,
			validateFunc: func(t *testing.T, clauses []*tx.Clause) {
				if len(clauses[0].Data()) == 0 {
					t.Error("Expected non-empty data")
				}
			},
		},
		{
			name: "clause without 'to' (contract creation)",
			clausesRaw: []any{
				map[string]any{
					"value": "0",
					"data":  "0x608060405234801561001057600080fd5b50",
				},
			},
			expectError:   false,
			expectedCount: 1,
			validateFunc: func(t *testing.T, clauses []*tx.Clause) {
				if clauses[0].To() != nil {
					t.Error("Expected nil 'to' address for contract creation")
				}
			},
		},
		{
			name: "multiple clauses",
			clausesRaw: []any{
				map[string]any{
					"to":    meshtests.TestAddress1,
					"value": "1000000000000000000",
					"data":  "0x",
				},
				map[string]any{
					"to":    meshtests.FirstSoloAddress,
					"value": "2000000000000000000",
					"data":  "0x",
				},
			},
			expectError:   false,
			expectedCount: 2,
		},
		{
			name: "clause with zero value",
			clausesRaw: []any{
				map[string]any{
					"to":    meshtests.TestAddress1,
					"value": "0",
					"data":  "0x",
				},
			},
			expectError:   false,
			expectedCount: 1,
			validateFunc: func(t *testing.T, clauses []*tx.Clause) {
				if clauses[0].Value().Cmp(big.NewInt(0)) != 0 {
					t.Errorf("Expected value 0, got %v", clauses[0].Value())
				}
			},
		},
		{
			name:        "invalid input - not an array",
			clausesRaw:  "invalid",
			expectError: true,
		},
		{
			name: "invalid input - clause not an object",
			clausesRaw: []any{
				"invalid",
			},
			expectError: true,
		},
		{
			name: "invalid 'to' address",
			clausesRaw: []any{
				map[string]any{
					"to":    "invalid_address",
					"value": "0",
					"data":  "0x",
				},
			},
			expectError: true,
		},
		{
			name: "invalid 'value'",
			clausesRaw: []any{
				map[string]any{
					"to":    meshtests.TestAddress1,
					"value": "not_a_number",
					"data":  "0x",
				},
			},
			expectError: true,
		},
		{
			name: "invalid 'data' - odd length hex",
			clausesRaw: []any{
				map[string]any{
					"to":    meshtests.TestAddress1,
					"value": "0",
					"data":  "0x123",
				},
			},
			expectError: true,
		},
		{
			name: "invalid 'data' - non-hex characters",
			clausesRaw: []any{
				map[string]any{
					"to":    meshtests.TestAddress1,
					"value": "0",
					"data":  "0xZZZZ",
				},
			},
			expectError: true,
		},
		{
			name: "clause with large value",
			clausesRaw: []any{
				map[string]any{
					"to":    meshtests.TestAddress1,
					"value": "1000000000000000000000000000",
					"data":  "0x",
				},
			},
			expectError:   false,
			expectedCount: 1,
			validateFunc: func(t *testing.T, clauses []*tx.Clause) {
				expected := new(big.Int)
				expected.SetString("1000000000000000000000000000", 10)
				if clauses[0].Value().Cmp(expected) != 0 {
					t.Errorf("Expected large value, got %v", clauses[0].Value())
				}
			},
		},
		{
			name: "clause with hex value format",
			clausesRaw: []any{
				map[string]any{
					"to":    meshtests.TestAddress1,
					"value": "0x0de0b6b3a7640000",
					"data":  "0x",
				},
			},
			expectError:   false,
			expectedCount: 1,
			validateFunc: func(t *testing.T, clauses []*tx.Clause) {
				expected := big.NewInt(1000000000000000000)
				if clauses[0].Value().Cmp(expected) != 0 {
					t.Errorf("Expected hex value parsed correctly, got %v", clauses[0].Value())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clauses, err := parser.ParseClausesFromOptions(tt.clausesRaw)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(clauses) != tt.expectedCount {
				t.Errorf("Expected %d clauses, got %d", tt.expectedCount, len(clauses))
				return
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, clauses)
			}
		})
	}
}
