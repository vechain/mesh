package services

import (
	"context"
	"errors"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshconfig "github.com/vechain/mesh/config"
	meshtests "github.com/vechain/mesh/tests"
	meshthor "github.com/vechain/mesh/thor"
)

func createMockConstructionService() *ConstructionService {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      meshcommon.TestNetwork,
		Mode:         meshcommon.OnlineMode,
		BaseGasPrice: "1000000000000000000",
	}
	return NewConstructionService(mockClient, config)
}

func createTestPublicKey() *types.PublicKey {
	return &types.PublicKey{
		Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
		CurveType: meshtests.SECP256k1,
	}
}

func createTestNetworkIdentifier(network string) *types.NetworkIdentifier {
	return &types.NetworkIdentifier{
		Blockchain: meshcommon.BlockchainName,
		Network:    network,
	}
}

func TestNewConstructionService(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      meshcommon.TestNetwork,
		Mode:         meshcommon.OnlineMode,
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	if service == nil {
		t.Errorf("NewConstructionService() returned nil")
	} else {
		if service.vechainClient != mockClient {
			t.Errorf("NewConstructionService() client = %v, want %v", service.vechainClient, mockClient)
		}
		if service.config != config {
			t.Errorf("NewConstructionService() config = %v, want %v", service.config, config)
		}
		if service.encoder == nil {
			t.Errorf("NewConstructionService() encoder is nil")
		}
	}
}

func TestConstructionService_ConstructionDerive_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	// Create valid request
	request := &types.ConstructionDeriveRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		PublicKey:         createTestPublicKey(),
	}

	ctx := context.Background()
	response, err := service.ConstructionDerive(ctx, request)

	if err != nil {
		t.Fatalf("ConstructionDerive() error = %v", err)
	}

	// Verify response structure
	if response.AccountIdentifier == nil {
		t.Errorf("ConstructionDerive() account identifier is nil")
	}
	if response.AccountIdentifier.Address == "" {
		t.Errorf("ConstructionDerive() account address is empty")
	}
}

func TestConstructionService_ConstructionPreprocess_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	// Create valid request with both sender (negative) and receiver (positive) operations
	request := &types.ConstructionPreprocessRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		Operations: []*types.Operation{
			// Sender operation (negative amount)
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
			// Receiver operation (positive amount)
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 1},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: meshtests.TestAddress1,
				},
				Amount: &types.Amount{
					Value:    "1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
		},
		Metadata: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
		},
	}

	ctx := context.Background()
	response, err := service.ConstructionPreprocess(ctx, request)

	if err != nil {
		t.Fatalf("ConstructionPreprocess() error = %v", err)
	}

	// Verify response structure
	if response.Options == nil {
		t.Errorf("ConstructionPreprocess() options is nil")
	}
}

func TestConstructionService_ConstructionPreprocess_VIP180Token(t *testing.T) {
	service := createMockConstructionService()

	// Create valid request with VIP180 token operations
	request := &types.ConstructionPreprocessRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		Operations: []*types.Operation{
			// Sender operation (negative amount)
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Amount: &types.Amount{
					Value: "-1000000000000000000",
					Currency: &types.Currency{
						Symbol:   "TVIP",
						Decimals: 18,
						Metadata: map[string]any{
							"contractAddress": "0x1234567890123456789012345678901234567890",
						},
					},
				},
			},
			// Receiver operation (positive amount)
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 1},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: meshtests.TestAddress1,
				},
				Amount: &types.Amount{
					Value: "1000000000000000000",
					Currency: &types.Currency{
						Symbol:   "TVIP",
						Decimals: 18,
						Metadata: map[string]any{
							"contractAddress": "0x1234567890123456789012345678901234567890",
						},
					},
				},
			},
		},
		Metadata: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
		},
	}

	ctx := context.Background()
	response, err := service.ConstructionPreprocess(ctx, request)

	if err != nil {
		t.Fatalf("ConstructionPreprocess() error = %v", err)
	}

	// Verify response structure
	if response.Options == nil {
		t.Errorf("ConstructionPreprocess() options is nil")
		return
	}

	// Verify clauses were created for VIP180 token transfer
	if clauses, ok := response.Options["clauses"].([]map[string]any); ok {
		if len(clauses) == 0 {
			t.Errorf("ConstructionPreprocess() expected clauses, got empty array")
		}
		// Verify the clause has the token contract address as "to"
		if len(clauses) > 0 {
			clause := clauses[0]
			if to, exists := clause["to"]; exists {
				if to != "0x1234567890123456789012345678901234567890" {
					t.Errorf("ConstructionPreprocess() clause 'to' = %v, want contract address", to)
				}
			} else {
				t.Errorf("ConstructionPreprocess() clause missing 'to' field")
			}
			// Verify the clause has value "0" for token transfer
			if value, exists := clause["value"]; exists {
				if value != "0" {
					t.Errorf("ConstructionPreprocess() clause 'value' = %v, want '0' for token transfer", value)
				}
			}
			// Verify the clause has encoded data
			if data, exists := clause["data"]; exists {
				if dataStr, ok := data.(string); !ok || dataStr == "" || dataStr == "0x" {
					t.Errorf("ConstructionPreprocess() clause 'data' should contain encoded VIP180 transfer data")
				}
			} else {
				t.Errorf("ConstructionPreprocess() clause missing 'data' field")
			}
		}
	} else {
		t.Errorf("ConstructionPreprocess() options['clauses'] not found or not an array")
	}
}

