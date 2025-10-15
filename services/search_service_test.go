package services

import (
	"context"
	"errors"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/math"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/api/transactions"
	"github.com/vechain/thor/v2/thor"
)

func TestSearchService_SearchTransactions_Success(t *testing.T) {
	// Create mock client
	mockClient := meshthor.NewMockVeChainClient()

	// Create mock transaction
	txHash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	mockTx := &transactions.Transaction{
		ID: txHash,
		Meta: &api.TxMeta{
			BlockNumber: 100,
			BlockID:     txHash,
		},
		Clauses: api.Clauses{
			{
				To:    &thor.Address{},
				Value: &math.HexOrDecimal256{},
				Data:  "0x",
			},
		},
	}
	mockClient.SetTransaction(mockTx)

	// Create mock transaction receipt
	mockReceipt := &api.Receipt{
		Meta: api.ReceiptMeta{
			BlockNumber: 100,
			BlockID:     txHash,
			TxID:        txHash,
			TxOrigin:    thor.Address{},
		},
		GasUsed:  21000,
		GasPayer: thor.Address{},
		Paid:     &math.HexOrDecimal256{},
		Reward:   &math.HexOrDecimal256{},
		Reverted: false,
		Outputs: []*api.Output{
			{
				ContractAddress: nil,
				Events:          []*api.Event{},
				Transfers:       []*api.Transfer{},
			},
		},
	}
	mockClient.SetReceipt(mockReceipt)

	// Create search service
	searchService := NewSearchService(mockClient)

	// Create request
	request := &types.SearchTransactionsRequest{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	ctx := context.Background()
	response, err := searchService.SearchTransactions(ctx, request)

	if err != nil {
		t.Fatalf("SearchTransactions() error = %v", err)
	}

	// Verify response structure
	if response.TotalCount != 1 {
		t.Errorf("Expected TotalCount 1, got %d", response.TotalCount)
	}

	if len(response.Transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(response.Transactions))
	}

	if response.Transactions[0].BlockIdentifier.Index != 100 {
		t.Errorf("Expected block index 100, got %d", response.Transactions[0].BlockIdentifier.Index)
	}

	if response.Transactions[0].Transaction.TransactionIdentifier.Hash != "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef" {
		t.Errorf("Expected transaction hash, got %s", response.Transactions[0].Transaction.TransactionIdentifier.Hash)
	}
}

func TestSearchService_SearchTransactions_MissingTransactionIdentifier(t *testing.T) {
	// Create mock client
	mockClient := meshthor.NewMockVeChainClient()

	// Create search service
	searchService := NewSearchService(mockClient)

	// Create request without transaction identifier
	request := &types.SearchTransactionsRequest{
		TransactionIdentifier: nil,
	}

	ctx := context.Background()
	_, err := searchService.SearchTransactions(ctx, request)

	if err == nil {
		t.Error("SearchTransactions() expected error for missing transaction identifier")
	}
}

func TestSearchService_SearchTransactions_EmptyTransactionHash(t *testing.T) {
	// Create mock client
	mockClient := meshthor.NewMockVeChainClient()

	// Create search service
	searchService := NewSearchService(mockClient)

	// Create request with empty transaction hash
	request := &types.SearchTransactionsRequest{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "",
		},
	}

	ctx := context.Background()
	_, err := searchService.SearchTransactions(ctx, request)

	if err == nil {
		t.Error("SearchTransactions() expected error for empty transaction hash")
	}
}

func TestSearchService_SearchTransactions_TransactionNotFound(t *testing.T) {
	// Create mock client
	mockClient := meshthor.NewMockVeChainClient()

	// Don't set any transaction in the mock client

	// Create search service
	searchService := NewSearchService(mockClient)

	// Create request
	request := &types.SearchTransactionsRequest{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	ctx := context.Background()
	response, err := searchService.SearchTransactions(ctx, request)

	if err != nil {
		t.Fatalf("SearchTransactions() error = %v", err)
	}

	// Verify empty response
	if response.TotalCount != 0 {
		t.Errorf("Expected TotalCount 0, got %d", response.TotalCount)
	}

	if len(response.Transactions) != 0 {
		t.Errorf("Expected 0 transactions, got %d", len(response.Transactions))
	}
}

func TestSearchService_SearchTransactions_ThorClientError(t *testing.T) {
	// Create mock client with error
	mockClient := meshthor.NewMockVeChainClient()
	mockClient.SetMockError(errors.New("thor client error"))

	// Create search service
	searchService := NewSearchService(mockClient)

	// Create request
	request := &types.SearchTransactionsRequest{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	ctx := context.Background()
	_, err := searchService.SearchTransactions(ctx, request)

	if err == nil {
		t.Error("SearchTransactions() expected error when thor client returns error")
	}
}
