package e2e

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshutils "github.com/vechain/mesh/utils"
)

// TestConstructionMempoolSolo tests the complete construction flow in solo mode
// This test follows the exact sequence from vechain/rosetta construction.mempool.solo.test
// It tests both legacy and dynamic transaction types in the same flow
func TestConstructionMempoolSolo(t *testing.T) {
	t.Log("Starting construction.mempool.solo test sequence...")

	// Get test configuration
	config := GetTestConfig()
	client := NewHTTPClient(config.BaseURL, config.TimeoutSeconds)
	networkIdentifier := CreateTestNetworkIdentifier(config.Network)

	// Test both transaction types in the same flow
	transactionTypes := []string{TransactionTypeLegacy, TransactionTypeDynamic}

	for _, transactionType := range transactionTypes {
		t.Run(transactionType+"Transaction", func(t *testing.T) {
			t.Logf("Testing %s transaction flow...", transactionType)
			testTransactionFlow(t, client, networkIdentifier, config, transactionType)
		})
	}

	t.Log("✅ All construction.mempool.solo test steps completed successfully!")
}

// testTransactionFlow tests the complete transaction flow for a specific transaction type
func testTransactionFlow(t *testing.T, client *HTTPClient, networkIdentifier *types.NetworkIdentifier, config *TestConfig, transactionType string) {
	// Step 1: Network List
	t.Log("Step 1: Testing /network/list")
	networkListResp, err := testNetworkList(client)
	if err != nil {
		t.Fatalf("Network list test failed: %v", err)
	}
	t.Logf("Network list response: %+v", networkListResp)

	// Step 2: Network Options
	t.Log("Step 2: Testing /network/options")
	networkOptionsResp, err := testNetworkOptions(client, networkIdentifier)
	if err != nil {
		t.Fatalf("Network options test failed: %v", err)
	}
	t.Logf("Network options response: %+v", networkOptionsResp)

	// Step 3: Network Status
	t.Log("Step 3: Testing /network/status")
	networkStatusResp, err := testNetworkStatus(client, networkIdentifier)
	if err != nil {
		t.Fatalf("Network status test failed: %v", err)
	}
	t.Logf("Network status response: %+v", networkStatusResp)

	// Step 4: Construction Preprocess
	t.Logf("Step 4: Testing /construction/preprocess for %s transaction", transactionType)
	preprocessResp, err := testConstructionPreprocess(client, networkIdentifier, config, transactionType)
	if err != nil {
		t.Fatalf("Construction preprocess test failed: %v", err)
	}
	t.Logf("Preprocess response: %+v", preprocessResp)

	// Step 5: Construction Metadata
	t.Logf("Step 5: Testing /construction/metadata for %s transaction", transactionType)
	metadataResp, err := testConstructionMetadata(client, networkIdentifier, preprocessResp, transactionType)
	if err != nil {
		t.Fatalf("Construction metadata test failed: %v", err)
	}
	t.Logf("Metadata response: %+v", metadataResp)

	// Validate transaction type in metadata
	if err := ValidateTransactionTypeInMetadata(metadataResp.Metadata, transactionType); err != nil {
		t.Fatalf("Metadata validation failed: %v", err)
	}

	// Step 6: Construction Payloads
	t.Logf("Step 6: Testing /construction/payloads for %s transaction", transactionType)
	payloadsResp, err := testConstructionPayloads(client, networkIdentifier, metadataResp, config, transactionType)
	if err != nil {
		t.Fatalf("Construction payloads test failed: %v", err)
	}
	t.Logf("Payloads response: %+v", payloadsResp)

	// Step 7: Sign the payload
	t.Log("Step 7: Signing payload")
	payloadHex := fmt.Sprintf("%x", payloadsResp.Payloads[0].Bytes)
	signature, err := meshutils.SignPayload(config.SenderPrivateKey, payloadHex)
	if err != nil {
		t.Fatalf("Failed to sign payload: %v", err)
	}
	t.Logf("Generated signature: %s", signature)

	// Step 8: Construction Combine
	t.Log("Step 8: Testing /construction/combine")
	combineResp, err := testConstructionCombine(client, networkIdentifier, payloadsResp, signature)
	if err != nil {
		t.Fatalf("Construction combine test failed: %v", err)
	}
	t.Logf("Combine response: %+v", combineResp)

	// Step 9: Construction Hash
	t.Log("Step 9: Testing /construction/hash")
	hashResp, err := testConstructionHash(client, networkIdentifier, combineResp)
	if err != nil {
		t.Fatalf("Construction hash test failed: %v", err)
	}
	t.Logf("Hash response: %+v", hashResp)

	// Step 10: Construction Submit
	t.Log("Step 10: Testing /construction/submit")
	submitResp, err := testConstructionSubmit(client, networkIdentifier, combineResp)
	if err != nil {
		t.Fatalf("Construction submit test failed: %v", err)
	}
	t.Logf("Submit response: %+v", submitResp)

	// Step 11: Mempool
	t.Log("Step 11: Testing /mempool")
	mempoolResp, err := testMempool(client, networkIdentifier)
	if err != nil {
		t.Fatalf("Mempool test failed: %v", err)
	}
	t.Logf("Mempool response: %+v", mempoolResp)

	// Step 12: Mempool Transaction
	t.Log("Step 12: Testing /mempool/transaction")
	mempoolTxResp, err := testMempoolTransaction(client, networkIdentifier, submitResp.TransactionIdentifier)
	if err != nil {
		t.Fatalf("Mempool transaction test failed: %v", err)
	}
	t.Logf("Mempool transaction response: %+v", mempoolTxResp)

	t.Logf("✅ All construction steps completed successfully for %s transaction!", transactionType)
}