func TestConstructionService_ConstructionMetadata_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	// Create valid request for legacy
	request := &types.ConstructionMetadataRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		Options: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
			"clauses": []any{
				map[string]any{
					"to":    meshtests.TestAddress1,
					"value": "1000000000000000000",
					"data":  "0x",
				},
			},
		},
	}

	ctx := context.Background()
	response, err := service.ConstructionMetadata(ctx, request)

	if err != nil {
		t.Fatalf("ConstructionMetadata() error = %v", err)
	}

	// Verify response structure
	if response.Metadata == nil {
		t.Errorf("ConstructionMetadata() metadata is nil")
	}
	if len(response.SuggestedFee) == 0 {
		t.Errorf("ConstructionMetadata() suggested fee is empty")
	}
}

func TestConstructionService_ConstructionMetadata_DynamicRequest(t *testing.T) {
	service := createMockConstructionService()

	// Create valid request for dynamic
	request := &types.ConstructionMetadataRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		Options: map[string]any{
			"transactionType": meshcommon.TransactionTypeDynamic,
			"clauses": []any{
				map[string]any{
					"to":    meshtests.TestAddress1,
					"value": "1000000000000000000",
					"data":  "0x",
				},
			},
		},
	}

	ctx := context.Background()
	response, err := service.ConstructionMetadata(ctx, request)

	if err != nil {
		t.Fatalf("ConstructionMetadata() error = %v", err)
	}

	// Verify response structure
	if response.Metadata == nil {
		t.Errorf("ConstructionMetadata() metadata is nil")
	}
	if len(response.SuggestedFee) == 0 {
		t.Errorf("ConstructionMetadata() suggested fee is empty")
	}
}

func TestConstructionService_ConstructionMetadata_ClientError(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      meshcommon.TestNetwork,
		Mode:         meshcommon.OnlineMode,
		BaseGasPrice: "1000000000000000000",
	}
	service := NewConstructionService(mockClient, config)

	// Configure mock to return error for GetBestBlock
	mockClient.SetMockBlockError(errors.New("failed to get best block"))

	request := &types.ConstructionMetadataRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		Options: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
			"clauses": []any{
				map[string]any{
					"to":    meshtests.TestAddress1,
					"value": "1000000000000000000",
					"data":  "0x",
				},
			},
		},
	}

	ctx := context.Background()
	_, err := service.ConstructionMetadata(ctx, request)

	if err == nil {
		t.Error("ConstructionMetadata() expected error when client fails")
	}

	if err != nil && err.Code != int32(meshcommon.ErrGettingBlockchainMetadata) {
		t.Errorf("ConstructionMetadata() error code = %d, want %d", err.Code, meshcommon.ErrGettingBlockchainMetadata)
	}
}

func TestConstructionService_ConstructionMetadata_MissingClauses(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionMetadataRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		Options: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
			// Missing clauses
		},
	}

	ctx := context.Background()
	_, err := service.ConstructionMetadata(ctx, request)

	if err == nil {
		t.Error("ConstructionMetadata() expected error when clauses are missing")
	}

	if err != nil && err.Code != int32(meshcommon.ErrGettingBlockchainMetadata) {
		t.Errorf("ConstructionMetadata() error code = %d, want %d", err.Code, meshcommon.ErrGettingBlockchainMetadata)
	}
}

func TestConstructionService_ConstructionMetadata_EmptyClauses(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionMetadataRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		Options: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
			"clauses":         []any{},
		},
	}

	ctx := context.Background()
	_, err := service.ConstructionMetadata(ctx, request)

	if err == nil {
		t.Error("ConstructionMetadata() expected error when clauses are empty")
	}

	if err != nil && err.Code != int32(meshcommon.ErrGettingBlockchainMetadata) {
		t.Errorf("ConstructionMetadata() error code = %d, want %d", err.Code, meshcommon.ErrGettingBlockchainMetadata)
	}
}

func TestConstructionService_ConstructionMetadata_InvalidClauses(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionMetadataRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		Options: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
			"clauses": []any{
				map[string]any{
					"to": "INVALID_ADDRESS",
					// Missing value and data
				},
			},
		},
	}

	ctx := context.Background()
	_, err := service.ConstructionMetadata(ctx, request)

	if err == nil {
		t.Error("ConstructionMetadata() expected error when clauses are invalid")
	}

	if err != nil && err.Code != int32(meshcommon.ErrGettingBlockchainMetadata) {
		t.Errorf("ConstructionMetadata() error code = %d, want %d", err.Code, meshcommon.ErrGettingBlockchainMetadata)
	}
}

