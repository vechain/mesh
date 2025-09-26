package services

import (
	"fmt"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshthor "github.com/vechain/mesh/thor"
	meshutils "github.com/vechain/mesh/utils"
	"github.com/vechain/thor/v2/api"
)

// BlockService handles block API endpoints
type BlockService struct {
	vechainClient meshthor.VeChainClientInterface
}

// NewBlockService creates a new block service
func NewBlockService(vechainClient meshthor.VeChainClientInterface) *BlockService {
	return &BlockService{
		vechainClient: vechainClient,
	}
}

// Block gets block information
func (b *BlockService) Block(w http.ResponseWriter, r *http.Request) {
	request, err := b.parseBlockRequest(r)
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrInvalidBlockIdentifierParameter), http.StatusBadRequest)
		return
	}

	block, err := b.getBlockByPartialIdentifier(*request.BlockIdentifier)
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetErrorWithMetadata(meshutils.ErrBlockNotFound, map[string]any{
			"error": err.Error(),
		}), http.StatusBadRequest)
		return
	}

	parent, err := b.getParentBlock(block)
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetErrorWithMetadata(meshutils.ErrBlockNotFound, map[string]any{
			"error": err.Error(),
		}), http.StatusInternalServerError)
		return
	}

	meshutils.WriteJSONResponse(w, b.buildBlockResponse(block, parent))
}

// BlockTransaction gets a specific transaction from a block
func (b *BlockService) BlockTransaction(w http.ResponseWriter, r *http.Request) {
	request, err := b.parseBlockTransactionRequest(r)
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrInvalidBlockIdentifierParameter), http.StatusBadRequest)
		return
	}

	block, err := b.getBlockByIdentifier(*request.BlockIdentifier)
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetErrorWithMetadata(meshutils.ErrBlockNotFound, map[string]any{
			"error": err.Error(),
		}), http.StatusBadRequest)
		return
	}

	// Get the full transaction data from the block
	foundTx, err := b.findTransactionInBlock(block, request.TransactionIdentifier.Hash)
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetErrorWithMetadata(meshutils.ErrTransactionNotFound, map[string]any{
			"transaction_identifier_hash": request.TransactionIdentifier.Hash,
		}), http.StatusBadRequest)
		return
	}

	meshutils.WriteJSONResponse(w, b.buildBlockTransactionResponse(foundTx))
}

// parseBlockRequest parses and validates a block request
func (b *BlockService) parseBlockRequest(r *http.Request) (*types.BlockRequest, error) {
	var request types.BlockRequest
	if err := meshutils.ParseJSONFromRequestContext(r, &request); err != nil {
		return nil, fmt.Errorf("invalid request body")
	}

	// Validate that a block identifier is provided
	var revision any
	if request.BlockIdentifier.Hash != nil && *request.BlockIdentifier.Hash != "" {
		revision = *request.BlockIdentifier.Hash
	} else if request.BlockIdentifier.Index != nil {
		revision = *request.BlockIdentifier.Index
	}

	if revision == nil {
		return nil, fmt.Errorf("block identifier is required")
	}

	return &request, nil
}

// parseBlockTransactionRequest parses and validates a block transaction request
func (b *BlockService) parseBlockTransactionRequest(r *http.Request) (*types.BlockTransactionRequest, error) {
	var request types.BlockTransactionRequest
	if err := meshutils.ParseJSONFromRequestContext(r, &request); err != nil {
		return nil, fmt.Errorf("invalid request body")
	}

	// Validate that a block identifier is provided
	var revision any
	if request.BlockIdentifier.Hash != "" {
		revision = request.BlockIdentifier.Hash
	} else if request.BlockIdentifier.Index != 0 {
		revision = request.BlockIdentifier.Index
	}

	if revision == nil {
		return nil, fmt.Errorf("block identifier is required")
	}

	// Validate that transaction identifier is provided
	if request.TransactionIdentifier.Hash == "" {
		return nil, fmt.Errorf("transaction identifier is required")
	}

	return &request, nil
}

