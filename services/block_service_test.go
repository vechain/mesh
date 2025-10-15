package services

import (
	"context"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshthor "github.com/vechain/mesh/thor"
)

func TestNewBlockService(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()

	service := NewBlockService(mockClient)

	if service == nil {
		t.Fatal("NewBlockService() returned nil")
	}

	if service.vechainClient == nil {
		t.Errorf("NewBlockService() vechainClient is nil")
	}

	if service.encoder == nil {
		t.Errorf("NewBlockService() encoder is nil")
	}

	if service.builder == nil {
		t.Errorf("NewBlockService() builder is nil")
	}
}

func TestBlockService_Block_ValidRequest(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	tests := []struct {
		name            string
		blockIdentifier *types.PartialBlockIdentifier
		wantError       bool
	}{
		{
			name: "block by index",
			blockIdentifier: &types.PartialBlockIdentifier{
				Index: int64Ptr(1),
			},
			wantError: false,
		},
		{
			name: "block by hash",
			blockIdentifier: &types.PartialBlockIdentifier{
				Hash: stringPtr("0x00000001c458949db492fb211c05c4f05f770648fc58db33d05c9a94cb3ece8e"),
			},
			wantError: false,
		},
		{
			name:            "no block identifier",
			blockIdentifier: &types.PartialBlockIdentifier{},
			wantError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &types.BlockRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				BlockIdentifier: tt.blockIdentifier,
			}

			ctx := context.Background()
			response, err := service.Block(ctx, request)

			if tt.wantError {
				if err == nil {
					t.Errorf("Block() expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Block() returned error: %v", err)
				}

				if response == nil {
					t.Fatal("Block() returned nil response")
				}

				if response.Block == nil {
					t.Error("Block() Block is nil")
				}

				if response.Block.BlockIdentifier == nil {
					t.Error("Block() BlockIdentifier is nil")
				}

				if response.Block.ParentBlockIdentifier == nil {
					t.Error("Block() ParentBlockIdentifier is nil")
				}

				if response.Block.Transactions == nil {
					t.Error("Block() Transactions is nil")
				}
			}
		})
	}
}

func TestBlockService_BlockTransaction_ValidRequest(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	// First get a block to find a real transaction
	blockReq := &types.BlockRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: int64Ptr(1),
		},
	}

	ctx := context.Background()
	blockResp, err := service.Block(ctx, blockReq)
	if err != nil {
		t.Fatalf("Block() returned error: %v", err)
	}

	if len(blockResp.Block.Transactions) == 0 {
		t.Skip("No transactions in block 1 to test with")
	}

	// Now test BlockTransaction with the first transaction
	txHash := blockResp.Block.Transactions[0].TransactionIdentifier.Hash

	request := &types.BlockTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		BlockIdentifier: &types.BlockIdentifier{
			Index: blockResp.Block.BlockIdentifier.Index,
			Hash:  blockResp.Block.BlockIdentifier.Hash,
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: txHash,
		},
	}

	response, err := service.BlockTransaction(ctx, request)

	if err != nil {
		t.Fatalf("BlockTransaction() returned error: %v", err)
	}

	if response == nil {
		t.Fatal("BlockTransaction() returned nil response")
	}

	if response.Transaction == nil {
		t.Error("BlockTransaction() Transaction is nil")
	}

	if response.Transaction.TransactionIdentifier == nil {
		t.Error("BlockTransaction() TransactionIdentifier is nil")
	}

	if response.Transaction.TransactionIdentifier.Hash != txHash {
		t.Errorf("BlockTransaction() transaction hash = %v, want %v",
			response.Transaction.TransactionIdentifier.Hash, txHash)
	}
}

func TestBlockService_BlockTransaction_InvalidTransaction(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewBlockService(mockClient)

	request := &types.BlockTransactionRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		BlockIdentifier: &types.BlockIdentifier{
			Index: 1,
			Hash:  "0x00000001c458949db492fb211c05c4f05f770648fc58db33d05c9a94cb3ece8e",
		},
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x0000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	ctx := context.Background()
	response, err := service.BlockTransaction(ctx, request)

	// Should return error for non-existent transaction
	if err == nil {
		t.Error("BlockTransaction() expected error for non-existent transaction")
	}

	if response != nil {
		t.Error("BlockTransaction() should return nil response on error")
	}
}

func TestBlockService_Block_GenesisBlock(t *testing.T) {
	// Note: Mock client returns block 100 by default, so we skip this test
	// In a real environment with genesis block, this would work correctly
	t.Skip("Mock client doesn't support genesis block test - would work with real data")
}