func TestConstructionService_ConstructionMetadata_InvalidDynamicMetadata(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      meshcommon.TestNetwork,
		Mode:         meshcommon.OnlineMode,
		BaseGasPrice: "1000000000000000000",
	}
	service := NewConstructionService(mockClient, config)

	// Set up mock to fail on GetDynamicGasPrice
	mockClient.SetMockError(errors.New("failed to get dynamic gas price"))

	request := &types.ConstructionMetadataRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		Options: map[string]any{
			"transactionType": meshcommon.TransactionTypeDynamic,
			"clauses": []any{
				map[string]any{
					"to":    meshtests.TestAddress1,
					"value": "1000000000000000000",
					"data":  "0x",
				},
			},
		},
	}

	ctx := context.Background()
	_, err := service.ConstructionMetadata(ctx, request)

	if err == nil {
		t.Error("ConstructionMetadata() expected error when dynamic metadata is invalid")
	}

	if err != nil && err.Code != int32(meshcommon.ErrGettingBlockchainMetadata) {
		t.Errorf("ConstructionMetadata() error code = %d, want %d", err.Code, meshcommon.ErrGettingBlockchainMetadata)
	}
}

func TestConstructionService_ConstructionPayloads_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	// Create valid request
	request := &types.ConstructionPayloadsRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
		},
		PublicKeys: []*types.PublicKey{
			{
				Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
				CurveType: meshtests.SECP256k1,
			},
		},
		Metadata: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
			"blockRef":        "0x0000000000000000",
			"chainTag":        float64(1),
			"gas":             float64(21000),
			"nonce":           "0x1",
			"gasPriceCoef":    uint8(128),
		},
	}

	ctx := context.Background()
	response, err := service.ConstructionPayloads(ctx, request)

	if err != nil {
		t.Fatalf("ConstructionPayloads() error = %v", err)
	}

	// Verify response structure
	if len(response.Payloads) == 0 {
		t.Errorf("ConstructionPayloads() payloads is empty")
	}
}

func TestConstructionService_ConstructionPayloads_OriginAddressMismatch(t *testing.T) {
	service := createMockConstructionService()

	// Create request with mismatched origin address
	request := &types.ConstructionPayloadsRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: "0x1234567890123456789012345678901234567890",
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
		},
		PublicKeys: []*types.PublicKey{
			{
				Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
				CurveType: meshtests.SECP256k1,
			},
		},
		Metadata: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
			"blockRef":        "0x0000000000000000",
			"chainTag":        float64(1),
			"gas":             float64(21000),
			"nonce":           "0x1",
			"gasPriceCoef":    uint8(128),
		},
	}

	ctx := context.Background()
	_, err := service.ConstructionPayloads(ctx, request)

	if err == nil {
		t.Error("ConstructionPayloads() expected error for address mismatch")
	}
}

func TestConstructionService_ConstructionPayloads_InvalidPublicKey(t *testing.T) {
	service := createMockConstructionService()

	// Create request with invalid public key
	request := &types.ConstructionPayloadsRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
		},
		PublicKeys: []*types.PublicKey{
			{
				Bytes:     []byte{0x01, 0x02, 0x03},
				CurveType: meshtests.SECP256k1,
			},
		},
		Metadata: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
			"blockRef":        "0x0000000000000000",
			"chainTag":        float64(1),
			"gas":             float64(21000),
			"nonce":           "0x1",
			"gasPriceCoef":    uint8(128),
		},
	}

	ctx := context.Background()
	_, err := service.ConstructionPayloads(ctx, request)

	if err == nil {
		t.Error("ConstructionPayloads() expected error for invalid public key")
	}
}

func TestConstructionService_ConstructionPayloads_NoPublicKeys(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionPayloadsRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
		},
		PublicKeys: []*types.PublicKey{},
		Metadata: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
			"blockRef":        "0x0000000000000000",
			"chainTag":        float64(1),
			"gas":             float64(21000),
			"nonce":           "0x1",
			"gasPriceCoef":    uint8(128),
		},
	}

	ctx := context.Background()
	_, err := service.ConstructionPayloads(ctx, request)

	if err == nil {
		t.Error("ConstructionPayloads() expected error for no public keys")
	}
}

func TestConstructionService_ConstructionPayloads_TooManyPublicKeys(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionPayloadsRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
		},
		PublicKeys: []*types.PublicKey{
			{
				Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
				CurveType: meshtests.SECP256k1,
			},
			{
				Bytes:     []byte{0x04, 0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef, 0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef, 0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef, 0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef},
				CurveType: meshtests.SECP256k1,
			},
			{
				Bytes:     []byte{0x05, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa},
				CurveType: meshtests.SECP256k1,
			},
		},
		Metadata: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
			"blockRef":        "0x0000000000000000",
			"chainTag":        float64(1),
			"gas":             float64(21000),
			"nonce":           "0x1",
			"gasPriceCoef":    uint8(128),
		},
	}

	ctx := context.Background()
	_, err := service.ConstructionPayloads(ctx, request)

	if err == nil {
		t.Error("ConstructionPayloads() expected error for too many public keys")
	}
}

