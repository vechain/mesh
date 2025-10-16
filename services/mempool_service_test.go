package services

import (
	"context"
	"errors"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
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

func TestMempoolService_Mempool_ValidRequest(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	// Create request
	request := &types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
	}

	ctx := context.Background()
	response, err := service.Mempool(ctx, request)

	if err != nil {
		t.Fatalf("Mempool() error = %v", err)
	}

	if response == nil {
		t.Error("Mempool() returned nil response")
	}
}

func TestMempoolService_Mempool_WithOriginFilter(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	// Create request with origin filter
	request := &types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		Metadata: map[string]any{
			"origin": meshtests.FirstSoloAddress,
		},
	}

	ctx := context.Background()
	response, err := service.Mempool(ctx, request)

	if err != nil {
		t.Fatalf("Mempool() error = %v", err)
	}

	if response == nil {
		t.Error("Mempool() returned nil response")
	}
}

func TestMempoolService_MempoolTransaction_ValidRequest(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	// Create request
	request := &types.MempoolTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	ctx := context.Background()
	response, err := service.MempoolTransaction(ctx, request)

	// This should return an error because VeChain doesn't support this operation
	if err != nil {
		t.Errorf("MempoolTransaction() should not return an error %v", err)
	}

	if response == nil {
		t.Error("MempoolTransaction() should not return nil response")
	}
}

func TestMempoolService_Mempool_ClientError(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	// Configure mock to return error
	mockClient.SetMockError(errors.New("failed to get mempool"))

	request := &types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
	}

	ctx := context.Background()
	_, err := service.Mempool(ctx, request)

	if err == nil {
		t.Error("Mempool() expected error when client fails")
	}

	if err != nil && err.Code != int32(meshcommon.ErrFailedToGetMempool) {
		t.Errorf("Mempool() error code = %d, want %d", err.Code, meshcommon.ErrFailedToGetMempool)
	}
}

func TestMempoolService_MempoolTransaction_NilTransactionIdentifier(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	request := &types.MempoolTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		TransactionIdentifier: nil,
	}

	ctx := context.Background()
	_, err := service.MempoolTransaction(ctx, request)

	if err == nil {
		t.Error("MempoolTransaction() expected error for nil transaction identifier")
	}

	if err != nil && err.Code != int32(meshcommon.ErrInvalidTransactionIdentifier) {
		t.Errorf("MempoolTransaction() error code = %d, want %d", err.Code, meshcommon.ErrInvalidTransactionIdentifier)
	}
}

func TestMempoolService_MempoolTransaction_EmptyHash(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	request := &types.MempoolTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "",
		},
	}

	ctx := context.Background()
	_, err := service.MempoolTransaction(ctx, request)

	if err == nil {
		t.Error("MempoolTransaction() expected error for empty hash")
	}

	if err != nil && err.Code != int32(meshcommon.ErrInvalidTransactionIdentifier) {
		t.Errorf("MempoolTransaction() error code = %d, want %d", err.Code, meshcommon.ErrInvalidTransactionIdentifier)
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
			Hash: "0xINVALID",
		},
	}

	ctx := context.Background()
	_, err := service.MempoolTransaction(ctx, request)

	if err == nil {
		t.Error("MempoolTransaction() expected error for invalid hash")
	}

	if err != nil && err.Code != int32(meshcommon.ErrInvalidTransactionHash) {
		t.Errorf("MempoolTransaction() error code = %d, want %d", err.Code, meshcommon.ErrInvalidTransactionHash)
	}
}

func TestMempoolService_MempoolTransaction_TransactionNotFound(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewMempoolService(mockClient)

	// Configure mock to return error for transaction not found
	mockClient.SetMockError(errors.New("transaction not found in mempool"))

	request := &types.MempoolTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	ctx := context.Background()
	_, err := service.MempoolTransaction(ctx, request)

	if err == nil {
		t.Error("MempoolTransaction() expected error when transaction not found")
	}

	if err != nil && err.Code != int32(meshcommon.ErrTransactionNotFoundInMempool) {
		t.Errorf("MempoolTransaction() error code = %d, want %d", err.Code, meshcommon.ErrTransactionNotFoundInMempool)
	}
}
