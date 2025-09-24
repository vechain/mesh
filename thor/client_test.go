package thor

import (
	"testing"

	"github.com/vechain/thor/v2/thor"
)

func TestNewVeChainClient(t *testing.T) {
	baseURL := "http://localhost:8669"
	client := NewVeChainClient(baseURL)

	if client == nil {
		t.Errorf("NewVeChainClient() returned nil")
	}
	if client.client == nil {
		t.Errorf("NewVeChainClient() client is nil")
	}
}

func TestVeChainClient_GetBestBlock(t *testing.T) {
	client := NewVeChainClient("http://localhost:8669")

	// This will fail because we don't have a real Thor node running
	// but we can test that the method exists and handles errors properly
	_, err := client.GetBestBlock()
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
	_, err := client.GetBlockByHash(blockHash)
	if err == nil {
		t.Errorf("GetBlockByHash() should return error when no Thor node is available")
	}
}

func TestVeChainClient_GetAccount(t *testing.T) {
	client := NewVeChainClient("http://localhost:8669")

	// Test with a valid account address
	address := "0xf077b491b355e64048ce21e3a6fc4751eeea77fa"
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

	// Test with a valid transaction
	txData := []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef}
	_, err := client.SubmitTransaction(txData)
	if err == nil {
		t.Errorf("SubmitTransaction() should return error when no Thor node is available")
	}
}

func TestVeChainClient_GetDynamicGasPrice(t *testing.T) {
	client := NewVeChainClient("http://localhost:8669")

	// This will fail because we don't have a real Thor node running
	_, err := client.GetDynamicGasPrice()
	if err == nil {
		t.Errorf("GetDynamicGasPrice() should return error when no Thor node is available")
	}
}

func TestVeChainClient_GetSyncProgress(t *testing.T) {
	client := NewVeChainClient("http://localhost:8669")

	// This will fail because we don't have a real Thor node running
	_, err := client.GetSyncProgress()
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
	client := NewVeChainClient("http://localhost:8669")

	// Test without origin filter
	_, err := client.GetMempoolTransactions(nil)
	if err == nil {
		t.Errorf("GetMempoolTransactions() should return error when no Thor node is available")
	}
}

func TestVeChainClient_GetMempoolTransaction(t *testing.T) {
	client := NewVeChainClient("http://localhost:8669")

	// Test with a valid transaction hash
	txHashStr := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	txHash, _ := thor.ParseBytes32(txHashStr)
	_, err := client.GetMempoolTransaction(&txHash)
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
