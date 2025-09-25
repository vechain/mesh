package utils

import (
	"context"
	"encoding/json"
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
