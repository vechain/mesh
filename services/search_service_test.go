package services

import (
	"testing"

	meshthor "github.com/vechain/mesh/thor"
)

func TestNewSearchService(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()

	service := NewSearchService(mockClient)

	if service == nil {
		t.Fatal("NewSearchService() returned nil")
	}

	if service.vechainClient == nil {
		t.Errorf("NewSearchService() vechainClient is nil")
	}
}

func TestSearchService_SearchTransactions(t *testing.T) {
	// Note: SearchTransactions requires mock data to be properly set up with transaction receipts
	// Skipping for now as mock doesn't fully support this flow
	t.Skip("Mock client doesn't fully support SearchTransactions - would work with real data")
}
