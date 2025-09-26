package e2e

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshutils "github.com/vechain/mesh/utils"
)

// Transaction types
const (
	TransactionTypeLegacy  = "legacy"
	TransactionTypeDynamic = "dynamic"
)

// Expected metadata fields based on the actual implementation
type ExpectedMetadata struct {
	TransactionType string
	BlockRef        string
	ChainTag        int64
	Gas             int64
	Nonce           string
	// Legacy specific fields
	GasPriceCoef *int64
	// Dynamic specific fields
	MaxFeePerGas         *string
	MaxPriorityFeePerGas *string
}

// HTTPClient wraps http.Client with test configuration
type HTTPClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPClient creates a new HTTP client for e2e tests
func NewHTTPClient(baseURL string, timeoutSeconds int) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
		baseURL: baseURL,
	}
}

// Post makes an HTTP POST request to the Mesh API
func (c *HTTPClient) Post(endpoint string, request any) (*http.Response, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	url := c.baseURL + endpoint
	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make request to %s: %v", url, err)
	}

	return resp, nil
}

// ParseResponse parses the HTTP response body into the target struct
func ParseResponse(resp *http.Response, target any) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return nil
}

// CreateTestNetworkIdentifier creates a network identifier for testing
func CreateTestNetworkIdentifier(network string) *types.NetworkIdentifier {
	return &types.NetworkIdentifier{
		Blockchain: "vechainthor",
		Network:    network,
	}
}

// CreateVETCurrency creates a VET currency definition
func CreateVETCurrency() *types.Currency {
	return meshutils.VETCurrency
}

// CreateVTHOCurrency creates a VTHO currency definition
func CreateVTHOCurrency() *types.Currency {
	return meshutils.VTHOCurrency
}

// CreateTransferOperations creates transfer operations for testing
func CreateTransferOperations(senderAddress, recipientAddress, amount string) []*types.Operation {
	return []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 0,
			},
			Type:   meshutils.OperationTypeTransfer,
			Status: meshutils.StringPtr(meshutils.OperationStatusNone),
			Account: &types.AccountIdentifier{
				Address: recipientAddress,
			},
			Amount: &types.Amount{
				Value:    amount,
				Currency: CreateVETCurrency(),
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 1,
			},
			Type:   meshutils.OperationTypeTransfer,
			Status: meshutils.StringPtr(meshutils.OperationStatusNone),
			Account: &types.AccountIdentifier{
				Address: senderAddress,
			},
			Amount: &types.Amount{
				Value:    "-" + amount,
				Currency: CreateVETCurrency(),
			},
		},
	}
}

// CreateTestPublicKey creates a test public key
func CreateTestPublicKey() *types.PublicKey {
	// Use the correct public key that corresponds to the private key
	hexStr := "03e32e5960781ce0b43d8c2952eeea4b95e286b1bb5f8c1f0c9f09983ba7141d2f"
	bytes, _ := hex.DecodeString(hexStr)

	return &types.PublicKey{
		Bytes:     bytes,
		CurveType: "secp256k1",
	}
}

// ValidateNetworkListResponse validates a network list response
func ValidateNetworkListResponse(response *types.NetworkListResponse) error {
	if len(response.NetworkIdentifiers) == 0 {
		return fmt.Errorf("no networks returned")
	}

	network := response.NetworkIdentifiers[0]
	if network.Blockchain != "vechainthor" {
		return fmt.Errorf("unexpected blockchain: %s", network.Blockchain)
	}

	return nil
}

// ValidateNetworkOptionsResponse validates a network options response
func ValidateNetworkOptionsResponse(response *types.NetworkOptionsResponse) error {
	if response.Version == nil {
		return fmt.Errorf("version not returned")
	}

	if response.Allow == nil {
		return fmt.Errorf("allow not returned")
	}

	return nil
}

// ValidateNetworkStatusResponse validates a network status response
func ValidateNetworkStatusResponse(response *types.NetworkStatusResponse) error {
	if response.CurrentBlockIdentifier == nil {
		return fmt.Errorf("current block identifier not returned")
	}

	return nil
}

