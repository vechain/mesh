package thor

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common/math"
	meshtests "github.com/vechain/mesh/tests"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/api/transactions"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/tx"
)

func TestNewVeChainClient(t *testing.T) {
	baseURL := "http://localhost:8669"
	client := NewVeChainClient(baseURL)

	if client == nil || client.client == nil {
		t.Errorf("NewVeChainClient() returned nil")
	}
}

func TestVeChainClient_GetBestBlock(t *testing.T) {
	client := NewVeChainClient("http://localhost:8669")

	// This will fail because we don't have a real Thor node running
	// but we can test that the method exists and handles errors properly
	_, err := client.GetBlock("best")
	if err == nil {
		t.Errorf("GetBestBlock() should return error when no Thor node is available")
	}
}

func TestVeChainClient_GetBlockByNumber(t *testing.T) {
	client := NewVeChainClient("http://localhost:8669")

	// Test with a valid block number
	_, err := client.GetBlockByNumber(100)
	if err == nil {
		t.Errorf("GetBlockByNumber() should return error when no Thor node is available")
	}
}

func TestVeChainClient_GetBlockByHash(t *testing.T) {
	client := NewVeChainClient("http://localhost:8669")

	// Test with a valid block hash
	blockHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	_, err := client.GetBlock(blockHash)
	if err == nil {
		t.Errorf("GetBlockByHash() should return error when no Thor node is available")
	}
}

func TestVeChainClient_GetAccount(t *testing.T) {
	client := NewVeChainClient("http://localhost:8669")

	// Test with a valid account address
	address := meshtests.FirstSoloAddress
	_, err := client.GetAccount(address)
	if err == nil {
		t.Errorf("GetAccount() should return error when no Thor node is available")
	}
}

func TestVeChainClient_GetChainID(t *testing.T) {
	client := NewVeChainClient("http://localhost:8669")

	// This will fail because we don't have a real Thor node running
	_, err := client.GetChainID()
	if err == nil {
		t.Errorf("GetChainID() should return error when no Thor node is available")
	}
}

func TestVeChainClient_SubmitTransaction(t *testing.T) {
	client := NewVeChainClient("http://localhost:8669")

	// Test with a valid transaction - create a proper transaction
	builder := tx.NewBuilder(tx.TypeLegacy)
	builder.ChainTag(0x27)
	blockRef := tx.BlockRef([8]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef})
	builder.BlockRef(blockRef)
	builder.Expiration(720)
	builder.Gas(21000)
	builder.GasPriceCoef(128)
	builder.Nonce(0x1234567890abcdef)

	validTx := builder.Build()

	_, err := client.SubmitTransaction(validTx)
	if err == nil {
		t.Errorf("SubmitTransaction() should return error when no Thor node is available")
	}
}

func TestVeChainClient_GetDynamicGasPrice(t *testing.T) {
	// Test with mock client to cover success path
	mockClient := NewMockVeChainClient()
	gasPrice, err := mockClient.GetDynamicGasPrice()
	if err != nil {
		t.Errorf("GetDynamicGasPrice() with mock client should not return error, got: %v", err)
	}
	if gasPrice == nil || gasPrice.BaseFee == nil || gasPrice.Reward == nil {
		t.Errorf("GetDynamicGasPrice() should return gas price or BaseFee and Reward are nil")
	}

	// Test with real client (will fail)
	client := NewVeChainClient("http://localhost:8669")
	_, err = client.GetDynamicGasPrice()
	if err == nil {
		t.Errorf("GetDynamicGasPrice() should return error when no Thor node is available")
	}
}

func TestVeChainClient_GetSyncProgress(t *testing.T) {
	// Test with mock client to cover success path
	mockClient := NewMockVeChainClient()
	progress, err := mockClient.GetSyncProgress()
	if err != nil {
		t.Errorf("GetSyncProgress() with mock client should not return error, got: %v", err)
	}
	if progress < 0 || progress > 1 {
		t.Errorf("GetSyncProgress() should return progress between 0 and 1, got: %v", progress)
	}

	// Test with real client (will fail)
	client := NewVeChainClient("http://localhost:8669")
	_, err = client.GetSyncProgress()
	if err == nil {
		t.Errorf("GetSyncProgress() should return error when no Thor node is available")
	}
}