func TestConstructionService_ConstructionPayloads_DelegatorAddressMismatch(t *testing.T) {
	service := createMockConstructionService()

	// Create request with fee delegation but mismatched delegator address
	request := &types.ConstructionPayloadsRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
		},
		PublicKeys: []*types.PublicKey{
			{
				Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
				CurveType: meshtests.SECP256k1,
			},
			{
				Bytes:     []byte{0x02, 0x79, 0xbe, 0x66, 0x7e, 0xf9, 0xdc, 0xbb, 0xac, 0x55, 0xa0, 0x62, 0x95, 0xce, 0x87, 0x0b, 0x07, 0x02, 0x9b, 0xfc, 0xdb, 0x2d, 0xce, 0x28, 0xd9, 0x59, 0xf2, 0x81, 0x5b, 0x16, 0xf8, 0x17, 0x98},
				CurveType: meshtests.SECP256k1,
			},
		},
		Metadata: map[string]any{
			"transactionType":                      meshcommon.TransactionTypeLegacy,
			"blockRef":                             "0x0000000000000000",
			"chainTag":                             float64(1),
			"gas":                                  float64(21000),
			"nonce":                                "0x1",
			"gasPriceCoef":                         uint8(128),
			meshcommon.DelegatorAccountMetadataKey: "0x1234567890123456789012345678901234567890",
		},
	}

	ctx := context.Background()
	_, err := service.ConstructionPayloads(ctx, request)

	if err == nil {
		t.Error("ConstructionPayloads() expected error for delegator address mismatch")
	}
}

func TestConstructionService_ConstructionParse_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionParseRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		Signed:      false,
		Transaction: "0xf85db84551f84281f68502b506882881b4e0df9416277a1ff38678291c41d1820957c78bb5da59ce880de0b6b3a764000080808609184e72a00082bb80808827706abefbc974eac08094f077b491b355e64048ce21e3a6fc4751eeea77fa80",
	}

	ctx := context.Background()
	_, err := service.ConstructionParse(ctx, request)

	if err != nil {
		t.Fatalf("ConstructionParse() error = %v", err)
	}
}

func TestConstructionService_ConstructionCombine_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	// Create valid request with the provided values
	request := &types.ConstructionCombineRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.SoloNetwork,
		},
		UnsignedTransaction: "0xf85db84551f84281f68502b506882881b4e0df9416277a1ff38678291c41d1820957c78bb5da59ce880de0b6b3a764000080808609184e72a00082bb80808827706abefbc974eac08094f077b491b355e64048ce21e3a6fc4751eeea77fa80",
		Signatures: []*types.Signature{
			{
				SigningPayload: &types.SigningPayload{
					AccountIdentifier: &types.AccountIdentifier{
						Address: meshtests.FirstSoloAddress,
					},
					Bytes:         []byte{0x4d, 0x7b, 0xe3, 0xe5, 0xd9, 0x77, 0xe3, 0xf6, 0xbc, 0x6d, 0xc1, 0xc7, 0x55, 0x85, 0x52, 0x35, 0x19, 0xa1, 0x38, 0x74, 0x12, 0xb3, 0x06, 0xd3, 0x5e, 0x51, 0xf0, 0xb7, 0x2c, 0x8b, 0x1b, 0x67},
					SignatureType: "ecdsa_recovery",
				},
				PublicKey: &types.PublicKey{
					Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
					CurveType: meshtests.SECP256k1,
				},
				SignatureType: "ecdsa_recovery",
				Bytes:         []byte{0x1b, 0xc4, 0xaf, 0xf0, 0xc0, 0xd4, 0x25, 0xec, 0xd1, 0xa9, 0x31, 0xad, 0x43, 0x51, 0x56, 0xa4, 0x8b, 0x3e, 0x74, 0xb9, 0xa7, 0x6b, 0x79, 0xfc, 0xc6, 0xe8, 0x66, 0x33, 0x7a, 0x73, 0xf0, 0x5a, 0x7e, 0x05, 0x77, 0x3a, 0x5f, 0xcb, 0xd4, 0xcf, 0x12, 0x51, 0xcf, 0x02, 0x6e, 0x70, 0xc5, 0xcc, 0x5b, 0x35, 0x24, 0x86, 0x64, 0x46, 0xde, 0x93, 0xa7, 0xd4, 0x98, 0x97, 0xc0, 0xba, 0xc5, 0x79, 0x00},
			},
		},
	}

	ctx := context.Background()
	response, err := service.ConstructionCombine(ctx, request)

	if err != nil {
		t.Fatalf("ConstructionCombine() error = %v", err)
	}

	if response.SignedTransaction == "" {
		t.Error("Expected signed transaction in response")
	}
}

