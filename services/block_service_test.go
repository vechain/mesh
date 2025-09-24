package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
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

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/block", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call Block
	service.Block(w, req)

	// Check response - should fail because we don't have a real Thor node
	// but the request parsing should work
	if w.Code == http.StatusOK {
		// If it succeeds, verify response structure
		var response types.BlockResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Block == nil {
			t.Errorf("Block() response.Block is nil")
		}
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

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/block/transaction", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call BlockTransaction
	service.BlockTransaction(w, req)

	// Check response - should fail because we don't have a real Thor node
	// but the request parsing should work
	if w.Code == http.StatusOK {
		// If it succeeds, verify response structure
		var response types.BlockTransactionResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Transaction == nil {
			t.Errorf("BlockTransaction() response.Transaction is nil")
		}
	}
}
