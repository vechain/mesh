package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshthor "github.com/vechain/mesh/thor"
	meshutils "github.com/vechain/mesh/utils"
)

// BlockService handles block API endpoints
type BlockService struct {
	vechainClient *meshthor.VeChainClient
}

// NewBlockService creates a new block service
func NewBlockService(vechainClient *meshthor.VeChainClient) *BlockService {
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

	response := b.buildBlockResponse(block, parent)
	b.writeJSONResponse(w, response)
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

	foundTx, err := b.findTransactionInBlock(block, request.TransactionIdentifier.Hash)
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetErrorWithMetadata(meshutils.ErrTransactionNotFound, map[string]any{
			"transaction_identifier_hash": request.TransactionIdentifier.Hash,
		}), http.StatusBadRequest)
		return
	}

	response := b.buildBlockTransactionResponse(foundTx)
	b.writeJSONResponse(w, response)
}

// parseBlockRequest parses and validates a block request
func (b *BlockService) parseBlockRequest(r *http.Request) (*types.BlockRequest, error) {
	var request types.BlockRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
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
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
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
func (b *BlockService) getBlockByIdentifier(blockIdentifier types.BlockIdentifier) (*meshthor.Block, error) {
	if blockIdentifier.Hash != "" {
		// Get block by hash
		return b.vechainClient.GetBlockByHash(blockIdentifier.Hash)
	} else if blockIdentifier.Index != 0 {
		// Get block by number
		return b.vechainClient.GetBlockByNumber(blockIdentifier.Index)
	}
	return nil, fmt.Errorf("invalid block identifier")
}

// getBlockByPartialIdentifier gets a block by its partial identifier (hash or index)
func (b *BlockService) getBlockByPartialIdentifier(blockIdentifier types.PartialBlockIdentifier) (*meshthor.Block, error) {
	if blockIdentifier.Hash != nil && *blockIdentifier.Hash != "" {
		// Get block by hash
		return b.vechainClient.GetBlockByHash(*blockIdentifier.Hash)
	} else if blockIdentifier.Index != nil {
		// Get block by number
		return b.vechainClient.GetBlockByNumber(*blockIdentifier.Index)
	}
	return nil, fmt.Errorf("invalid block identifier")
}

// getParentBlock gets the parent block of the given block
func (b *BlockService) getParentBlock(block *meshthor.Block) (*meshthor.Block, error) {
	if block.Number == 0 {
		// For genesis block, parent is itself
		return block, nil
	}
	return b.vechainClient.GetBlockByHash(block.ParentID)
}

// findTransactionInBlock finds a specific transaction in a block
func (b *BlockService) findTransactionInBlock(block *meshthor.Block, txHash string) (*meshthor.Transaction, error) {
	for _, tx := range block.Transactions {
		if tx.ID == txHash {
			return &tx, nil
		}
	}
	return nil, fmt.Errorf("transaction %s not found in block %s", txHash, block.ID)
}

// buildBlockResponse builds the response for a block request
func (b *BlockService) buildBlockResponse(block, parent *meshthor.Block) *types.BlockResponse {
	blockIdentifier := &types.BlockIdentifier{
		Index: block.Number,
		Hash:  block.ID,
	}

	parentBlockIdentifier := &types.BlockIdentifier{
		Index: parent.Number,
		Hash:  parent.ID,
	}

	// Process transactions
	var transactions []*types.Transaction
	var otherTransactions []*types.TransactionIdentifier

	for _, tx := range block.Transactions {
		// Convert to common format and parse operations
		meshTx := meshutils.ConvertMeshThorTransactionToMeshTransaction(tx)
		operations := meshutils.ParseTransactionOperations(meshTx, "Success")

		if len(operations) > 0 {
			// Transaction has operations, include it in transactions
			transaction := meshutils.BuildRosettaTransaction(meshTx, operations)
			transactions = append(transactions, transaction)
		} else {
			// Transaction has no operations, add to other_transactions
			otherTransactions = append(otherTransactions, &types.TransactionIdentifier{
				Hash: tx.ID,
			})
		}
	}

	// Create response structure
	meshBlock := &types.Block{
		BlockIdentifier:       blockIdentifier,
		ParentBlockIdentifier: parentBlockIdentifier,
		Timestamp:             block.Timestamp * 1000, // Convert to milliseconds
		Transactions:          transactions,
	}

	response := &types.BlockResponse{
		Block: meshBlock,
	}

	// Add other_transactions if there are any
	if len(otherTransactions) > 0 {
		response.OtherTransactions = otherTransactions
	}

	return response
}

// buildBlockTransactionResponse builds the response for a block transaction request
func (b *BlockService) buildBlockTransactionResponse(tx *meshthor.Transaction) *types.BlockTransactionResponse {
	// Convert to common format and parse operations
	meshTx := meshutils.ConvertMeshThorTransactionToMeshTransaction(*tx)
	operations := meshutils.ParseTransactionOperations(meshTx, "Success")
	rosettaTx := meshutils.BuildRosettaTransaction(meshTx, operations)

	return &types.BlockTransactionResponse{
		Transaction: rosettaTx,
	}
}

// writeJSONResponse writes a JSON response
func (b *BlockService) writeJSONResponse(w http.ResponseWriter, response any) {
	meshutils.WriteJSONResponse(w, response)
}
