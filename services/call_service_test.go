package services

import (
	"context"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshconfig "github.com/vechain/mesh/config"
	meshthor "github.com/vechain/mesh/thor"
)

func TestNewCallService(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{}

	service := NewCallService(mockClient, config)

	if service == nil {
		t.Fatal("NewCallService() returned nil")
	}

	if service.vechainClient == nil {
		t.Errorf("NewCallService() vechainClient is nil")
	}

	if service.config != config {
		t.Errorf("NewCallService() config = %v, want %v", service.config, config)
	}
}

func TestCallService_Call_UnsupportedMethod(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{}
	service := NewCallService(mockClient, config)

	request := &types.CallRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		Method:     "unsupported_method",
		Parameters: map[string]any{},
	}

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	// Should return error for unsupported method
	if err == nil {
		t.Error("Call() expected error for unsupported method")
	}
}

func TestCallService_Call_InspectClauses(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{}
	service := NewCallService(mockClient, config)

	request := &types.CallRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		Method: meshcommon.CallMethodInspectClauses,
		Parameters: map[string]any{
			"clauses": []any{
				map[string]any{
					"to":    "0x0000000000000000000000000000000000000000",
					"value": "0x0",
					"data":  "0x",
				},
			},
		},
	}

	ctx := context.Background()
	response, err := service.Call(ctx, request)

	if err != nil {
		t.Fatalf("Call() returned error: %v", err)
	}

	if response == nil {
		t.Fatal("Call() returned nil response")
	}

	if response.Result == nil {
		t.Error("Call() Result is nil")
	}

	if !response.Idempotent {
		t.Error("Call() Idempotent should be true for InspectClauses")
	}
}

func TestCallService_Call_MissingClauses(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{}
	service := NewCallService(mockClient, config)

	request := &types.CallRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		Method:     meshcommon.CallMethodInspectClauses,
		Parameters: map[string]any{},
	}

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	// Should return error for missing clauses
	if err == nil {
		t.Error("Call() expected error for missing clauses")
	}
}