func TestVeChainClient_GetPeers(t *testing.T) {
	client := NewVeChainClient("http://localhost:8669")

	// This will fail because we don't have a real Thor node running
	_, err := client.GetPeers()
	if err == nil {
		t.Errorf("GetPeers() should return error when no Thor node is available")
	}
}

func TestVeChainClient_GetMempoolTransactions(t *testing.T) {
	// Test with mock client to cover success path
	mockClient := NewMockVeChainClient()
	txs, err := mockClient.GetMempoolTransactions(nil)
	if err != nil {
		t.Errorf("GetMempoolTransactions() with mock client should not return error, got: %v", err)
	}
	if txs == nil {
		t.Errorf("GetMempoolTransactions() should return transaction list")
	}

	// Test with real client (will fail)
	client := NewVeChainClient("http://localhost:8669")
	_, err = client.GetMempoolTransactions(nil)
	if err == nil {
		t.Errorf("GetMempoolTransactions() should return error when no Thor node is available")
	}
}

func TestVeChainClient_GetMempoolTransaction(t *testing.T) {
	// Test with mock client to cover success path
	mockClient := NewMockVeChainClient()
	txHashStr := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	txHash, _ := thor.ParseBytes32(txHashStr)
	tx, err := mockClient.GetMempoolTransaction(&txHash)
	if err != nil {
		t.Errorf("GetMempoolTransaction() with mock client should not return error, got: %v", err)
	}
	if tx == nil {
		t.Errorf("GetMempoolTransaction() should return transaction")
	}

	// Test with real client (will fail)
	client := NewVeChainClient("http://localhost:8669")
	_, err = client.GetMempoolTransaction(&txHash)
	if err == nil {
		t.Errorf("GetMempoolTransaction() should return error when no Thor node is available")
	}
}

func TestVeChainClient_GetMempoolStatus(t *testing.T) {
	client := NewVeChainClient("http://localhost:8669")

	// This will fail because we don't have a real Thor node running
	_, err := client.GetMempoolStatus()
	if err == nil {
		t.Errorf("GetMempoolStatus() should return error when no Thor node is available")
	}
}

func TestVeChainClient_CallContract(t *testing.T) {
	client := NewVeChainClient("http://localhost:8669")

	// Test with valid contract call parameters
	contractAddress := "0x0000000000000000000000000000456e65726779"
	data := "0x1234567890abcdef"
	_, err := client.CallContract(contractAddress, data)
	if err == nil {
		t.Errorf("CallContract() should return error when no Thor node is available")
	}
}

func TestVeChainClient_GetTransaction(t *testing.T) {
	client := NewVeChainClient("http://localhost:8669")

	// Test with invalid transaction ID
	_, err := client.GetTransaction("invalid-tx-id")
	if err == nil {
		t.Errorf("GetTransaction() should return error when no Thor node is available")
	}
}

func TestVeChainClient_GetTransactionReceipt(t *testing.T) {
	client := NewVeChainClient("http://localhost:8669")

	// Test with invalid transaction ID
	_, err := client.GetTransactionReceipt("invalid-tx-id")
	if err == nil {
		t.Errorf("GetTransactionReceipt() should return error when no Thor node is available")
	}
}

// Tests for MockVeChainClient methods with 0% coverage
func TestMockVeChainClient_GetBlock(t *testing.T) {
	mockClient := NewMockVeChainClient()

	// Test with no mock block set
	block, err := mockClient.GetBlock("best")
	if err != nil {
		t.Errorf("GetBlock() error = %v, want nil", err)
	}
	if block == nil {
		t.Errorf("GetBlock() returned nil block")
	}
}

