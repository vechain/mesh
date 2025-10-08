package thor

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	meshtests "github.com/vechain/mesh/tests"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/api/transactions"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/thorclient"
	"github.com/vechain/thor/v2/tx"
)

// MockVeChainClient is a mock client for tests that simulates VeChain responses
type MockVeChainClient struct {
	// Mock configuration
	MockBestBlock      *api.JSONExpandedBlock
	MockBlock          *api.JSONExpandedBlock
	MockAccount        *api.Account
	MockChainID        int
	MockGasPrice       *DynamicGasPrice
	MockSyncProgress   float64
	MockPeers          []Peer
	MockMempoolTxs     []*thor.Bytes32
	MockMempoolTx      *transactions.Transaction
	MockMempoolStatus  *api.Status
	MockCallResult     string
	MockCallResults    []string
	MockCallIndex      int
	MockBlockByNumber  *api.JSONExpandedBlock
	MockTransaction    *transactions.Transaction
	MockReceipt        *api.Receipt
	MockInspectClauses []*api.CallResult

	// Simulated errors
	MockError error
}

// NewMockVeChainClient creates a new mock client
func NewMockVeChainClient() *MockVeChainClient {
	return &MockVeChainClient{
		MockBestBlock: &api.JSONExpandedBlock{
			JSONBlockSummary: &api.JSONBlockSummary{
				Number: 100,
				ID: func() thor.Bytes32 {
					hash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
					return hash
				}(),
				Size: 1024,
				ParentID: func() thor.Bytes32 {
					hash, _ := thor.ParseBytes32("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
					return hash
				}(),
				Timestamp: uint64(time.Now().Unix()),
				GasLimit:  10000000,
				GasUsed:   5000000,
				Beneficiary: func() thor.Address {
					addr, _ := thor.ParseAddress(meshtests.FirstSoloAddress)
					return addr
				}(),
				TotalScore: 1000,
				TxsRoot: func() thor.Bytes32 {
					hash, _ := thor.ParseBytes32("0x1111111111111111111111111111111111111111111111111111111111111111")
					return hash
				}(),
				StateRoot: func() thor.Bytes32 {
					hash, _ := thor.ParseBytes32("0x2222222222222222222222222222222222222222222222222222222222222222")
					return hash
				}(),
				ReceiptsRoot: func() thor.Bytes32 {
					hash, _ := thor.ParseBytes32("0x3333333333333333333333333333333333333333333333333333333333333333")
					return hash
				}(),
			},
			Transactions: []*api.JSONEmbeddedTx{
				{
					ID: func() thor.Bytes32 {
						hash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
						return hash
					}(),
					Origin: func() thor.Address {
						addr, _ := thor.ParseAddress(meshtests.FirstSoloAddress)
						return addr
					}(),
					Gas:          21000,
					GasPriceCoef: func() *uint8 { v := uint8(128); return &v }(),
					Nonce: func() math.HexOrDecimal64 {
						return math.HexOrDecimal64(1)
					}(),
					Clauses: []*api.JSONClause{
						{
							To: func() *thor.Address {
								addr, _ := thor.ParseAddress(meshtests.TestAddress1)
								return &addr
							}(),
							Value: func() math.HexOrDecimal256 {
								val, _ := new(big.Int).SetString("1000000000000000000", 10) // 1 VET
								return math.HexOrDecimal256(*val)
							}(),
							Data: "0x",
						},
					},
				},
			},
		},
		MockBlock: &api.JSONExpandedBlock{
			JSONBlockSummary: &api.JSONBlockSummary{
				Number: 100,
				ID: func() thor.Bytes32 {
					hash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
					return hash
				}(),
				Size: 1024,
				ParentID: func() thor.Bytes32 {
					hash, _ := thor.ParseBytes32("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
					return hash
				}(),
				Timestamp: uint64(time.Now().Unix()),
				GasLimit:  10000000,
				GasUsed:   5000000,
				Beneficiary: func() thor.Address {
					addr, _ := thor.ParseAddress(meshtests.FirstSoloAddress)
					return addr
				}(),
				TotalScore: 1000,
				TxsRoot: func() thor.Bytes32 {
					hash, _ := thor.ParseBytes32("0x1111111111111111111111111111111111111111111111111111111111111111")
					return hash
				}(),
				StateRoot: func() thor.Bytes32 {
					hash, _ := thor.ParseBytes32("0x2222222222222222222222222222222222222222222222222222222222222222")
					return hash
				}(),
				ReceiptsRoot: func() thor.Bytes32 {
					hash, _ := thor.ParseBytes32("0x3333333333333333333333333333333333333333333333333333333333333333")
					return hash
				}(),
			},
			Transactions: []*api.JSONEmbeddedTx{
				{
					ID: func() thor.Bytes32 {
						hash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
						return hash
					}(),
					Origin: func() thor.Address {
						addr, _ := thor.ParseAddress(meshtests.FirstSoloAddress)
						return addr
					}(),
					Gas:          21000,
					GasPriceCoef: func() *uint8 { v := uint8(128); return &v }(),
					Nonce: func() math.HexOrDecimal64 {
						return math.HexOrDecimal64(1)
					}(),
					Clauses: []*api.JSONClause{
						{
							To: func() *thor.Address {
								addr, _ := thor.ParseAddress(meshtests.TestAddress1)
								return &addr
							}(),
							Value: func() math.HexOrDecimal256 {
								val, _ := new(big.Int).SetString("1000000000000000000", 10) // 1 VET
								return math.HexOrDecimal256(*val)
							}(),
							Data: "0x",
						},
					},
				},
			},
		},
		MockAccount: &api.Account{
			Balance: func() *math.HexOrDecimal256 {
				val, _ := new(big.Int).SetString("1000000000000000000000", 10)
				hexVal := math.HexOrDecimal256(*val)
				return &hexVal
			}(), // 1000 VET
			Energy: func() *math.HexOrDecimal256 {
				val, _ := new(big.Int).SetString("500000000000000000000", 10)
				hexVal := math.HexOrDecimal256(*val)
				return &hexVal
			}(), // 500 VTHO
			HasCode: false,
		},
		MockChainID: 1,
		MockGasPrice: &DynamicGasPrice{
			BaseFee: big.NewInt(1000000000000000000), // 1 VTHO
			Reward:  big.NewInt(500000000000000000),  // 0.5 VTHO
		},
		MockSyncProgress: 1.0,
		MockPeers: []Peer{
			{
				PeerID:      "peer1",
				BestBlockID: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
		},
		MockMempoolTxs: []*thor.Bytes32{
			func() *thor.Bytes32 {
				hash, _ := thor.ParseBytes32("0x1111111111111111111111111111111111111111111111111111111111111111")
				return &hash
			}(),
		},
		MockMempoolTx: &transactions.Transaction{
			ID: func() thor.Bytes32 {
				hash, _ := thor.ParseBytes32("0x1111111111111111111111111111111111111111111111111111111111111111")
				return hash
			}(),
			Type:       0,
			ChainTag:   0x27,
			BlockRef:   "0x1234567890abcdef",
			Expiration: 720,
			Clauses: api.Clauses{
				{
					To: func() *thor.Address {
						addr, _ := thor.ParseAddress(meshtests.TestAddress1)
						return &addr
					}(),
					Value: func() *math.HexOrDecimal256 {
						val, _ := new(big.Int).SetString("1000000000000000000", 10)
						hexVal := math.HexOrDecimal256(*val)
						return &hexVal
					}(), // 1 VET
					Data: "0x",
				},
			},
			GasPriceCoef: func() *uint8 { val := uint8(0); return &val }(),
			Gas:          21000,
			Origin: func() thor.Address {
				addr, _ := thor.ParseAddress(meshtests.FirstSoloAddress)
				return addr
			}(),
			Nonce:     func() math.HexOrDecimal64 { val := math.HexOrDecimal64(0x1234567890abcdef); return val }(),
			DependsOn: nil,
			Size:      200,
			Meta: &api.TxMeta{
				BlockID: func() thor.Bytes32 {
					hash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
					return hash
				}(),
				BlockNumber:    100,
				BlockTimestamp: uint64(time.Now().Unix()),
			},
		},
		MockMempoolStatus: &api.Status{
			Amount: 100,
		},
		MockCallResult: "0x0000000000000000000000000000000000000000000000000000000000000001",
	}
}

// Implement the VeChainClient interface

func (m *MockVeChainClient) GetBlock(revision string) (*api.JSONExpandedBlock, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}

	// Handle "best" revision
	if revision == "best" {
		return m.MockBlock, nil
	}

	// Handle numeric revision
	if m.MockBlockByNumber != nil {
		return m.MockBlockByNumber, nil
	}

	return m.MockBlock, nil
}

func (m *MockVeChainClient) GetBlockByNumber(blockNumber int64) (*api.JSONExpandedBlock, error) {
	// Convert to hex format and use GetBlock
	revision := fmt.Sprintf("0x%x", blockNumber)
	return m.GetBlock(revision)
}

func (m *MockVeChainClient) GetAccount(address string) (*api.Account, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}
	return m.MockAccount, nil
}

func (m *MockVeChainClient) GetAccountAtRevision(address string, revision string) (*api.Account, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}
	return m.MockAccount, nil
}

