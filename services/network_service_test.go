package services

import (
	"context"
	"testing"

	meshcommon "github.com/vechain/mesh/common"

	"github.com/coinbase/rosetta-sdk-go/types"
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
	config.Network = "test"

	mockClient := meshthor.NewMockVeChainClient()
	service := NewNetworkService(mockClient, config)

	// Create request
	req := &types.MetadataRequest{}

	// Call NetworkList
	ctx := context.Background()
	response, err := service.NetworkList(ctx, req)

	// Check error
	if err != nil {
		t.Fatalf("NetworkList() returned error: %v", err)
	}

	// Verify response structure
	if len(response.NetworkIdentifiers) == 0 {
		t.Errorf("NetworkList() returned no networks")
	}

	network := response.NetworkIdentifiers[0]
	if network.Blockchain != meshcommon.BlockchainName {
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

	// Create request
	request := &types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
	}

	// Call NetworkOptions
	ctx := context.Background()
	response, err := service.NetworkOptions(ctx, request)

	// Check error
	if err != nil {
		t.Fatalf("NetworkOptions() returned error: %v", err)
	}

	// Verify response structure
	if response.Version == nil {
		t.Errorf("NetworkOptions() Version is nil")
	}

	if response.Allow == nil {
		t.Errorf("NetworkOptions() Allow is nil")
	}

	// Check operation types
	expectedOpTypes := []string{
		meshcommon.OperationTypeTransfer,
		meshcommon.OperationTypeFee,
		meshcommon.OperationTypeFeeDelegation,
		meshcommon.OperationTypeContractCall,
	}

	if len(response.Allow.OperationTypes) != len(expectedOpTypes) {
		t.Errorf("NetworkOptions() OperationTypes length = %v, want %v", len(response.Allow.OperationTypes), len(expectedOpTypes))
	}

	// Check operation statuses
	expectedOpStatuses := 2 // Succeeded, Reverted
	if len(response.Allow.OperationStatuses) != expectedOpStatuses {
		t.Errorf("NetworkOptions() OperationStatuses length = %v, want %v", len(response.Allow.OperationStatuses), expectedOpStatuses)
	}

	// Check balance exemptions
	if len(response.Allow.BalanceExemptions) == 0 {
		t.Errorf("NetworkOptions() BalanceExemptions is empty")
	}

	// Verify VTHO has dynamic exemption
	vthoExemption := response.Allow.BalanceExemptions[0]
	if vthoExemption.Currency.Symbol != meshcommon.VTHOCurrency.Symbol {
		t.Errorf("NetworkOptions() first exemption currency = %v, want %v", vthoExemption.Currency.Symbol, meshcommon.VTHOCurrency.Symbol)
	}

	if vthoExemption.ExemptionType != types.BalanceDynamic {
		t.Errorf("NetworkOptions() VTHO exemption type = %v, want %v", vthoExemption.ExemptionType, types.BalanceDynamic)
	}

	// Check historical balance lookup
	if !response.Allow.HistoricalBalanceLookup {
		t.Errorf("NetworkOptions() HistoricalBalanceLookup = false, want true")
	}

	// Check call methods
	if len(response.Allow.CallMethods) == 0 {
		t.Errorf("NetworkOptions() CallMethods is empty")
	}
}

func TestNetworkService_NetworkStatus(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{}
	service := NewNetworkService(mockClient, config)

	// Create request
	request := &types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
	}

	// Call NetworkStatus
	ctx := context.Background()
	response, err := service.NetworkStatus(ctx, request)

	// Check error
	if err != nil {
		t.Fatalf("NetworkStatus() returned error: %v", err)
	}

	// Verify response structure
	if response.CurrentBlockIdentifier == nil {
		t.Errorf("NetworkStatus() CurrentBlockIdentifier is nil")
	}

	if response.GenesisBlockIdentifier == nil {
		t.Errorf("NetworkStatus() GenesisBlockIdentifier is nil")
	}

	if response.SyncStatus == nil {
		t.Errorf("NetworkStatus() SyncStatus is nil")
	}

	// Check that genesis block exists (mock returns block 100 as genesis)
	if response.GenesisBlockIdentifier.Index < 0 {
		t.Errorf("NetworkStatus() GenesisBlockIdentifier.Index = %v, should be >= 0", response.GenesisBlockIdentifier.Index)
	}

	// Check sync status fields
	if response.SyncStatus.CurrentIndex == nil {
		t.Errorf("NetworkStatus() SyncStatus.CurrentIndex is nil")
	}

	if response.SyncStatus.TargetIndex == nil {
		t.Errorf("NetworkStatus() SyncStatus.TargetIndex is nil")
	}

	if response.SyncStatus.Synced == nil {
		t.Errorf("NetworkStatus() SyncStatus.Synced is nil")
	}

	// Check peers
	if response.Peers == nil {
		t.Errorf("NetworkStatus() Peers is nil")
	}
}