func TestMockVeChainClient_GetAccount(t *testing.T) {
	mockClient := NewMockVeChainClient()

	// Test with no mock account set
	account, err := mockClient.GetAccount(meshtests.FirstSoloAddress)
	if err != nil {
		t.Errorf("GetAccount() error = %v, want nil", err)
	}
	if account == nil {
		t.Errorf("GetAccount() returned nil account")
	}
}

func TestMockVeChainClient_GetChainID(t *testing.T) {
	mockClient := NewMockVeChainClient()

	chainID, err := mockClient.GetChainID()
	if err != nil {
		t.Errorf("GetChainID() error = %v, want nil", err)
	}
	if chainID == 0 {
		t.Errorf("GetChainID() returned 0")
	}
}

func TestMockVeChainClient_SubmitTransaction(t *testing.T) {
	mockClient := NewMockVeChainClient()

	// Create a mock transaction
	mockTx := &tx.Transaction{}
	txID, err := mockClient.SubmitTransaction(mockTx)
	if err != nil {
		t.Errorf("SubmitTransaction() error = %v, want nil", err)
	}
	if txID == "" {
		t.Errorf("SubmitTransaction() returned empty tx ID")
	}
}

func TestMockVeChainClient_GetPeers(t *testing.T) {
	mockClient := NewMockVeChainClient()

	peers, err := mockClient.GetPeers()
	if err != nil {
		t.Errorf("GetPeers() error = %v, want nil", err)
	}
	if peers == nil {
		t.Errorf("GetPeers() returned nil peers")
	}
}

func TestMockVeChainClient_GetMempoolStatus(t *testing.T) {
	mockClient := NewMockVeChainClient()

	status, err := mockClient.GetMempoolStatus()
	if err != nil {
		t.Errorf("GetMempoolStatus() error = %v, want nil", err)
	}
	if status == nil {
		t.Errorf("GetMempoolStatus() returned nil status")
	}
}

func TestMockVeChainClient_CallContract(t *testing.T) {
	mockClient := NewMockVeChainClient()

	result, err := mockClient.CallContract("0x1234567890123456789012345678901234567890", "0x1234")
	if err != nil {
		t.Errorf("CallContract() error = %v, want nil", err)
	}
	if result == "" {
		t.Errorf("CallContract() returned empty result")
	}
}

func TestMockVeChainClient_SetMockError(t *testing.T) {
	mockClient := NewMockVeChainClient()

	// Test setting mock error
	mockClient.SetMockError(fmt.Errorf("test error"))

	// Test that error is returned
	_, err := mockClient.GetAccount(meshtests.FirstSoloAddress)
	if err == nil {
		t.Errorf("GetAccount() should return error after SetMockError")
	}
}

func TestMockVeChainClient_SetMockAccount(t *testing.T) {
	mockClient := NewMockVeChainClient()

	// Test setting mock account
	balance := math.HexOrDecimal256{}
	err := balance.UnmarshalText([]byte("1000000000000000000"))
	if err != nil {
		t.Errorf("SetMockAccount() error = %v, want nil", err)
	}
	energy := math.HexOrDecimal256{}
	err = energy.UnmarshalText([]byte("1000000"))
	if err != nil {
		t.Errorf("SetMockAccount() error = %v, want nil", err)
	}

	mockAccount := &api.Account{
		Balance: &balance,
		Energy:  &energy,
	}
	mockClient.SetMockAccount(mockAccount)

	// Test that account is returned
	account, err := mockClient.GetAccount(meshtests.FirstSoloAddress)
	if err != nil {
		t.Errorf("GetAccount() error = %v, want nil", err)
	}
	if account == nil {
		t.Errorf("GetAccount() returned nil account")
	}
}

