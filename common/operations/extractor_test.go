package operations

import (
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshthor "github.com/vechain/mesh/thor"

	meshtests "github.com/vechain/mesh/tests"
)

func TestGetStringFromOptions(t *testing.T) {
	extractor := NewOperationsExtractor()
	options := map[string]any{
		"key1": "value1",
		"key2": 123,
		"key3": nil,
	}

	// Test existing string value
	result := extractor.GetStringFromOptions(options, "key1")
	if result != "value1" {
		t.Errorf("Expected 'value1', got %s", result)
	}

	// Test non-existing key
	result = extractor.GetStringFromOptions(options, "nonexistent")
	if result != meshcommon.TransactionTypeDynamic {
		t.Errorf("Expected 'dynamic', got %s", result)
	}

	// Test non-string value
	result = extractor.GetStringFromOptions(options, "key2")
	if result != meshcommon.TransactionTypeDynamic {
		t.Errorf("Expected 'dynamic' for non-string value, got %s", result)
	}

	// Test nil value
	result = extractor.GetStringFromOptions(options, "key3")
	if result != meshcommon.TransactionTypeDynamic {
		t.Errorf("Expected 'dynamic' for nil value, got %s", result)
	}
}

func TestGetTxOrigins(t *testing.T) {
	operations := []*types.Operation{
		{
			Type: meshcommon.OperationTypeFee,
			Account: &types.AccountIdentifier{
				Address: "0x1234567890123456789012345678901234567890",
			},
		},
		{
			Type: meshcommon.OperationTypeTransfer,
			Account: &types.AccountIdentifier{
				Address: "0x0987654321098765432109876543210987654321",
			},
			Amount: &types.Amount{
				Value: "-1000000000000000000", // Negative value for sending
			},
		},
	}

	extractor := NewOperationsExtractor()
	origins := extractor.GetTxOrigins(operations)

	if len(origins) != 2 {
		t.Errorf("Expected 2 origins, got %d", len(origins))
		return // Don't continue if we don't have the expected number of origins
	}

	// Function converts to lowercase
	expected1 := "0x1234567890123456789012345678901234567890"
	expected2 := "0x0987654321098765432109876543210987654321"

	if origins[0] != expected1 {
		t.Errorf("Expected first origin to be %s, got %s", expected1, origins[0])
	}

	if origins[1] != expected2 {
		t.Errorf("Expected second origin to be %s, got %s", expected2, origins[1])
	}
}