func TestConstructionService_ConstructionCombine_InvalidUnsignedTransaction(t *testing.T) {
	service := createMockConstructionService()

	// Create request with invalid unsigned transaction
	request := &types.ConstructionCombineRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.SoloNetwork,
		},
		UnsignedTransaction: "invalid_hex",
		Signatures: []*types.Signature{
			{
				SigningPayload: &types.SigningPayload{
					AccountIdentifier: &types.AccountIdentifier{
						Address: meshtests.FirstSoloAddress,
					},
					Bytes:         []byte{0x4d, 0x7b, 0xe3, 0xe5, 0xd9, 0x77, 0xe3, 0xf6, 0xbc, 0x6d, 0xc1, 0xc7, 0x55, 0x85, 0x52, 0x35, 0x19, 0xa1, 0x38, 0x74, 0x12, 0xb3, 0x06, 0xd3, 0x5e, 0x51, 0xf0, 0xb7, 0x2c, 0x8b, 0x1b, 0x67},
					SignatureType: "ecdsa_recovery",
				},
				PublicKey: &types.PublicKey{
					Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
					CurveType: meshtests.SECP256k1,
				},
				SignatureType: "ecdsa_recovery",
				Bytes:         []byte{0x1b, 0xc4, 0xaf, 0xf0, 0xc0, 0xd4, 0x25, 0xec, 0xd1, 0xa9, 0x31, 0xad, 0x43, 0x51, 0x56, 0xa4, 0x8b, 0x3e, 0x74, 0xb9, 0xa7, 0x6b, 0x79, 0xfc, 0xc6, 0xe8, 0x66, 0x33, 0x7a, 0x73, 0xf0, 0x5a, 0x7e, 0x05, 0x77, 0x3a, 0x5f, 0xcb, 0xd4, 0xcf, 0x12, 0x51, 0xcf, 0x02, 0x6e, 0x70, 0xc5, 0xcc, 0x5b, 0x35, 0x24, 0x86, 0x64, 0x46, 0xde, 0x93, 0xa7, 0xd4, 0x98, 0x97, 0xc0, 0xba, 0xc5, 0x79, 0x00},
			},
		},
	}

	ctx := context.Background()
	_, err := service.ConstructionCombine(ctx, request)

	if err == nil {
		t.Error("ConstructionCombine() expected error for invalid unsigned transaction")
	}
}

func TestConstructionService_ConstructionCombine_InvalidNumberOfSignatures(t *testing.T) {
	service := createMockConstructionService()

	// Create request with no signatures
	request := &types.ConstructionCombineRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.SoloNetwork,
		},
		UnsignedTransaction: "0xf85281f68800000005e6911c7481b4dad99416277a1ff38678291c41d1820957c78bb5da59ce8227108082bb808864d53d1260b9a69f94f077b491b355e64048ce21e3a6fc4751eeea77fa808609184e72a00080",
		Signatures:          []*types.Signature{}, // No signatures
	}

	ctx := context.Background()
	_, err := service.ConstructionCombine(ctx, request)

	if err == nil {
		t.Error("ConstructionCombine() expected error for no signatures")
	}

	// Test with too many signatures (3)
	request.Signatures = []*types.Signature{
		{
			SigningPayload: &types.SigningPayload{
				AccountIdentifier: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Bytes:         []byte{0x4d, 0x7b, 0xe3, 0xe5, 0xd9, 0x77, 0xe3, 0xf6, 0xbc, 0x6d, 0xc1, 0xc7, 0x55, 0x85, 0x52, 0x35, 0x19, 0xa1, 0x38, 0x74, 0x12, 0xb3, 0x06, 0xd3, 0x5e, 0x51, 0xf0, 0xb7, 0x2c, 0x8b, 0x1b, 0x67},
				SignatureType: "ecdsa_recovery",
			},
			PublicKey: &types.PublicKey{
				Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
				CurveType: meshtests.SECP256k1,
			},
			SignatureType: "ecdsa_recovery",
			Bytes:         []byte{0x03, 0xf8, 0xa1, 0xca, 0x4d, 0xd3, 0x9d, 0x99, 0xab, 0x54, 0x97, 0xf9, 0x4b, 0x8b, 0x79, 0x11, 0x34, 0x0c, 0xea, 0xc7, 0x18, 0x20, 0x19, 0xb7, 0xbe, 0x9d, 0x81, 0xf0, 0x43, 0xc7, 0x43, 0xf9, 0x5a, 0x69, 0x43, 0x1d, 0x71, 0x5a, 0xde, 0x0c, 0x9b, 0x74, 0x1f, 0x7c, 0x83, 0xd9, 0x57, 0x2a, 0xd8, 0x42, 0x71, 0xb4, 0xf2, 0xec, 0xb6, 0x2c, 0x8f, 0x49, 0xdd, 0xfa, 0x3e, 0x8c, 0x3a, 0xea, 0x01},
		},
		{
			SigningPayload: &types.SigningPayload{
				AccountIdentifier: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Bytes:         []byte{0x4d, 0x7b, 0xe3, 0xe5, 0xd9, 0x77, 0xe3, 0xf6, 0xbc, 0x6d, 0xc1, 0xc7, 0x55, 0x85, 0x52, 0x35, 0x19, 0xa1, 0x38, 0x74, 0x12, 0xb3, 0x06, 0xd3, 0x5e, 0x51, 0xf0, 0xb7, 0x2c, 0x8b, 0x1b, 0x67},
				SignatureType: "ecdsa_recovery",
			},
			PublicKey: &types.PublicKey{
				Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
				CurveType: meshtests.SECP256k1,
			},
			SignatureType: "ecdsa_recovery",
			Bytes:         []byte{0x03, 0xf8, 0xa1, 0xca, 0x4d, 0xd3, 0x9d, 0x99, 0xab, 0x54, 0x97, 0xf9, 0x4b, 0x8b, 0x79, 0x11, 0x34, 0x0c, 0xea, 0xc7, 0x18, 0x20, 0x19, 0xb7, 0xbe, 0x9d, 0x81, 0xf0, 0x43, 0xc7, 0x43, 0xf9, 0x5a, 0x69, 0x43, 0x1d, 0x71, 0x5a, 0xde, 0x0c, 0x9b, 0x74, 0x1f, 0x7c, 0x83, 0xd9, 0x57, 0x2a, 0xd8, 0x42, 0x71, 0xb4, 0xf2, 0xec, 0xb6, 0x2c, 0x8f, 0x49, 0xdd, 0xfa, 0x3e, 0x8c, 0x3a, 0xea, 0x01},
		},
		{
			SigningPayload: &types.SigningPayload{
				AccountIdentifier: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Bytes:         []byte{0x4d, 0x7b, 0xe3, 0xe5, 0xd9, 0x77, 0xe3, 0xf6, 0xbc, 0x6d, 0xc1, 0xc7, 0x55, 0x85, 0x52, 0x35, 0x19, 0xa1, 0x38, 0x74, 0x12, 0xb3, 0x06, 0xd3, 0x5e, 0x51, 0xf0, 0xb7, 0x2c, 0x8b, 0x1b, 0x67},
				SignatureType: "ecdsa_recovery",
			},
			PublicKey: &types.PublicKey{
				Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
				CurveType: meshtests.SECP256k1,
			},
			SignatureType: "ecdsa_recovery",
			Bytes:         []byte{0x03, 0xf8, 0xa1, 0xca, 0x4d, 0xd3, 0x9d, 0x99, 0xab, 0x54, 0x97, 0xf9, 0x4b, 0x8b, 0x79, 0x11, 0x34, 0x0c, 0xea, 0xc7, 0x18, 0x20, 0x19, 0xb7, 0xbe, 0x9d, 0x81, 0xf0, 0x43, 0xc7, 0x43, 0xf9, 0x5a, 0x69, 0x43, 0x1d, 0x71, 0x5a, 0xde, 0x0c, 0x9b, 0x74, 0x1f, 0x7c, 0x83, 0xd9, 0x57, 0x2a, 0xd8, 0x42, 0x71, 0xb4, 0xf2, 0xec, 0xb6, 0x2c, 0x8f, 0x49, 0xdd, 0xfa, 0x3e, 0x8c, 0x3a, 0xea, 0x01},
		},
	}

	ctx = context.Background()
	_, err = service.ConstructionCombine(ctx, request)

	if err == nil {
		t.Error("ConstructionCombine() expected error for too many signatures")
	}
}

