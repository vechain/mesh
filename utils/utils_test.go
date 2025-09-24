package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
)

func TestWriteJSONResponse(t *testing.T) {
	tests := []struct {
		name     string
		response any
		wantCode int
	}{
		{
			name:     "simple string response",
			response: "test",
			wantCode: http.StatusOK,
		},
		{
			name: "complex struct response",
			response: map[string]any{
				"message": "success",
				"data":    []string{"item1", "item2"},
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "nil response",
			response: nil,
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteJSONResponse(w, tt.response)

			if w.Code != tt.wantCode {
				t.Errorf("WriteJSONResponse() status code = %v, want %v", w.Code, tt.wantCode)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("WriteJSONResponse() content type = %v, want application/json", contentType)
			}

			// Verify response can be unmarshaled
			var result any
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Errorf("WriteJSONResponse() produced invalid JSON: %v", err)
			}
		})
	}
}

func TestWriteErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		err        *types.Error
		statusCode int
		wantCode   int
	}{
		{
			name: "valid error",
			err: &types.Error{
				Code:    1001,
				Message: "Test error",
			},
			statusCode: http.StatusBadRequest,
			wantCode:   http.StatusBadRequest,
		},
		{
			name:       "nil error with default",
			err:        nil,
			statusCode: http.StatusInternalServerError,
			wantCode:   http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteErrorResponse(w, tt.err, tt.statusCode)

			if w.Code != tt.wantCode {
				t.Errorf("WriteErrorResponse() status code = %v, want %v", w.Code, tt.wantCode)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("WriteErrorResponse() content type = %v, want application/json", contentType)
			}

			// Verify error response structure
			var result map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Errorf("WriteErrorResponse() produced invalid JSON: %v", err)
			}

			if _, exists := result["error"]; !exists {
				t.Errorf("WriteErrorResponse() missing 'error' field in response")
			}
		})
	}
}

func TestGenerateNonce(t *testing.T) {
	// Test multiple calls to ensure randomness
	nonces := make(map[string]bool)

	for i := 0; i < 100; i++ {
		nonce, err := GenerateNonce()
		if err != nil {
			t.Errorf("GenerateNonce() error = %v", err)
			return
		}

		// Check format (should start with 0x and be 18 characters total)
		if !strings.HasPrefix(nonce, "0x") {
			t.Errorf("GenerateNonce() = %v, want prefix '0x'", nonce)
		}

		if len(nonce) != 18 {
			t.Errorf("GenerateNonce() = %v, want length 18, got %d", nonce, len(nonce))
		}

		// Check uniqueness
		if nonces[nonce] {
			t.Errorf("GenerateNonce() produced duplicate nonce: %v", nonce)
		}
		nonces[nonce] = true
	}
}

func TestStringPtr(t *testing.T) {
	tests := []struct {
		name string
		s    string
	}{
		{"empty string", ""},
		{"simple string", "test"},
		{"string with spaces", "hello world"},
		{"string with special chars", "test@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := StringPtr(tt.s)
			if ptr == nil {
				t.Errorf("StringPtr() returned nil pointer")
				return
			}
			if *ptr != tt.s {
				t.Errorf("StringPtr() = %v, want %v", *ptr, tt.s)
			}
		})
	}
}

