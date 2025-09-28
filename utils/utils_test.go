package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/vechain/mesh/config"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/thor"
	thorTx "github.com/vechain/thor/v2/tx"
)

func TestWriteJSONResponse(t *testing.T) {
	// Test successful JSON response
	w := httptest.NewRecorder()
	response := map[string]string{"message": "test"}

	WriteJSONResponse(w, response)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("WriteJSONResponse() status = %v, want %v", w.Code, http.StatusOK)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("WriteJSONResponse() Content-Type = %v, want application/json", contentType)
	}

	// Check response body
	var result map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &result)
	if err != nil {
		t.Errorf("WriteJSONResponse() failed to unmarshal response: %v", err)
	}

	if result["message"] != "test" {
		t.Errorf("WriteJSONResponse() message = %v, want test", result["message"])
	}

	// Test with invalid JSON (should not happen in practice, but test error handling)
	w2 := httptest.NewRecorder()
	invalidResponse := make(chan int) // channels cannot be marshaled to JSON

	WriteJSONResponse(w2, invalidResponse)

	// Should return 500 error for invalid JSON
	if w2.Code != http.StatusInternalServerError {
		t.Errorf("WriteJSONResponse() status = %v, want %v", w2.Code, http.StatusInternalServerError)
	}
}

func TestParseJSONFromRequestContext(t *testing.T) {
	// Create test request data
	request := types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
	}
	requestBody, _ := json.Marshal(request)

	// Create request with body in context
	req := httptest.NewRequest("POST", "/network/status", nil)
	ctx := context.WithValue(req.Context(), RequestBodyKey, requestBody)
	req = req.WithContext(ctx)

	// Test parsing from context
	var parsedRequest types.NetworkRequest
	err := ParseJSONFromRequestContext(req, &parsedRequest)
	if err != nil {
		t.Fatalf("ParseJSONFromRequestContext failed: %v", err)
	}

	// Verify parsed data
	if parsedRequest.NetworkIdentifier.Blockchain != "vechainthor" {
		t.Errorf("Expected blockchain 'vechainthor', got '%s'", parsedRequest.NetworkIdentifier.Blockchain)
	}
	if parsedRequest.NetworkIdentifier.Network != "test" {
		t.Errorf("Expected network 'test', got '%s'", parsedRequest.NetworkIdentifier.Network)
	}
}

func TestParseJSONFromRequestContext_NoBodyInContext(t *testing.T) {
	// Create request WITHOUT body in context
	req := httptest.NewRequest("POST", "/network/status", nil)

	// Test parsing should fail
	var parsedRequest types.NetworkRequest
	err := ParseJSONFromRequestContext(req, &parsedRequest)
	if err == nil {
		t.Error("Expected error when no body in context, but got nil")
	}

	// Verify error message
	expectedError := "request body not found in context - middleware may not be properly configured"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestParseJSONFromRequestContext_InvalidJSON(t *testing.T) {
	// Create request with invalid JSON in context
	invalidJSON := []byte(`{"invalid": json}`)
	req := httptest.NewRequest("POST", "/network/status", nil)
	ctx := context.WithValue(req.Context(), RequestBodyKey, invalidJSON)
	req = req.WithContext(ctx)

	// Test parsing should fail
	var parsedRequest types.NetworkRequest
	err := ParseJSONFromRequestContext(req, &parsedRequest)
	if err == nil {
		t.Error("Expected error when parsing invalid JSON, but got nil")
	}
}

func TestWriteErrorResponse(t *testing.T) {
	// Test error response with valid error
	error := GetError(ErrInvalidRequestBody)
	w := httptest.NewRecorder()

	WriteErrorResponse(w, error, http.StatusBadRequest)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}

	if result["error"] == nil {
		t.Error("Expected error field in response")
	}
}

func TestWriteErrorResponse_NilError(t *testing.T) {
	// Test error response with nil error (should use default)
	w := httptest.NewRecorder()

	WriteErrorResponse(w, nil, http.StatusInternalServerError)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}

	if result["error"] == nil {
		t.Error("Expected error field in response")
	}
}