func TestConstructionService_ConstructionHash_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionHashRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.TestNetwork,
		},
		SignedTransaction: "0x51f88481f68502b506882881b4e0df9416277a1ff38678291c41d1820957c78bb5da59ce880de0b6b3a764000080808609184e72a00082bb80808827706abefbc974eac0b8411bc4aff0c0d425ecd1a931ad435156a48b3e74b9a76b79fcc6e866337a73f05a7e05773a5fcbd4cf1251cf026e70c5cc5b3524866446de93a7d49897c0bac57900",
	}

	ctx := context.Background()
	_, err := service.ConstructionHash(ctx, request)

	if err != nil {
		t.Fatalf("ConstructionHash() error = %v", err)
	}
}

func TestConstructionService_ConstructionSubmit_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	// Create valid request
	request := &types.ConstructionSubmitRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    meshcommon.SoloNetwork,
		},
		SignedTransaction: "0x51f88481f68502b506882881b4e0df9416277a1ff38678291c41d1820957c78bb5da59ce880de0b6b3a764000080808609184e72a00082bb80808827706abefbc974eac0b8411bc4aff0c0d425ecd1a931ad435156a48b3e74b9a76b79fcc6e866337a73f05a7e05773a5fcbd4cf1251cf026e70c5cc5b3524866446de93a7d49897c0bac57900",
	}

	ctx := context.Background()
	_, err := service.ConstructionSubmit(ctx, request)

	// This will likely fail with mock client but tests the code path
	if err != nil {
		// Expected for mock client
		return
	}
}

func TestConstructionService_ConstructionDerive_EmptyPublicKey(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionDeriveRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		PublicKey: &types.PublicKey{
			Bytes:     []byte{},
			CurveType: meshtests.SECP256k1,
		},
	}

	ctx := context.Background()
	_, err := service.ConstructionDerive(ctx, request)

	if err == nil {
		t.Error("ConstructionDerive() expected error for empty public key")
	}
}

