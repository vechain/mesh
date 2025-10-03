package thor

import (
	"fmt"

	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/api/transactions"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/thorclient"
	"github.com/vechain/thor/v2/tx"
)

// MockThorClient is a mock implementation of ThorClientInterface
type MockThorClient struct {
	expandedBlockFunc      func(revision string) (*api.JSONExpandedBlock, error)
	accountFunc            func(address *thor.Address, opts ...thorclient.Option) (*api.Account, error)
	chainTagFunc           func() (byte, error)
	sendTransactionFunc    func(tx *tx.Transaction) (*api.SendTxResult, error)
	feesHistoryFunc        func(blockCount uint32, newestBlock string, rewardPercentiles []float64) (*api.FeesHistory, error)
	peersFunc              func() ([]*api.PeerStats, error)
	txPoolFunc             func(expanded bool, origin *thor.Address) (any, error)
	txPoolStatusFunc       func() (*api.Status, error)
	inspectClausesFunc     func(batchCallData *api.BatchCallData, options ...thorclient.Option) ([]*api.CallResult, error)
	transactionFunc        func(txHash *thor.Bytes32, opts ...thorclient.Option) (*transactions.Transaction, error)
	transactionReceiptFunc func(txHash *thor.Bytes32, opts ...thorclient.Option) (*api.Receipt, error)
}

// NewMockThorClient creates a new mock thor client
func NewMockThorClient() *MockThorClient {
	return &MockThorClient{}
}

func (m *MockThorClient) ExpandedBlock(revision string) (*api.JSONExpandedBlock, error) {
	if m.expandedBlockFunc != nil {
		return m.expandedBlockFunc(revision)
	}
	return nil, fmt.Errorf("mock not configured")
}

func (m *MockThorClient) Account(address *thor.Address, opts ...thorclient.Option) (*api.Account, error) {
	if m.accountFunc != nil {
		return m.accountFunc(address, opts...)
	}
	return nil, fmt.Errorf("mock not configured")
}

func (m *MockThorClient) ChainTag() (byte, error) {
	if m.chainTagFunc != nil {
		return m.chainTagFunc()
	}
	return 0, fmt.Errorf("mock not configured")
}

func (m *MockThorClient) SendTransaction(tx *tx.Transaction) (*api.SendTxResult, error) {
	if m.sendTransactionFunc != nil {
		return m.sendTransactionFunc(tx)
	}
	return nil, fmt.Errorf("mock not configured")
}

func (m *MockThorClient) FeesHistory(blockCount uint32, newestBlock string, rewardPercentiles []float64) (*api.FeesHistory, error) {
	if m.feesHistoryFunc != nil {
		return m.feesHistoryFunc(blockCount, newestBlock, rewardPercentiles)
	}
	return nil, fmt.Errorf("mock not configured")
}

func (m *MockThorClient) Peers() ([]*api.PeerStats, error) {
	if m.peersFunc != nil {
		return m.peersFunc()
	}
	return nil, fmt.Errorf("mock not configured")
}

func (m *MockThorClient) TxPool(expanded bool, origin *thor.Address) (any, error) {
	if m.txPoolFunc != nil {
		return m.txPoolFunc(expanded, origin)
	}
	return nil, fmt.Errorf("mock not configured")
}

func (m *MockThorClient) TxPoolStatus() (*api.Status, error) {
	if m.txPoolStatusFunc != nil {
		return m.txPoolStatusFunc()
	}
	return nil, fmt.Errorf("mock not configured")
}

func (m *MockThorClient) InspectClauses(batchCallData *api.BatchCallData, options ...thorclient.Option) ([]*api.CallResult, error) {
	if m.inspectClausesFunc != nil {
		return m.inspectClausesFunc(batchCallData, options...)
	}
	return nil, fmt.Errorf("mock not configured")
}

func (m *MockThorClient) Transaction(txHash *thor.Bytes32, opts ...thorclient.Option) (*transactions.Transaction, error) {
	if m.transactionFunc != nil {
		return m.transactionFunc(txHash, opts...)
	}
	return nil, fmt.Errorf("mock not configured")
}

func (m *MockThorClient) TransactionReceipt(txHash *thor.Bytes32, opts ...thorclient.Option) (*api.Receipt, error) {
	if m.transactionReceiptFunc != nil {
		return m.transactionReceiptFunc(txHash, opts...)
	}
	return nil, fmt.Errorf("mock not configured")
}

// Setter methods for configuring mock behavior
func (m *MockThorClient) SetExpandedBlockFunc(f func(revision string) (*api.JSONExpandedBlock, error)) {
	m.expandedBlockFunc = f
}

func (m *MockThorClient) SetAccountFunc(f func(address *thor.Address, opts ...thorclient.Option) (*api.Account, error)) {
	m.accountFunc = f
}

func (m *MockThorClient) SetChainTagFunc(f func() (byte, error)) {
	m.chainTagFunc = f
}

func (m *MockThorClient) SetSendTransactionFunc(f func(tx *tx.Transaction) (*api.SendTxResult, error)) {
	m.sendTransactionFunc = f
}

func (m *MockThorClient) SetFeesHistoryFunc(f func(blockCount uint32, newestBlock string, rewardPercentiles []float64) (*api.FeesHistory, error)) {
	m.feesHistoryFunc = f
}

func (m *MockThorClient) SetPeersFunc(f func() ([]*api.PeerStats, error)) {
	m.peersFunc = f
}

func (m *MockThorClient) SetTxPoolFunc(f func(expanded bool, origin *thor.Address) (any, error)) {
	m.txPoolFunc = f
}

func (m *MockThorClient) SetTxPoolStatusFunc(f func() (*api.Status, error)) {
	m.txPoolStatusFunc = f
}

func (m *MockThorClient) SetInspectClausesFunc(f func(batchCallData *api.BatchCallData, options ...thorclient.Option) ([]*api.CallResult, error)) {
	m.inspectClausesFunc = f
}

func (m *MockThorClient) SetTransactionFunc(f func(txHash *thor.Bytes32, opts ...thorclient.Option) (*transactions.Transaction, error)) {
	m.transactionFunc = f
}

func (m *MockThorClient) SetTransactionReceiptFunc(f func(txHash *thor.Bytes32, opts ...thorclient.Option) (*api.Receipt, error)) {
	m.transactionReceiptFunc = f
}
