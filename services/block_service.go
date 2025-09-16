package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshclient "github.com/vechain/mesh/client"
	meshmodels "github.com/vechain/mesh/models"
	meshutils "github.com/vechain/mesh/utils"
)

// BlockService handles block API endpoints
type BlockService struct {
	vechainClient *meshclient.VeChainClient
}

// NewBlockService creates a new block service
func NewBlockService(vechainClient *meshclient.VeChainClient) *BlockService {
	return &BlockService{
		vechainClient: vechainClient,
	}
}

// Block gets block information
func (b *BlockService) Block(w http.ResponseWriter, r *http.Request) {
	request, err := b.parseBlockRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	block, err := b.getBlockByPartialIdentifier(*request.BlockIdentifier)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if block == nil {
		http.Error(w, "Block not found", http.StatusNotFound)
		return
	}

	parent, err := b.getParentBlock(block)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := b.buildBlockResponse(block, parent)
	b.writeJSONResponse(w, response)
}

// BlockTransaction gets a specific transaction from a block
func (b *BlockService) BlockTransaction(w http.ResponseWriter, r *http.Request) {
	request, err := b.parseBlockTransactionRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	block, err := b.getBlockByIdentifier(*request.BlockIdentifier)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if block == nil {
		http.Error(w, "Block not found", http.StatusNotFound)
		return
	}

	foundTx, err := b.findTransactionInBlock(block, request.TransactionIdentifier.Hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := b.buildBlockTransactionResponse(foundTx)
	b.writeJSONResponse(w, response)
}

// parseTransactionOperations parses a transaction and returns its operations
// This analyzes the transaction data to extract meaningful operations
func (b *BlockService) parseTransactionOperations(tx meshclient.Transaction) []*types.Operation {
	var operations []*types.Operation

	// Check if this is a meaningful transaction
	// A transaction is meaningful if it has:
	// 1. Value transfer (VET) in clauses
	// 2. Contract interaction in clauses
	// 3. Energy transfer (VTHO) - gas usage

	hasValueTransfer := false
	hasContractInteraction := false
	hasEnergyTransfer := false

	// Analyze clauses for value transfers and contract interactions
	for _, clause := range tx.Clauses {
		// Check for value transfer (VET)
		if clause.Value != "" && clause.Value != "0" {
			hasValueTransfer = true
		}

		// Check for contract interaction (has data or calls a contract)
		if clause.Data != "" || (clause.To != "" && clause.To != "0x0000000000000000000000000000000000000000") {
			hasContractInteraction = true
		}
	}

	// Check for energy transfer (VTHO) - gas usage
	if tx.Gas > 0 {
		hasEnergyTransfer = true
	}

	// If no meaningful operations, return empty array
	if !hasValueTransfer && !hasContractInteraction && !hasEnergyTransfer {
		return operations
	}

	operationIndex := 0

	// Process each clause for value transfers and contract interactions
	for clauseIndex, clause := range tx.Clauses {
		// Add VET transfer operation if there's value transfer in this clause
		if clause.Value != "" && clause.Value != "0" {
			// Sender operation (negative amount)
			senderOp := &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: int64(operationIndex),
				},
				Type:   "Transfer",
				Status: meshutils.StringPtr("Success"),
				Account: &types.AccountIdentifier{
					Address: tx.Origin,
				},
				Amount: &types.Amount{
					Value:    "-" + clause.Value, // Negative for sender
					Currency: meshmodels.VETCurrency,
				},
				Metadata: map[string]any{
					"clauseIndex": clauseIndex,
				},
			}
			operations = append(operations, senderOp)
			operationIndex++

			// Receiver operation (positive amount) - only if there's a recipient
			if clause.To != "" && clause.To != "0x0000000000000000000000000000000000000000" {
				receiverOp := &types.Operation{
					OperationIdentifier: &types.OperationIdentifier{
						Index: int64(operationIndex),
					},
					Type:   "Transfer",
					Status: meshutils.StringPtr("Success"),
					Account: &types.AccountIdentifier{
						Address: clause.To,
					},
					Amount: &types.Amount{
						Value:    clause.Value, // Positive for receiver
						Currency: meshmodels.VETCurrency,
					},
					Metadata: map[string]any{
						"clauseIndex": clauseIndex,
					},
				}
				operations = append(operations, receiverOp)
				operationIndex++
			}
		}

		// Add contract interaction operation if there's contract interaction in this clause
		if clause.Data != "" || (clause.To != "" && clause.To != "0x0000000000000000000000000000000000000000") {
			contractOp := &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: int64(operationIndex),
				},
				Type:   "ContractCall",
				Status: meshutils.StringPtr("Success"),
				Account: &types.AccountIdentifier{
					Address: tx.Origin,
				},
				Amount: &types.Amount{
					Value:    "0", // Contract calls don't transfer value in the operation itself
					Currency: meshmodels.VETCurrency,
				},
				Metadata: map[string]any{
					"to":          clause.To,
					"data":        clause.Data,
					"clauseIndex": clauseIndex,
				},
			}
			operations = append(operations, contractOp)
			operationIndex++
		}
	}

	// Add energy (VTHO) transfer operation if there's gas usage
	if hasEnergyTransfer {
		// Energy is consumed by the transaction
		energyOp := &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(operationIndex),
			},
			Type:   "EnergyTransfer",
			Status: meshutils.StringPtr("Success"),
			Account: &types.AccountIdentifier{
				Address: tx.Origin,
			},
			Amount: &types.Amount{
				Value:    "-" + fmt.Sprintf("%d", tx.Gas), // Negative for energy consumption
				Currency: meshmodels.VTHOCurrency,
			},
			Metadata: map[string]any{
				"gasUsed": tx.Gas,
			},
		}
		operations = append(operations, energyOp)
	}

	return operations
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
func (b *BlockService) getBlockByIdentifier(blockIdentifier types.BlockIdentifier) (*meshclient.Block, error) {
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
func (b *BlockService) getBlockByPartialIdentifier(blockIdentifier types.PartialBlockIdentifier) (*meshclient.Block, error) {
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
func (b *BlockService) getParentBlock(block *meshclient.Block) (*meshclient.Block, error) {
	if block.Number == 0 {
		// For genesis block, parent is itself
		return block, nil
	}
	return b.vechainClient.GetBlockByNumber(block.Number - 1)
}

// findTransactionInBlock finds a specific transaction in a block
func (b *BlockService) findTransactionInBlock(block *meshclient.Block, txHash string) (*meshclient.Transaction, error) {
	for _, tx := range block.Transactions {
		if tx.ID == txHash {
			return &tx, nil
		}
	}
	return nil, fmt.Errorf("transaction %s not found in block %s", txHash, block.ID)
}

// buildBlockResponse builds the response for a block request
func (b *BlockService) buildBlockResponse(block, parent *meshclient.Block) map[string]any {
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
	var otherTransactions []map[string]string

	for _, tx := range block.Transactions {
		operations := b.parseTransactionOperations(tx)

		if len(operations) > 0 {
			// Transaction has operations, include it in transactions
			transaction := b.buildRosettaTransaction(tx, operations)
			transactions = append(transactions, transaction)
		} else {
			// Transaction has no operations, add to other_transactions
			otherTransactions = append(otherTransactions, map[string]string{
				"hash": tx.ID,
			})
		}
	}

	// Create response structure
	response := map[string]any{
		"block": map[string]any{
			"block_identifier":        blockIdentifier,
			"parent_block_identifier": parentBlockIdentifier,
			"timestamp":               block.Timestamp * 1000, // Convert to milliseconds
			"transactions":            transactions,
		},
	}

	// Add other_transactions if there are any
	if len(otherTransactions) > 0 {
		response["other_transactions"] = otherTransactions
	}

	return response
}

// buildBlockTransactionResponse builds the response for a block transaction request
func (b *BlockService) buildBlockTransactionResponse(tx *meshclient.Transaction) map[string]any {
	operations := b.parseTransactionOperations(*tx)
	rosettaTx := b.buildRosettaTransaction(*tx, operations)

	return map[string]any{
		"transaction": rosettaTx,
	}
}

// buildRosettaTransaction builds a Rosetta transaction from a VeChain transaction
func (b *BlockService) buildRosettaTransaction(tx meshclient.Transaction, operations []*types.Operation) *types.Transaction {
	return &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: tx.ID,
		},
		Operations: operations,
		Metadata: map[string]any{
			"chainTag":     tx.ChainTag,
			"blockRef":     tx.BlockRef,
			"expiration":   tx.Expiration,
			"gas":          tx.Gas,
			"gasPriceCoef": tx.GasPriceCoef,
			"size":         tx.Size,
		},
	}
}

// writeJSONResponse writes a JSON response
func (b *BlockService) writeJSONResponse(w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
