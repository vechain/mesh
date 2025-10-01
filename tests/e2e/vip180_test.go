package e2e

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common"
	meshcommon "github.com/vechain/mesh/common"
	meshcrypto "github.com/vechain/mesh/common/crypto"
	meshvip180 "github.com/vechain/mesh/common/vip180"
	meshthor "github.com/vechain/mesh/thor"
	thorTx "github.com/vechain/thor/v2/tx"
)

// TestVIP180Solo tests the complete VIP180 flow in solo mode
func TestVIP180Solo(t *testing.T) {
	t.Log("Starting VIP180 E2E test in solo mode...")

	// Get test configuration
	config := GetVIP180TestConfig()
	thorClient := meshthor.NewVeChainClient(config.ThorURL)
	networkIdentifier := CreateTestNetworkIdentifier(config.Network)

	// Step 1: Deploy VIP180 contract using Thor Solo API
	t.Log("Step 1: Deploying VIP180 contract using Thor Solo API")
	contractAddress, err := deployVIP180Contract(thorClient, config)
	if err != nil {
		t.Fatalf("Failed to deploy VIP180 contract: %v", err)
	}
	config.ContractAddress = contractAddress
	t.Logf("✅ VIP180 contract deployed at address: %s", contractAddress)

	// Step 2: Validate contract deployment by calling a method
	t.Log("Step 2: Validating contract deployment")
	if err := validateContractDeployment(thorClient, config); err != nil {
		t.Fatalf("Failed to validate contract deployment: %v", err)
	}
	t.Log("✅ Contract deployment validated")

	client := NewHTTPClient(config.BaseURL, config.TimeoutSeconds)
	// Step 3: Test VIP180 transfer using Mesh construction endpoints
	t.Log("Step 3: Testing VIP180 transfer using Mesh construction endpoints")
	if err := testVIP180Transfer(t, client, networkIdentifier, config); err != nil {
		t.Fatalf("Failed to test VIP180 transfer: %v", err)
	}

	t.Log("✅ All VIP180 E2E test steps completed successfully!")
}

// deployVIP180Contract deploys the VIP180 contract using Thor Solo API
func deployVIP180Contract(client *meshthor.VeChainClient, config *VIP180TestConfig) (string, error) {
	// Create deployment transaction using Thor transaction builder
	deploymentTx, err := createDeploymentTransaction(client, config)
	if err != nil {
		return "", fmt.Errorf("failed to create deployment transaction: %v", err)
	}

	// Sign the transaction
	signedTx, err := signTransaction(deploymentTx, config.SenderPrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign deployment transaction: %v", err)
	}

	// Submit using VeChain client (same as Mesh uses)
	fmt.Printf("Submitting signed transaction using VeChain client\n")
	txHash, err := client.SubmitTransaction(signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to submit deployment transaction: %v", err)
	}
	fmt.Printf("Transaction submitted with hash: %s\n", txHash)

	// Wait for transaction to be mined and get contract address
	contractAddress, err := waitForContractDeployment(client, txHash)
	if err != nil {
		return "", fmt.Errorf("failed to get contract address: %v", err)
	}

	return contractAddress, nil
}

// validateContractDeployment validates that the contract was deployed correctly
func validateContractDeployment(client *meshthor.VeChainClient, config *VIP180TestConfig) error {
	// Create VIP180 contract wrapper
	vip180Contract, err := meshvip180.NewVIP180Contract(config.ContractAddress, client)
	if err != nil {
		return fmt.Errorf("failed to create VIP180 contract wrapper: %v", err)
	}

	// Test calling a view function (name)
	name, err := vip180Contract.Name()
	if err != nil {
		return fmt.Errorf("failed to call name() function: %v", err)
	}

	if name != config.TokenName {
		return fmt.Errorf("unexpected token name: got %s, want %s", name, config.TokenName)
	}

	// Test calling another view function (totalSupply)
	totalSupply, err := vip180Contract.TotalSupply()
	if err != nil {
		return fmt.Errorf("failed to call totalSupply() function: %v", err)
	}

	if totalSupply.String() != config.TokenTotalSupply {
		return fmt.Errorf("unexpected total supply: got %s, want %s", totalSupply.String(), config.TokenTotalSupply)
	}

	return nil
}

