package services

import (
	"encoding/json"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshclient "github.com/vechain/mesh/client"
)

// MempoolService handles mempool API endpoints
type MempoolService struct {
	vechainClient *meshclient.VeChainClient
}

// NewMempoolService creates a new mempool service
func NewMempoolService(vechainClient *meshclient.VeChainClient) *MempoolService {
	return &MempoolService{
		vechainClient: vechainClient,
	}
}

// Mempool gets mempool information
func (m *MempoolService) Mempool(w http.ResponseWriter, r *http.Request) {
	// For now, return an empty mempool
	// TODO: Implement actual mempool data retrieval from VeChain
	response := &types.MempoolResponse{
		TransactionIdentifiers: []*types.TransactionIdentifier{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// MempoolTransaction gets a specific transaction from the mempool
func (m *MempoolService) MempoolTransaction(w http.ResponseWriter, r *http.Request) {
	var request types.MempoolTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// For now, return a not found error
	// TODO: Implement actual mempool transaction retrieval from VeChain
	http.Error(w, "Transaction not found in mempool", http.StatusNotFound)
}