// testNetworkList tests the /network/list endpoint
func testNetworkList(client *HTTPClient) (*types.NetworkListResponse, error) {
	request := &types.MetadataRequest{}

	resp, err := client.Post("/network/list", request)
	if err != nil {
		return nil, err
	}

	var response types.NetworkListResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, err
	}

	// Validate response
	if err := ValidateNetworkListResponse(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// testNetworkOptions tests the /network/options endpoint
func testNetworkOptions(client *HTTPClient, networkIdentifier *types.NetworkIdentifier) (*types.NetworkOptionsResponse, error) {
	request := &types.NetworkRequest{
		NetworkIdentifier: networkIdentifier,
	}

	resp, err := client.Post("/network/options", request)
	if err != nil {
		return nil, err
	}

	var response types.NetworkOptionsResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, err
	}

	// Validate response
	if err := ValidateNetworkOptionsResponse(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// testNetworkStatus tests the /network/status endpoint
func testNetworkStatus(client *HTTPClient, networkIdentifier *types.NetworkIdentifier) (*types.NetworkStatusResponse, error) {
	request := &types.NetworkRequest{
		NetworkIdentifier: networkIdentifier,
	}

	resp, err := client.Post("/network/status", request)
	if err != nil {
		return nil, err
	}

	var response types.NetworkStatusResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, err
	}

	// Validate response
	if err := ValidateNetworkStatusResponse(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// testConstructionPreprocess tests the /construction/preprocess endpoint
func testConstructionPreprocess(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, config *TestConfig, transactionType string) (*types.ConstructionPreprocessResponse, error) {
	var operations []*types.Operation

	if transactionType == TransactionTypeLegacy {
		operations = CreateLegacyTransactionOperations(config.SenderAddress, config.RecipientAddress, config.TransferAmount)
	} else {
		operations = CreateDynamicTransactionOperations(config.SenderAddress, config.RecipientAddress, config.TransferAmount)
	}

	request := &types.ConstructionPreprocessRequest{
		NetworkIdentifier: networkIdentifier,
		Operations:        operations,
	}

	resp, err := client.Post("/construction/preprocess", request)
	if err != nil {
		return nil, err
	}

	var response types.ConstructionPreprocessResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, err
	}

	// Validate response
	if err := ValidateConstructionPreprocessResponse(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// testConstructionMetadata tests the /construction/metadata endpoint
func testConstructionMetadata(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, preprocessResp *types.ConstructionPreprocessResponse, transactionType string) (*types.ConstructionMetadataResponse, error) {
	request := &types.ConstructionMetadataRequest{
		NetworkIdentifier: networkIdentifier,
		Options:           preprocessResp.Options,
	}

	// Add transaction type to options
	if request.Options == nil {
		request.Options = make(map[string]any)
	}
	request.Options["transactionType"] = transactionType

	resp, err := client.Post("/construction/metadata", request)
	if err != nil {
		return nil, err
	}

	var response types.ConstructionMetadataResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, err
	}

	// Validate response
	if err := ValidateConstructionMetadataResponse(&response); err != nil {
		return nil, err
	}

	// Validate type-specific fields
	if transactionType == TransactionTypeLegacy {
		if err := ValidateLegacyMetadataFields(response.Metadata); err != nil {
			return nil, err
		}
	} else if transactionType == TransactionTypeDynamic {
		if err := ValidateDynamicMetadataFields(response.Metadata); err != nil {
			return nil, err
		}
	}

	return &response, nil
}

// testConstructionPayloads tests the /construction/payloads endpoint
func testConstructionPayloads(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, metadataResp *types.ConstructionMetadataResponse, config *TestConfig, transactionType string) (*types.ConstructionPayloadsResponse, error) {
	var operations []*types.Operation

	if transactionType == TransactionTypeLegacy {
		operations = CreateLegacyTransactionOperations(config.SenderAddress, config.RecipientAddress, config.TransferAmount)
	} else {
		operations = CreateDynamicTransactionOperations(config.SenderAddress, config.RecipientAddress, config.TransferAmount)
	}

	request := &types.ConstructionPayloadsRequest{
		NetworkIdentifier: networkIdentifier,
		Operations:        operations,
		Metadata:          metadataResp.Metadata,
		PublicKeys:        []*types.PublicKey{CreateTestPublicKey()},
	}

	resp, err := client.Post("/construction/payloads", request)
	if err != nil {
		return nil, err
	}

	var response types.ConstructionPayloadsResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, err
	}

	// Validate response
	if err := ValidateConstructionPayloadsResponse(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// testConstructionCombine tests the /construction/combine endpoint
func testConstructionCombine(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, payloadsResp *types.ConstructionPayloadsResponse, signature string) (*types.ConstructionCombineResponse, error) {
	request := &types.ConstructionCombineRequest{
		NetworkIdentifier:   networkIdentifier,
		UnsignedTransaction: payloadsResp.UnsignedTransaction,
		Signatures: []*types.Signature{
			{
				SigningPayload: payloadsResp.Payloads[0],
				PublicKey:      CreateTestPublicKey(),
				SignatureType:  "ecdsa_recovery",
				Bytes:          func() []byte { b, _ := hex.DecodeString(signature); return b }(),
			},
		},
	}

	resp, err := client.Post("/construction/combine", request)
	if err != nil {
		return nil, err
	}

	var response types.ConstructionCombineResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, err
	}

	// Validate response
	if err := ValidateConstructionCombineResponse(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// testConstructionHash tests the /construction/hash endpoint
func testConstructionHash(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, combineResp *types.ConstructionCombineResponse) (*types.TransactionIdentifierResponse, error) {
	request := &types.ConstructionHashRequest{
		NetworkIdentifier: networkIdentifier,
		SignedTransaction: combineResp.SignedTransaction,
	}

	resp, err := client.Post("/construction/hash", request)
	if err != nil {
		return nil, err
	}

	var response types.TransactionIdentifierResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, err
	}

	// Validate response
	if err := ValidateTransactionIdentifierResponse(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// testConstructionSubmit tests the /construction/submit endpoint
func testConstructionSubmit(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, combineResp *types.ConstructionCombineResponse) (*types.TransactionIdentifierResponse, error) {
	request := &types.ConstructionSubmitRequest{
		NetworkIdentifier: networkIdentifier,
		SignedTransaction: combineResp.SignedTransaction,
	}

	resp, err := client.Post("/construction/submit", request)
	if err != nil {
		return nil, err
	}

	var response types.TransactionIdentifierResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, err
	}

	// Validate response
	if err := ValidateTransactionIdentifierResponse(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// testMempool tests the /mempool endpoint
func testMempool(client *HTTPClient, networkIdentifier *types.NetworkIdentifier) (*types.MempoolResponse, error) {
	request := &types.NetworkRequest{
		NetworkIdentifier: networkIdentifier,
	}

	resp, err := client.Post("/mempool", request)
	if err != nil {
		return nil, err
	}

	var response types.MempoolResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// testMempoolTransaction tests the /mempool/transaction endpoint
func testMempoolTransaction(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, txID *types.TransactionIdentifier) (*types.MempoolTransactionResponse, error) {
	request := &types.MempoolTransactionRequest{
		NetworkIdentifier:     networkIdentifier,
		TransactionIdentifier: txID,
	}

	resp, err := client.Post("/mempool/transaction", request)
	if err != nil {
		return nil, err
	}

	var response types.MempoolTransactionResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, err
	}

	// Validate response
	if err := ValidateMempoolTransactionResponse(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