func (m *MockVeChainClient) GetChainID() (int, error) {
	if m.MockError != nil {
		return 0, m.MockError
	}
	return m.MockChainID, nil
}

func (m *MockVeChainClient) SubmitTransaction(vechainTx *tx.Transaction) (string, error) {
	if m.MockError != nil {
		return "", m.MockError
	}
	// Simulate transaction hash
	return "0x2222222222222222222222222222222222222222222222222222222222222222", nil
}

func (m *MockVeChainClient) GetDynamicGasPrice() (*DynamicGasPrice, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}
	return m.MockGasPrice, nil
}

func (m *MockVeChainClient) GetSyncProgress() (float64, error) {
	if m.MockError != nil {
		return 0, m.MockError
	}
	return m.MockSyncProgress, nil
}

func (m *MockVeChainClient) GetPeers() ([]Peer, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}
	return m.MockPeers, nil
}

func (m *MockVeChainClient) GetMempoolTransactions(origin *thor.Address) ([]*thor.Bytes32, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}
	return m.MockMempoolTxs, nil
}

func (m *MockVeChainClient) GetMempoolTransaction(txID *thor.Bytes32) (*transactions.Transaction, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}
	return m.MockMempoolTx, nil
}

func (m *MockVeChainClient) GetMempoolStatus() (*api.Status, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}
	return m.MockMempoolStatus, nil
}

