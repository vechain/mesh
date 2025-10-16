package services

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshoperations "github.com/vechain/mesh/common/operations"
	meshtx "github.com/vechain/mesh/common/tx"
	meshthor "github.com/vechain/mesh/thor"
)

// SearchService handles search API endpoints
type SearchService struct {
	vechainClient meshthor.VeChainClientInterface
	encoder       *meshtx.MeshTransactionEncoder
	clauseParser  *meshoperations.ClauseParser
}

// NewSearchService creates a new search service
func NewSearchService(vechainClient meshthor.VeChainClientInterface) *SearchService {
	return &SearchService{
		vechainClient: vechainClient,
		encoder:       meshtx.NewMeshTransactionEncoder(vechainClient),
		clauseParser:  meshoperations.NewClauseParser(vechainClient, meshoperations.NewOperationsExtractor()),
	}
}

// SearchTransactions handles the /search/transactions endpoint.
// This implementation only supports searching by transaction hash.
// It requires a TransactionIdentifier with a valid hash and returns
// exactly one transaction with its operations and block information.
// Searching by other criteria is not currently supported.
func (s *SearchService) SearchTransactions(
	ctx context.Context,
	req *types.SearchTransactionsRequest,
) (*types.SearchTransactionsResponse, *types.Error) {
	// Validate transaction identifier
	if req.TransactionIdentifier == nil || req.TransactionIdentifier.Hash == "" {
		return nil, meshcommon.GetError(meshcommon.ErrInvalidRequestParameters)
	}

	txID := req.TransactionIdentifier.Hash

	// Get transaction to get the clauses
	tx, err := s.vechainClient.GetTransaction(txID)
	if err != nil {
		return nil, meshcommon.GetError(meshcommon.ErrTransactionNotFound)
	}

	// Get transaction receipt to check status
	txReceipt, err := s.vechainClient.GetTransactionReceipt(txID)
	if err != nil {
		return nil, meshcommon.GetError(meshcommon.ErrTransactionNotFound)
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

	operations, err := s.clauseParser.ParseOperationsFromAPIClauses(tx.Clauses, tx.Origin.String(), delegatorAddr, txReceipt.GasUsed, &status)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInternalServerError, map[string]any{
			"error": err.Error(),
		})
	}

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
	return &types.SearchTransactionsResponse{
		Transactions: []*types.BlockTransaction{blockTransaction},
		TotalCount:   1,
	}, nil
}
