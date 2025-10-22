package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
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

// Request body validation is handled by SDK asserter

func TestBlockService_Block_ValidRequest(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	// Create request with valid block identifier
	request := &types.BlockRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: func() *int64 { i := int64(100); return &i }(),
		},
	}

	ctx := context.Background()
	response, err := service.Block(ctx, request)

	if err != nil {
		t.Fatalf("Block() error = %v", err)
	}

	if response.Block == nil {
		t.Errorf("Block() response.Block is nil")
	}
}

func TestBlockService_BlockTransaction_ValidRequest(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	// Create request with valid block and transaction identifiers
	request := &types.BlockTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(100),
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	ctx := context.Background()
	response, err := service.BlockTransaction(ctx, request)

	if err != nil {
		t.Fatalf("BlockTransaction() error = %v", err)
	}

	if response.Transaction == nil {
		t.Errorf("BlockTransaction() response.Transaction is nil")
	}
}

func TestBlockService_BlockTransaction_InvalidBlockIdentifier(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	request := &types.BlockTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		BlockIdentifier: &types.BlockIdentifier{},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	ctx := context.Background()
	_, err := service.BlockTransaction(ctx, request)

	if err == nil {
		t.Error("BlockTransaction() expected error for invalid block identifier")
	}

	if err != nil && err.Code != int32(meshcommon.ErrBlockNotFound) {
		t.Errorf("BlockTransaction() error code = %d, want %d", err.Code, meshcommon.ErrBlockNotFound)
	}
}

func TestBlockService_BlockTransaction_NilTransactionIdentifier(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	request := &types.BlockTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(100),
		},
		TransactionIdentifier: nil,
	}

	ctx := context.Background()
	_, err := service.BlockTransaction(ctx, request)

	if err == nil {
		t.Error("BlockTransaction() expected error for nil transaction identifier")
	}

	if err != nil && err.Code != int32(meshcommon.ErrInvalidRequestBody) {
		t.Errorf("BlockTransaction() error code = %d, want %d", err.Code, meshcommon.ErrInvalidRequestBody)
	}
}

func TestBlockService_BlockTransaction_EmptyTransactionHash(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	request := &types.BlockTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(100),
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "",
		},
	}

	ctx := context.Background()
	_, err := service.BlockTransaction(ctx, request)

	if err == nil {
		t.Error("BlockTransaction() expected error for empty transaction hash")
	}

	if err != nil && err.Code != int32(meshcommon.ErrInvalidRequestBody) {
		t.Errorf("BlockTransaction() error code = %d, want %d", err.Code, meshcommon.ErrInvalidRequestBody)
	}
}

func TestBlockService_BlockTransaction_BlockNotFound(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	mockClient.SetMockError(errors.New("block not found"))

	request := &types.BlockTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(999999),
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	ctx := context.Background()
	_, err := service.BlockTransaction(ctx, request)

	if err == nil {
		t.Error("BlockTransaction() expected error when block not found")
	}

	if err != nil && err.Code != int32(meshcommon.ErrBlockNotFound) {
		t.Errorf("BlockTransaction() error code = %d, want %d", err.Code, meshcommon.ErrBlockNotFound)
	}
}

func TestBlockService_BlockTransaction_TransactionNotFoundInBlock(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	request := &types.BlockTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(100),
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		},
	}

	ctx := context.Background()
	_, err := service.BlockTransaction(ctx, request)

	if err == nil {
		t.Error("BlockTransaction() expected error when transaction not found in block")
	}

	if err != nil && err.Code != int32(meshcommon.ErrTransactionNotFound) {
		t.Errorf("BlockTransaction() error code = %d, want %d", err.Code, meshcommon.ErrTransactionNotFound)
	}
}

func TestBlockService_Block_WithHashIdentifier(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	// Create request with hash identifier
	request := &types.BlockRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Hash: func() *string { h := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"; return &h }(),
		},
	}

	ctx := context.Background()
	response, err := service.Block(ctx, request)

	if err != nil {
		t.Fatalf("Block() error = %v", err)
	}

	if response.Block == nil {
		t.Errorf("Block() response.Block is nil")
	}
}

func TestBlockService_Block_WithBothIdentifiers(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	// Create request with both index and hash identifiers
	request := &types.BlockRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: func() *int64 { i := int64(100); return &i }(),
			Hash:  func() *string { h := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"; return &h }(),
		},
	}

	ctx := context.Background()
	response, err := service.Block(ctx, request)

	if err != nil {
		t.Fatalf("Block() error = %v", err)
	}

	if response.Block == nil {
		t.Errorf("Block() response.Block is nil")
	}
}

func TestBlockService_Block_ErrorCases(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	t.Run("Block not found", func(t *testing.T) {
		// Set up mock to return error
		mockClient.SetMockError(fmt.Errorf("block not found"))

		request := &types.BlockRequest{
			NetworkIdentifier: &types.NetworkIdentifier{
				Blockchain: meshcommon.BlockchainName,
				Network:    meshcommon.TestNetwork,
			},
			BlockIdentifier: &types.PartialBlockIdentifier{
				Index: func() *int64 { i := int64(12345); return &i }(),
			},
		}

		ctx := context.Background()
		_, err := service.Block(ctx, request)

		if err == nil {
			t.Error("Block() expected error but got none")
		}
	})
}

func TestBlockService_Block_TimestampOverflow(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	// Force block timestamp such that timestamp*1000 > MaxInt64
	if mockClient.MockBlock != nil && mockClient.MockBlock.JSONBlockSummary != nil {
		mockClient.MockBlock.Timestamp = uint64(math.MaxInt64/1000 + 1)
	}

	request := &types.BlockRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: func() *int64 { i := int64(100); return &i }(),
		},
	}

	_, err := service.Block(context.Background(), request)
	if err == nil {
		t.Fatalf("Block() expected error on timestamp overflow")
	}
	if err.Code != int32(meshcommon.ErrInternalServerError) {
		t.Errorf("Block() error code = %d, want %d", err.Code, meshcommon.ErrInternalServerError)
	}
}

func TestBlockService_Block_WithHashBlockIdentifier(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	// Create request with hash block identifier
	request := &types.BlockTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(100),
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	ctx := context.Background()
	response, err := service.BlockTransaction(ctx, request)

	if err != nil {
		t.Fatalf("BlockTransaction() error = %v", err)
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
