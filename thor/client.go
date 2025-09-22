package thor

import (
	"fmt"

	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/thorclient"
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

// GetTransaction gets a transaction by its hash
// Note: This is a simplified implementation that returns basic transaction info
// In a real implementation, you would need to fetch the full transaction details
func (c *VeChainClient) GetTransaction(txHash string) (*Transaction, error) {
	// For now, we'll return a basic transaction structure
	// In a real implementation, you would fetch the full transaction from the blockchain
	return &Transaction{
		ID: txHash,
		// Other fields would be populated from the actual transaction data
	}, nil
}

// GetTransactionReceipt gets a transaction receipt by transaction hash
// Note: This is a simplified implementation that returns basic receipt info
// In a real implementation, you would need to fetch the full receipt details
func (c *VeChainClient) GetTransactionReceipt(txHash string) (*TransactionReceipt, error) {
	// For now, we'll return a basic receipt structure
	// In a real implementation, you would fetch the full receipt from the blockchain
	return &TransactionReceipt{
		BlockID: "unknown", // This would be populated from the actual receipt
		// Other fields would be populated from the actual receipt data
	}, nil
}

// TransactionReceipt represents a VeChain transaction receipt
type TransactionReceipt struct {
	BlockID        string `json:"blockID"`
	BlockNumber    int64  `json:"blockNumber"`
	BlockTimestamp int64  `json:"blockTimestamp"`
	GasUsed        int64  `json:"gasUsed"`
	GasPayer       string `json:"gasPayer"`
	Paid           string `json:"paid"`
	Reward         string `json:"reward"`
	Reverted       bool   `json:"reverted"`
	Meta           struct {
		BlockID        string `json:"blockID"`
		BlockNumber    int64  `json:"blockNumber"`
		BlockTimestamp int64  `json:"blockTimestamp"`
		TxID           string `json:"txID"`
		TxOrigin       string `json:"txOrigin"`
	} `json:"meta"`
	Outputs []struct {
		ContractAddress string `json:"contractAddress"`
		Events          []struct {
			Address string   `json:"address"`
			Topics  []string `json:"topics"`
			Data    string   `json:"data"`
		} `json:"events"`
		Transfers []struct {
			Sender    string `json:"sender"`
			Recipient string `json:"recipient"`
			Amount    string `json:"amount"`
		} `json:"transfers"`
	} `json:"outputs"`
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
