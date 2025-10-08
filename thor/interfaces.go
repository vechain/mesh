package thor

import (
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/api/transactions"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/thorclient"
	"github.com/vechain/thor/v2/tx"
)

// VeChainClientInterface defines a common interface
type VeChainClientInterface interface {
	GetBlock(revision string) (*api.JSONExpandedBlock, error)
	GetBlockByNumber(blockNumber int64) (*api.JSONExpandedBlock, error)
	GetAccount(address string) (*api.Account, error)
	GetAccountAtRevision(address string, revision string) (*api.Account, error)
	GetChainID() (int, error)
	SubmitTransaction(vechainTx *tx.Transaction) (string, error)
	GetDynamicGasPrice() (*DynamicGasPrice, error)
	GetSyncProgress() (float64, error)
	GetPeers() ([]Peer, error)
	GetMempoolTransactions(origin *thor.Address) ([]*thor.Bytes32, error)
	GetMempoolTransaction(txID *thor.Bytes32) (*transactions.Transaction, error)
	GetMempoolStatus() (*api.Status, error)
	CallContract(contractAddress, callData string) (string, error)
	GetTransaction(txID string) (*transactions.Transaction, error)
	GetTransactionReceipt(txID string) (*api.Receipt, error)
	InspectClauses(batchCallData *api.BatchCallData, options ...thorclient.Option) ([]*api.CallResult, error)
}

// ThorClientInterface defines the interface for thorclient.Client methods we use
type ThorClientInterface interface {
	ExpandedBlock(revision string) (*api.JSONExpandedBlock, error)
	Account(address *thor.Address, opts ...thorclient.Option) (*api.Account, error)
	ChainTag() (byte, error)
	SendTransaction(tx *tx.Transaction) (*api.SendTxResult, error)
	FeesHistory(blockCount uint32, newestBlock string, rewardPercentiles []float64) (*api.FeesHistory, error)
	Peers() ([]*api.PeerStats, error)
	TxPool(expanded bool, origin *thor.Address) (any, error)
	TxPoolStatus() (*api.Status, error)
	InspectClauses(batchCallData *api.BatchCallData, options ...thorclient.Option) ([]*api.CallResult, error)
	Transaction(txHash *thor.Bytes32, opts ...thorclient.Option) (*transactions.Transaction, error)
	TransactionReceipt(txHash *thor.Bytes32, opts ...thorclient.Option) (*api.Receipt, error)
}