// ValidateConstructionPreprocessResponse validates a construction preprocess response
func ValidateConstructionPreprocessResponse(response *types.ConstructionPreprocessResponse) error {
	if response.Options == nil {
		return fmt.Errorf("options not returned")
	}

	// Validate clauses in options
	clauses, ok := response.Options["clauses"].([]any)
	if !ok {
		return fmt.Errorf("clauses not found in options")
	}

	if len(clauses) == 0 {
		return fmt.Errorf("no clauses found in options")
	}

	// Validate each clause structure
	for i, clause := range clauses {
		clauseMap, ok := clause.(map[string]any)
		if !ok {
			return fmt.Errorf("clause %d is not a valid object", i)
		}

		// Validate required clause fields
		if _, ok := clauseMap["to"].(string); !ok {
			return fmt.Errorf("clause %d missing 'to' field", i)
		}

		if _, ok := clauseMap["value"].(string); !ok {
			return fmt.Errorf("clause %d missing 'value' field", i)
		}

		if _, ok := clauseMap["data"].(string); !ok {
			return fmt.Errorf("clause %d missing 'data' field", i)
		}
	}

	// Validate required public keys
	if len(response.RequiredPublicKeys) == 0 {
		return fmt.Errorf("no required public keys returned")
	}

	for i, pubKey := range response.RequiredPublicKeys {
		if pubKey.Address == "" {
			return fmt.Errorf("required public key %d missing address", i)
		}
	}

	return nil
}

// ValidateConstructionMetadataResponse validates a construction metadata response
func ValidateConstructionMetadataResponse(response *types.ConstructionMetadataResponse) error {
	if response.Metadata == nil {
		return fmt.Errorf("metadata not returned")
	}

	// Validate suggested fee
	if len(response.SuggestedFee) == 0 {
		return fmt.Errorf("suggested_fee not returned")
	}

	fee := response.SuggestedFee[0]
	if fee.Currency == nil {
		return fmt.Errorf("suggested_fee currency not returned")
	}

	if fee.Currency.Symbol != meshutils.VTHOCurrency.Symbol {
		return fmt.Errorf("expected VTHO currency, got %s", fee.Currency.Symbol)
	}

	if fee.Currency.Decimals != meshutils.VTHOCurrency.Decimals {
		return fmt.Errorf("expected %d decimals, got %d", meshutils.VTHOCurrency.Decimals, fee.Currency.Decimals)
	}

	// Validate VTHO contract address
	if fee.Currency.Metadata == nil {
		return fmt.Errorf("VTHO currency metadata not returned")
	}

	contractAddr, ok := fee.Currency.Metadata["contractAddress"].(string)
	expectedContractAddr := meshutils.VTHOCurrency.Metadata["contractAddress"].(string)
	if !ok || contractAddr != expectedContractAddr {
		return fmt.Errorf("invalid VTHO contract address: %v", fee.Currency.Metadata["contractAddress"])
	}

	// Validate fee value is negative
	if !strings.HasPrefix(fee.Value, "-") {
		return fmt.Errorf("expected negative fee value, got %s", fee.Value)
	}

	// Validate metadata fields based on transaction type
	return ValidateMetadataFields(response.Metadata)
}

// ValidateMetadataFields validates the metadata fields based on transaction type
func ValidateMetadataFields(metadata map[string]any) error {
	// Check transaction type
	transactionType, ok := metadata["transactionType"].(string)
	if !ok {
		return fmt.Errorf("transactionType not found in metadata")
	}

	if transactionType != TransactionTypeLegacy && transactionType != TransactionTypeDynamic {
		return fmt.Errorf("invalid transaction type: %s", transactionType)
	}

	// Validate common fields
	if _, ok := metadata["blockRef"].(string); !ok {
		return fmt.Errorf("blockRef not found or invalid in metadata")
	}

	if chainTag, ok := metadata["chainTag"].(float64); !ok || chainTag <= 0 {
		return fmt.Errorf("chainTag not found or invalid in metadata")
	}

	if gas, ok := metadata["gas"].(float64); !ok || gas <= 0 {
		return fmt.Errorf("gas not found or invalid in metadata")
	}

	if _, ok := metadata["nonce"].(string); !ok {
		return fmt.Errorf("nonce not found or invalid in metadata")
	}

	// Validate type-specific fields
	switch transactionType {
	case TransactionTypeLegacy:
		if gasPriceCoef, ok := metadata["gasPriceCoef"].(float64); !ok || gasPriceCoef < 0 || gasPriceCoef > 255 {
			return fmt.Errorf("gasPriceCoef not found or invalid for legacy transaction")
		}
	case TransactionTypeDynamic:
		if _, ok := metadata["maxFeePerGas"].(string); !ok {
			return fmt.Errorf("maxFeePerGas not found for dynamic transaction")
		}
		if _, ok := metadata["maxPriorityFeePerGas"].(string); !ok {
			return fmt.Errorf("maxPriorityFeePerGas not found for dynamic transaction")
		}
	}

	return nil
}

