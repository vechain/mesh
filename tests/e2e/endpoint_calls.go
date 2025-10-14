package e2e

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
)

// testNetworkList tests the /network/list endpoint
func testNetworkList(client *HTTPClient) (*types.NetworkListResponse, error) {
	request := &types.MetadataRequest{}

	resp, err := client.Post(meshcommon.NetworkListEndpoint, request)
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

	resp, err := client.Post(meshcommon.NetworkOptionsEndpoint, request)
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

	resp, err := client.Post(meshcommon.NetworkStatusEndpoint, request)
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

// testConstructionDerive tests the construction/derive endpoint
func testConstructionDerive(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, publicKey *types.PublicKey) (*types.ConstructionDeriveResponse, error) {
	request := &types.ConstructionDeriveRequest{
		NetworkIdentifier: networkIdentifier,
		PublicKey:         publicKey,
	}

	resp, err := client.Post(meshcommon.ConstructionDeriveEndpoint, request)
	if err != nil {
		return nil, err
	}

	var deriveResp types.ConstructionDeriveResponse
	if err := ParseResponse(resp, &deriveResp); err != nil {
		return nil, err
	}

	return &deriveResp, nil
}

// testConstructionPreprocess tests the /construction/preprocess endpoint
func testConstructionPreprocess(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, operations []*types.Operation, config *TestConfig, transactionType string) (*types.ConstructionPreprocessResponse, error) {
	if operations == nil {
		if transactionType == meshcommon.TransactionTypeLegacy {
			operations = CreateLegacyTransactionOperations(config.SenderAddress, config.RecipientAddress, config.TransferAmount)
		} else {
			operations = CreateDynamicTransactionOperations(config.SenderAddress, config.RecipientAddress, config.TransferAmount)
		}
	}

	request := &types.ConstructionPreprocessRequest{
		NetworkIdentifier: networkIdentifier,
		Operations:        operations,
	}

	resp, err := client.Post(meshcommon.ConstructionPreprocessEndpoint, request)
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

	resp, err := client.Post(meshcommon.ConstructionMetadataEndpoint, request)
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
	switch transactionType {
	case meshcommon.TransactionTypeLegacy:
		if err := ValidateLegacyMetadataFields(response.Metadata); err != nil {
			return nil, err
		}
	case meshcommon.TransactionTypeDynamic:
		if err := ValidateDynamicMetadataFields(response.Metadata); err != nil {
			return nil, err
		}
	}

	return &response, nil
}

// testConstructionPayloadsWithMetadata tests construction/payloads with provided metadata
func testConstructionPayloadsWithMetadata(
	client *HTTPClient,
	networkIdentifier *types.NetworkIdentifier,
	operations []*types.Operation,
	publicKeys []*types.PublicKey,
	metadata map[string]any,
) (*types.ConstructionPayloadsResponse, error) {
	request := &types.ConstructionPayloadsRequest{
		NetworkIdentifier: networkIdentifier,
		Operations:        operations,
		PublicKeys:        publicKeys,
		Metadata:          metadata,
	}

	resp, err := client.Post(meshcommon.ConstructionPayloadsEndpoint, request)
	if err != nil {
		return nil, err
	}

	var payloadsResp types.ConstructionPayloadsResponse
	if err := ParseResponse(resp, &payloadsResp); err != nil {
		return nil, err
	}

	return &payloadsResp, nil
}

// testConstructionPayloads tests the /construction/payloads endpoint
func testConstructionPayloads(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, metadataResp *types.ConstructionMetadataResponse, config *TestConfig, transactionType string) (*types.ConstructionPayloadsResponse, error) {
	var operations []*types.Operation

	if transactionType == meshcommon.TransactionTypeLegacy {
		operations = CreateLegacyTransactionOperations(config.SenderAddress, config.RecipientAddress, config.TransferAmount)
	} else {
		operations = CreateDynamicTransactionOperations(config.SenderAddress, config.RecipientAddress, config.TransferAmount)
	}

	response, err := testConstructionPayloadsWithMetadata(client, networkIdentifier, operations, []*types.PublicKey{CreateTestPublicKey()}, metadataResp.Metadata)
	if err != nil {
		return nil, err
	}

	// Validate response
	if err := ValidateConstructionPayloadsResponse(response); err != nil {
		return nil, err
	}

	return response, nil
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

	resp, err := client.Post(meshcommon.ConstructionCombineEndpoint, request)
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

// testConstructionCombineWithSignatures tests the /construction/combine endpoint with multiple signatures
func testConstructionCombineWithSignatures(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, payloadsResp *types.ConstructionPayloadsResponse, signatures []*types.Signature) (*types.ConstructionCombineResponse, error) {
	request := &types.ConstructionCombineRequest{
		NetworkIdentifier:   networkIdentifier,
		UnsignedTransaction: payloadsResp.UnsignedTransaction,
		Signatures:          signatures,
	}

	resp, err := client.Post(meshcommon.ConstructionCombineEndpoint, request)
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

	resp, err := client.Post(meshcommon.ConstructionHashEndpoint, request)
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

	resp, err := client.Post(meshcommon.ConstructionSubmitEndpoint, request)
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

	resp, err := client.Post(meshcommon.MempoolEndpoint, request)
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

	resp, err := client.Post(meshcommon.MempoolTransactionEndpoint, request)
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

// testBlock tests the /block endpoint
func testBlock(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, blockIdentifier *types.PartialBlockIdentifier) (*types.BlockResponse, error) {
	request := types.BlockRequest{
		NetworkIdentifier: networkIdentifier,
		BlockIdentifier:   blockIdentifier,
	}

	resp, err := client.Post(meshcommon.BlockEndpoint, request)
	if err != nil {
		return nil, fmt.Errorf("failed to make block request: %v", err)
	}

	var response types.BlockResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to parse block response: %v", err)
	}

	return &response, nil
}

// testBlockTransaction tests the /block/transaction endpoint
func testBlockTransaction(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, blockIdentifier *types.BlockIdentifier, transactionIdentifier *types.TransactionIdentifier) (*types.BlockTransactionResponse, error) {
	request := types.BlockTransactionRequest{
		NetworkIdentifier:     networkIdentifier,
		BlockIdentifier:       blockIdentifier,
		TransactionIdentifier: transactionIdentifier,
	}

	resp, err := client.Post(meshcommon.BlockTransactionEndpoint, request)
	if err != nil {
		return nil, fmt.Errorf("failed to make block transaction request: %v", err)
	}

	var response types.BlockTransactionResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to parse block transaction response: %v", err)
	}

	return &response, nil
}

// testEventsBlocks tests the /events/blocks endpoint
func testEventsBlocks(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, offset *int64, limit *int64) (*types.EventsBlocksResponse, error) {
	request := types.EventsBlocksRequest{
		NetworkIdentifier: networkIdentifier,
		Offset:            offset,
		Limit:             limit,
	}

	resp, err := client.Post(meshcommon.EventsBlocksEndpoint, request)
	if err != nil {
		return nil, fmt.Errorf("failed to make events blocks request: %v", err)
	}

	var response types.EventsBlocksResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to parse events blocks response: %v", err)
	}

	return &response, nil
}

