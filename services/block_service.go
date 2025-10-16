package services

import (
	"context"
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshtx "github.com/vechain/mesh/common/tx"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/api"
)

// BlockService handles block API endpoints
type BlockService struct {
	vechainClient meshthor.VeChainClientInterface
	encoder       *meshtx.MeshTransactionEncoder
	builder       *meshtx.TransactionBuilder
}

// NewBlockService creates a new block service
func NewBlockService(vechainClient meshthor.VeChainClientInterface) *BlockService {
	return &BlockService{
		vechainClient: vechainClient,
		encoder:       meshtx.NewMeshTransactionEncoder(vechainClient),
		builder:       meshtx.NewTransactionBuilder(),
	}
}

// Block gets block information
func (b *BlockService) Block(
	ctx context.Context,
	req *types.BlockRequest,
) (*types.BlockResponse, *types.Error) {
	block, err := b.getBlockByPartialIdentifier(*req.BlockIdentifier)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrBlockNotFound, map[string]any{
			"error": err.Error(),
		})
	}

	parent, err := b.getParentBlock(block)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrBlockNotFound, map[string]any{
			"error": err.Error(),
		})
	}

	response, err := b.buildMeshBlock(block, parent)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInternalServerError, map[string]any{
			"error": err.Error(),
		})
	}

	return response, nil
}

// BlockTransaction gets a specific transaction from a block
func (b *BlockService) BlockTransaction(
	ctx context.Context,
	req *types.BlockTransactionRequest,
) (*types.BlockTransactionResponse, *types.Error) {
	// Validate that transaction identifier is provided
	if req.TransactionIdentifier == nil || req.TransactionIdentifier.Hash == "" {
		return nil, meshcommon.GetError(meshcommon.ErrInvalidRequestBody)
	}

	block, err := b.getBlockByIdentifier(*req.BlockIdentifier)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrBlockNotFound, map[string]any{
			"error": err.Error(),
		})
	}

	// Get the full transaction data from the block
	foundTx, err := b.findTransactionInBlock(block, req.TransactionIdentifier.Hash)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrTransactionNotFound, map[string]any{
			"transaction_identifier_hash": req.TransactionIdentifier.Hash,
		})
	}

	response, err := b.buildBlockTransactionResponse(foundTx)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInternalServerError, map[string]any{
			"error": err.Error(),
		})
	}

	return response, nil
}

// getBlockByIdentifier gets a block by its identifier (hash or index)
func (b *BlockService) getBlockByIdentifier(blockIdentifier types.BlockIdentifier) (*api.JSONExpandedBlock, error) {
	if blockIdentifier.Hash != "" {
		// Get block by hash
		return b.vechainClient.GetBlock(blockIdentifier.Hash)
	} else if blockIdentifier.Index != 0 {
		// Get block by number
		return b.vechainClient.GetBlockByNumber(blockIdentifier.Index)
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
		return b.vechainClient.GetBlockByNumber(*blockIdentifier.Index)
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

// buildMeshBlock builds the response for a block request
func (b *BlockService) buildMeshBlock(block, parent *api.JSONExpandedBlock) (*types.BlockResponse, error) {
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
		operations, err := b.encoder.ParseTransactionOperationsFromAPI(tx)
		if err != nil {
			return nil, err
		}

		if len(operations) > 0 {
			transaction := b.builder.BuildMeshTransactionFromAPI(tx, operations)
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

	return response, nil
}

// buildBlockTransactionResponse builds the response for a block transaction request
func (b *BlockService) buildBlockTransactionResponse(tx *api.JSONEmbeddedTx) (*types.BlockTransactionResponse, error) {
	operations, err := b.encoder.ParseTransactionOperationsFromAPI(tx)
	if err != nil {
		return nil, err
	}
	meshTx := b.builder.BuildMeshTransactionFromAPI(tx, operations)

	return &types.BlockTransactionResponse{
		Transaction: meshTx,
	}, nil
}
