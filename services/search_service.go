package services

import (
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshhttp "github.com/vechain/mesh/common/http"
	meshoperations "github.com/vechain/mesh/common/operations"
	meshtx "github.com/vechain/mesh/common/tx"
	meshthor "github.com/vechain/mesh/thor"
)

// SearchService handles search API endpoints
type SearchService struct {
	requestHandler  *meshhttp.RequestHandler
	responseHandler *meshhttp.ResponseHandler
	vechainClient   meshthor.VeChainClientInterface
	encoder         *meshtx.MeshTransactionEncoder
	clauseParser    *meshoperations.ClauseParser
}

// NewSearchService creates a new search service
func NewSearchService(vechainClient meshthor.VeChainClientInterface) *SearchService {
	return &SearchService{
		requestHandler:  meshhttp.NewRequestHandler(),
		responseHandler: meshhttp.NewResponseHandler(),
		vechainClient:   vechainClient,
		encoder:         meshtx.NewMeshTransactionEncoder(vechainClient),
		clauseParser:    meshoperations.NewClauseParser(vechainClient, meshoperations.NewOperationsExtractor()),
	}
}

// SearchTransactions handles the /search/transactions endpoint
func (s *SearchService) SearchTransactions(w http.ResponseWriter, r *http.Request) {
	var request types.SearchTransactionsRequest
	if err := s.requestHandler.ParseJSONFromContext(r, &request); err != nil {
		s.responseHandler.WriteErrorResponse(w, meshcommon.GetError(meshcommon.ErrInvalidRequestBody), http.StatusBadRequest)
		return
	}

	// Validate transaction identifier
	if request.TransactionIdentifier == nil || request.TransactionIdentifier.Hash == "" {
		s.responseHandler.WriteErrorResponse(w, meshcommon.GetError(meshcommon.ErrInvalidRequestParameters), http.StatusBadRequest)
		return
	}

	txID := request.TransactionIdentifier.Hash

	// Get transaction to get the clauses
	tx, err := s.vechainClient.GetTransaction(txID)
	if err != nil {
		s.responseHandler.WriteErrorResponse(w, meshcommon.GetError(meshcommon.ErrTransactionNotFound), http.StatusBadRequest)
		return
	}

	// Get transaction receipt to check status
	txReceipt, err := s.vechainClient.GetTransactionReceipt(txID)
	if err != nil {
		s.responseHandler.WriteErrorResponse(w, meshcommon.GetError(meshcommon.ErrTransactionNotFound), http.StatusBadRequest)
		return
	}

	// Create block identifier
	blockIdentifier := &types.BlockIdentifier{
		Index: int64(txReceipt.Meta.BlockNumber),
		Hash:  txReceipt.Meta.BlockID.String(),
	}

	// Convert transaction to operations
	status := meshcommon.OperationStatusSucceeded
	if txReceipt.Reverted {
		status = meshcommon.OperationStatusReverted
	}

	var delegatorAddr string
	if tx.Delegator != nil && !tx.Delegator.IsZero() {
		delegatorAddr = tx.Delegator.String()
	}

	operations := s.clauseParser.ParseOperationsFromAPIClauses(tx.Clauses, tx.Origin.String(), delegatorAddr, txReceipt.GasUsed, &status)

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

	s.responseHandler.WriteJSONResponse(w, response)
}