// ValidateConstructionPayloadsResponse validates a construction payloads response
func ValidateConstructionPayloadsResponse(response *types.ConstructionPayloadsResponse) error {
	if response.UnsignedTransaction == "" {
		return fmt.Errorf("unsigned transaction not returned")
	}

	// Validate unsigned transaction format (should be hex with 0x prefix)
	if !strings.HasPrefix(response.UnsignedTransaction, "0x") {
		return fmt.Errorf("unsigned transaction should start with 0x")
	}

	if len(response.Payloads) == 0 {
		return fmt.Errorf("no payloads returned")
	}

	// Validate each payload
	for i, payload := range response.Payloads {
		if payload.AccountIdentifier == nil {
			return fmt.Errorf("payload %d missing account identifier", i)
		}

		if payload.AccountIdentifier.Address == "" {
			return fmt.Errorf("payload %d missing address", i)
		}

		if len(payload.Bytes) == 0 {
			return fmt.Errorf("payload %d missing bytes", i)
		}

		if payload.SignatureType == "" {
			return fmt.Errorf("payload %d missing signature type", i)
		}

		if payload.SignatureType != "ecdsa_recovery" {
			return fmt.Errorf("payload %d invalid signature type: %s", i, payload.SignatureType)
		}
	}

	return nil
}

// ValidateConstructionCombineResponse validates a construction combine response
func ValidateConstructionCombineResponse(response *types.ConstructionCombineResponse) error {
	if response.SignedTransaction == "" {
		return fmt.Errorf("signed transaction not returned")
	}

	// Validate signed transaction format (should be hex with 0x prefix)
	if !strings.HasPrefix(response.SignedTransaction, "0x") {
		return fmt.Errorf("signed transaction should start with 0x")
	}

	// Validate that signed transaction is different from unsigned (should be longer)
	if len(response.SignedTransaction) < 10 {
		return fmt.Errorf("signed transaction seems too short")
	}

	return nil
}

// ValidateTransactionIdentifierResponse validates a transaction identifier response
func ValidateTransactionIdentifierResponse(response *types.TransactionIdentifierResponse) error {
	if response.TransactionIdentifier == nil {
		return fmt.Errorf("transaction identifier not returned")
	}

	if response.TransactionIdentifier.Hash == "" {
		return fmt.Errorf("transaction hash not returned")
	}

	// Validate transaction hash format (should be hex with 0x prefix)
	if !strings.HasPrefix(response.TransactionIdentifier.Hash, "0x") {
		return fmt.Errorf("transaction hash should start with 0x")
	}

	// Validate hash length (should be 66 characters for 32-byte hash)
	if len(response.TransactionIdentifier.Hash) != 66 {
		return fmt.Errorf("transaction hash should be 66 characters long, got %d", len(response.TransactionIdentifier.Hash))
	}

	return nil
}

// ValidateMempoolTransactionResponse validates a mempool transaction response
func ValidateMempoolTransactionResponse(response *types.MempoolTransactionResponse) error {
	if response.Transaction == nil {
		return fmt.Errorf("transaction not returned")
	}

	if response.Transaction.TransactionIdentifier == nil {
		return fmt.Errorf("transaction identifier not returned")
	}

	if response.Transaction.TransactionIdentifier.Hash == "" {
		return fmt.Errorf("transaction hash not returned")
	}

	// Validate transaction hash format
	if !strings.HasPrefix(response.Transaction.TransactionIdentifier.Hash, "0x") {
		return fmt.Errorf("transaction hash should start with 0x")
	}

	if len(response.Transaction.TransactionIdentifier.Hash) != 66 {
		return fmt.Errorf("transaction hash should be 66 characters long, got %d", len(response.Transaction.TransactionIdentifier.Hash))
	}

	// Validate operations exist
	if len(response.Transaction.Operations) == 0 {
		return fmt.Errorf("no operations in transaction")
	}

	// Validate each operation
	for i, op := range response.Transaction.Operations {
		if op.OperationIdentifier == nil {
			return fmt.Errorf("operation %d missing operation identifier", i)
		}

		if op.Type == "" {
			return fmt.Errorf("operation %d missing type", i)
		}

		if op.Account == nil {
			return fmt.Errorf("operation %d missing account", i)
		}

		if op.Account.Address == "" {
			return fmt.Errorf("operation %d missing account address", i)
		}

		if op.Amount == nil {
			return fmt.Errorf("operation %d missing amount", i)
		}

		if op.Amount.Value == "" {
			return fmt.Errorf("operation %d missing amount value", i)
		}

		if op.Amount.Currency == nil {
			return fmt.Errorf("operation %d missing currency", i)
		}
	}

	return nil
}

