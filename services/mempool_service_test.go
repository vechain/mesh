package services

import (
	"context"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
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

func TestMempoolService_Mempool(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	request := &types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
	}

	ctx := context.Background()
	response, err := service.Mempool(ctx, request)

	if err != nil {
		t.Fatalf("Mempool() returned error: %v", err)
	}

	if response == nil {
		t.Fatal("Mempool() returned nil response")
	}

	if response.TransactionIdentifiers == nil {
		t.Error("Mempool() TransactionIdentifiers is nil")
	}
}

func TestMempoolService_MempoolTransaction(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	// Valid transaction hash from mock
	request := &types.MempoolTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x8e5c7a4b971c2d028e3ba9ba7e56e78d1ff75ddb68d4a26a1c5e36aef70e1a3c",
		},
	}

	ctx := context.Background()
	response, err := service.MempoolTransaction(ctx, request)

	if err != nil {
		t.Fatalf("MempoolTransaction() returned error: %v", err)
	}

	if response == nil {
		t.Fatal("MempoolTransaction() returned nil response")
	}

	if response.Transaction == nil {
		t.Error("MempoolTransaction() Transaction is nil")
	}
}

func TestMempoolService_MempoolTransaction_InvalidHash(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	request := &types.MempoolTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "invalid",
		},
	}

	ctx := context.Background()
	_, err := service.MempoolTransaction(ctx, request)

	// Should return error for invalid hash format
	if err == nil {
		t.Error("MempoolTransaction() expected error for invalid hash")
	}
}
