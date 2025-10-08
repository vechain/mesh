package thor

import (
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/api/transactions"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/thorclient"
	"github.com/vechain/thor/v2/tx"
)

type VeChainClient struct {
	client ThorClientInterface
}

// NewVeChainClient creates a new VeChain client
func NewVeChainClient(baseURL string) *VeChainClient {
	client := thorclient.New(baseURL)
	return &VeChainClient{
		client: client,
	}
}

// NewVeChainClientWithMock creates a new VeChain client with a custom ThorClientInterface (for testing)
func NewVeChainClientWithMock(client ThorClientInterface) *VeChainClient {
	return &VeChainClient{
		client: client,
	}
}

// GetBlock fetches a block by its revision
func (c *VeChainClient) GetBlock(revision string) (*api.JSONExpandedBlock, error) {
	return c.client.ExpandedBlock(revision)
}

// GetBlockByNumber fetches a block by its number
func (c *VeChainClient) GetBlockByNumber(blockNumber int64) (*api.JSONExpandedBlock, error) {
	// Convert block number to hex format with 0x prefix
	revision := fmt.Sprintf("0x%x", blockNumber)
	return c.GetBlock(revision)
}

// GetAccount fetches account details by address at the latest block
func (c *VeChainClient) GetAccount(address string) (*api.Account, error) {
	return c.GetAccountAtRevision(address, "")
}

// GetAccountAtRevision fetches account details by address at a specific block revision
func (c *VeChainClient) GetAccountAtRevision(address string, revision string) (*api.Account, error) {
	addr, err := thor.ParseAddress(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Get account information with optional revision
	var account *api.Account
	if revision != "" {
		account, err = c.client.Account(&addr, thorclient.Revision(revision))
	} else {
		account, err = c.client.Account(&addr)
	}
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
func (c *VeChainClient) SubmitTransaction(vechainTx *tx.Transaction) (string, error) {

	// Submit transaction using the VeChain client
	result, err := c.client.SendTransaction(vechainTx)
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
	bestBlock, err := c.GetBlock("best")
	if err != nil {
		return 0, err
	}

	// Get genesis block
	genesisBlock, err := c.GetBlock("0")
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

// getTransactions call the endpoint node/txpool
func (c *VeChainClient) getTransactions(origin *thor.Address, expanded bool) (any, error) {
	return c.client.TxPool(expanded, origin)
}

// GetMempoolTransactions returns all pending transactions in the mempool
func (c *VeChainClient) GetMempoolTransactions(origin *thor.Address) ([]*thor.Bytes32, error) {
	// Get transaction pool with expanded=false to get only transaction IDs
	txPool, err := c.getTransactions(origin, false)
	if err != nil {
		return nil, err
	}

	if txIDs, ok := txPool.([]thor.Bytes32); ok {
		var result []*thor.Bytes32
		for _, txID := range txIDs {
			result = append(result, &txID)
		}
		return result, nil
	}

	return []*thor.Bytes32{}, nil
}

// GetMempoolTransaction returns a specific transaction from the mempool
func (c *VeChainClient) GetMempoolTransaction(txID *thor.Bytes32) (*transactions.Transaction, error) {
	// Get all expanded transactions from mempool (no origin filter)
	txPool, err := c.getTransactions(nil, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get mempool transactions: %w", err)
	}

	var txs []*transactions.Transaction
	if txList, ok := txPool.([]transactions.Transaction); ok {
		for i := range txList {
			txs = append(txs, &txList[i])
		}
	} else {
		return nil, fmt.Errorf("unexpected response type from TxPool: %T", txPool)
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

// GetTransaction fetches a transaction by its ID
func (c *VeChainClient) GetTransaction(txID string) (*transactions.Transaction, error) {
	txHash, err := thor.ParseBytes32(txID)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction ID format: %w", err)
	}

	tx, err := c.client.Transaction(&txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	return tx, nil
}

// GetTransactionReceipt fetches a transaction receipt by transaction ID
func (c *VeChainClient) GetTransactionReceipt(txID string) (*api.Receipt, error) {
	txHash, err := thor.ParseBytes32(txID)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction ID format: %w", err)
	}

	receipt, err := c.client.TransactionReceipt(&txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}
	return receipt, nil
}

// InspectClauses simulates execution of clauses without submitting a transaction
func (c *VeChainClient) InspectClauses(batchCallData *api.BatchCallData, options ...thorclient.Option) ([]*api.CallResult, error) {
	results, err := c.client.InspectClauses(batchCallData, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect clauses: %w", err)
	}
	return results, nil
}