func TestGetTokenCurrencyFromContractAddressWithClient(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()

	tests := []struct {
		name             string
		contractAddr     string
		expectedSymbol   string
		expectedDecimals int32
		expectError      bool
	}{
		{
			name:             "VTHO token",
			contractAddr:     meshcommon.VTHOContractAddress,
			expectedSymbol:   "VTHO",
			expectedDecimals: 18,
			expectError:      false,
		},
		{
			name:             "Unknown token with successful contract calls",
			contractAddr:     "0x1234567890123456789012345678901234567890",
			expectedSymbol:   "USDT",
			expectedDecimals: 6,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock responses for unknown tokens
			if tt.contractAddr != meshcommon.VTHOContractAddress {
				// Mock symbol call (first call) and decimals call (second call)
				mockClient.SetMockCallResults([]string{
					"0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000045553445400000000000000000000000000000000000000000000000000000000", // Symbol: "USDT"
					"0x0000000000000000000000000000000000000000000000000000000000000006", // Decimals: 6
				})
			}

			extractor := NewOperationsExtractor()
			result, err := extractor.GetTokenCurrencyFromContractAddress(tt.contractAddr, mockClient)

			if tt.expectError {
				if err == nil {
					t.Errorf("GetTokenCurrencyFromContractAddressWithClient() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Errorf("GetTokenCurrencyFromContractAddressWithClient() error = %v, want nil", err)
				return
			}

			if result.Symbol != tt.expectedSymbol {
				t.Errorf("GetTokenCurrencyFromContractAddressWithClient() Symbol = %v, want %v", result.Symbol, tt.expectedSymbol)
			}

			if result.Decimals != tt.expectedDecimals {
				t.Errorf("GetTokenCurrencyFromContractAddressWithClient() Decimals = %v, want %v", result.Decimals, tt.expectedDecimals)
			}
		})
	}
}

func TestGetVETOperations(t *testing.T) {
	tests := []struct {
		name       string
		operations []*types.Operation
		expected   []map[string]string
	}{
		{
			name:       "empty operations",
			operations: []*types.Operation{},
			expected:   []map[string]string{},
		},
		{
			name: "VET transfer operation",
			operations: []*types.Operation{
				{
					Type: meshcommon.OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: meshtests.TestAddress1,
					},
					Amount: &types.Amount{
						Value:    "1000000000000000000",
						Currency: meshcommon.VETCurrency,
					},
				},
			},
			expected: []map[string]string{
				{
					"value": "1000000000000000000",
					"to":    meshtests.TestAddress1,
				},
			},
		},
		{
			name: "negative VET operation (should be ignored)",
			operations: []*types.Operation{
				{
					Type: meshcommon.OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: meshtests.TestAddress1,
					},
					Amount: &types.Amount{
						Value:    "-1000000000000000000",
						Currency: meshcommon.VETCurrency,
					},
				},
			},
			expected: []map[string]string{},
		},
		{
			name: "non-VET operation (should be ignored)",
			operations: []*types.Operation{
				{
					Type: meshcommon.OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: meshtests.TestAddress1,
					},
					Amount: &types.Amount{
						Value:    "1000000000000000000",
						Currency: meshcommon.VTHOCurrency,
					},
				},
			},
			expected: []map[string]string{},
		},
		{
			name: "non-transfer operation (should be ignored)",
			operations: []*types.Operation{
				{
					Type: meshcommon.OperationTypeFee,
					Account: &types.AccountIdentifier{
						Address: meshtests.TestAddress1,
					},
					Amount: &types.Amount{
						Value:    "1000000000000000000",
						Currency: meshcommon.VETCurrency,
					},
				},
			},
			expected: []map[string]string{},
		},
		{
			name: "multiple VET operations",
			operations: []*types.Operation{
				{
					Type: meshcommon.OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: meshtests.TestAddress1,
					},
					Amount: &types.Amount{
						Value:    "1000000000000000000",
						Currency: meshcommon.VETCurrency,
					},
				},
				{
					Type: meshcommon.OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: meshtests.FirstSoloAddress,
					},
					Amount: &types.Amount{
						Value:    "2000000000000000000",
						Currency: meshcommon.VETCurrency,
					},
				},
			},
			expected: []map[string]string{
				{
					"value": "1000000000000000000",
					"to":    meshtests.TestAddress1,
				},
				{
					"value": "2000000000000000000",
					"to":    meshtests.FirstSoloAddress,
				},
			},
		},
	}

	extractor := NewOperationsExtractor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.GetVETOperations(tt.operations)

			if len(result) != len(tt.expected) {
				t.Errorf("GetVETOperations() length = %v, want %v", len(result), len(tt.expected))
				return
			}

			for i, op := range result {
				if op["value"] != tt.expected[i]["value"] {
					t.Errorf("GetVETOperations() [%d] value = %v, want %v", i, op["value"], tt.expected[i]["value"])
				}
				if op["to"] != tt.expected[i]["to"] {
					t.Errorf("GetVETOperations() [%d] to = %v, want %v", i, op["to"], tt.expected[i]["to"])
				}
			}
		})
	}
}

