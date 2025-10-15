package services

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshoperations "github.com/vechain/mesh/common/operations"
	meshtx "github.com/vechain/mesh/common/tx"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/thor"
)

// MempoolService handles mempool API endpoints
type MempoolService struct {
	vechainClient meshthor.VeChainClientInterface
	encoder       *meshtx.MeshTransactionEncoder
	builder       *meshtx.TransactionBuilder
	clauseParser  *meshoperations.ClauseParser
}

// NewMempoolService creates a new mempool service
func NewMempoolService(vechainClient meshthor.VeChainClientInterface) *MempoolService {
	return &MempoolService{
		vechainClient: vechainClient,
		encoder:       meshtx.NewMeshTransactionEncoder(vechainClient),
		builder:       meshtx.NewTransactionBuilder(),
		clauseParser:  meshoperations.NewClauseParser(vechainClient, meshoperations.NewOperationsExtractor()),
	}
}

// Mempool gets mempool information
func (m *MempoolService) Mempool(
	ctx context.Context,
	req *types.NetworkRequest,
) (*types.MempoolResponse, *types.Error) {
	// Parse origin address from metadata if provided
	var origin *thor.Address
	if req.Metadata != nil {
		if originStr, ok := req.Metadata["origin"].(string); ok && originStr != "" {
			if parsedOrigin, err := thor.ParseAddress(originStr); err == nil {
				origin = &parsedOrigin
			}
		}
	}

	// Get all pending transactions from the mempool
	txIDs, err := m.vechainClient.GetMempoolTransactions(origin)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrFailedToGetMempool, map[string]any{
			"error": err.Error(),
		})
	}

	// Convert to Mesh format
	var transactionIdentifiers []*types.TransactionIdentifier
	for _, txID := range txIDs {
		transactionIdentifiers = append(transactionIdentifiers, &types.TransactionIdentifier{
			Hash: txID.String(),
		})
	}

	return &types.MempoolResponse{
		TransactionIdentifiers: transactionIdentifiers,
	}, nil
}

// MempoolTransaction gets a specific transaction from the mempool
func (m *MempoolService) MempoolTransaction(
	ctx context.Context,
	req *types.MempoolTransactionRequest,
) (*types.MempoolTransactionResponse, *types.Error) {
	// Validate required fields
	if req.TransactionIdentifier == nil || req.TransactionIdentifier.Hash == "" {
		return nil, meshcommon.GetError(meshcommon.ErrInvalidTransactionIdentifier)
	}

	// Parse transaction hash
	txID, err := thor.ParseBytes32(req.TransactionIdentifier.Hash)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInvalidTransactionHash, map[string]any{
			"error": err.Error(),
		})
	}

	// Get transaction from mempool
	tx, err := m.vechainClient.GetMempoolTransaction(&txID)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrTransactionNotFoundInMempool, map[string]any{
			"error": err.Error(),
		})
	}
	status := meshcommon.OperationStatusPending

	var delegatorAddr string
	if tx.Delegator != nil && !tx.Delegator.IsZero() {
		delegatorAddr = tx.Delegator.String()
	}

	operations, err := m.clauseParser.ParseOperationsFromAPIClauses(tx.Clauses, tx.Origin.String(), delegatorAddr, tx.Gas, &status)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInternalServerError, map[string]any{
			"error": err.Error(),
		})
	}
	meshTx := m.builder.BuildMeshTransactionFromTransaction(tx, operations)

	// Build the response
	return &types.MempoolTransactionResponse{
		Transaction: meshTx,
	}, nil
}
