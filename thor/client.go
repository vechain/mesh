package thor

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/api/transactions"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/thorclient"
	"github.com/vechain/thor/v2/tx"
)

// VeChainClient handles communication with VeChain RPC
type VeChainClient struct {
	client *thorclient.Client
}

// Use native Thor types instead of duplicating them

// NewVeChainClient creates a new VeChain client
func NewVeChainClient(baseURL string) *VeChainClient {
	client := thorclient.New(baseURL)
	return &VeChainClient{
		client: client,
	}
}

// GetBestBlock fetches the latest block from VeChain
func (c *VeChainClient) GetBestBlock() (*api.JSONExpandedBlock, error) {
	block, err := c.client.ExpandedBlock("best")
	if err != nil {
		return nil, fmt.Errorf("failed to get best block: %w", err)
	}
	return block, nil
}

// GetBlockByNumber fetches a block by its number
func (c *VeChainClient) GetBlockByNumber(blockNumber int64) (*api.JSONExpandedBlock, error) {
	revision := fmt.Sprintf("%x", blockNumber)
	block, err := c.client.ExpandedBlock(revision)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by number: %w", err)
	}
	return block, nil
}

// GetBlockByHash fetches a block by its hash
func (c *VeChainClient) GetBlockByHash(blockHash string) (*api.JSONExpandedBlock, error) {
	block, err := c.client.ExpandedBlock(blockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by hash: %w", err)
	}
	return block, nil
}