func (m *MockVeChainClient) CallContract(contractAddress, callData string) (string, error) {
	if m.MockError != nil {
		return "", m.MockError
	}

	// If MockCallResults is set, use it for multiple responses
	if len(m.MockCallResults) > 0 {
		if m.MockCallIndex >= len(m.MockCallResults) {
			return "", fmt.Errorf("no more mock call results available")
		}
		result := m.MockCallResults[m.MockCallIndex]
		m.MockCallIndex++
		return result, nil
	}

	// Fallback to single MockCallResult
	return m.MockCallResult, nil
}

// Methods to configure the mock in tests

// SetMockError configures a simulated error
func (m *MockVeChainClient) SetMockError(err error) {
	m.MockError = err
}

// SetMockAccount configures the simulated account
func (m *MockVeChainClient) SetMockAccount(account *api.Account) {
	m.MockAccount = account
}

// SetMockBlock configures the simulated block
func (m *MockVeChainClient) SetMockBlock(block *api.JSONExpandedBlock) {
	m.MockBlock = block
}

// SetMockMempoolTx configures the simulated mempool transaction
func (m *MockVeChainClient) SetMockMempoolTx(tx *transactions.Transaction) {
	m.MockMempoolTx = tx
}

func (m *MockVeChainClient) SetMockCallResult(result string) {
	m.MockCallResult = result
}

// SetMockCallResults configures multiple call results for consecutive calls
func (m *MockVeChainClient) SetMockCallResults(results []string) {
	m.MockCallResults = results
	m.MockCallIndex = 0
}

// SetBlockByNumber configures the simulated block by number
func (m *MockVeChainClient) SetBlockByNumber(block *api.JSONExpandedBlock) {
	m.MockBlockByNumber = block
}

// GetTransaction simulates getting a transaction by ID
func (m *MockVeChainClient) GetTransaction(txID string) (*transactions.Transaction, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}
	return m.MockTransaction, nil
}

// GetTransactionReceipt simulates getting a transaction receipt by ID
func (m *MockVeChainClient) GetTransactionReceipt(txID string) (*api.Receipt, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}
	return m.MockReceipt, nil
}

// SetTransaction configures the simulated transaction
func (m *MockVeChainClient) SetTransaction(tx *transactions.Transaction) {
	m.MockTransaction = tx
}

// SetReceipt configures the simulated receipt
func (m *MockVeChainClient) SetReceipt(receipt *api.Receipt) {
	m.MockReceipt = receipt
}

// InspectClauses simulates inspecting clauses
func (m *MockVeChainClient) InspectClauses(batchCallData *api.BatchCallData, options ...thorclient.Option) ([]*api.CallResult, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}

	// If MockInspectClauses is set, return it
	if m.MockInspectClauses != nil {
		return m.MockInspectClauses, nil
	}

	// Default mock response
	return []*api.CallResult{
		{
			Data:      "0x0000000000000000000000000000000000000000000000000000000000000001",
			Events:    []*api.Event{},
			Transfers: []*api.Transfer{},
			GasUsed:   21000,
			Reverted:  false,
			VMError:   "",
		},
	}, nil
}

// SetInspectClausesResult configures the simulated inspect clauses result
func (m *MockVeChainClient) SetInspectClausesResult(results []*api.CallResult) {
	m.MockInspectClauses = results
}

// InspectClausesWithRevision simulates inspecting clauses with a specific revision
func (m *MockVeChainClient) InspectClausesWithRevision(batchCallData *api.BatchCallData, revision string) ([]*api.CallResult, error) {
	// For mock purposes, we ignore the revision and return the same results
	return m.InspectClauses(batchCallData)
}