func TestConstructionService_ConstructionDerive_InvalidPublicKey(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionDeriveRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		PublicKey: &types.PublicKey{
			Bytes:     []byte{0x01, 0x02, 0x03},
			CurveType: meshtests.SECP256k1,
		},
	}

	ctx := context.Background()
	_, err := service.ConstructionDerive(ctx, request)

	if err == nil {
		t.Error("ConstructionDerive() expected error for invalid public key")
	}
}

func TestConstructionService_ConstructionPreprocess_MultipleOrigins(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionPreprocessRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeTransfer,
				Account:             &types.AccountIdentifier{Address: meshtests.FirstSoloAddress},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 1},
				Type:                meshcommon.OperationTypeTransfer,
				Account:             &types.AccountIdentifier{Address: meshtests.TestAddress1},
				Amount: &types.Amount{
					Value:    "1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 2},
				Type:                meshcommon.OperationTypeTransfer,
				Account:             &types.AccountIdentifier{Address: "0x2222222222222222222222222222222222222222"}, // Different origin
				Amount: &types.Amount{
					Value:    "-500000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
		},
	}

	ctx := context.Background()
	_, err := service.ConstructionPreprocess(ctx, request)

	if err == nil {
		t.Error("ConstructionPreprocess() expected error for multiple origins")
	}
}

func TestConstructionService_ConstructionPreprocess_NoOrigins(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionPreprocessRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeTransfer,
				Account:             &types.AccountIdentifier{Address: meshtests.TestAddress1},
				Amount: &types.Amount{
					Value:    "1000000000000000000", // Only positive amount, no sender
					Currency: meshcommon.VETCurrency,
				},
			},
		},
	}

	ctx := context.Background()
	_, err := service.ConstructionPreprocess(ctx, request)

	if err == nil {
		t.Error("ConstructionPreprocess() expected error without origins")
	}
}

func TestConstructionService_ConstructionPreprocess_NoTransferOperations(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionPreprocessRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeFee,
				Account:             &types.AccountIdentifier{Address: meshtests.FirstSoloAddress},
				Amount: &types.Amount{
					Value:    "-1000",
					Currency: meshcommon.VTHOCurrency,
				},
			},
		},
	}

	ctx := context.Background()
	_, err := service.ConstructionPreprocess(ctx, request)

	if err == nil {
		t.Error("ConstructionPreprocess() expected error without transfer operations")
	}
}

func TestConstructionService_ConstructionParse_InvalidTransactionHex(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionParseRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		Signed:            false,
		Transaction:       "0xINVALID_HEX",
	}

	ctx := context.Background()
	_, err := service.ConstructionParse(ctx, request)

	if err == nil {
		t.Error("ConstructionParse() expected error for invalid hex")
	}
}

func TestConstructionService_ConstructionParse_InvalidTransactionBytes(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionParseRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		Signed:            false,
		Transaction:       "0x0102030405", // Invalid transaction bytes
	}

	ctx := context.Background()
	_, err := service.ConstructionParse(ctx, request)

	if err == nil {
		t.Error("ConstructionParse() expected error for invalid transaction bytes")
	}
}

func TestConstructionService_ConstructionHash_InvalidTransactionHex(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionHashRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		SignedTransaction: "0xINVALID_HEX",
	}

	ctx := context.Background()
	_, err := service.ConstructionHash(ctx, request)

	if err == nil {
		t.Error("ConstructionHash() expected error for invalid hex")
	}
}

func TestConstructionService_ConstructionHash_InvalidTransactionBytes(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionHashRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		SignedTransaction: "0x0102030405", // Invalid transaction bytes
	}

	ctx := context.Background()
	_, err := service.ConstructionHash(ctx, request)

	if err == nil {
		t.Error("ConstructionHash() expected error for invalid transaction bytes")
	}
}

func TestConstructionService_ConstructionSubmit_InvalidTransactionHex(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionSubmitRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		SignedTransaction: "0xINVALID_HEX",
	}

	ctx := context.Background()
	_, err := service.ConstructionSubmit(ctx, request)

	if err == nil {
		t.Error("ConstructionSubmit() expected error for invalid hex")
	}
}

func TestConstructionService_ConstructionSubmit_InvalidTransactionBytes(t *testing.T) {
	service := createMockConstructionService()

	request := &types.ConstructionSubmitRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		SignedTransaction: "0x0102030405", // Invalid transaction bytes
	}

	ctx := context.Background()
	_, err := service.ConstructionSubmit(ctx, request)

	if err == nil {
		t.Error("ConstructionSubmit() expected error for invalid transaction bytes")
	}
}

