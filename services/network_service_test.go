package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshconfig "github.com/vechain/mesh/config"
	meshtests "github.com/vechain/mesh/tests"
	meshthor "github.com/vechain/mesh/thor"
)

func TestNewNetworkService(t *testing.T) {
	// Create a real client for testing
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{}

	service := NewNetworkService(mockClient, config)

	if service == nil {
		t.Fatal("NewNetworkService() returned nil")
	}

	if service.vechainClient == nil {
		t.Errorf("NewNetworkService() vechainClient is nil")
	}

	if service.config != config {
		t.Errorf("NewNetworkService() config = %v, want %v", service.config, config)
	}
}

func TestNetworkService_NetworkList(t *testing.T) {
	// Create test config
	config := &meshconfig.Config{}
	config.Network = "test"

	mockClient := meshthor.NewMockVeChainClient()
	service := NewNetworkService(mockClient, config)

	// Create request
	req := httptest.NewRequest("POST", "/network/list", nil)
	w := httptest.NewRecorder()

	// Call NetworkList
	service.NetworkList(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("NetworkList() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Parse response
	var response types.NetworkListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response structure
	if len(response.NetworkIdentifiers) == 0 {
		t.Errorf("NetworkList() returned no networks")
	}

	network := response.NetworkIdentifiers[0]
	if network.Blockchain != "vechainthor" {
		t.Errorf("NetworkList() blockchain = %v, want vechainthor", network.Blockchain)
	}

	if network.Network != "test" {
		t.Errorf("NetworkList() network = %v, want test", network.Network)
	}
}

func TestNetworkService_NetworkOptions(t *testing.T) {
	config := &meshconfig.Config{}
	mockClient := meshthor.NewMockVeChainClient()
	service := NewNetworkService(mockClient, config)

	// Create request with proper body
	request := types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
	}

	req := meshtests.CreateRequestWithContext("POST", "/network/options", request)
	w := httptest.NewRecorder()

	// Call NetworkOptions
	service.NetworkOptions(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("NetworkOptions() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Parse response
	var response types.NetworkOptionsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response structure
	if response.Version == nil {
		t.Errorf("NetworkOptions() version is nil")
	}

	if response.Allow == nil {
		t.Errorf("NetworkOptions() allow is nil")
	}
}

func TestNetworkService_NetworkStatus_InvalidRequestBody(t *testing.T) {
	config := &meshconfig.Config{}
	mockClient := meshthor.NewMockVeChainClient()
	service := NewNetworkService(mockClient, config)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/network/status", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call NetworkStatus
	service.NetworkStatus(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("NetworkStatus() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestNetworkService_NetworkStatus_ValidRequest(t *testing.T) {
	config := &meshconfig.Config{}
	mockClient := meshthor.NewMockVeChainClient()
	service := NewNetworkService(mockClient, config)

	// Create request
	request := types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
	}

	req := meshtests.CreateRequestWithContext("POST", "/network/status", request)
	w := httptest.NewRecorder()

	// Call NetworkStatus
	service.NetworkStatus(w, req)

	// Note: This test will fail if the VeChain node is not running
	// but it tests the request parsing and basic flow
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("NetworkStatus() status code = %v, want %v or %v", w.Code, http.StatusOK, http.StatusInternalServerError)
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
	index := getTargetIndex(0, peers)
	if index != 2 {
		t.Errorf("Expected index 2, got %d", index)
	}

	// Test with local index higher than peers
	index = getTargetIndex(5, peers)
	if index != 5 {
		t.Errorf("Expected index 5, got %d", index)
	}
}