// testVIP180Transfer tests VIP180 transfer using Mesh construction endpoints
func testVIP180Transfer(t *testing.T, client *HTTPClient, networkIdentifier *types.NetworkIdentifier, config *VIP180TestConfig) error {
	// Create VIP180 transfer operations
	operations := createVIP180TransferOperations(config)

	// Test construction preprocess
	t.Log("Testing /construction/preprocess for VIP180 transfer")
	preprocessResp, err := testConstructionPreprocess(client, networkIdentifier, operations, config.TestConfig, meshcommon.TransactionTypeLegacy)
	if err != nil {
		return fmt.Errorf("construction preprocess test failed: %v", err)
	}

	// Test construction metadata
	t.Log("Testing /construction/metadata for VIP180 transfer")
	metadataResp, err := testConstructionMetadata(client, networkIdentifier, preprocessResp, meshcommon.TransactionTypeLegacy)
	if err != nil {
		return fmt.Errorf("construction metadata test failed: %v", err)
	}

	// Test construction payloads
	t.Log("Testing /construction/payloads for VIP180 transfer")
	payloadsResp, err := testConstructionPayloads(client, networkIdentifier, metadataResp, config.TestConfig, meshcommon.TransactionTypeLegacy)
	if err != nil {
		return fmt.Errorf("construction payloads test failed: %v", err)
	}

	// Sign the payload
	t.Log("Signing VIP180 transfer payload")
	payloadHex := fmt.Sprintf("%x", payloadsResp.Payloads[0].Bytes)
	signature, err := meshcrypto.NewSigningHandler(config.SenderPrivateKey).SignPayload(payloadHex)
	if err != nil {
		return fmt.Errorf("failed to sign payload: %v", err)
	}

	// Test construction combine
	t.Log("Testing /construction/combine for VIP180 transfer")
	combineResp, err := testConstructionCombine(client, networkIdentifier, payloadsResp, signature)
	if err != nil {
		return fmt.Errorf("construction combine test failed: %v", err)
	}

	// Test construction hash
	t.Log("Testing /construction/hash for VIP180 transfer")
	_, err = testConstructionHash(client, networkIdentifier, combineResp)
	if err != nil {
		return fmt.Errorf("construction hash test failed: %v", err)
	}

	// Test construction submit
	t.Log("Testing /construction/submit for VIP180 transfer")
	submitResp, err := testConstructionSubmit(client, networkIdentifier, combineResp)
	if err != nil {
		return fmt.Errorf("construction submit test failed: %v", err)
	}

	t.Logf("✅ VIP180 transfer submitted with hash: %s", submitResp.TransactionIdentifier.Hash)

	// Wait for transaction to be confirmed using the existing search function
	t.Log("Waiting for VIP180 transfer to be confirmed...")
	retries := 3
	delayInSeconds := 1
	searchResp, err := testSearchTransactionsWithRetry(client, networkIdentifier, submitResp.TransactionIdentifier, retries, delayInSeconds)
	if err != nil {
		return fmt.Errorf("failed to confirm VIP180 transfer: %v", err)
	}
	if err := ValidateSearchTransactionsResponse(searchResp); err != nil {
		return fmt.Errorf("search transactions response validation failed: %v", err)
	}
	t.Logf("Search transactions response: total count = %d, transactions count = %d",
		searchResp.TotalCount, len(searchResp.Transactions))

	t.Log("✅ VIP180 transfer confirmed in blockchain")
	return nil
}

