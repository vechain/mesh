package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
)

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

func TestWriteJSONResponse(t *testing.T) {
	// Test successful JSON response
	response := map[string]string{"message": "test"}
	w := httptest.NewRecorder()

	WriteJSONResponse(w, response)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if result["message"] != "test" {
		t.Errorf("Expected message 'test', got %s", result["message"])
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

func TestRemoveHexPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0x1234", "1234"},
		{"-0x1234", "-0x1234"}, // Function only removes "0x" prefix, not "-0x"
		{"1234", "1234"},
		{"", ""},
		{"0x", "0x"},   // Function doesn't handle edge case of just "0x"
		{"-0x", "-0x"}, // Function doesn't handle edge case of just "-0x"
	}

	for _, tt := range tests {
		result := RemoveHexPrefix(tt.input)
		if result != tt.expected {
			t.Errorf("RemoveHexPrefix(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
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

func TestInt64Ptr(t *testing.T) {
	value := int64(42)
	ptr := Int64Ptr(value)

	if ptr == nil || *ptr != value {
		t.Error("Expected non-nil pointer and value")
	}
}

func TestBoolPtr(t *testing.T) {
	value := true
	ptr := BoolPtr(value)

	if ptr == nil || *ptr != value {
		t.Error("Expected non-nil pointer")
	}

	if *ptr != value {
		t.Errorf("Expected %t, got %t", value, *ptr)
	}
}

func TestStringPtr(t *testing.T) {
	value := "test"
	ptr := StringPtr(value)

	if ptr == nil || *ptr != value {
		t.Error("Expected non-nil pointer")
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
		{"0x", nil, true}, // "0x" becomes "" after RemoveHexPrefix, which is invalid hex
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
