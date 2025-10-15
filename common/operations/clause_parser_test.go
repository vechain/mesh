package operations

import (
	"fmt"
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

// Mock ClauseValue that returns error on MarshalText
type errorClauseValue struct{}

func (e errorClauseValue) MarshalText() ([]byte, error) {
	return nil, fmt.Errorf("mock marshal error")
}

// Mock ClauseData with error ClauseValue
type errorClauseData struct {
	to    *thor.Address
	data  string
	value ClauseValue
}

func (e errorClauseData) GetValue() ClauseValue {
	return e.value
}

func (e errorClauseData) GetTo() *thor.Address {
	return e.to
}

func (e errorClauseData) GetData() string {
	return e.data
}

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
			hasValueTransfer, hasContractInteraction, hasEnergyTransfer, err := parser.analyzeClauses(tt.clauseData, tt.gas)
			if err != nil {
				t.Errorf("analyzeClauses() error = %v", err)
			}

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

func TestMeshTransactionEncoder_createEnergyTransferOperation(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	tests := []struct {
		name           string
		operationIndex int
		originAddr     string
		delegatorAddr  string
		gas            uint64
		status         *string
		expectedType   string
	}{
		{
			name:           "basic energy transfer without delegation",
			operationIndex: 0,
			originAddr:     meshtests.FirstSoloAddress,
			delegatorAddr:  "",
			gas:            21000,
			status:         &testStatus,
			expectedType:   meshcommon.OperationTypeFee,
		},
		{
			name:           "energy transfer with fee delegation",
			operationIndex: 1,
			originAddr:     meshtests.TestAddress1,
			delegatorAddr:  meshtests.FirstSoloAddress,
			gas:            25000,
			status:         &testStatus,
			expectedType:   meshcommon.OperationTypeFeeDelegation,
		},
		{
			name:           "zero gas without delegation",
			operationIndex: 2,
			originAddr:     meshtests.TestAddress1,
			delegatorAddr:  "",
			gas:            0,
			status:         &testStatus,
			expectedType:   meshcommon.OperationTypeFee,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation := parser.createEnergyTransferOperation(tt.operationIndex, tt.originAddr, tt.delegatorAddr, tt.gas, tt.status)

			if operation.OperationIdentifier.Index != int64(tt.operationIndex) {
				t.Errorf("createEnergyTransferOperation() index = %v, want %v", operation.OperationIdentifier.Index, tt.operationIndex)
			}
			if operation.Type != tt.expectedType {
				t.Errorf("createEnergyTransferOperation() type = %v, want %v", operation.Type, tt.expectedType)
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

			// Check fee_delegator_account in metadata for delegation
			if tt.delegatorAddr != "" {
				delegatorInMetadata, ok := operation.Metadata["fee_delegator_account"].(string)
				if !ok {
					t.Errorf("createEnergyTransferOperation() fee_delegator_account not found in metadata")
				}
				if delegatorInMetadata != tt.delegatorAddr {
					t.Errorf("createEnergyTransferOperation() fee_delegator_account = %v, want %v", delegatorInMetadata, tt.delegatorAddr)
				}
			} else {
				if _, exists := operation.Metadata["fee_delegator_account"]; exists {
					t.Errorf("createEnergyTransferOperation() fee_delegator_account should not exist in metadata without delegation")
				}
			}
		})
	}
}

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
			operations, err := parser.ParseTransactionOperationsFromJSONClauses(tt.clauses, tt.originAddr, "", tt.gas, tt.status)
			if err != nil {
				t.Errorf("ParseTransactionOperationsFromJSONClauses() error = %v", err)
			}

			if len(operations) != tt.expectedOps {
				t.Errorf("parseTransactionOperationsFromClauses() operations length = %v, want %v", len(operations), tt.expectedOps)
			}
		})
	}
}

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
			operations, err := parser.ParseOperationsFromAPIClauses(tt.clauses, tt.originAddr, "", tt.gas, tt.status)
			if err != nil {
				t.Errorf("ParseOperationsFromAPIClauses() error = %v", err)
			}

			if len(operations) != tt.expectedOps {
				t.Errorf("ParseTransactionOperationsFromTransactionClauses() operations length = %v, want %v", len(operations), tt.expectedOps)
			}
		})
	}
}

