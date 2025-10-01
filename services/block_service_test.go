package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	meshcommon "github.com/vechain/mesh/common"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshtests "github.com/vechain/mesh/tests"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/thor"
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
	req := httptest.NewRequest("POST", meshcommon.BlockEndpoint, bytes.NewBufferString("invalid json"))
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
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: func() *int64 { i := int64(100); return &i }(),
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.BlockEndpoint, request)
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
	req := httptest.NewRequest("POST", meshcommon.BlockTransactionEndpoint, bytes.NewBufferString("invalid json"))
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
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(100),
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.BlockTransactionEndpoint, request)
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
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Hash: func() *string { h := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"; return &h }(),
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.BlockEndpoint, request)
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
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: func() *int64 { i := int64(100); return &i }(),
			Hash:  func() *string { h := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"; return &h }(),
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.BlockEndpoint, request)
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

func TestBlockService_Block_ErrorCases(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	t.Run("Invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", meshcommon.BlockEndpoint, bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		service.Block(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Block() status = %v, want %v", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("Block not found", func(t *testing.T) {
		// Set up mock to return error
		mockClient.SetMockError(fmt.Errorf("block not found"))

		request := types.BlockRequest{
			NetworkIdentifier: &types.NetworkIdentifier{
				Blockchain: meshcommon.BlockchainName,
				Network:    "test",
			},
			BlockIdentifier: &types.PartialBlockIdentifier{
				Index: func() *int64 { i := int64(12345); return &i }(),
			},
		}

		req := meshtests.CreateRequestWithContext("POST", meshcommon.BlockEndpoint, request)
		w := httptest.NewRecorder()

		service.Block(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Block() status = %v, want %v", w.Code, http.StatusBadRequest)
		}
	})
}

func TestBlockService_Block_WithHashBlockIdentifier(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	// Create request with hash block identifier
	request := types.BlockTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(100),
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.BlockTransactionEndpoint, request)
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

func TestBlockService_getBlockByIdentifier(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	t.Run("Get block by number", func(t *testing.T) {
		// Set up mock block
		mockBlock := &api.JSONExpandedBlock{
			JSONBlockSummary: &api.JSONBlockSummary{
				Number: 100,
				ID:     thor.Bytes32{},
			},
		}
		mockClient.SetMockBlock(mockBlock)

		// Test getting block by number
		blockIdentifier := types.BlockIdentifier{
			Index: 100,
		}
		block, err := service.getBlockByIdentifier(blockIdentifier)
		if err != nil {
			t.Errorf("getBlockByIdentifier() error = %v, want nil", err)
		}
		if block == nil {
			t.Errorf("getBlockByIdentifier() returned nil block")
		}
	})

	t.Run("Get block by hash", func(t *testing.T) {
		// Set up mock block
		mockBlock := &api.JSONExpandedBlock{
			JSONBlockSummary: &api.JSONBlockSummary{
				Number: 100,
				ID:     thor.Bytes32{},
			},
		}
		mockClient.SetMockBlock(mockBlock)

		// Test getting block by hash
		blockIdentifier := types.BlockIdentifier{
			Hash: "0x1234567890abcdef",
		}
		block, err := service.getBlockByIdentifier(blockIdentifier)
		if err != nil {
			t.Errorf("getBlockByIdentifier() error = %v, want nil", err)
		}
		if block == nil {
			t.Errorf("getBlockByIdentifier() returned nil block")
		}
	})

	t.Run("Error case - block not found", func(t *testing.T) {
		// Set up mock error
		mockClient.SetMockError(fmt.Errorf("block not found"))

		// Test error case
		blockIdentifier := types.BlockIdentifier{
			Index: 999999,
		}
		block, err := service.getBlockByIdentifier(blockIdentifier)
		if err == nil {
			t.Errorf("getBlockByIdentifier() should return error when block not found")
		}
		if block != nil {
			t.Errorf("getBlockByIdentifier() should return nil block when error occurs")
		}
	})

	t.Run("Error case - invalid identifier", func(t *testing.T) {
		// Test with empty identifier
		blockIdentifier := types.BlockIdentifier{}
		block, err := service.getBlockByIdentifier(blockIdentifier)
		if err == nil {
			t.Errorf("getBlockByIdentifier() should return error for invalid identifier")
		}
		if block != nil {
			t.Errorf("getBlockByIdentifier() should return nil block for invalid identifier")
		}
	})
}

func TestBlockService_getBlockByPartialIdentifier(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	t.Run("Get block by number", func(t *testing.T) {
		// Set up mock block
		mockBlock := &api.JSONExpandedBlock{
			JSONBlockSummary: &api.JSONBlockSummary{
				Number: 100,
				ID:     thor.Bytes32{},
			},
		}
		mockClient.SetMockBlock(mockBlock)

		// Test getting block by number
		index := int64(100)
		blockIdentifier := types.PartialBlockIdentifier{
			Index: &index,
		}
		block, err := service.getBlockByPartialIdentifier(blockIdentifier)
		if err != nil {
			t.Errorf("getBlockByPartialIdentifier() error = %v, want nil", err)
		}
		if block == nil {
			t.Errorf("getBlockByPartialIdentifier() returned nil block")
		}
	})

	t.Run("Get block by hash", func(t *testing.T) {
		// Set up mock block
		mockBlock := &api.JSONExpandedBlock{
			JSONBlockSummary: &api.JSONBlockSummary{
				Number: 100,
				ID:     thor.Bytes32{},
			},
		}
		mockClient.SetMockBlock(mockBlock)

		// Test getting block by hash
		hash := "0x1234567890abcdef"
		blockIdentifier := types.PartialBlockIdentifier{
			Hash: &hash,
		}
		block, err := service.getBlockByPartialIdentifier(blockIdentifier)
		if err != nil {
			t.Errorf("getBlockByPartialIdentifier() error = %v, want nil", err)
		}
		if block == nil {
			t.Errorf("getBlockByPartialIdentifier() returned nil block")
		}
	})

	t.Run("Error case - invalid identifier", func(t *testing.T) {
		// Test with empty identifier
		blockIdentifier := types.PartialBlockIdentifier{}
		block, err := service.getBlockByPartialIdentifier(blockIdentifier)
		if err == nil {
			t.Errorf("getBlockByPartialIdentifier() should return error for invalid identifier")
		}
		if block != nil {
			t.Errorf("getBlockByPartialIdentifier() should return nil block for invalid identifier")
		}
	})

	t.Run("Error case - block not found", func(t *testing.T) {
		// Set up mock error
		mockClient.SetMockError(fmt.Errorf("block not found"))

		// Test error case
		index := int64(999999)
		blockIdentifier := types.PartialBlockIdentifier{
			Index: &index,
		}
		block, err := service.getBlockByPartialIdentifier(blockIdentifier)
		if err == nil {
			t.Errorf("getBlockByPartialIdentifier() should return error when block not found")
		}
		if block != nil {
			t.Errorf("getBlockByPartialIdentifier() should return nil block when error occurs")
		}
	})
}