// createDeploymentTransaction creates a transaction to deploy the VIP180 contract
func createDeploymentTransaction(client *meshthor.VeChainClient, config *VIP180TestConfig) (*thorTx.Transaction, error) {
	// Create constructor data with parameters: name, symbol, decimals, bridge
	constructorData := encodeConstructorData(config)

	// Combine bytecode with constructor data
	fullBytecode := append(config.ContractBytecode, constructorData...)

	// Get the best block to use as BlockRef
	bestBlock, err := client.GetBlock("best")
	if err != nil {
		return nil, fmt.Errorf("failed to get best block: %v", err)
	}

	// Create BlockRef from the best block ID (use first 8 bytes)
	blockRefBytes := common.FromHex(bestBlock.ID.String())
	if len(blockRefBytes) < 8 {
		return nil, fmt.Errorf("invalid block ID length: %d", len(blockRefBytes))
	}
	blockRef := thorTx.BlockRef{}
	copy(blockRef[:], blockRefBytes[:8])

	// Create deployment clause
	clause := thorTx.NewClause(nil)
	clause = clause.WithValue(big.NewInt(0))
	clause = clause.WithData(fullBytecode)

	// Create transaction builder for legacy transaction
	builder := thorTx.NewBuilder(thorTx.TypeLegacy)
	builder.ChainTag(0xf6) // Solo network chain tag
	builder.BlockRef(blockRef)
	builder.Expiration(32)
	builder.Gas(2000000)
	builder.GasPriceCoef(0)
	builder.Nonce(0)
	builder.Clause(clause)

	return builder.Build(), nil
}

// signTransaction signs a transaction using the private key
func signTransaction(tx *thorTx.Transaction, privateKey string) (*thorTx.Transaction, error) {
	signature, err := meshcrypto.NewSigningHandler(privateKey).SignPayload(tx.SigningHash().String())
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return nil, fmt.Errorf("error decoding signature: %v", err)
	}
	signedTx := tx.WithSignature(signatureBytes)

	return signedTx, nil
}

// waitForContractDeployment waits for the contract deployment transaction to be mined
func waitForContractDeployment(vechainClient *meshthor.VeChainClient, txHash string) (string, error) {
	// Wait for transaction to be mined (retry 10 times with 2 seconds delay)
	maxRetries := 10
	delaySeconds := 2

	for attempt := 1; attempt <= maxRetries; attempt++ {
		receipt, err := vechainClient.GetTransactionReceipt(txHash)
		if err == nil && receipt != nil {
			// Check if transaction was reverted
			if receipt.Reverted {
				return "", fmt.Errorf("contract deployment transaction was reverted")
			}

			// Extract contract address from the first output
			if len(receipt.Outputs) == 0 {
				return "", fmt.Errorf("no outputs found in transaction receipt")
			}

			contractAddress := receipt.Outputs[0].ContractAddress
			if contractAddress == nil {
				return "", fmt.Errorf("no contract address found in transaction receipt")
			}

			return contractAddress.String(), nil
		}

		if attempt < maxRetries {
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		}
	}

	return "", fmt.Errorf("transaction not mined after %d attempts", maxRetries)
}

// createVIP180TransferOperations creates operations for VIP180 transfer
func createVIP180TransferOperations(config *VIP180TestConfig) []*types.Operation {
	transferAmount := "1000000000000000000" // 1 token with 18 decimals

	// Create VIP180 currency
	vip180Currency := &types.Currency{
		Symbol:   config.TokenSymbol,
		Decimals: int32(config.TokenDecimals),
		Metadata: map[string]any{
			"contractAddress": config.ContractAddress,
		},
	}

	status := meshcommon.OperationStatusNone
	return []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 0,
			},
			Type:   meshcommon.OperationTypeTransfer,
			Status: &status,
			Account: &types.AccountIdentifier{
				Address: config.RecipientAddress,
			},
			Amount: &types.Amount{
				Value:    transferAmount,
				Currency: vip180Currency,
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 1,
			},
			Type:   meshcommon.OperationTypeTransfer,
			Status: &status,
			Account: &types.AccountIdentifier{
				Address: config.SenderAddress,
			},
			Amount: &types.Amount{
				Value:    "-" + transferAmount,
				Currency: vip180Currency,
			},
		},
	}
}
