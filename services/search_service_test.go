package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/stretchr/testify/assert"
	meshtests "github.com/vechain/mesh/tests"
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
	request := types.SearchTransactionsRequest{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	// Create HTTP request
	req := meshtests.CreateRequestWithContext(meshtests.POSTMethod, "/search/transactions", request)
	w := httptest.NewRecorder()

	// Call the handler
	searchService.SearchTransactions(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Parse response
	var response types.SearchTransactionsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
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

func TestSearchService_SearchTransactions_InvalidRequestBody(t *testing.T) {
	// Create mock client
	mockClient := meshthor.NewMockVeChainClient()

	// Create search service
	searchService := NewSearchService(mockClient)

	// Create invalid JSON request
	req := httptest.NewRequest(meshtests.POSTMethod, "/search/transactions", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call the handler
	searchService.SearchTransactions(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSearchService_SearchTransactions_MissingTransactionIdentifier(t *testing.T) {
	// Create mock client
	mockClient := meshthor.NewMockVeChainClient()

	// Create search service
	searchService := NewSearchService(mockClient)

	// Create request without transaction identifier
	request := types.SearchTransactionsRequest{}

	// Create HTTP request
	req := meshtests.CreateRequestWithContext(meshtests.POSTMethod, "/search/transactions", request)
	w := httptest.NewRecorder()

	// Call the handler
	searchService.SearchTransactions(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSearchService_SearchTransactions_EmptyTransactionHash(t *testing.T) {
	// Create mock client
	mockClient := meshthor.NewMockVeChainClient()

	// Create search service
	searchService := NewSearchService(mockClient)

	// Create request with empty transaction hash
	request := types.SearchTransactionsRequest{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "",
		},
	}

	// Create HTTP request
	req := meshtests.CreateRequestWithContext(meshtests.POSTMethod, "/search/transactions", request)
	w := httptest.NewRecorder()

	// Call the handler
	searchService.SearchTransactions(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSearchService_SearchTransactions_TransactionNotFound(t *testing.T) {
	// Create mock client with error
	mockClient := meshthor.NewMockVeChainClient()
	mockClient.SetMockError(assert.AnError)

	// Create search service
	searchService := NewSearchService(mockClient)

	// Create request
	request := types.SearchTransactionsRequest{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	// Create HTTP request
	req := meshtests.CreateRequestWithContext(meshtests.POSTMethod, "/search/transactions", request)
	w := httptest.NewRecorder()

	// Call the handler
	searchService.SearchTransactions(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSearchService_SearchTransactions_ThorClientError(t *testing.T) {
	// Create mock client with error
	mockClient := meshthor.NewMockVeChainClient()
	mockClient.SetMockError(assert.AnError)

	// Create search service
	searchService := NewSearchService(mockClient)

	// Create request
	request := types.SearchTransactionsRequest{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	// Create HTTP request
	req := meshtests.CreateRequestWithContext(meshtests.POSTMethod, "/search/transactions", request)
	w := httptest.NewRecorder()

	// Call the handler
	searchService.SearchTransactions(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