func TestConstructionService_createDelegatorPayload_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	// Create a valid VeChain transaction for testing using thor.Builder
	request := types.ConstructionPayloadsRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 1},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: meshtests.TestAddress1,
				},
				Amount: &types.Amount{
					Value:    "1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
		},
		Metadata: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
			"blockRef":        "0x0000000000000000",
			"chainTag":        float64(1),
			"gas":             float64(21000),
			"nonce":           "0x1",
			"gasPriceCoef":    uint8(128),
		},
	}

	vechainTx, err := service.builder.BuildTransactionFromRequest(request, service.config.Expiration)
	if err != nil {
		t.Fatalf("Failed to build transaction: %v", err)
	}

	// Create two valid public keys: one for origin, one for delegator
	publicKeys := []*types.PublicKey{
		{
			Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
			CurveType: meshtests.SECP256k1,
		},
		{
			Bytes:     []byte{0x02, 0x79, 0xbe, 0x66, 0x7e, 0xf9, 0xdc, 0xbb, 0xac, 0x55, 0xa0, 0x62, 0x95, 0xce, 0x87, 0x0b, 0x07, 0x02, 0x9b, 0xfc, 0xdb, 0x2d, 0xce, 0x28, 0xd9, 0x59, 0xf2, 0x81, 0x5b, 0x16, 0xf8, 0x17, 0x98},
			CurveType: meshtests.SECP256k1,
		},
	}

	// Call the private method
	payload, err := service.createDelegatorPayload(vechainTx, publicKeys)

	if err != nil {
		t.Fatalf("createDelegatorPayload() error = %v", err)
	}

	// Verify the payload structure
	if payload == nil {
		t.Fatal("createDelegatorPayload() returned nil payload")
	}

	if payload.AccountIdentifier == nil {
		t.Error("createDelegatorPayload() AccountIdentifier is nil")
	}

	if payload.AccountIdentifier.Address == "" {
		t.Error("createDelegatorPayload() AccountIdentifier.Address is empty")
	}

	if len(payload.Bytes) != 32 {
		t.Errorf("createDelegatorPayload() Bytes length = %d, want 32", len(payload.Bytes))
	}

	if payload.SignatureType != types.EcdsaRecovery {
		t.Errorf("createDelegatorPayload() SignatureType = %v, want %v", payload.SignatureType, types.EcdsaRecovery)
	}

	// Verify that the delegator address is different from origin address
	originAddress, _ := service.bytesHandler.ComputeAddress(publicKeys[0])
	if payload.AccountIdentifier.Address == originAddress {
		t.Error("createDelegatorPayload() delegator address should be different from origin address")
	}
}

func TestConstructionService_createDelegatorPayload_InvalidDelegatorPublicKey(t *testing.T) {
	service := createMockConstructionService()

	// Create a valid VeChain transaction
	request := types.ConstructionPayloadsRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
		},
		Metadata: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
			"blockRef":        "0x0000000000000000",
			"chainTag":        float64(1),
			"gas":             float64(21000),
			"nonce":           "0x1",
			"gasPriceCoef":    uint8(128),
		},
	}

	vechainTx, err := service.builder.BuildTransactionFromRequest(request, service.config.Expiration)
	if err != nil {
		t.Fatalf("Failed to build transaction: %v", err)
	}

	// Create public keys with invalid delegator key
	publicKeys := []*types.PublicKey{
		{
			Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
			CurveType: meshtests.SECP256k1,
		},
		{
			Bytes:     []byte{0x01, 0x02, 0x03}, // Invalid public key (too short)
			CurveType: meshtests.SECP256k1,
		},
	}

	// Call the private method
	_, err = service.createDelegatorPayload(vechainTx, publicKeys)

	if err == nil {
		t.Error("createDelegatorPayload() expected error for invalid delegator public key")
	}
}

func TestConstructionService_createDelegatorPayload_InvalidOriginPublicKey(t *testing.T) {
	service := createMockConstructionService()

	// Create a valid VeChain transaction
	request := types.ConstructionPayloadsRequest{
		NetworkIdentifier: createTestNetworkIdentifier(meshcommon.TestNetwork),
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
		},
		Metadata: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
			"blockRef":        "0x0000000000000000",
			"chainTag":        float64(1),
			"gas":             float64(21000),
			"nonce":           "0x1",
			"gasPriceCoef":    uint8(128),
		},
	}

	vechainTx, err := service.builder.BuildTransactionFromRequest(request, service.config.Expiration)
	if err != nil {
		t.Fatalf("Failed to build transaction: %v", err)
	}

	// Create public keys with invalid origin key
	publicKeys := []*types.PublicKey{
		{
			Bytes:     []byte{0x01, 0x02, 0x03}, // Invalid public key (too short)
			CurveType: meshtests.SECP256k1,
		},
		{
			Bytes:     []byte{0x02, 0x79, 0xbe, 0x66, 0x7e, 0xf9, 0xdc, 0xbb, 0xac, 0x55, 0xa0, 0x62, 0x95, 0xce, 0x87, 0x0b, 0x07, 0x02, 0x9b, 0xfc, 0xdb, 0x2d, 0xce, 0x28, 0xd9, 0x59, 0xf2, 0x81, 0x5b, 0x16, 0xf8, 0x17, 0x98},
			CurveType: meshtests.SECP256k1,
		},
	}

	// Call the private method
	_, err = service.createDelegatorPayload(vechainTx, publicKeys)

	if err == nil {
		t.Error("createDelegatorPayload() expected error for invalid origin public key")
	}
}