// GetAccount fetches account details by address
func (c *VeChainClient) GetAccount(address string) (*api.Account, error) {
	addr, err := thor.ParseAddress(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Get account information
	account, err := c.client.Account(&addr)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return account, nil
}

// GetChainID gets the chain ID
func (c *VeChainClient) GetChainID() (int, error) {
	chainTag, err := c.client.ChainTag()
	if err != nil {
		return 0, fmt.Errorf("failed to get chain ID: %w", err)
	}

	return int(chainTag), nil
}

// SubmitTransaction submits a raw transaction to the VeChain network
func (c *VeChainClient) SubmitTransaction(rawTx []byte) (string, error) {
	// Decode the raw transaction bytes into a VeChain transaction
	var vechainTx tx.Transaction
	stream := rlp.NewStream(bytes.NewReader(rawTx), 0)
	if err := vechainTx.DecodeRLP(stream); err != nil {
		return "", fmt.Errorf("failed to decode transaction: %w", err)
	}

	// Submit transaction using the VeChain client
	result, err := c.client.SendTransaction(&vechainTx)
	if err != nil {
		return "", fmt.Errorf("failed to submit transaction: %w", err)
	}

	return result.ID.String(), nil
}

// DynamicGasPrice represents the dynamic gas price information
type DynamicGasPrice struct {
	BaseFee *big.Int
	Reward  *big.Int
}

// GetDynamicGasPrice gets the current dynamic gas price from the network
func (c *VeChainClient) GetDynamicGasPrice() (*DynamicGasPrice, error) {
	feesHistory, err := c.client.FeesHistory(1, "best", []float64{50})
	if err != nil {
		return nil, fmt.Errorf("failed to get fees history: %w", err)
	}

	// Extract base fee and reward from fees history
	var baseFee *big.Int
	if len(feesHistory.BaseFeePerGas) > 0 {
		baseFee = feesHistory.BaseFeePerGas[0].ToInt()
	} else {
		// Fallback to 0 if no base fee data available
		baseFee = big.NewInt(0)
	}

	var reward *big.Int
	if len(feesHistory.Reward) > 0 && len(feesHistory.Reward[0]) > 0 {
		reward = feesHistory.Reward[0][0].ToInt()
	} else {
		// Fallback to 0 if no reward data available
		reward = big.NewInt(0)
	}

	return &DynamicGasPrice{
		BaseFee: baseFee,
		Reward:  reward,
	}, nil
}

// GetSyncProgress returns the current sync progress (0.0 to 1.0)
func (c *VeChainClient) GetSyncProgress() (float64, error) {
	// Get best block (head)
	bestBlock, err := c.GetBestBlock()
	if err != nil {
		return 0, err
	}

	// Get genesis block
	genesisBlock, err := c.GetBlockByNumber(0)
	if err != nil {
		return 0, err
	}

	nowTsMs := float64(time.Now().UnixMilli())
	headTsMs := float64(bestBlock.Timestamp * 1000)
	genesisTsMs := float64(genesisBlock.Timestamp * 1000)

	// If the head block is recent (within 30 seconds), consider it fully synced
	if nowTsMs-headTsMs < 30*1000 {
		return 1.0, nil
	}

	// Calculate sync progress based on time difference
	progress := (headTsMs - genesisTsMs) / (nowTsMs - genesisTsMs)

	// Return NaN if progress is negative (shouldn't happen in normal conditions)
	if progress < 0 {
		return math.NaN(), nil
	}

	return progress, nil
}

// GetPeers returns the list of connected peers
func (c *VeChainClient) GetPeers() ([]Peer, error) {
	peers, err := c.client.Peers()
	if err != nil {
		return nil, err
	}

	result := make([]Peer, len(peers))
	for i, peer := range peers {
		result[i] = Peer{
			PeerID:      peer.PeerID,
			BestBlockID: peer.BestBlockID.String(),
		}
	}
	return result, nil
}

// Peer represents a connected peer
type Peer struct {
	PeerID      string
	BestBlockID string
}

// GetMempoolTransactions returns all pending transactions in the mempool
func (c *VeChainClient) GetMempoolTransactions(origin *thor.Address) ([]*thor.Bytes32, error) {
	// Get transaction pool with expanded=false to get only transaction IDs
	txPool, err := c.client.TxPool(false, origin)
	if err != nil {
		return nil, err
	}

	// Convert the result to []*thor.Bytes32
	// The TxPool method returns 'any', so we need to type assert it
	if txIDs, ok := txPool.([]*thor.Bytes32); ok {
		return txIDs, nil
	}

	// If the type assertion fails, try to convert from []thor.Bytes32
	if txIDs, ok := txPool.([]thor.Bytes32); ok {
		var result []*thor.Bytes32
		for _, txID := range txIDs {
			result = append(result, &txID)
		}
		return result, nil
	}

	// If neither type assertion works, return empty slice
	return []*thor.Bytes32{}, nil
}

// GetMempoolTransactionsExpanded returns all pending transactions in the mempool with full transaction data
func (c *VeChainClient) GetMempoolTransactionsExpanded() ([]*transactions.Transaction, error) {
	// Get transaction pool with expanded=true to get full transaction objects
	txPool, err := c.client.TxPool(true, nil)
	if err != nil {
		return nil, err
	}

	// Convert the result to []*transactions.Transaction
	if txs, ok := txPool.([]*transactions.Transaction); ok {
		return txs, nil
	}

	// If the type assertion fails, return empty slice
	return []*transactions.Transaction{}, nil
}

// GetMempoolTransaction returns a specific transaction from the mempool
func (c *VeChainClient) GetMempoolTransaction(txID *thor.Bytes32) (*transactions.Transaction, error) {
	// Get all expanded transactions from mempool and find the specific one
	txs, err := c.GetMempoolTransactionsExpanded()
	if err != nil {
		return nil, err
	}

	// Find the transaction with the matching ID
	for _, tx := range txs {
		if tx.ID == *txID {
			return tx, nil
		}
	}

	return nil, fmt.Errorf("transaction not found in mempool")
}

// GetMempoolStatus returns the current status of the transaction pool
func (c *VeChainClient) GetMempoolStatus() (*api.Status, error) {
	return c.client.TxPoolStatus()
}

// CallContract makes a contract call and returns the result
func (c *VeChainClient) CallContract(contractAddress, callData string) (string, error) {
	// Parse contract address
	contractAddr, err := thor.ParseAddress(contractAddress)
	if err != nil {
		return "", fmt.Errorf("invalid contract address: %w", err)
	}

	// Create batch call data for InspectClauses
	batchCallData := &api.BatchCallData{
		Clauses: []*api.Clause{
			{
				To:   &contractAddr,
				Data: callData,
			},
		},
	}

	// Make the call using InspectClauses
	results, err := c.client.InspectClauses(batchCallData)
	if err != nil {
		return "", fmt.Errorf("failed to call contract: %w", err)
	}

	if len(results) == 0 {
		return "", fmt.Errorf("no results returned from contract call")
	}

	// Return the result data from the first clause
	return results[0].Data, nil
}
