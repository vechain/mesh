package services

import (
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshthor "github.com/vechain/mesh/thor"
	meshutils "github.com/vechain/mesh/utils"
)

// SearchService handles search API endpoints
type SearchService struct {
	vechainClient meshthor.VeChainClientInterface
	encoder       *meshutils.MeshTransactionEncoder
}

// NewSearchService creates a new search service
func NewSearchService(vechainClient meshthor.VeChainClientInterface) *SearchService {
	return &SearchService{
		vechainClient: vechainClient,
		encoder:       meshutils.NewMeshTransactionEncoder(vechainClient),
	}
}

// SearchTransactions handles the /search/transactions endpoint
func (s *SearchService) SearchTransactions(w http.ResponseWriter, r *http.Request) {
	var request types.SearchTransactionsRequest
	if err := meshutils.ParseJSONFromRequestContext(r, &request); err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrInvalidRequestBody), http.StatusBadRequest)
		return
	}

	// Validate transaction identifier
	if request.TransactionIdentifier == nil || request.TransactionIdentifier.Hash == "" {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrInvalidRequestParameters), http.StatusBadRequest)
		return
	}

	txID := request.TransactionIdentifier.Hash

	// Get transaction to get the clauses
	tx, err := s.vechainClient.GetTransaction(txID)
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrTransactionNotFound), http.StatusBadRequest)
		return
	}

	// Get transaction receipt to check status
	txReceipt, err := s.vechainClient.GetTransactionReceipt(txID)
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrTransactionNotFound), http.StatusBadRequest)
		return
	}

	// Create block identifier
	blockIdentifier := &types.BlockIdentifier{
		Index: int64(txReceipt.Meta.BlockNumber),
		Hash:  txReceipt.Meta.BlockID.String(),
	}

	// Convert transaction to operations
	status := meshutils.OperationStatusSucceeded
	if txReceipt.Reverted {
		status = meshutils.OperationStatusReverted
	}
	operations := s.encoder.ParseTransactionOperationsFromTransactionClauses(tx.Clauses, tx.Origin.String(), txReceipt.GasUsed, &status)

	// Create transaction identifier
	transactionIdentifier := &types.TransactionIdentifier{
		Hash: txID,
	}

	// Create transaction with operations
	transaction := &types.Transaction{
		TransactionIdentifier: transactionIdentifier,
		Operations:            operations,
	}

	// Create block transaction
	blockTransaction := &types.BlockTransaction{
		BlockIdentifier: blockIdentifier,
		Transaction:     transaction,
	}

	// Create response
	response := types.SearchTransactionsResponse{
		Transactions: []*types.BlockTransaction{blockTransaction},
		TotalCount:   1,
	}

	meshutils.WriteJSONResponse(w, response)
}