// testSearchTransactions tests the /search/transactions endpoint
func testSearchTransactions(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, transactionIdentifier *types.TransactionIdentifier) (*types.SearchTransactionsResponse, error) {
	request := types.SearchTransactionsRequest{
		NetworkIdentifier:     networkIdentifier,
		TransactionIdentifier: transactionIdentifier,
	}

	resp, err := client.Post(meshcommon.SearchTransactionsEndpoint, request)
	if err != nil {
		return nil, fmt.Errorf("failed to make search transactions request: %v", err)
	}

	var response types.SearchTransactionsResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to parse search transactions response: %v", err)
	}

	return &response, nil
}

// testSearchTransactionsWithRetry tests the /search/transactions endpoint with retry logic
func testSearchTransactionsWithRetry(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, transactionIdentifier *types.TransactionIdentifier, maxRetries int, delaySeconds int) (*types.SearchTransactionsResponse, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		response, err := testSearchTransactions(client, networkIdentifier, transactionIdentifier)
		if err == nil {
			return response, nil
		}

		lastErr = err

		if attempt < maxRetries {
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		}
	}

	return nil, fmt.Errorf("search transactions failed after %d attempts, last error: %v", maxRetries, lastErr)
}

// testConstructionParse tests the /construction/parse endpoint
func testConstructionParse(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, transaction []byte, signed bool) (*types.ConstructionParseResponse, error) {
	request := &types.ConstructionParseRequest{
		NetworkIdentifier: networkIdentifier,
		Signed:            signed,
		Transaction:       string(transaction),
	}

	resp, err := client.Post(meshcommon.ConstructionParseEndpoint, request)
	if err != nil {
		return nil, err
	}

	var response types.ConstructionParseResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, err
	}

	// Validate response
	if err := ValidateConstructionParseResponse(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// testCall tests the /call endpoint
func testCall(client *HTTPClient, networkIdentifier *types.NetworkIdentifier, method string, parameters map[string]any) (*types.CallResponse, error) {
	request := &types.CallRequest{
		NetworkIdentifier: networkIdentifier,
		Method:            method,
		Parameters:        parameters,
	}

	resp, err := client.Post(meshcommon.CallEndpoint, request)
	if err != nil {
		return nil, err
	}

	var response types.CallResponse
	if err := ParseResponse(resp, &response); err != nil {
		return nil, err
	}

	// Basic validation
	if response.Result == nil {
		return nil, fmt.Errorf("call response result is nil")
	}

	return &response, nil
}