// CreateLegacyTransactionOperations creates operations for legacy transaction testing
func CreateLegacyTransactionOperations(senderAddress, recipientAddress, amount string) []*types.Operation {
	operations := CreateTransferOperations(senderAddress, recipientAddress, amount)

	// Add transaction type to operations metadata for legacy
	for _, op := range operations {
		if op.Metadata == nil {
			op.Metadata = make(map[string]any)
		}
		op.Metadata["transactionType"] = TransactionTypeLegacy
	}

	return operations
}

// CreateDynamicTransactionOperations creates operations for dynamic transaction testing
func CreateDynamicTransactionOperations(senderAddress, recipientAddress, amount string) []*types.Operation {
	operations := CreateTransferOperations(senderAddress, recipientAddress, amount)

	// Add transaction type to operations metadata for dynamic
	for _, op := range operations {
		if op.Metadata == nil {
			op.Metadata = make(map[string]any)
		}
		op.Metadata["transactionType"] = TransactionTypeDynamic
	}

	return operations
}

// CreatePreprocessRequestWithTransactionType creates a preprocess request with specific transaction type
func CreatePreprocessRequestWithTransactionType(networkIdentifier *types.NetworkIdentifier, operations []*types.Operation, transactionType string) *types.ConstructionPreprocessRequest {
	request := &types.ConstructionPreprocessRequest{
		NetworkIdentifier: networkIdentifier,
		Operations:        operations,
	}

	return request
}

// ValidateTransactionTypeInMetadata validates that the metadata contains the expected transaction type
func ValidateTransactionTypeInMetadata(metadata map[string]any, expectedType string) error {
	actualType, ok := metadata["transactionType"].(string)
	if !ok {
		return fmt.Errorf("transactionType not found in metadata")
	}

	if actualType != expectedType {
		return fmt.Errorf("expected transaction type %s, got %s", expectedType, actualType)
	}

	return nil
}

// ValidateLegacyMetadataFields validates legacy-specific metadata fields
func ValidateLegacyMetadataFields(metadata map[string]any) error {
	// Validate gasPriceCoef field exists and is valid
	gasPriceCoef, ok := metadata["gasPriceCoef"].(float64)
	if !ok {
		return fmt.Errorf("gasPriceCoef not found in legacy metadata")
	}

	if gasPriceCoef < 0 || gasPriceCoef > 255 {
		return fmt.Errorf("gasPriceCoef out of range (0-255): %f", gasPriceCoef)
	}

	// Ensure dynamic fields are not present
	if _, exists := metadata["maxFeePerGas"]; exists {
		return fmt.Errorf("maxFeePerGas should not be present in legacy metadata")
	}

	if _, exists := metadata["maxPriorityFeePerGas"]; exists {
		return fmt.Errorf("maxPriorityFeePerGas should not be present in legacy metadata")
	}

	return nil
}

// ValidateDynamicMetadataFields validates dynamic-specific metadata fields
func ValidateDynamicMetadataFields(metadata map[string]any) error {
	// Validate maxFeePerGas field exists and is valid
	maxFeePerGas, ok := metadata["maxFeePerGas"].(string)
	if !ok {
		return fmt.Errorf("maxFeePerGas not found in dynamic metadata")
	}

	if maxFeePerGas == "" {
		return fmt.Errorf("maxFeePerGas cannot be empty")
	}

	// Validate maxPriorityFeePerGas field exists and is valid
	maxPriorityFeePerGas, ok := metadata["maxPriorityFeePerGas"].(string)
	if !ok {
		return fmt.Errorf("maxPriorityFeePerGas not found in dynamic metadata")
	}

	if maxPriorityFeePerGas == "" {
		return fmt.Errorf("maxPriorityFeePerGas cannot be empty")
	}

	// Ensure legacy fields are not present
	if _, exists := metadata["gasPriceCoef"]; exists {
		return fmt.Errorf("gasPriceCoef should not be present in dynamic metadata")
	}

	return nil
}
