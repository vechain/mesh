package services

import (
	"encoding/json"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshthor "github.com/vechain/mesh/thor"
	meshutils "github.com/vechain/mesh/utils"
	"github.com/vechain/thor/v2/thor"
)

// MempoolService handles mempool API endpoints
type MempoolService struct {
	vechainClient meshthor.VeChainClientInterface
}

// NewMempoolService creates a new mempool service
func NewMempoolService(vechainClient meshthor.VeChainClientInterface) *MempoolService {
	return &MempoolService{
		vechainClient: vechainClient,
	}
}

// Mempool gets mempool information
func (m *MempoolService) Mempool(w http.ResponseWriter, r *http.Request) {
	// Parse request body to get metadata (origin filter)
	var requestBody map[string]any
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrInvalidRequestBody), http.StatusBadRequest)
		return
	}

	// Parse origin address from metadata if provided
	var origin *thor.Address
	if metadata, ok := requestBody["metadata"].(map[string]any); ok {
		if originStr, ok := metadata["origin"].(string); ok && originStr != "" {
			if parsedOrigin, err := thor.ParseAddress(originStr); err == nil {
				origin = &parsedOrigin
			}
		}
	}

	// Get all pending transactions from the mempool
	txIDs, err := m.vechainClient.GetMempoolTransactions(origin)
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetErrorWithMetadata(meshutils.ErrFailedToGetMempool, map[string]any{
			"error": err.Error(),
		}), http.StatusInternalServerError)
		return
	}

	// Convert to Mesh format
	var transactionIdentifiers []*types.TransactionIdentifier
	for _, txID := range txIDs {
		transactionIdentifiers = append(transactionIdentifiers, &types.TransactionIdentifier{
			Hash: txID.String(),
		})
	}

	response := &types.MempoolResponse{
		TransactionIdentifiers: transactionIdentifiers,
	}

	meshutils.WriteJSONResponse(w, response)
}

// MempoolTransaction gets a specific transaction from the mempool
func (m *MempoolService) MempoolTransaction(w http.ResponseWriter, r *http.Request) {
	var request types.MempoolTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrInvalidRequestBody), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if request.TransactionIdentifier == nil || request.TransactionIdentifier.Hash == "" {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrInvalidTransactionIdentifier), http.StatusBadRequest)
		return
	}

	// Parse transaction hash
	txID, err := thor.ParseBytes32(request.TransactionIdentifier.Hash)
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetErrorWithMetadata(meshutils.ErrInvalidTransactionHash, map[string]any{
			"error": err.Error(),
		}), http.StatusBadRequest)
		return
	}

	// Get transaction from mempool
	tx, err := m.vechainClient.GetMempoolTransaction(&txID)
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetErrorWithMetadata(meshutils.ErrTransactionNotFoundInMempool, map[string]any{
			"error": err.Error(),
		}), http.StatusNotFound)
		return
	}

	// Parse operations directly from transactions.Transaction
	operations := meshutils.ParseTransactionOperationsFromTransactions(tx)
	meshTx := meshutils.BuildMeshTransactionFromTransactions(tx, operations)

	// Build the response
	response := &types.MempoolTransactionResponse{
		Transaction: meshTx,
	}

	meshutils.WriteJSONResponse(w, response)
}