func TestGenerateNonce(t *testing.T) {
	nonce, err := GenerateNonce()
	if err != nil {
		t.Errorf("GenerateNonce() unexpected error: %v", err)
	}

	if len(nonce) != 18 { // "0x" + 16 hex chars = 18
		t.Errorf("Expected nonce length 18, got %d", len(nonce))
	}

	// Test that nonces are different
	nonce2, err := GenerateNonce()
	if err != nil {
		t.Errorf("GenerateNonce() unexpected error: %v", err)
	}
	if nonce == nonce2 {
		t.Error("Expected different nonces, but got the same")
	}
}

func TestGetStringFromOptions(t *testing.T) {
	options := map[string]any{
		"key1": "value1",
		"key2": 123,
		"key3": nil,
	}

	// Test existing string value
	result := GetStringFromOptions(options, "key1")
	if result != "value1" {
		t.Errorf("Expected 'value1', got %s", result)
	}

	// Test non-existing key
	result = GetStringFromOptions(options, "nonexistent")
	if result != "dynamic" {
		t.Errorf("Expected 'dynamic', got %s", result)
	}

	// Test non-string value
	result = GetStringFromOptions(options, "key2")
	if result != "dynamic" {
		t.Errorf("Expected 'dynamic' for non-string value, got %s", result)
	}

	// Test nil value
	result = GetStringFromOptions(options, "key3")
	if result != "dynamic" {
		t.Errorf("Expected 'dynamic' for nil value, got %s", result)
	}
}