func TestMockVeChainClient_SetMockBlock(t *testing.T) {
	mockClient := NewMockVeChainClient()

	// Test setting mock block
	mockBlock := &api.JSONExpandedBlock{
		JSONBlockSummary: &api.JSONBlockSummary{
			Number: 100,
		},
	}
	mockClient.SetMockBlock(mockBlock)

	// Test that block is returned
	block, err := mockClient.GetBlock("best")
	if err != nil {
		t.Errorf("GetBlock() error = %v, want nil", err)
	}
	if block == nil {
		t.Errorf("GetBlock() returned nil block")
	}
}

func TestMockVeChainClient_SetMockMempoolTx(t *testing.T) {
	mockClient := NewMockVeChainClient()

	// Test setting mock mempool tx
	mockTx := &transactions.Transaction{}
	mockClient.SetMockMempoolTx(mockTx)

	// Test that mempool tx is returned
	address := thor.Address{}
	txs, err := mockClient.GetMempoolTransactions(&address)
	if err != nil {
		t.Errorf("GetMempoolTransactions() error = %v, want nil", err)
	}
	if len(txs) == 0 {
		t.Errorf("GetMempoolTransactions() returned empty list")
	}
}

func TestMockVeChainClient_SetMockCallResult(t *testing.T) {
	mockClient := NewMockVeChainClient()

	// Test setting mock call result
	mockClient.SetMockCallResult("0x1234567890abcdef")

	// Test that call result is returned
	result, err := mockClient.CallContract("0x1234567890123456789012345678901234567890", "0x1234")
	if err != nil {
		t.Errorf("CallContract() error = %v, want nil", err)
	}
	if result != "0x1234567890abcdef" {
		t.Errorf("CallContract() result = %v, want 0x1234567890abcdef", result)
	}
}

func TestMockVeChainClient_SetMockCallResults(t *testing.T) {
	mockClient := NewMockVeChainClient()

	// Test setting multiple mock call results
	results := []string{"0x1111", "0x2222", "0x3333"}
	mockClient.SetMockCallResults(results)

	// Test that results are returned sequentially
	for i, expected := range results {
		result, err := mockClient.CallContract("0x1234567890123456789012345678901234567890", "0x1234")
		if err != nil {
			t.Errorf("CallContract() [%d] error = %v, want nil", i, err)
		}
		if result != expected {
			t.Errorf("CallContract() [%d] result = %v, want %v", i, result, expected)
		}
	}
}

func TestMockVeChainClient_SetBlockByNumber(t *testing.T) {
	mockClient := NewMockVeChainClient()

	// Test setting block by number
	mockBlock := &api.JSONExpandedBlock{
		JSONBlockSummary: &api.JSONBlockSummary{
			Number: 100,
		},
	}
	mockClient.SetBlockByNumber(mockBlock)

	// Test that block is returned
	block, err := mockClient.GetBlock("100")
	if err != nil {
		t.Errorf("GetBlock() error = %v, want nil", err)
	}
	if block == nil {
		t.Errorf("GetBlock() returned nil block")
	}
}

func TestMockVeChainClient_SetTransaction(t *testing.T) {
	mockClient := NewMockVeChainClient()

	// Test setting mock transaction
	mockTx := &transactions.Transaction{}
	mockClient.SetTransaction(mockTx)

	// Test that transaction is returned
	tx, err := mockClient.GetTransaction("0x1234567890abcdef")
	if err != nil {
		t.Errorf("GetTransaction() error = %v, want nil", err)
	}
	if tx == nil {
		t.Errorf("GetTransaction() returned nil transaction")
	}
}

func TestMockVeChainClient_SetReceipt(t *testing.T) {
	mockClient := NewMockVeChainClient()

	// Test setting mock receipt
	mockReceipt := &api.Receipt{
		GasUsed: 21000,
	}
	mockClient.SetReceipt(mockReceipt)

	// Test that receipt is returned
	receipt, err := mockClient.GetTransactionReceipt("0x1234567890abcdef")
	if err != nil {
		t.Errorf("GetTransactionReceipt() error = %v, want nil", err)
	}
	if receipt == nil {
		t.Errorf("GetTransactionReceipt() returned nil receipt")
	}
}
