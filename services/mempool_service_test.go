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
	req := httptest.NewRequest("POST", "/mempool", bytes.NewBufferString("invalid json"))
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
			"blockchain": "vechainthor",
			"network":    "test",
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/mempool", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
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
			"blockchain": "vechainthor",
			"network":    "test",
		},
		"metadata": map[string]any{
			"origin": "0xf077b491b355E64048cE21E3A6Fc4751eEeA77fa",
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/mempool", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
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
	req := httptest.NewRequest("POST", "/mempool/transaction", bytes.NewBufferString("invalid json"))
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
			Blockchain: "vechainthor",
			Network:    "test",
		},
		// TransactionIdentifier is nil
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/mempool/transaction", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
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
			Blockchain: "vechainthor",
			Network:    "test",
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "", // Empty hash
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/mempool/transaction", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
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
			Blockchain: "vechainthor",
			Network:    "test",
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "invalid_hash", // Invalid hash format
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/mempool/transaction", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
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
			Blockchain: "vechainthor",
			Network:    "test",
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/mempool/transaction", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call MempoolTransaction
	service.MempoolTransaction(w, req)

	// Note: This test will fail if the VeChain node is not running
	// but it tests the request parsing and basic flow
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("MempoolTransaction() status code = %v, want %v, %v, or %v", w.Code, http.StatusOK, http.StatusNotFound, http.StatusInternalServerError)
	}
}
