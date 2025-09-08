package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
)

// BlockService handles block API endpoints
type BlockService struct {
	vechainClient *VeChainClient
}

// NewBlockService creates a new block service
func NewBlockService(vechainClient *VeChainClient) *BlockService {
	return &BlockService{
		vechainClient: vechainClient,
	}
}

// Block gets block information
func (b *BlockService) Block(w http.ResponseWriter, r *http.Request) {
	var request types.BlockRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get block by identifier
	var block *Block
	var err error

	// For now, always get the best block
	// TODO: Implement proper block lookup by hash and index
	block, err = b.vechainClient.GetBestBlock()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get best block: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to Rosetta format
	blockIdentifier := &types.BlockIdentifier{
		Index: block.Number,
		Hash:  block.ID,
	}

	parentBlockIdentifier := &types.BlockIdentifier{
		Index: block.Number - 1,
		Hash:  block.ParentID,
	}

	// Convert transactions to operations
	var transactions []*types.Transaction
	for i, tx := range block.Transactions {
		// Create operations for each transaction
		var operations []*types.Operation

		// Add a basic transfer operation (simplified)
		operation := &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(i),
			},
			Type:   "Transfer",
			Status: stringPtr("Success"),
			Account: &types.AccountIdentifier{
				Address: tx.Origin,
			},
			Amount: &types.Amount{
				Value: "0", // This would need to be calculated from actual transaction data
				Currency: &types.Currency{
					Symbol:   "VET",
					Decimals: 18,
				},
			},
		}
		operations = append(operations, operation)

		transaction := &types.Transaction{
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
		transactions = append(transactions, transaction)
	}

	response := &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier:       blockIdentifier,
			ParentBlockIdentifier: parentBlockIdentifier,
			Timestamp:             block.Timestamp * 1000, // Convert to milliseconds
			Transactions:          transactions,
			Metadata: map[string]any{
				"size":         block.Size,
				"gasLimit":     block.GasLimit,
				"gasUsed":      block.GasUsed,
				"beneficiary":  block.Beneficiary,
				"totalScore":   block.TotalScore,
				"txsRoot":      block.TxsRoot,
				"txsFeatures":  block.TxsFeatures,
				"stateRoot":    block.StateRoot,
				"receiptsRoot": block.ReceiptsRoot,
				"signer":       block.Signer,
				"isTrunk":      block.IsTrunk,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// BlockTransaction gets a specific transaction from a block
func (b *BlockService) BlockTransaction(w http.ResponseWriter, r *http.Request) {
	var request types.BlockTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get block first
	var block *Block
	var err error

	// For now, always get the best block
	// TODO: Implement proper block lookup by hash and index
	block, err = b.vechainClient.GetBestBlock()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get best block: %v", err), http.StatusInternalServerError)
		return
	}

	// Find the specific transaction
	var foundTx *Transaction
	for _, tx := range block.Transactions {
		if tx.ID == request.TransactionIdentifier.Hash {
			foundTx = &tx
			break
		}
	}

	if foundTx == nil {
		http.Error(w, "Transaction not found in block", http.StatusNotFound)
		return
	}

	// Convert to Rosetta format
	var operations []*types.Operation

	// Add a basic transfer operation (simplified)
	operation := &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{
			Index: 0,
		},
		Type:   "Transfer",
		Status: stringPtr("Success"),
		Account: &types.AccountIdentifier{
			Address: foundTx.Origin,
		},
		Amount: &types.Amount{
			Value: "0", // This would need to be calculated from actual transaction data
			Currency: &types.Currency{
				Symbol:   "VET",
				Decimals: 18,
			},
		},
	}
	operations = append(operations, operation)

	transaction := &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: foundTx.ID,
		},
		Operations: operations,
		Metadata: map[string]any{
			"chainTag":     foundTx.ChainTag,
			"blockRef":     foundTx.BlockRef,
			"expiration":   foundTx.Expiration,
			"gas":          foundTx.Gas,
			"gasPriceCoef": foundTx.GasPriceCoef,
			"size":         foundTx.Size,
		},
	}

	response := &types.BlockTransactionResponse{
		Transaction: transaction,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
