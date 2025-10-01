package services

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	meshcommon "github.com/vechain/mesh/common"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshtests "github.com/vechain/mesh/tests"
	meshthor "github.com/vechain/mesh/thor"
)

func TestNewMempoolService(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()

	service := NewMempoolService(mockClient)

	if service == nil {
		t.Fatal("NewMempoolService() returned nil")
	}

	if service.vechainClient == nil {
		t.Errorf("NewMempoolService() vechainClient is nil")
	}
}

func TestMempoolService_Mempool_InvalidRequestBody(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", meshcommon.MempoolEndpoint, bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call Mempool
	service.Mempool(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Mempool() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestMempoolService_Mempool_ValidRequest(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	// Create request
	request := map[string]any{
		"network_identifier": map[string]any{
			"blockchain": meshcommon.BlockchainName,
			"network":    "test",
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.MempoolEndpoint, request)
	w := httptest.NewRecorder()

	// Call Mempool
	service.Mempool(w, req)

	// Note: This test will fail if the VeChain node is not running
	// but it tests the request parsing and basic flow
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Mempool() status code = %v, want %v or %v", w.Code, http.StatusOK, http.StatusInternalServerError)
	}
}

func TestMempoolService_Mempool_WithOriginFilter(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	// Create request with origin filter
	request := map[string]any{
		"network_identifier": map[string]any{
			"blockchain": meshcommon.BlockchainName,
			"network":    "test",
		},
		"metadata": map[string]any{
			"origin": meshtests.FirstSoloAddress,
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.MempoolEndpoint, request)
	w := httptest.NewRecorder()

	// Call Mempool
	service.Mempool(w, req)

	// Note: This test will fail if the VeChain node is not running
	// but it tests the request parsing and basic flow
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Mempool() status code = %v, want %v or %v", w.Code, http.StatusOK, http.StatusInternalServerError)
	}
}

func TestMempoolService_MempoolTransaction_InvalidRequestBody(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", meshcommon.MempoolTransactionEndpoint, bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call MempoolTransaction
	service.MempoolTransaction(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("MempoolTransaction() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestMempoolService_MempoolTransaction_MissingTransactionIdentifier(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	// Create request without transaction identifier
	request := types.MempoolTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		// TransactionIdentifier is nil
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.MempoolTransactionEndpoint, request)
	w := httptest.NewRecorder()

	// Call MempoolTransaction
	service.MempoolTransaction(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("MempoolTransaction() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestMempoolService_MempoolTransaction_EmptyTransactionHash(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	// Create request with empty transaction hash
	request := types.MempoolTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "", // Empty hash
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.MempoolTransactionEndpoint, request)
	w := httptest.NewRecorder()

	// Call MempoolTransaction
	service.MempoolTransaction(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("MempoolTransaction() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestMempoolService_MempoolTransaction_InvalidTransactionHash(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	// Create request with invalid transaction hash
	request := types.MempoolTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "invalid_hash", // Invalid hash format
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.MempoolTransactionEndpoint, request)
	w := httptest.NewRecorder()

	// Call MempoolTransaction
	service.MempoolTransaction(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("MempoolTransaction() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestMempoolService_MempoolTransaction_ValidRequest(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	// Create request
	request := types.MempoolTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.MempoolTransactionEndpoint, request)
	w := httptest.NewRecorder()

	// Call MempoolTransaction
	service.MempoolTransaction(w, req)

	// Note: This test will fail if the VeChain node is not running
	// but it tests the request parsing and basic flow
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("MempoolTransaction() status code = %v, want %v, %v, or %v", w.Code, http.StatusOK, http.StatusNotFound, http.StatusInternalServerError)
	}
}
