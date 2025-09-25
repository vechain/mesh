package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshtests "github.com/vechain/mesh/tests"
	meshthor "github.com/vechain/mesh/thor"
)

func TestNewBlockService(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	if service == nil || service.vechainClient != mockClient {
		t.Errorf("NewBlockService() returned nil or client mismatch")
	}
}

func TestBlockService_Block_InvalidRequestBody(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/block", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call Block
	service.Block(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Block() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestBlockService_Block_ValidRequest(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	// Create request with valid block identifier
	request := types.BlockRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: func() *int64 { i := int64(100); return &i }(),
		},
	}

	req := meshtests.CreateRequestWithContext("POST", "/block", request)
	w := httptest.NewRecorder()

	// Call Block
	service.Block(w, req)

	// Should succeed with mock client
	if w.Code != http.StatusOK {
		t.Errorf("Block() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify response structure
	var response types.BlockResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Block == nil {
		t.Errorf("Block() response.Block is nil")
	}
}

func TestBlockService_BlockTransaction_InvalidRequestBody(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/block/transaction", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call BlockTransaction
	service.BlockTransaction(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("BlockTransaction() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestBlockService_BlockTransaction_ValidRequest(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	// Create request with valid block and transaction identifiers
	request := types.BlockTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(100),
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	req := meshtests.CreateRequestWithContext("POST", "/block/transaction", request)
	w := httptest.NewRecorder()

	// Call BlockTransaction
	service.BlockTransaction(w, req)

	// Should succeed with mock client
	if w.Code != http.StatusOK {
		t.Errorf("BlockTransaction() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify response structure
	var response types.BlockTransactionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Transaction == nil {
		t.Errorf("BlockTransaction() response.Transaction is nil")
	}
}

func TestBlockService_Block_WithHashIdentifier(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	// Create request with hash identifier
	request := types.BlockRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Hash: func() *string { h := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"; return &h }(),
		},
	}

	req := meshtests.CreateRequestWithContext("POST", "/block", request)
	w := httptest.NewRecorder()

	// Call Block
	service.Block(w, req)

	// Should succeed with mock client
	if w.Code != http.StatusOK {
		t.Errorf("Block() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify response structure
	var response types.BlockResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Block == nil {
		t.Errorf("Block() response.Block is nil")
	}
}

func TestBlockService_Block_WithBothIdentifiers(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	// Create request with both index and hash identifiers
	request := types.BlockRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: func() *int64 { i := int64(100); return &i }(),
			Hash:  func() *string { h := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"; return &h }(),
		},
	}

	req := meshtests.CreateRequestWithContext("POST", "/block", request)
	w := httptest.NewRecorder()

	// Call Block
	service.Block(w, req)

	// Should succeed with mock client
	if w.Code != http.StatusOK {
		t.Errorf("Block() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify response structure
	var response types.BlockResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Block == nil {
		t.Errorf("Block() response.Block is nil")
	}
}

func TestBlockService_Block_WithHashBlockIdentifier(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	// Create request with hash block identifier
	request := types.BlockTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(100),
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	req := meshtests.CreateRequestWithContext("POST", "/block/transaction", request)
	w := httptest.NewRecorder()

	// Call BlockTransaction
	service.BlockTransaction(w, req)

	// Should succeed with mock client
	if w.Code != http.StatusOK {
		t.Errorf("BlockTransaction() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify response structure
	var response types.BlockTransactionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Transaction == nil {
		t.Errorf("BlockTransaction() response.Transaction is nil")
	}
}