// getBlockByIdentifier gets a block by its identifier (hash or index)
func (b *BlockService) getBlockByIdentifier(blockIdentifier types.BlockIdentifier) (*api.JSONExpandedBlock, error) {
	if blockIdentifier.Hash != "" {
		// Get block by hash
		return b.vechainClient.GetBlock(blockIdentifier.Hash)
	} else if blockIdentifier.Index != 0 {
		// Get block by number
		return b.vechainClient.GetBlock(fmt.Sprintf("%x", blockIdentifier.Index))
	}
	return nil, fmt.Errorf("invalid block identifier")
}

// getBlockByPartialIdentifier gets a block by its partial identifier (hash or index)
func (b *BlockService) getBlockByPartialIdentifier(blockIdentifier types.PartialBlockIdentifier) (*api.JSONExpandedBlock, error) {
	if blockIdentifier.Hash != nil && *blockIdentifier.Hash != "" {
		// Get block by hash
		return b.vechainClient.GetBlock(*blockIdentifier.Hash)
	} else if blockIdentifier.Index != nil {
		// Get block by number
		return b.vechainClient.GetBlock(fmt.Sprintf("%x", *blockIdentifier.Index))
	}
	return nil, fmt.Errorf("invalid block identifier")
}

// getParentBlock gets the parent block of the given block
func (b *BlockService) getParentBlock(block *api.JSONExpandedBlock) (*api.JSONExpandedBlock, error) {
	if block.Number == 0 {
		// For genesis block, parent is itself
		return block, nil
	}
	return b.vechainClient.GetBlock(block.ParentID.String())
}

// findTransactionInBlock finds a specific transaction in a block
func (b *BlockService) findTransactionInBlock(block *api.JSONExpandedBlock, txHash string) (*api.JSONEmbeddedTx, error) {
	// Now we have full transaction data in JSONExpandedBlock
	for _, tx := range block.Transactions {
		if tx.ID.String() == txHash {
			return tx, nil
		}
	}
	return nil, fmt.Errorf("transaction %s not found in block %s", txHash, block.ID.String())
}

// buildBlockResponse builds the response for a block request
func (b *BlockService) buildBlockResponse(block, parent *api.JSONExpandedBlock) *types.BlockResponse {
	blockIdentifier := &types.BlockIdentifier{
		Index: int64(block.Number),
		Hash:  block.ID.String(),
	}

	parentBlockIdentifier := &types.BlockIdentifier{
		Index: int64(parent.Number),
		Hash:  parent.ID.String(),
	}

	var transactions []*types.Transaction
	var otherTransactions []*types.TransactionIdentifier

	for _, tx := range block.Transactions {
		operations := meshutils.ParseTransactionOperationsFromAPI(tx)

		if len(operations) > 0 {
			transaction := meshutils.BuildMeshTransactionFromAPI(tx, operations)
			transactions = append(transactions, transaction)
		} else {
			otherTransactions = append(otherTransactions, &types.TransactionIdentifier{
				Hash: tx.ID.String(),
			})
		}
	}

	meshBlock := &types.Block{
		BlockIdentifier:       blockIdentifier,
		ParentBlockIdentifier: parentBlockIdentifier,
		Timestamp:             int64(block.Timestamp) * 1000, // Convert to milliseconds
		Transactions:          transactions,
	}

	response := &types.BlockResponse{
		Block: meshBlock,
	}

	if len(otherTransactions) > 0 {
		response.OtherTransactions = otherTransactions
	}

	return response
}

// buildBlockTransactionResponse builds the response for a block transaction request
func (b *BlockService) buildBlockTransactionResponse(tx *api.JSONEmbeddedTx) *types.BlockTransactionResponse {
	operations := meshutils.ParseTransactionOperationsFromAPI(tx)
	meshTx := meshutils.BuildMeshTransactionFromAPI(tx, operations)

	return &types.BlockTransactionResponse{
		Transaction: meshTx,
	}
}