func TestHexToDecimal(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"0x1", "1", false},
		{"0xa", "10", false},
		{"0xff", "255", false},
		{"0x100", "256", false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		result, err := HexToDecimal(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("HexToDecimal(%s) expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("HexToDecimal(%s) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("HexToDecimal(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		}
	}
}

func TestComputeAddress(t *testing.T) {
	// Test with valid public key
	publicKey := &types.PublicKey{
		Bytes:     []byte{0x02, 0xd9, 0x92, 0xbd, 0x20, 0x3d, 0x2b, 0xf8, 0x88, 0x38, 0x90, 0x89, 0xdb, 0x13, 0xd2, 0xd0, 0x80, 0x7c, 0x16, 0x97, 0x09, 0x1d, 0xe3, 0x77, 0x99, 0x8e, 0xfe, 0x6c, 0xf6, 0x0d, 0x66, 0xfb, 0xb3},
		CurveType: "secp256k1",
	}

	address, err := ComputeAddress(publicKey)
	if err != nil {
		t.Errorf("ComputeAddress() unexpected error: %v", err)
	}

	if address == "" {
		t.Error("Expected non-empty address")
	}

	// Test with invalid public key
	invalidPublicKey := &types.PublicKey{
		Bytes:     []byte("invalid"),
		CurveType: "secp256k1",
	}

	_, err = ComputeAddress(invalidPublicKey)
	if err == nil {
		t.Error("Expected error for invalid public key")
	}
}

func TestGetTxOrigins(t *testing.T) {
	operations := []*types.Operation{
		{
			Type: OperationTypeFee,
			Account: &types.AccountIdentifier{
				Address: "0x1234567890123456789012345678901234567890",
			},
		},
		{
			Type: OperationTypeTransfer,
			Account: &types.AccountIdentifier{
				Address: "0x0987654321098765432109876543210987654321",
			},
			Amount: &types.Amount{
				Value: "-1000000000000000000", // Negative value for sending
			},
		},
	}

	origins := GetTxOrigins(operations)

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

func TestGetTargetIndex(t *testing.T) {
	peers := []Peer{
		{
			BestBlockID: "0000000000000001",
		},
		{
			BestBlockID: "0000000000000002",
		},
	}

	// Test with local index 0
	index := GetTargetIndex(0, peers)
	if index != 2 {
		t.Errorf("Expected index 2, got %d", index)
	}

	// Test with local index higher than peers
	index = GetTargetIndex(5, peers)
	if index != 5 {
		t.Errorf("Expected index 5, got %d", index)
	}
}

func TestDecodeHexStringWithPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected []byte
		hasError bool
	}{
		{"0x1234", []byte{0x12, 0x34}, false},
		{"0xff", []byte{0xff}, false},
		{"invalid", nil, true},
		{"", []byte{}, false}, // Empty string is valid hex (returns empty byte slice)
	}

	for _, tt := range tests {
		result, err := DecodeHexStringWithPrefix(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("DecodeHexStringWithPrefix(%s) expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("DecodeHexStringWithPrefix(%s) unexpected error: %v", tt.input, err)
			}
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("DecodeHexStringWithPrefix(%s) = %v, expected %v", tt.input, result, tt.expected)
			}
		}
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
			contractAddr:     "0x0000000000000000000000000000456e65726779",
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
			if tt.contractAddr != "0x0000000000000000000000000000456e65726779" {
				// Mock symbol call (first call) and decimals call (second call)
				mockClient.SetMockCallResults([]string{
					"0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000045553445400000000000000000000000000000000000000000000000000000000", // Symbol: "USDT"
					"0x0000000000000000000000000000000000000000000000000000000000000006", // Decimals: 6
				})
			}

			result, err := GetTokenCurrencyFromContractAddress(tt.contractAddr, mockClient)

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
					Type: OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: "0x16277a1ff38678291c41d1820957c78bb5da59ce",
					},
					Amount: &types.Amount{
						Value:    "1000000000000000000",
						Currency: VETCurrency,
					},
				},
			},
			expected: []map[string]string{
				{
					"value": "1000000000000000000",
					"to":    "0x16277a1ff38678291c41d1820957c78bb5da59ce",
				},
			},
		},
		{
			name: "negative VET operation (should be ignored)",
			operations: []*types.Operation{
				{
					Type: OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: "0x16277a1ff38678291c41d1820957c78bb5da59ce",
					},
					Amount: &types.Amount{
						Value:    "-1000000000000000000",
						Currency: VETCurrency,
					},
				},
			},
			expected: []map[string]string{},
		},
		{
			name: "non-VET operation (should be ignored)",
			operations: []*types.Operation{
				{
					Type: OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: "0x16277a1ff38678291c41d1820957c78bb5da59ce",
					},
					Amount: &types.Amount{
						Value:    "1000000000000000000",
						Currency: VTHOCurrency,
					},
				},
			},
			expected: []map[string]string{},
		},
		{
			name: "non-transfer operation (should be ignored)",
			operations: []*types.Operation{
				{
					Type: OperationTypeFee,
					Account: &types.AccountIdentifier{
						Address: "0x16277a1ff38678291c41d1820957c78bb5da59ce",
					},
					Amount: &types.Amount{
						Value:    "1000000000000000000",
						Currency: VETCurrency,
					},
				},
			},
			expected: []map[string]string{},
		},
		{
			name: "multiple VET operations",
			operations: []*types.Operation{
				{
					Type: OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: "0x16277a1ff38678291c41d1820957c78bb5da59ce",
					},
					Amount: &types.Amount{
						Value:    "1000000000000000000",
						Currency: VETCurrency,
					},
				},
				{
					Type: OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
					},
					Amount: &types.Amount{
						Value:    "2000000000000000000",
						Currency: VETCurrency,
					},
				},
			},
			expected: []map[string]string{
				{
					"value": "1000000000000000000",
					"to":    "0x16277a1ff38678291c41d1820957c78bb5da59ce",
				},
				{
					"value": "2000000000000000000",
					"to":    "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetVETOperations(tt.operations)

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
	// Create a mock config with some tokens
	mockConfig := &config.Config{
		TokenList: []types.Currency{
			{
				Symbol:   "VTHO",
				Decimals: 18,
				Metadata: map[string]any{
					"contractAddress": "0x0000000000000000000000000000456e65726779",
				},
			},
			{
				Symbol:   "TOKEN",
				Decimals: 18,
				Metadata: map[string]any{
					"contractAddress": "0x1234567890123456789012345678901234567890",
				},
			},
		},
	}

	tests := []struct {
		name                 string
		operations           []*types.Operation
		config               *config.Config
		expectedRegistered   []map[string]string
		expectedUnregistered []string
	}{
		{
			name:                 "empty operations",
			operations:           []*types.Operation{},
			config:               mockConfig,
			expectedRegistered:   []map[string]string{},
			expectedUnregistered: []string{},
		},
		{
			name: "registered token operation",
			operations: []*types.Operation{
				{
					Type: OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: "0x16277a1ff38678291c41d1820957c78bb5da59ce",
					},
					Amount: &types.Amount{
						Value: "1000000000000000000",
						Currency: &types.Currency{
							Symbol:   "VTHO",
							Decimals: 18,
							Metadata: map[string]any{
								"contractAddress": "0x0000000000000000000000000000456e65726779",
							},
						},
					},
				},
			},
			config: mockConfig,
			expectedRegistered: []map[string]string{
				{
					"token": "0x0000000000000000000000000000456e65726779",
					"value": "1000000000000000000",
					"to":    "0x16277a1ff38678291c41d1820957c78bb5da59ce",
				},
			},
			expectedUnregistered: []string{},
		},
		{
			name: "unregistered token operation",
			operations: []*types.Operation{
				{
					Type: OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: "0x16277a1ff38678291c41d1820957c78bb5da59ce",
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
			config:               mockConfig,
			expectedRegistered:   []map[string]string{},
			expectedUnregistered: []string{"0x9999999999999999999999999999999999999999"},
		},
		{
			name: "negative token operation (should be ignored)",
			operations: []*types.Operation{
				{
					Type: OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: "0x16277a1ff38678291c41d1820957c78bb5da59ce",
					},
					Amount: &types.Amount{
						Value: "-1000000000000000000",
						Currency: &types.Currency{
							Symbol:   "VTHO",
							Decimals: 18,
							Metadata: map[string]any{
								"contractAddress": "0x0000000000000000000000000000456e65726779",
							},
						},
					},
				},
			},
			config:               mockConfig,
			expectedRegistered:   []map[string]string{},
			expectedUnregistered: []string{},
		},
		{
			name: "non-transfer operation (should be ignored)",
			operations: []*types.Operation{
				{
					Type: OperationTypeFee,
					Account: &types.AccountIdentifier{
						Address: "0x16277a1ff38678291c41d1820957c78bb5da59ce",
					},
					Amount: &types.Amount{
						Value: "1000000000000000000",
						Currency: &types.Currency{
							Symbol:   "VTHO",
							Decimals: 18,
							Metadata: map[string]any{
								"contractAddress": "0x0000000000000000000000000000456e65726779",
							},
						},
					},
				},
			},
			config:               mockConfig,
			expectedRegistered:   []map[string]string{},
			expectedUnregistered: []string{},
		},
		{
			name: "no config (assume all registered)",
			operations: []*types.Operation{
				{
					Type: OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: "0x16277a1ff38678291c41d1820957c78bb5da59ce",
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
			config: nil,
			expectedRegistered: []map[string]string{
				{
					"token": "0x9999999999999999999999999999999999999999",
					"value": "1000000000000000000",
					"to":    "0x16277a1ff38678291c41d1820957c78bb5da59ce",
				},
			},
			expectedUnregistered: []string{},
		},
		{
			name: "mixed registered and unregistered tokens",
			operations: []*types.Operation{
				{
					Type: OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: "0x16277a1ff38678291c41d1820957c78bb5da59ce",
					},
					Amount: &types.Amount{
						Value: "1000000000000000000",
						Currency: &types.Currency{
							Symbol:   "VTHO",
							Decimals: 18,
							Metadata: map[string]any{
								"contractAddress": "0x0000000000000000000000000000456e65726779",
							},
						},
					},
				},
				{
					Type: OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
					},
					Amount: &types.Amount{
						Value: "2000000000000000000",
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
			config: mockConfig,
			expectedRegistered: []map[string]string{
				{
					"token": "0x0000000000000000000000000000456e65726779",
					"value": "1000000000000000000",
					"to":    "0x16277a1ff38678291c41d1820957c78bb5da59ce",
				},
			},
			expectedUnregistered: []string{"0x9999999999999999999999999999999999999999"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registered, unregistered := GetTokensOperations(tt.operations, tt.config)

			if len(registered) != len(tt.expectedRegistered) {
				t.Errorf("GetTokensOperations() registered length = %v, want %v", len(registered), len(tt.expectedRegistered))
				return
			}

			if len(unregistered) != len(tt.expectedUnregistered) {
				t.Errorf("GetTokensOperations() unregistered length = %v, want %v", len(unregistered), len(tt.expectedUnregistered))
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

			// Check unregistered tokens
			for i, unreg := range unregistered {
				if unreg != tt.expectedUnregistered[i] {
					t.Errorf("GetTokensOperations() unregistered[%d] = %v, want %v", i, unreg, tt.expectedUnregistered[i])
				}
			}
		})
	}
}

func TestMeshTransactionEncoder_parseTransactionSignersAndOperations(t *testing.T) {
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
					toAddr, _ := thor.ParseAddress("0x16277a1ff38678291c41d1820957c78bb5da59ce")
					value := new(big.Int)
					value.SetString("1000000000000000000", 10) // 1 VET

					thorClause := thorTx.NewClause(&toAddr)
					thorClause = thorClause.WithValue(value)
					thorClause = thorClause.WithData([]byte{})
					builder.Clause(thorClause)

					return builder.Build()
				}(),
				Origin: func() []byte {
					addr, _ := thor.ParseAddress("0xf077b491b355e64048ce21e3a6fc4751eeea77fa")
					return addr.Bytes()
				}(),
				Delegator: []byte{},
			},
			expectedOps:     4,
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
					addr, _ := thor.ParseAddress("0xf077b491b355e64048ce21e3a6fc4751eeea77fa")
					return addr.Bytes()
				}(),
				Delegator: func() []byte {
					addr, _ := thor.ParseAddress("0x16277a1ff38678291c41d1820957c78bb5da59ce")
					return addr.Bytes()
				}(),
			},
			expectedOps:     4,
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
					addr, _ := thor.ParseAddress("0xf077b491b355e64048ce21e3a6fc4751eeea77fa")
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
			operations, signers := encoder.parseTransactionSignersAndOperations(tt.meshTx)

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

func TestCreateTransactionBuilder(t *testing.T) {
	t.Run("Legacy transaction with valid gasPriceCoef", func(t *testing.T) {
		metadata := map[string]any{
			"gasPriceCoef": float64(100),
		}

		builder, err := createTransactionBuilder("legacy", metadata)
		if err != nil {
			t.Errorf("createTransactionBuilder() error = %v, want nil", err)
		}
		if builder == nil {
			t.Errorf("createTransactionBuilder() returned nil builder")
		}
	})

	t.Run("Legacy transaction with uint8 gasPriceCoef", func(t *testing.T) {
		metadata := map[string]any{
			"gasPriceCoef": uint8(100),
		}

		builder, err := createTransactionBuilder("legacy", metadata)
		if err != nil {
			t.Errorf("createTransactionBuilder() error = %v, want nil", err)
		}
		if builder == nil {
			t.Errorf("createTransactionBuilder() returned nil builder")
		}
	})

	t.Run("Legacy transaction with invalid gasPriceCoef type", func(t *testing.T) {
		metadata := map[string]any{
			"gasPriceCoef": "invalid",
		}

		builder, err := createTransactionBuilder("legacy", metadata)
		if err == nil {
			t.Errorf("createTransactionBuilder() should return error for invalid gasPriceCoef type")
		}
		if builder != nil {
			t.Errorf("createTransactionBuilder() should return nil builder when error occurs")
		}
	})

	t.Run("Dynamic fee transaction", func(t *testing.T) {
		metadata := map[string]any{
			"maxFeePerGas":         "1000000000",
			"maxPriorityFeePerGas": "1000000000",
		}

		builder, err := createTransactionBuilder("dynamic", metadata)
		if err != nil {
			t.Errorf("createTransactionBuilder() error = %v, want nil", err)
		}
		if builder == nil {
			t.Errorf("createTransactionBuilder() returned nil builder")
		}
	})
}
