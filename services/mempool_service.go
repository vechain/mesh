package services

import (
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshhttp "github.com/vechain/mesh/common/http"
	meshoperations "github.com/vechain/mesh/common/operations"
	meshtx "github.com/vechain/mesh/common/tx"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/thor"
)

// MempoolService handles mempool API endpoints
type MempoolService struct {
	requestHandler  *meshhttp.RequestHandler
	responseHandler *meshhttp.ResponseHandler
	vechainClient   meshthor.VeChainClientInterface
	encoder         *meshtx.MeshTransactionEncoder
	builder         *meshtx.TransactionBuilder
	clauseParser    *meshoperations.ClauseParser
}

// NewMempoolService creates a new mempool service
func NewMempoolService(vechainClient meshthor.VeChainClientInterface) *MempoolService {
	return &MempoolService{
		requestHandler:  meshhttp.NewRequestHandler(),
		responseHandler: meshhttp.NewResponseHandler(),
		vechainClient:   vechainClient,
		encoder:         meshtx.NewMeshTransactionEncoder(vechainClient),
		builder:         meshtx.NewTransactionBuilder(),
		clauseParser:    meshoperations.NewClauseParser(vechainClient, meshoperations.NewOperationsExtractor()),
	}
}

// Mempool gets mempool information
func (m *MempoolService) Mempool(w http.ResponseWriter, r *http.Request) {
	// Parse request body to get metadata (origin filter)
	var requestBody map[string]any
	if err := m.requestHandler.ParseJSONFromContext(r, &requestBody); err != nil {
		m.responseHandler.WriteErrorResponse(w, meshcommon.GetError(meshcommon.ErrInvalidRequestBody), http.StatusBadRequest)
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
		m.responseHandler.WriteErrorResponse(w, meshcommon.GetErrorWithMetadata(meshcommon.ErrFailedToGetMempool, map[string]any{
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

	m.responseHandler.WriteJSONResponse(w, response)
}

// MempoolTransaction gets a specific transaction from the mempool
func (m *MempoolService) MempoolTransaction(w http.ResponseWriter, r *http.Request) {
	var request types.MempoolTransactionRequest
	if err := m.requestHandler.ParseJSONFromContext(r, &request); err != nil {
		m.responseHandler.WriteErrorResponse(w, meshcommon.GetError(meshcommon.ErrInvalidRequestBody), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if request.TransactionIdentifier == nil || request.TransactionIdentifier.Hash == "" {
		m.responseHandler.WriteErrorResponse(w, meshcommon.GetError(meshcommon.ErrInvalidTransactionIdentifier), http.StatusBadRequest)
		return
	}

	// Parse transaction hash
	txID, err := thor.ParseBytes32(request.TransactionIdentifier.Hash)
	if err != nil {
		m.responseHandler.WriteErrorResponse(w, meshcommon.GetErrorWithMetadata(meshcommon.ErrInvalidTransactionHash, map[string]any{
			"error": err.Error(),
		}), http.StatusBadRequest)
		return
	}

	// Get transaction from mempool
	tx, err := m.vechainClient.GetMempoolTransaction(&txID)
	if err != nil {
		m.responseHandler.WriteErrorResponse(w, meshcommon.GetErrorWithMetadata(meshcommon.ErrTransactionNotFoundInMempool, map[string]any{
			"error": err.Error(),
		}), http.StatusNotFound)
		return
	}
	status := meshcommon.OperationStatusPending

	var delegatorAddr string
	if tx.Delegator != nil && !tx.Delegator.IsZero() {
		delegatorAddr = tx.Delegator.String()
	}

	operations, err := m.clauseParser.ParseOperationsFromAPIClauses(tx.Clauses, tx.Origin.String(), delegatorAddr, tx.Gas, &status)
	if err != nil {
		m.responseHandler.WriteErrorResponse(w, meshcommon.GetErrorWithMetadata(meshcommon.ErrInternalServerError, map[string]any{
			"error": err.Error(),
		}), http.StatusInternalServerError)
		return
	}
	meshTx := m.builder.BuildMeshTransactionFromTransaction(tx, operations)

	// Build the response
	response := &types.MempoolTransactionResponse{
		Transaction: meshTx,
	}

	m.responseHandler.WriteJSONResponse(w, response)
}
