package services

import (
	"context"
	"math"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshconfig "github.com/vechain/mesh/config"
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
	config.Network = meshcommon.TestNetwork

	mockClient := meshthor.NewMockVeChainClient()
	service := NewNetworkService(mockClient, config)

	// Create request
	request := &types.MetadataRequest{}

	ctx := context.Background()
	response, err := service.NetworkList(ctx, request)

	if err != nil {
		t.Fatalf("NetworkList() error = %v", err)
	}

	// Verify response structure
	if len(response.NetworkIdentifiers) == 0 {
		t.Errorf("NetworkList() returned no networks")
	}

	network := response.NetworkIdentifiers[0]
	if network.Blockchain != meshcommon.BlockchainName {
		t.Errorf("NetworkList() blockchain = %v, want vechainthor", network.Blockchain)
	}

	if network.Network != meshcommon.TestNetwork {
		t.Errorf("NetworkList() network = %v, want test", network.Network)
	}
}

func TestNetworkService_NetworkOptions(t *testing.T) {
	config := &meshconfig.Config{}
	mockClient := meshthor.NewMockVeChainClient()
	service := NewNetworkService(mockClient, config)

	// Create request with proper body
	request := &types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
	}

	ctx := context.Background()
	response, err := service.NetworkOptions(ctx, request)

	if err != nil {
		t.Fatalf("NetworkOptions() error = %v", err)
	}

	// Verify response structure
	if response.Version == nil {
		t.Errorf("NetworkOptions() version is nil")
	}

	if response.Allow == nil {
		t.Errorf("NetworkOptions() allow is nil")
	}
}

func TestNetworkService_NetworkStatus_ValidRequest(t *testing.T) {
	config := &meshconfig.Config{}
	mockClient := meshthor.NewMockVeChainClient()
	service := NewNetworkService(mockClient, config)

	// Create request
	request := &types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
	}

	ctx := context.Background()
	response, err := service.NetworkStatus(ctx, request)

	// Note: This test will succeed with mock client
	if err != nil {
		t.Fatalf("NetworkStatus() error = %v", err)
	}

	if response == nil {
		t.Error("NetworkStatus() returned nil response")
	}
	if response.SyncStatus == nil || response.SyncStatus.Synced == nil || !*response.SyncStatus.Synced {
		t.Errorf("NetworkStatus() expected synced=true, got %+v", response.SyncStatus)
	}
	if response.CurrentBlockTimestamp <= 0 {
		t.Errorf("NetworkStatus() expected positive CurrentBlockTimestamp, got %d", response.CurrentBlockTimestamp)
	}
}

func TestNetworkService_NetworkStatus_TimestampOverflow(t *testing.T) {
	config := &meshconfig.Config{}
	mockClient := meshthor.NewMockVeChainClient()
	// Force best block timestamp such that timestamp*1000 > MaxInt64
	if mockClient.MockBlock != nil && mockClient.MockBlock.JSONBlockSummary != nil {
		mockClient.MockBlock.JSONBlockSummary.Timestamp = uint64(math.MaxInt64/1000 + 1)
	}
	service := NewNetworkService(mockClient, config)

	request := &types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
	}

	ctx := context.Background()
	_, err := service.NetworkStatus(ctx, request)
	if err == nil {
		t.Fatalf("NetworkStatus() expected error for timestamp overflow")
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