func TestClauseParser_ParseOperationsWithFeeDelegation(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	tests := []struct {
		name          string
		clauses       api.Clauses
		originAddr    string
		delegatorAddr string
		gas           uint64
		status        *string
		expectedOps   int
		validateFunc  func(*testing.T, []*types.Operation)
	}{
		{
			name: "VET transfer with fee delegation",
			clauses: api.Clauses{
				createTestClause(
					createTestAddress(meshtests.TestAddress1),
					big.NewInt(1000000000000000000),
					"",
				),
			},
			originAddr:    meshtests.FirstSoloAddress,
			delegatorAddr: "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
			gas:           21000,
			status:        &testStatus,
			expectedOps:   3, // sender, receiver, fee delegation
			validateFunc: func(t *testing.T, ops []*types.Operation) {
				// Check that last operation is fee delegation
				if len(ops) < 3 {
					t.Fatalf("Expected at least 3 operations, got %d", len(ops))
				}
				feeOp := ops[len(ops)-1]
				if feeOp.Type != meshcommon.OperationTypeFeeDelegation {
					t.Errorf("Expected OperationTypeFeeDelegation, got %s", feeOp.Type)
				}
				// Check account is origin
				if feeOp.Account.Address != meshtests.FirstSoloAddress {
					t.Errorf("Expected account to be origin %s, got %s", meshtests.FirstSoloAddress, feeOp.Account.Address)
				}
				// Check delegator in metadata
				delegator, ok := feeOp.Metadata["fee_delegator_account"].(string)
				if !ok {
					t.Error("Expected fee_delegator_account in metadata")
				}
				if delegator != "0xf077b491b355e64048ce21e3a6fc4751eeea77fa" {
					t.Errorf("Expected delegator 0xf077b491b355e64048ce21e3a6fc4751eeea77fa, got %s", delegator)
				}
			},
		},
		{
			name: "VET transfer without fee delegation",
			clauses: api.Clauses{
				createTestClause(
					createTestAddress(meshtests.TestAddress1),
					big.NewInt(1000000000000000000),
					"",
				),
			},
			originAddr:    meshtests.FirstSoloAddress,
			delegatorAddr: "",
			gas:           21000,
			status:        &testStatus,
			expectedOps:   3, // sender, receiver, fee
			validateFunc: func(t *testing.T, ops []*types.Operation) {
				// Check that last operation is regular fee
				if len(ops) < 3 {
					t.Fatalf("Expected at least 3 operations, got %d", len(ops))
				}
				feeOp := ops[len(ops)-1]
				if feeOp.Type != meshcommon.OperationTypeFee {
					t.Errorf("Expected OperationTypeFee, got %s", feeOp.Type)
				}
				// Check account is origin
				if feeOp.Account.Address != meshtests.FirstSoloAddress {
					t.Errorf("Expected account to be origin %s, got %s", meshtests.FirstSoloAddress, feeOp.Account.Address)
				}
				// Check no delegator in metadata
				if _, exists := feeOp.Metadata["fee_delegator_account"]; exists {
					t.Error("fee_delegator_account should not exist in metadata without delegation")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operations, err := parser.ParseOperationsFromAPIClauses(tt.clauses, tt.originAddr, tt.delegatorAddr, tt.gas, tt.status)
			if err != nil {
				t.Errorf("ParseOperationsFromAPIClauses() error = %v", err)
			}

			if len(operations) != tt.expectedOps {
				t.Errorf("ParseOperationsFromAPIClauses() operations length = %v, want %v", len(operations), tt.expectedOps)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, operations)
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
func TestParseHexOrDecimal256(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	tests := []struct {
		name        string
		input       string
		expectError bool
		expectedStr string
	}{
		{
			name:        "decimal value",
			input:       "1000000000000000000",
			expectError: false,
			expectedStr: "1000000000000000000",
		},
		{
			name:        "hex value with 0x prefix",
			input:       "0xde0b6b3a7640000",
			expectError: false,
			expectedStr: "1000000000000000000",
		},
		{
			name:        "zero value",
			input:       "0",
			expectError: false,
			expectedStr: "0",
		},
		{
			name:        "large value",
			input:       "1000000000000000000000000",
			expectError: false,
			expectedStr: "1000000000000000000000000",
		},
		{
			name:        "invalid value - empty string",
			input:       "",
			expectError: false,
		},
		{
			name:        "invalid value - non-numeric",
			input:       "abc",
			expectError: true,
		},
		{
			name:        "invalid value - invalid hex",
			input:       "0xzzzz",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseHexOrDecimal256(tt.input)

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

			if result == nil {
				t.Error("Expected non-nil result")
				return
			}

			// Convert to big.Int for comparison
			bigIntResult := (*big.Int)(result)
			expectedBigInt := new(big.Int)
			expectedBigInt.SetString(tt.expectedStr, 10)

			if bigIntResult.Cmp(expectedBigInt) != 0 {
				t.Errorf("Expected value %s, got %s", tt.expectedStr, bigIntResult.String())
			}
		})
	}
}

func TestParseAPIClause(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	tests := []struct {
		name        string
		input       map[string]any
		expectError bool
		validate    func(t *testing.T, clause *api.Clause)
	}{
		{
			name: "valid clause with all fields",
			input: map[string]any{
				"to":    "0x7567d83b7b8d80addcb281a71d54fc7b3364ffed",
				"value": "1000000000000000000",
				"data":  "0xa9059cbb",
			},
			expectError: false,
			validate: func(t *testing.T, clause *api.Clause) {
				if clause.To == nil {
					t.Error("Expected non-nil 'to' address")
				}
				if clause.To.String() != "0x7567d83b7b8d80addcb281a71d54fc7b3364ffed" {
					t.Errorf("Expected specific 'to' address, got %s", clause.To.String())
				}
				if clause.Value == nil {
					t.Error("Expected non-nil value")
				}
				if clause.Data != "0xa9059cbb" {
					t.Errorf("Expected data 0xa9059cbb, got %s", clause.Data)
				}
			},
		},
		{
			name: "valid clause with nil 'to' (contract creation)",
			input: map[string]any{
				"to":    nil,
				"value": "0",
				"data":  "0x608060",
			},
			expectError: false,
			validate: func(t *testing.T, clause *api.Clause) {
				if clause.To != nil {
					t.Error("Expected nil 'to' address for contract creation")
				}
			},
		},
		{
			name: "valid clause with zero value",
			input: map[string]any{
				"to":    "0x7567d83b7b8d80addcb281a71d54fc7b3364ffed",
				"value": "0",
				"data":  "0x",
			},
			expectError: false,
			validate: func(t *testing.T, clause *api.Clause) {
				bigIntValue := (*big.Int)(clause.Value)
				if bigIntValue.Cmp(big.NewInt(0)) != 0 {
					t.Errorf("Expected zero value, got %s", bigIntValue.String())
				}
			},
		},
		{
			name: "invalid - 'to' is not a string",
			input: map[string]any{
				"to":    123,
				"value": "0",
				"data":  "0x",
			},
			expectError: true,
		},
		{
			name: "invalid - 'to' address is invalid",
			input: map[string]any{
				"to":    "0xinvalid",
				"value": "0",
				"data":  "0x",
			},
			expectError: true,
		},
		{
			name: "invalid - missing 'value'",
			input: map[string]any{
				"to":   "0x7567d83b7b8d80addcb281a71d54fc7b3364ffed",
				"data": "0x",
			},
			expectError: true,
		},
		{
			name: "invalid - 'value' is not a string",
			input: map[string]any{
				"to":    "0x7567d83b7b8d80addcb281a71d54fc7b3364ffed",
				"value": 123,
				"data":  "0x",
			},
			expectError: true,
		},
		{
			name: "invalid - 'value' is invalid number",
			input: map[string]any{
				"to":    "0x7567d83b7b8d80addcb281a71d54fc7b3364ffed",
				"value": "invalid",
				"data":  "0x",
			},
			expectError: true,
		},
		{
			name: "invalid - missing 'data'",
			input: map[string]any{
				"to":    "0x7567d83b7b8d80addcb281a71d54fc7b3364ffed",
				"value": "0",
			},
			expectError: true,
		},
		{
			name: "invalid - 'data' is not a string",
			input: map[string]any{
				"to":    "0x7567d83b7b8d80addcb281a71d54fc7b3364ffed",
				"value": "0",
				"data":  123,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseAPIClause(tt.input)

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

			if result == nil {
				t.Error("Expected non-nil result")
				return
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestClauseParser_ErrorHandlingInGetClauseValue(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	// Create a clause with a value that will fail to marshal
	errorClause := errorClauseData{
		to:    createTestAddress(meshtests.TestAddress1),
		data:  "0x",
		value: errorClauseValue{},
	}

	t.Run("error in getClauseValue propagates in ParseTransactionOperationsFromClauseData", func(t *testing.T) {
		clauseData := []ClauseData{errorClause}
		operations, err := parser.ParseTransactionOperationsFromClauseData(clauseData, meshtests.FirstSoloAddress, "", 0, &testStatus)

		if err == nil {
			t.Error("Expected error from MarshalText failure but got none")
		}

		if operations != nil {
			t.Errorf("Expected nil operations on error, got %v", operations)
		}

		if err.Error() != "failed to marshal clause value: mock marshal error" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("error in getClauseValue propagates in analyzeClauses", func(t *testing.T) {
		clauseData := []ClauseData{errorClause}
		hasValue, hasContract, hasEnergy, err := parser.analyzeClauses(clauseData, 0)

		if err == nil {
			t.Error("Expected error from MarshalText failure but got none")
		}

		if hasValue || hasContract || hasEnergy {
			t.Error("Expected all flags to be false on error")
		}

		if err.Error() != "failed to marshal clause value: mock marshal error" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})
}

func TestClauseParser_ErrorHandlingInAnalyzeClauses(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	errorClause := errorClauseData{
		to:    createTestAddress(meshtests.TestAddress1),
		data:  "0x1234",
		value: errorClauseValue{},
	}

	t.Run("error in analyzeClauses stops processing", func(t *testing.T) {
		// Mix valid and invalid clauses
		clauseData := []ClauseData{
			JSONClauseAdapter{Clause: createTestJSONClause(
				createTestAddress(meshtests.TestAddress1),
				big.NewInt(1000000000000000000),
				"0x",
			)},
			errorClause, // This should cause error
		}

		operations, err := parser.ParseTransactionOperationsFromClauseData(clauseData, meshtests.FirstSoloAddress, "", 0, &testStatus)

		if err == nil {
			t.Error("Expected error but got none")
		}

		if operations != nil {
			t.Errorf("Expected nil operations on error, got %v operations", len(operations))
		}
	})
}

func TestClauseParser_ParseTransactionOperationsFromJSONClauses_ErrorHandling(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	t.Run("empty clauses returns empty operations without error", func(t *testing.T) {
		operations, err := parser.ParseTransactionOperationsFromJSONClauses([]*api.JSONClause{}, meshtests.FirstSoloAddress, "", 0, &testStatus)

		if err != nil {
			t.Errorf("Expected no error for empty clauses, got: %v", err)
		}

		if len(operations) != 0 {
			t.Errorf("Expected 0 operations for empty clauses, got %v", len(operations))
		}
	})
}

func TestClauseParser_ParseOperationsFromAPIClauses_ErrorHandling(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	t.Run("empty API clauses returns empty operations without error", func(t *testing.T) {
		operations, err := parser.ParseOperationsFromAPIClauses(api.Clauses{}, meshtests.FirstSoloAddress, "", 0, &testStatus)

		if err != nil {
			t.Errorf("Expected no error for empty clauses, got: %v", err)
		}

		if len(operations) != 0 {
			t.Errorf("Expected 0 operations for empty clauses, got %v", len(operations))
		}
	})
}

func TestClauseParser_GetClauseValue_DirectTest(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	parser := NewClauseParser(mockClient, NewOperationsExtractor())

	tests := []struct {
		name        string
		clause      ClauseData
		expectError bool
		expectedVal string
	}{
		{
			name: "valid clause with positive value",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				createTestAddress(meshtests.TestAddress1),
				big.NewInt(1000000000000000000),
				"0x",
			)},
			expectError: false,
			expectedVal: "1000000000000000000",
		},
		{
			name: "valid clause with zero value",
			clause: JSONClauseAdapter{Clause: createTestJSONClause(
				createTestAddress(meshtests.TestAddress1),
				big.NewInt(0),
				"0x",
			)},
			expectError: false,
			expectedVal: "0",
		},
		{
			name: "invalid clause with marshal error",
			clause: errorClauseData{
				to:    createTestAddress(meshtests.TestAddress1),
				data:  "0x",
				value: errorClauseValue{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := parser.getClauseValue(tt.clause)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if value.String() != tt.expectedVal {
				t.Errorf("Expected value %s, got %s", tt.expectedVal, value.String())
			}
		})
	}
}