func TestGetTokensOperations(t *testing.T) {
	tests := []struct {
		name               string
		operations         []*types.Operation
		expectedRegistered []map[string]string
	}{
		{
			name:               "empty operations",
			operations:         []*types.Operation{},
			expectedRegistered: []map[string]string{},
		},
		{
			name: "registered token operation",
			operations: []*types.Operation{
				{
					Type: meshcommon.OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: meshtests.TestAddress1,
					},
					Amount: &types.Amount{
						Value: "1000000000000000000",
						Currency: &types.Currency{
							Symbol:   "VTHO",
							Decimals: 18,
							Metadata: map[string]any{
								"contractAddress": meshcommon.VTHOContractAddress,
							},
						},
					},
				},
			},
			expectedRegistered: []map[string]string{
				{
					"token": meshcommon.VTHOContractAddress,
					"value": "1000000000000000000",
					"to":    meshtests.TestAddress1,
				},
			},
		},
		{
			name: "negative token operation (should be ignored)",
			operations: []*types.Operation{
				{
					Type: meshcommon.OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: meshtests.TestAddress1,
					},
					Amount: &types.Amount{
						Value: "-1000000000000000000",
						Currency: &types.Currency{
							Symbol:   "VTHO",
							Decimals: 18,
							Metadata: map[string]any{
								"contractAddress": meshcommon.VTHOContractAddress,
							},
						},
					},
				},
			},
			expectedRegistered: []map[string]string{},
		},
		{
			name: "non-transfer operation (should be ignored)",
			operations: []*types.Operation{
				{
					Type: meshcommon.OperationTypeFee,
					Account: &types.AccountIdentifier{
						Address: meshtests.TestAddress1,
					},
					Amount: &types.Amount{
						Value: "1000000000000000000",
						Currency: &types.Currency{
							Symbol:   "VTHO",
							Decimals: 18,
							Metadata: map[string]any{
								"contractAddress": meshcommon.VTHOContractAddress,
							},
						},
					},
				},
			},
		},
		{
			name: "no config (assume all registered)",
			operations: []*types.Operation{
				{
					Type: meshcommon.OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: meshtests.TestAddress1,
					},
					Amount: &types.Amount{
						Value: "1000000000000000000",
						Currency: &types.Currency{
							Symbol:   "UNKNOWN",
							Decimals: 18,
							Metadata: map[string]any{
								"contractAddress": "0x9999999999999999999999999999999999999999",
							},
						},
					},
				},
			},
			expectedRegistered: []map[string]string{
				{
					"token": "0x9999999999999999999999999999999999999999",
					"value": "1000000000000000000",
					"to":    meshtests.TestAddress1,
				},
			},
		},
	}

	extractor := NewOperationsExtractor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registered := extractor.GetTokensOperations(tt.operations)

			if len(registered) != len(tt.expectedRegistered) {
				t.Errorf("GetTokensOperations() registered length = %v, want %v", len(registered), len(tt.expectedRegistered))
				return
			}

			// Check registered tokens
			for i, reg := range registered {
				expected := tt.expectedRegistered[i]
				if reg["token"] != expected["token"] {
					t.Errorf("GetTokensOperations() registered[%d] token = %v, want %v", i, reg["token"], expected["token"])
				}
				if reg["value"] != expected["value"] {
					t.Errorf("GetTokensOperations() registered[%d] value = %v, want %v", i, reg["value"], expected["value"])
				}
				if reg["to"] != expected["to"] {
					t.Errorf("GetTokensOperations() registered[%d] to = %v, want %v", i, reg["to"], expected["to"])
				}
			}
		})
	}
}

func TestGetFeeDelegatorAccount(t *testing.T) {
	extractor := NewOperationsExtractor()

	tests := []struct {
		name     string
		metadata map[string]any
		expected string
	}{
		{
			name: "valid fee delegator account",
			metadata: map[string]any{
				meshcommon.DelegatorAccountMetadataKey: meshtests.TestAddress1,
			},
			expected: meshtests.TestAddress1,
		},
		{
			name:     "missing fee delegator account",
			metadata: map[string]any{},
			expected: "",
		},
		{
			name:     "nil metadata",
			metadata: nil,
			expected: "",
		},
		{
			name: "fee delegator account with uppercase (should be lowercased)",
			metadata: map[string]any{
				meshcommon.DelegatorAccountMetadataKey: "0xABCDEF1234567890ABCDEF1234567890ABCDEF12",
			},
			expected: "0xabcdef1234567890abcdef1234567890abcdef12",
		},
		{
			name: "wrong type for fee delegator account",
			metadata: map[string]any{
				meshcommon.DelegatorAccountMetadataKey: 12345,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.GetFeeDelegatorAccount(tt.metadata)
			if result != tt.expected {
				t.Errorf("GetFeeDelegatorAccount() = %v, want %v", result, tt.expected)
			}
		})
	}
}
