package thor

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/thorclient"
	"github.com/vechain/thor/v2/tx"
)

// VeChainClient handles communication with VeChain RPC
type VeChainClient struct {
	client *thorclient.Client
}

// Block represents a VeChain block
type Block struct {
	Number       int64         `json:"number"`
	ID           string        `json:"id"`
	Size         int           `json:"size"`
	ParentID     string        `json:"parentID"`
	Timestamp    int64         `json:"timestamp"`
	GasLimit     int64         `json:"gasLimit"`
	Beneficiary  string        `json:"beneficiary"`
	GasUsed      int64         `json:"gasUsed"`
	TotalScore   int64         `json:"totalScore"`
	TxsRoot      string        `json:"txsRoot"`
	TxsFeatures  int           `json:"txsFeatures"`
	StateRoot    string        `json:"stateRoot"`
	ReceiptsRoot string        `json:"receiptsRoot"`
	Signer       string        `json:"signer"`
	IsTrunk      bool          `json:"isTrunk"`
	Transactions []Transaction `json:"transactions"`
}

// Transaction represents a VeChain transaction
type Transaction struct {
	ID           string          `json:"id"`
	ChainTag     int             `json:"chainTag"`
	BlockRef     string          `json:"blockRef"`
	Expiration   int64           `json:"expiration"`
	Clauses      []Clause        `json:"clauses"`
	GasPriceCoef int             `json:"gasPriceCoef"`
	Gas          int64           `json:"gas"`
	Origin       string          `json:"origin"`
	Delegator    string          `json:"delegator,omitempty"`
	Nonce        string          `json:"nonce"`
	DependsOn    string          `json:"dependsOn,omitempty"`
	Size         int             `json:"size"`
	Meta         TransactionMeta `json:"meta"`
}

// Clause represents a VeChain transaction clause
type Clause struct {
	To    string `json:"to"`
	Value string `json:"value"`
	Data  string `json:"data"`
}

// TransactionMeta represents transaction metadata
type TransactionMeta struct {
	BlockID        string `json:"blockID"`
	BlockNumber    int64  `json:"blockNumber"`
	BlockTimestamp int64  `json:"blockTimestamp"`
}

// Account represents a VeChain account
type Account struct {
	Balance string `json:"balance"`
	Energy  string `json:"energy"`
	HasCode bool   `json:"hasCode"`
}

// NewVeChainClient creates a new VeChain client
func NewVeChainClient(baseURL string) *VeChainClient {
	client := thorclient.New(baseURL)
	return &VeChainClient{
		client: client,
	}
}

// GetBestBlock fetches the latest block from VeChain
func (c *VeChainClient) GetBestBlock() (*Block, error) {
	block, err := c.client.Block("best")
	if err != nil {
		return nil, fmt.Errorf("failed to get best block: %w", err)
	}
	return c.convertBlock(block), nil
}

// GetBlockByNumber fetches a block by its number
func (c *VeChainClient) GetBlockByNumber(blockNumber int64) (*Block, error) {
	revision := fmt.Sprintf("%x", blockNumber)
	block, err := c.client.Block(revision)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by number: %w", err)
	}
	return c.convertBlock(block), nil
}

// GetBlockByHash fetches a block by its hash
func (c *VeChainClient) GetBlockByHash(blockHash string) (*Block, error) {
	block, err := c.client.Block(blockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by hash: %w", err)
	}
	return c.convertBlock(block), nil
}

// GetAccount fetches account details by address
func (c *VeChainClient) GetAccount(address string) (*Account, error) {
	addr, err := thor.ParseAddress(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Get account information
	account, err := c.client.Account(&addr)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Check if account has code
	codeResult, err := c.client.AccountCode(&addr)
	if err != nil {
		return nil, fmt.Errorf("failed to get account code: %w", err)
	}

	// Convert HexOrDecimal256 to string
	balanceBytes, _ := account.Balance.MarshalText()
	energyBytes, _ := account.Energy.MarshalText()

	return &Account{
		Balance: string(balanceBytes),
		Energy:  string(energyBytes),
		HasCode: len(codeResult.Code) > 0,
	}, nil
}

// GetChainID gets the chain ID
func (c *VeChainClient) GetChainID() (int, error) {
	chainTag, err := c.client.ChainTag()
	if err != nil {
		return 0, fmt.Errorf("failed to get chain ID: %w", err)
	}

	return int(chainTag), nil
}

// convertBlock converts thorclient block to our Block type
func (c *VeChainClient) convertBlock(block *api.JSONCollapsedBlock) *Block {
	return &Block{
		Number:       int64(block.Number),
		ID:           block.ID.String(),
		Size:         int(block.Size),
		ParentID:     block.ParentID.String(),
		Timestamp:    int64(block.Timestamp),
		GasLimit:     int64(block.GasLimit),
		Beneficiary:  block.Beneficiary.String(),
		GasUsed:      int64(block.GasUsed),
		TotalScore:   int64(block.TotalScore),
		TxsRoot:      block.TxsRoot.String(),
		TxsFeatures:  int(block.TxsFeatures),
		StateRoot:    block.StateRoot.String(),
		ReceiptsRoot: block.ReceiptsRoot.String(),
		Signer:       block.Signer.String(),
		IsTrunk:      block.IsTrunk,
		Transactions: convertTransactionsFromAPI(block.Transactions),
	}
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

// convertTransactionsFromAPI converts API transactions to our Transaction type
func convertTransactionsFromAPI(apiTxs []thor.Bytes32) []Transaction {
	transactions := make([]Transaction, len(apiTxs))
	for i, txID := range apiTxs {
		transactions[i] = Transaction{
			ID: txID.String(),
			// Other fields would need to be fetched separately if needed
		}
	}
	return transactions
}

// GetSyncProgress returns the current sync progress (0.0 to 1.0)
func (c *VeChainClient) GetSyncProgress() (float64, error) {
	// For now, return 1.0 (fully synced) as a placeholder
	// TODO: Implement proper sync progress detection
	return 1.0, nil
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