func TestGetStringFromOptions(t *testing.T) {
	tests := []struct {
		name         string
		options      map[string]any
		key          string
		defaultValue string
		expected     string
	}{
		{
			name:         "valid string value",
			options:      map[string]any{"test": "value"},
			key:          "test",
			defaultValue: "default",
			expected:     "value",
		},
		{
			name:         "missing key",
			options:      map[string]any{"other": "value"},
			key:          "test",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "non-string value",
			options:      map[string]any{"test": 123},
			key:          "test",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "nil options",
			options:      nil,
			key:          "test",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStringFromOptions(tt.options, tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetStringFromOptions() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHexToDecimal(t *testing.T) {
	tests := []struct {
		name     string
		hexStr   string
		expected string
		hasError bool
	}{
		{
			name:     "valid hex with 0x prefix",
			hexStr:   "0x1a",
			expected: "26",
			hasError: false,
		},
		{
			name:     "valid hex without 0x prefix",
			hexStr:   "1a",
			expected: "26",
			hasError: false,
		},
		{
			name:     "zero value",
			hexStr:   "0x0",
			expected: "0",
			hasError: false,
		},
		{
			name:     "large value",
			hexStr:   "0xde0b6b3a7640000",
			expected: "1000000000000000000",
			hasError: false,
		},
		{
			name:     "invalid hex",
			hexStr:   "0xgg",
			expected: "",
			hasError: true,
		},
		{
			name:     "empty string",
			hexStr:   "",
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HexToDecimal(tt.hexStr)

			if tt.hasError {
				if err == nil {
					t.Errorf("HexToDecimal() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("HexToDecimal() unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("HexToDecimal() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestRemoveHexPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with 0x prefix", "0x1a", "1a"},
		{"without 0x prefix", "1a", "1a"},
		{"with 0X prefix", "0X1a", "0X1a"}, // Function only removes lowercase 0x
		{"empty string", "", ""},
		{"just 0x", "0x", "0x"}, // Function only removes lowercase 0x
		{"just 0X", "0X", "0X"}, // Function only removes lowercase 0x
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveHexPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveHexPrefix() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDecodeHexStringWithPrefix(t *testing.T) {
	tests := []struct {
		name     string
		hexStr   string
		expected []byte
		hasError bool
	}{
		{
			name:     "valid hex with 0x prefix",
			hexStr:   "0x1a2b",
			expected: []byte{0x1a, 0x2b},
			hasError: false,
		},
		{
			name:     "valid hex without 0x prefix",
			hexStr:   "1a2b",
			expected: []byte{0x1a, 0x2b},
			hasError: false,
		},
		{
			name:     "empty string",
			hexStr:   "",
			expected: []byte{},
			hasError: false,
		},
		{
			name:     "invalid hex",
			hexStr:   "0xgg",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeHexStringWithPrefix(tt.hexStr)

			if tt.hasError {
				if err == nil {
					t.Errorf("DecodeHexStringWithPrefix() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("DecodeHexStringWithPrefix() unexpected error: %v", err)
				}
				if !bytes.Equal(result, tt.expected) {
					t.Errorf("DecodeHexStringWithPrefix() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestGetTxOrigins(t *testing.T) {
	tests := []struct {
		name       string
		operations []*types.Operation
		expected   []string
	}{
		{
			name: "transfer operations with negative amounts",
			operations: []*types.Operation{
				{
					Type: OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: "0x123",
					},
					Amount: &types.Amount{
						Value: "-100",
					},
				},
				{
					Type: OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: "0x456",
					},
					Amount: &types.Amount{
						Value: "100",
					},
				},
			},
			expected: []string{"0x123"},
		},
		{
			name: "fee operations",
			operations: []*types.Operation{
				{
					Type: OperationTypeFee,
					Account: &types.AccountIdentifier{
						Address: "0x789",
					},
				},
			},
			expected: []string{"0x789"},
		},
		{
			name: "mixed operations",
			operations: []*types.Operation{
				{
					Type: OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: "0x123",
					},
					Amount: &types.Amount{
						Value: "-100",
					},
				},
				{
					Type: OperationTypeFee,
					Account: &types.AccountIdentifier{
						Address: "0x456",
					},
				},
			},
			expected: []string{"0x123", "0x456"},
		},
		{
			name:       "empty operations",
			operations: []*types.Operation{},
			expected:   []string{},
		},
		{
			name: "operations without accounts",
			operations: []*types.Operation{
				{
					Type:    OperationTypeTransfer,
					Account: nil,
				},
			},
			expected: []string{},
		},
		{
			name: "duplicate addresses",
			operations: []*types.Operation{
				{
					Type: OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: "0x123",
					},
					Amount: &types.Amount{
						Value: "-100",
					},
				},
				{
					Type: OperationTypeFee,
					Account: &types.AccountIdentifier{
						Address: "0x123",
					},
				},
			},
			expected: []string{"0x123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTxOrigins(tt.operations)

			if len(result) != len(tt.expected) {
				t.Errorf("GetTxOrigins() length = %v, want %v", len(result), len(tt.expected))
				return
			}

			for i, addr := range result {
				if addr != tt.expected[i] {
					t.Errorf("GetTxOrigins()[%d] = %v, want %v", i, addr, tt.expected[i])
				}
			}
		})
	}
}

// Test currency constants
func TestCurrencyConstants(t *testing.T) {
	// Test VETCurrency
	if VETCurrency.Symbol != "VET" {
		t.Errorf("VETCurrency.Symbol = %v, want VET", VETCurrency.Symbol)
	}
	if VETCurrency.Decimals != 18 {
		t.Errorf("VETCurrency.Decimals = %v, want 18", VETCurrency.Decimals)
	}

	// Test VTHOCurrency
	if VTHOCurrency.Symbol != "VTHO" {
		t.Errorf("VTHOCurrency.Symbol = %v, want VTHO", VTHOCurrency.Symbol)
	}
	if VTHOCurrency.Decimals != 18 {
		t.Errorf("VTHOCurrency.Decimals = %v, want 18", VTHOCurrency.Decimals)
	}
	if VTHOCurrency.Metadata == nil {
		t.Errorf("VTHOCurrency.Metadata is nil")
	} else {
		contractAddr, exists := VTHOCurrency.Metadata["contractAddress"]
		if !exists {
			t.Errorf("VTHOCurrency.Metadata missing contractAddress")
		} else {
			expectedAddr := "0x0000000000000000000000000000456E65726779"
			if contractAddr != expectedAddr {
				t.Errorf("VTHOCurrency.Metadata.contractAddress = %v, want %v", contractAddr, expectedAddr)
			}
		}
	}
}

// Test operation type constants
func TestOperationTypeConstants(t *testing.T) {
	expectedTypes := map[string]string{
		"OperationTypeNone":          "None",
		"OperationTypeTransfer":      "Transfer",
		"OperationTypeFee":           "Fee",
		"OperationTypeFeeDelegation": "FeeDelegation",
	}

	actualTypes := map[string]string{
		"OperationTypeNone":          OperationTypeNone,
		"OperationTypeTransfer":      OperationTypeTransfer,
		"OperationTypeFee":           OperationTypeFee,
		"OperationTypeFeeDelegation": OperationTypeFeeDelegation,
	}

	for name, expected := range expectedTypes {
		if actualTypes[name] != expected {
			t.Errorf("%s = %v, want %v", name, actualTypes[name], expected)
		}
	}
}

func TestInt64Ptr(t *testing.T) {
	value := int64(42)
	ptr := Int64Ptr(value)
	if ptr == nil {
		t.Errorf("Int64Ptr() returned nil")
	}
	if *ptr != value {
		t.Errorf("Int64Ptr() = %v, want %v", *ptr, value)
	}
}

func TestBoolPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected bool
	}{
		{"true", true, true},
		{"false", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := BoolPtr(tt.input)
			if ptr == nil {
				t.Errorf("BoolPtr() returned nil")
			}
			if *ptr != tt.expected {
				t.Errorf("BoolPtr() = %v, want %v", *ptr, tt.expected)
			}
		})
	}
}

func TestGetTargetIndex(t *testing.T) {
	tests := []struct {
		name       string
		localIndex int64
		peers      []Peer
		expected   int64
	}{
		{
			name:       "no peers",
			localIndex: 100,
			peers:      []Peer{},
			expected:   100,
		},
		{
			name:       "peers with lower block numbers",
			localIndex: 100,
			peers: []Peer{
				{PeerID: "peer1", BestBlockID: "0000000000000050"}, // 80 in decimal
				{PeerID: "peer2", BestBlockID: "0000000000000060"}, // 96 in decimal
			},
			expected: 100,
		},
		{
			name:       "peers with higher block numbers",
			localIndex: 100,
			peers: []Peer{
				{PeerID: "peer1", BestBlockID: "0000000000000080"}, // 128 in decimal
				{PeerID: "peer2", BestBlockID: "00000000000000A0"}, // 160 in decimal
			},
			expected: 160,
		},
		{
			name:       "peers with invalid block IDs",
			localIndex: 100,
			peers: []Peer{
				{PeerID: "peer1", BestBlockID: "invalid"},
				{PeerID: "peer2", BestBlockID: "short"},
			},
			expected: 100,
		},
		{
			name:       "mixed valid and invalid peers",
			localIndex: 100,
			peers: []Peer{
				{PeerID: "peer1", BestBlockID: "invalid"},
				{PeerID: "peer2", BestBlockID: "0000000000000080"}, // 128 in decimal
			},
			expected: 128,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTargetIndex(tt.localIndex, tt.peers)
			if result != tt.expected {
				t.Errorf("GetTargetIndex() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestComputeAddress(t *testing.T) {
	// Test with a valid public key
	validPubKey := &types.PublicKey{
		Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
		CurveType: "secp256k1",
	}

	address, err := ComputeAddress(validPubKey)
	if err != nil {
		t.Errorf("ComputeAddress() error = %v", err)
	}
	if address != "0xf077b491b355e64048ce21e3a6fc4751eeea77fa" {
		t.Errorf("ComputeAddress() = %v, want 0xf077b491b355e64048ce21e3a6fc4751eeea77fa", address)
	}

	// Test with invalid public key
	invalidPubKey := &types.PublicKey{
		Bytes:     []byte{0x00}, // Invalid public key
		CurveType: "secp256k1",
	}

	_, err = ComputeAddress(invalidPubKey)
	if err == nil {
		t.Errorf("ComputeAddress() with invalid key should return error")
	}
}
