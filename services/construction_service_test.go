package services

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshhttp "github.com/vechain/mesh/common/http"
	meshconfig "github.com/vechain/mesh/config"
	meshtests "github.com/vechain/mesh/tests"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/thor"
	thorTx "github.com/vechain/thor/v2/tx"
)

func createMockConstructionService() *ConstructionService {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000",
	}
	return NewConstructionService(mockClient, config)
}

func createTestPublicKey() *types.PublicKey {
	return &types.PublicKey{
		Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
		CurveType: "secp256k1",
	}
}

func createTestNetworkIdentifier(network string) *types.NetworkIdentifier {
	return &types.NetworkIdentifier{
		Blockchain: meshcommon.BlockchainName,
		Network:    network,
	}
}

func makeHTTPRequest(method, url string, body []byte) (*httptest.ResponseRecorder, *http.Request) {
	req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Simulate middleware by adding request body to context
	ctx := context.WithValue(req.Context(), meshhttp.RequestBodyKey, body)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	return w, req
}

func TestNewConstructionService(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
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

func TestConstructionService_ConstructionDerive_InvalidRequestBody(t *testing.T) {
	service := createMockConstructionService()

	// Create request with invalid JSON
	w, req := makeHTTPRequest("POST", meshcommon.ConstructionDeriveEndpoint, []byte("invalid json"))

	// Call ConstructionDerive
	service.ConstructionDerive(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionDerive() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionDerive_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	// Create valid request
	request := types.ConstructionDeriveRequest{
		NetworkIdentifier: createTestNetworkIdentifier("test"),
		PublicKey:         createTestPublicKey(),
	}

	requestBody, _ := json.Marshal(request)
	w, req := makeHTTPRequest("POST", meshcommon.ConstructionDeriveEndpoint, requestBody)

	// Call ConstructionDerive
	service.ConstructionDerive(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("ConstructionDerive() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Parse response
	var response types.ConstructionDeriveResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response structure
	if response.AccountIdentifier == nil {
		t.Errorf("ConstructionDerive() account identifier is nil")
	}
	if response.AccountIdentifier.Address == "" {
		t.Errorf("ConstructionDerive() account address is empty")
	}
}

func TestConstructionService_ConstructionPreprocess_InvalidRequestBody(t *testing.T) {
	service := createMockConstructionService()

	// Create request with invalid JSON
	w, req := makeHTTPRequest("POST", meshcommon.ConstructionPreprocessEndpoint, []byte("invalid json"))

	// Call ConstructionPreprocess
	service.ConstructionPreprocess(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionPreprocess() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionPreprocess_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	// Create valid request with both sender (negative) and receiver (positive) operations
	request := types.ConstructionPreprocessRequest{
		NetworkIdentifier: createTestNetworkIdentifier("test"),
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

	requestBody, _ := json.Marshal(request)
	w, req := makeHTTPRequest("POST", meshcommon.ConstructionPreprocessEndpoint, requestBody)

	// Call ConstructionPreprocess
	service.ConstructionPreprocess(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("ConstructionPreprocess() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Parse response
	var response types.ConstructionPreprocessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response structure
	if response.Options == nil {
		t.Errorf("ConstructionPreprocess() options is nil")
	}
}

func TestConstructionService_ConstructionPreprocess_VIP180Token(t *testing.T) {
	service := createMockConstructionService()

	// Create valid request with VIP180 token operations
	request := types.ConstructionPreprocessRequest{
		NetworkIdentifier: createTestNetworkIdentifier("test"),
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

	requestBody, _ := json.Marshal(request)
	w, req := makeHTTPRequest("POST", meshcommon.ConstructionPreprocessEndpoint, requestBody)

	// Call ConstructionPreprocess
	service.ConstructionPreprocess(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("ConstructionPreprocess() status code = %v, want %v. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	// Parse response
	var response types.ConstructionPreprocessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response structure
	if response.Options == nil {
		t.Errorf("ConstructionPreprocess() options is nil")
		return
	}

	// Verify clauses were created for VIP180 token transfer
	if clauses, ok := response.Options["clauses"].([]any); ok {
		if len(clauses) == 0 {
			t.Errorf("ConstructionPreprocess() expected clauses, got empty array")
		}
		// Verify the clause has the token contract address as "to"
		if len(clauses) > 0 {
			if clause, ok := clauses[0].(map[string]any); ok {
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
		}
	} else {
		t.Errorf("ConstructionPreprocess() options['clauses'] not found or not an array")
	}
}

func TestConstructionService_ConstructionMetadata_InvalidRequestBody(t *testing.T) {
	service := createMockConstructionService()

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", meshcommon.ConstructionMetadataEndpoint, bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call ConstructionMetadata
	service.ConstructionMetadata(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionMetadata() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionMetadata_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	// Create valid request for legacy
	request := types.ConstructionMetadataRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		Options: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.ConstructionMetadataEndpoint, request)
	w := httptest.NewRecorder()

	// Call ConstructionMetadata
	service.ConstructionMetadata(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("ConstructionMetadata() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Parse response
	var response types.ConstructionMetadataResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
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
	request := types.ConstructionMetadataRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		Options: map[string]any{
			"transactionType": meshcommon.TransactionTypeDynamic,
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.ConstructionMetadataEndpoint, request)
	w := httptest.NewRecorder()

	// Call ConstructionMetadata
	service.ConstructionMetadata(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("ConstructionMetadata() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Parse response
	var response types.ConstructionMetadataResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response structure
	if response.Metadata == nil {
		t.Errorf("ConstructionMetadata() metadata is nil")
	}
	if len(response.SuggestedFee) == 0 {
		t.Errorf("ConstructionMetadata() suggested fee is empty")
	}
}

func TestConstructionService_ConstructionPayloads_InvalidRequestBody(t *testing.T) {
	service := createMockConstructionService()

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", meshcommon.ConstructionPayloadsEndpoint, bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call ConstructionPayloads
	service.ConstructionPayloads(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionPayloads() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionPayloads_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	// Create valid request
	request := types.ConstructionPayloadsRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
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
				CurveType: "secp256k1",
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

	req := meshtests.CreateRequestWithContext("POST", meshcommon.ConstructionPayloadsEndpoint, request)
	w := httptest.NewRecorder()

	// Call ConstructionPayloads
	service.ConstructionPayloads(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("ConstructionPayloads() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Parse response
	var response types.ConstructionPayloadsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response structure
	if len(response.Payloads) == 0 {
		t.Errorf("ConstructionPayloads() payloads is empty")
	}
}

func TestConstructionService_ConstructionPayloads_OriginAddressMismatch(t *testing.T) {
	service := createMockConstructionService()

	// Create request with mismatched origin address
	request := types.ConstructionPayloadsRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: "0x1234567890123456789012345678901234567890", // Different address
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
				CurveType: "secp256k1",
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

	req := meshtests.CreateRequestWithContext("POST", meshcommon.ConstructionPayloadsEndpoint, request)
	w := httptest.NewRecorder()

	// Call ConstructionPayloads
	service.ConstructionPayloads(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionPayloads() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionPayloads_InvalidPublicKey(t *testing.T) {
	service := createMockConstructionService()

	// Create request with invalid public key
	request := types.ConstructionPayloadsRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
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
				Bytes:     []byte{0x01, 0x02, 0x03}, // Invalid public key (too short)
				CurveType: "secp256k1",
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

	req := meshtests.CreateRequestWithContext("POST", meshcommon.ConstructionPayloadsEndpoint, request)
	w := httptest.NewRecorder()

	// Call ConstructionPayloads
	service.ConstructionPayloads(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionPayloads() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionPayloads_DelegatorAddressMismatch(t *testing.T) {
	service := createMockConstructionService()

	// Create request with fee delegation but mismatched delegator address
	request := types.ConstructionPayloadsRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
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
				CurveType: "secp256k1",
			},
			{
				Bytes:     []byte{0x02, 0x79, 0xbe, 0x66, 0x7e, 0xf9, 0xdc, 0xbb, 0xac, 0x55, 0xa0, 0x62, 0x95, 0xce, 0x87, 0x0b, 0x07, 0x02, 0x9b, 0xfc, 0xdb, 0x2d, 0xce, 0x28, 0xd9, 0x59, 0xf2, 0x81, 0x5b, 0x16, 0xf8, 0x17, 0x98},
				CurveType: "secp256k1",
			},
		},
		Metadata: map[string]any{
			"transactionType":       meshcommon.TransactionTypeLegacy,
			"blockRef":              "0x0000000000000000",
			"chainTag":              float64(1),
			"gas":                   float64(21000),
			"nonce":                 "0x1",
			"gasPriceCoef":          uint8(128),
			"fee_delegator_account": "0x1234567890123456789012345678901234567890", // Different from public key
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.ConstructionPayloadsEndpoint, request)
	w := httptest.NewRecorder()

	// Call ConstructionPayloads
	service.ConstructionPayloads(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionPayloads() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionParse_InvalidRequestBody(t *testing.T) {
	service := createMockConstructionService()

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", meshcommon.ConstructionParseEndpoint, bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call ConstructionParse
	service.ConstructionParse(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionParse() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionParse_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	request := types.ConstructionParseRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		Signed:      false,
		Transaction: "0xf85db84551f84281f68502b506882881b4e0df9416277a1ff38678291c41d1820957c78bb5da59ce880de0b6b3a764000080808609184e72a00082bb80808827706abefbc974eac08094f077b491b355e64048ce21e3a6fc4751eeea77fa80",
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.ConstructionParseEndpoint, request)
	w := httptest.NewRecorder()

	service.ConstructionParse(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ConstructionParse() unexpected status code = %v", w.Code)
	}
}

func TestConstructionService_ConstructionCombine_InvalidRequestBody(t *testing.T) {
	service := createMockConstructionService()

	req := httptest.NewRequest("POST", meshcommon.ConstructionCombineEndpoint, bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	service.ConstructionCombine(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionCombine() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionCombine_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	// Create valid request with the provided values
	request := types.ConstructionCombineRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "solo",
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
					CurveType: "secp256k1",
				},
				SignatureType: "ecdsa_recovery",
				Bytes:         []byte{0x1b, 0xc4, 0xaf, 0xf0, 0xc0, 0xd4, 0x25, 0xec, 0xd1, 0xa9, 0x31, 0xad, 0x43, 0x51, 0x56, 0xa4, 0x8b, 0x3e, 0x74, 0xb9, 0xa7, 0x6b, 0x79, 0xfc, 0xc6, 0xe8, 0x66, 0x33, 0x7a, 0x73, 0xf0, 0x5a, 0x7e, 0x05, 0x77, 0x3a, 0x5f, 0xcb, 0xd4, 0xcf, 0x12, 0x51, 0xcf, 0x02, 0x6e, 0x70, 0xc5, 0xcc, 0x5b, 0x35, 0x24, 0x86, 0x64, 0x46, 0xde, 0x93, 0xa7, 0xd4, 0x98, 0x97, 0xc0, 0xba, 0xc5, 0x79, 0x00},
			},
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.ConstructionCombineEndpoint, request)
	w := httptest.NewRecorder()

	service.ConstructionCombine(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ConstructionCombine() expected status code 200, got %v. Response: %s", w.Code, w.Body.String())
	}

	// Verify response structure
	var response types.ConstructionCombineResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	if response.SignedTransaction == "" {
		t.Errorf("Expected signed transaction in response")
	}
}

func TestConstructionService_ConstructionCombine_InvalidUnsignedTransaction(t *testing.T) {
	service := createMockConstructionService()

	// Create request with invalid unsigned transaction
	request := types.ConstructionCombineRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "solo",
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
					CurveType: "secp256k1",
				},
				SignatureType: "ecdsa_recovery",
				Bytes:         []byte{0x1b, 0xc4, 0xaf, 0xf0, 0xc0, 0xd4, 0x25, 0xec, 0xd1, 0xa9, 0x31, 0xad, 0x43, 0x51, 0x56, 0xa4, 0x8b, 0x3e, 0x74, 0xb9, 0xa7, 0x6b, 0x79, 0xfc, 0xc6, 0xe8, 0x66, 0x33, 0x7a, 0x73, 0xf0, 0x5a, 0x7e, 0x05, 0x77, 0x3a, 0x5f, 0xcb, 0xd4, 0xcf, 0x12, 0x51, 0xcf, 0x02, 0x6e, 0x70, 0xc5, 0xcc, 0x5b, 0x35, 0x24, 0x86, 0x64, 0x46, 0xde, 0x93, 0xa7, 0xd4, 0x98, 0x97, 0xc0, 0xba, 0xc5, 0x79, 0x00},
			},
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.ConstructionCombineEndpoint, request)
	w := httptest.NewRecorder()

	service.ConstructionCombine(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionCombine() expected status code 400, got %v. Response: %s", w.Code, w.Body.String())
	}
}

func TestConstructionService_ConstructionCombine_InvalidNumberOfSignatures(t *testing.T) {
	service := createMockConstructionService()

	// Create request with no signatures
	request := types.ConstructionCombineRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "solo",
		},
		UnsignedTransaction: "0xf85281f68800000005e6911c7481b4dad99416277a1ff38678291c41d1820957c78bb5da59ce8227108082bb808864d53d1260b9a69f94f077b491b355e64048ce21e3a6fc4751eeea77fa808609184e72a00080",
		Signatures:          []*types.Signature{}, // No signatures
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.ConstructionCombineEndpoint, request)
	w := httptest.NewRecorder()

	service.ConstructionCombine(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionCombine() expected status code 400, got %v. Response: %s", w.Code, w.Body.String())
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
				CurveType: "secp256k1",
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
				CurveType: "secp256k1",
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
				CurveType: "secp256k1",
			},
			SignatureType: "ecdsa_recovery",
			Bytes:         []byte{0x03, 0xf8, 0xa1, 0xca, 0x4d, 0xd3, 0x9d, 0x99, 0xab, 0x54, 0x97, 0xf9, 0x4b, 0x8b, 0x79, 0x11, 0x34, 0x0c, 0xea, 0xc7, 0x18, 0x20, 0x19, 0xb7, 0xbe, 0x9d, 0x81, 0xf0, 0x43, 0xc7, 0x43, 0xf9, 0x5a, 0x69, 0x43, 0x1d, 0x71, 0x5a, 0xde, 0x0c, 0x9b, 0x74, 0x1f, 0x7c, 0x83, 0xd9, 0x57, 0x2a, 0xd8, 0x42, 0x71, 0xb4, 0xf2, 0xec, 0xb6, 0x2c, 0x8f, 0x49, 0xdd, 0xfa, 0x3e, 0x8c, 0x3a, 0xea, 0x01},
		},
	}

	req = meshtests.CreateRequestWithContext("POST", meshcommon.ConstructionCombineEndpoint, request)
	w = httptest.NewRecorder()

	service.ConstructionCombine(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionCombine() expected status code 400, got %v. Response: %s", w.Code, w.Body.String())
	}
}

func TestConstructionService_ConstructionHash_InvalidRequestBody(t *testing.T) {
	service := createMockConstructionService()

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", meshcommon.ConstructionHashEndpoint, bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call ConstructionHash
	service.ConstructionHash(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionHash() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionHash_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	request := types.ConstructionHashRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		SignedTransaction: "0x51f88481f68502b506882881b4e0df9416277a1ff38678291c41d1820957c78bb5da59ce880de0b6b3a764000080808609184e72a00082bb80808827706abefbc974eac0b8411bc4aff0c0d425ecd1a931ad435156a48b3e74b9a76b79fcc6e866337a73f05a7e05773a5fcbd4cf1251cf026e70c5cc5b3524866446de93a7d49897c0bac57900",
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.ConstructionHashEndpoint, request)
	w := httptest.NewRecorder()

	service.ConstructionHash(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ConstructionHash() unexpected status code = %v", w.Code)
	}
}

func TestConstructionService_ConstructionSubmit_InvalidRequestBody(t *testing.T) {
	service := createMockConstructionService()

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", meshcommon.ConstructionSubmitEndpoint, bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call ConstructionSubmit
	service.ConstructionSubmit(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionSubmit() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionSubmit_ValidRequest(t *testing.T) {
	service := createMockConstructionService()

	// Create valid request
	request := types.ConstructionSubmitRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "solo",
		},
		SignedTransaction: "0x51f88481f68502b506882881b4e0df9416277a1ff38678291c41d1820957c78bb5da59ce880de0b6b3a764000080808609184e72a00082bb80808827706abefbc974eac0b8411bc4aff0c0d425ecd1a931ad435156a48b3e74b9a76b79fcc6e866337a73f05a7e05773a5fcbd4cf1251cf026e70c5cc5b3524866446de93a7d49897c0bac57900",
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.ConstructionSubmitEndpoint, request)
	w := httptest.NewRecorder()

	// Call ConstructionSubmit
	service.ConstructionSubmit(w, req)

	// Check response - this will likely fail but covers more code paths
	if w.Code != http.StatusOK {
		t.Errorf("ConstructionSubmit() unexpected status code = %v", w.Code)
	}
}

func TestConstructionService_createDelegatorPayload(t *testing.T) {
	service := createMockConstructionService()

	// Create a valid VeChain transaction using the builder
	builder := thorTx.NewBuilder(thorTx.TypeLegacy)
	builder.ChainTag(0x27)
	blockRef := thorTx.BlockRef([8]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef})
	builder.BlockRef(blockRef)
	builder.Expiration(720)
	builder.Gas(21000)
	builder.GasPriceCoef(0)
	builder.Nonce(0x1234567890abcdef)

	// Add a clause
	toAddr, _ := thor.ParseAddress(meshtests.TestAddress1)
	value := new(big.Int)
	value.SetString("1000000000000000000", 10) // 1 VET

	thorClause := thorTx.NewClause(&toAddr)
	thorClause = thorClause.WithValue(value)
	thorClause = thorClause.WithData([]byte{})
	builder.Clause(thorClause)

	// Build the transaction
	vechainTx := builder.Build()

	// Create public keys for origin and delegator
	publicKeys := []*types.PublicKey{
		{
			Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
			CurveType: "secp256k1",
		},
		{
			Bytes:     []byte{0x02, 0x79, 0xbe, 0x66, 0x7e, 0xf9, 0xdc, 0xbb, 0xac, 0x55, 0xa0, 0x62, 0x95, 0xce, 0x87, 0x0b, 0x07, 0x02, 0x9b, 0xfc, 0xdb, 0x2d, 0xce, 0x28, 0xd9, 0x59, 0xf2, 0x81, 0x5b, 0x16, 0xf8, 0x17, 0x98},
			CurveType: "secp256k1",
		},
	}

	// Test createDelegatorPayload with valid transaction
	payload, err := service.createDelegatorPayload(vechainTx, publicKeys)
	if err != nil {
		t.Errorf("createDelegatorPayload() error = %v", err)
	}
	if payload.AccountIdentifier == nil {
		t.Errorf("createDelegatorPayload() returned nil AccountIdentifier")
	}
	if payload.Bytes == nil {
		t.Errorf("createDelegatorPayload() returned nil Bytes")
	}
	if payload.SignatureType != "ecdsa_recovery" {
		t.Errorf("createDelegatorPayload() SignatureType = %v, want ecdsa_recovery", payload.SignatureType)
	}

	// Test with invalid public key (should return error)
	invalidPublicKeys := []*types.PublicKey{
		{
			Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
			CurveType: "secp256k1",
		},
		{
			Bytes:     []byte{0x00}, // Invalid public key
			CurveType: "secp256k1",
		},
	}

	// Test with invalid public key - this should return an error
	_, err = service.createDelegatorPayload(vechainTx, invalidPublicKeys)
	if err == nil {
		t.Errorf("createDelegatorPayload() with invalid public key should return error")
	}
}

func TestConstructionService_getFeeDelegatorAccount(t *testing.T) {
	service := createMockConstructionService()

	// Test with valid fee delegator account
	metadata := map[string]any{
		"fee_delegator_account": meshtests.TestAddress1,
	}

	account := service.getFeeDelegatorAccount(metadata)
	if account != meshtests.TestAddress1 {
		t.Errorf("getFeeDelegatorAccount() = %v, want %s", account, meshtests.TestAddress1)
	}

	// Test with missing fee delegator account
	metadataEmpty := map[string]any{}
	accountEmpty := service.getFeeDelegatorAccount(metadataEmpty)
	if accountEmpty != "" {
		t.Errorf("getFeeDelegatorAccount() with empty metadata = %v, want empty string", accountEmpty)
	}

	// Test with nil metadata
	accountNil := service.getFeeDelegatorAccount(nil)
	if accountNil != "" {
		t.Errorf("getFeeDelegatorAccount() with nil metadata = %v, want empty string", accountNil)
	}
}

func TestConstructionService_calculateGas(t *testing.T) {
	service := createMockConstructionService()

	tests := []struct {
		name        string
		options     map[string]any
		expectError bool
		validate    func(*testing.T, uint64)
	}{
		{
			name:        "no clauses - returns base gas with 20% buffer",
			options:     map[string]any{},
			expectError: false,
			validate: func(t *testing.T, gas uint64) {
				expected := uint64(21000 * 1.2)
				if gas != expected {
					t.Errorf("Expected %d, got %d", expected, gas)
				}
			},
		},
		{
			name: "single clause without data - uses IntrinsicGas with 20% buffer",
			options: map[string]any{
				"clauses": []any{
					map[string]any{
						"to":    "0x1234567890123456789012345678901234567890",
						"value": "0",
						"data":  "0x",
					},
				},
			},
			expectError: false,
			validate: func(t *testing.T, gas uint64) {
				expected := uint64(21000 * 1.2)
				if gas != expected {
					t.Errorf("Expected %d, got %d", expected, gas)
				}
			},
		},
		{
			name: "clause with data - gas increases due to data cost",
			options: map[string]any{
				"clauses": []any{
					map[string]any{
						"to":    "0x1234567890123456789012345678901234567890",
						"value": "0",
						"data":  "0xa9059cbb0000000000000000000000001234567890123456789012345678901234567890000000000000000000000000000000000000000000000000000000000000000a",
					},
				},
			},
			expectError: false,
			validate: func(t *testing.T, gas uint64) {
				baseGas := uint64(21000 * 1.2)
				if gas <= baseGas {
					t.Errorf("Gas with data (%d) should be > base gas (%d)", gas, baseGas)
				}
			},
		},
		{
			name: "multiple clauses - gas scales with clause count",
			options: map[string]any{
				"clauses": []any{
					map[string]any{
						"to":    "0x1234567890123456789012345678901234567890",
						"value": "0",
						"data":  "0x",
					},
					map[string]any{
						"to":    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
						"value": "0",
						"data":  "0x",
					},
				},
			},
			expectError: false,
			validate: func(t *testing.T, gas uint64) {
				expected := uint64(37000 * 1.2)
				if gas != expected {
					t.Errorf("Expected %d, got %d", expected, gas)
				}
			},
		},
		{
			name: "empty clauses array - returns base gas with buffer",
			options: map[string]any{
				"clauses": []any{},
			},
			expectError: false,
			validate: func(t *testing.T, gas uint64) {
				expected := uint64(21000 * 1.2)
				if gas != expected {
					t.Errorf("Expected %d, got %d", expected, gas)
				}
			},
		},
		{
			name: "invalid clauses format - returns error",
			options: map[string]any{
				"clauses": "invalid_format",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gas, err := service.calculateGas(tt.options)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, gas)
			}
		})
	}
}
func TestConstructionService_ConstructionDerive_EmptyPublicKey(t *testing.T) {
	service := createMockConstructionService()

	request := types.ConstructionDeriveRequest{
		NetworkIdentifier: createTestNetworkIdentifier("test"),
		PublicKey: &types.PublicKey{
			Bytes:     []byte{}, // Empty bytes
			CurveType: "secp256k1",
		},
	}

	requestBody, _ := json.Marshal(request)
	w, req := makeHTTPRequest("POST", meshcommon.ConstructionDeriveEndpoint, requestBody)

	service.ConstructionDerive(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionDerive() with empty public key status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionDerive_InvalidPublicKey(t *testing.T) {
	service := createMockConstructionService()

	request := types.ConstructionDeriveRequest{
		NetworkIdentifier: createTestNetworkIdentifier("test"),
		PublicKey: &types.PublicKey{
			Bytes:     []byte{0x01, 0x02, 0x03}, // Invalid public key
			CurveType: "secp256k1",
		},
	}

	requestBody, _ := json.Marshal(request)
	w, req := makeHTTPRequest("POST", meshcommon.ConstructionDeriveEndpoint, requestBody)

	service.ConstructionDerive(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionDerive() with invalid public key status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionPreprocess_MultipleOrigins(t *testing.T) {
	service := createMockConstructionService()

	request := types.ConstructionPreprocessRequest{
		NetworkIdentifier: createTestNetworkIdentifier("test"),
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

	requestBody, _ := json.Marshal(request)
	w, req := makeHTTPRequest("POST", meshcommon.ConstructionPreprocessEndpoint, requestBody)

	service.ConstructionPreprocess(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionPreprocess() with multiple origins status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionPreprocess_NoOrigins(t *testing.T) {
	service := createMockConstructionService()

	request := types.ConstructionPreprocessRequest{
		NetworkIdentifier: createTestNetworkIdentifier("test"),
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

	requestBody, _ := json.Marshal(request)
	w, req := makeHTTPRequest("POST", meshcommon.ConstructionPreprocessEndpoint, requestBody)

	service.ConstructionPreprocess(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionPreprocess() without origins status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionPreprocess_NoTransferOperations(t *testing.T) {
	service := createMockConstructionService()

	request := types.ConstructionPreprocessRequest{
		NetworkIdentifier: createTestNetworkIdentifier("test"),
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

	requestBody, _ := json.Marshal(request)
	w, req := makeHTTPRequest("POST", meshcommon.ConstructionPreprocessEndpoint, requestBody)

	service.ConstructionPreprocess(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionPreprocess() without transfer operations status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionParse_InvalidTransactionHex(t *testing.T) {
	service := createMockConstructionService()

	request := types.ConstructionParseRequest{
		NetworkIdentifier: createTestNetworkIdentifier("test"),
		Signed:            false,
		Transaction:       "0xINVALID_HEX",
	}

	requestBody, _ := json.Marshal(request)
	w, req := makeHTTPRequest("POST", meshcommon.ConstructionParseEndpoint, requestBody)

	service.ConstructionParse(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionParse() with invalid hex status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionParse_InvalidTransactionBytes(t *testing.T) {
	service := createMockConstructionService()

	request := types.ConstructionParseRequest{
		NetworkIdentifier: createTestNetworkIdentifier("test"),
		Signed:            false,
		Transaction:       "0x0102030405", // Invalid transaction bytes
	}

	requestBody, _ := json.Marshal(request)
	w, req := makeHTTPRequest("POST", meshcommon.ConstructionParseEndpoint, requestBody)

	service.ConstructionParse(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionParse() with invalid transaction bytes status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionHash_InvalidTransactionHex(t *testing.T) {
	service := createMockConstructionService()

	request := types.ConstructionHashRequest{
		NetworkIdentifier: createTestNetworkIdentifier("test"),
		SignedTransaction: "0xINVALID_HEX",
	}

	requestBody, _ := json.Marshal(request)
	w, req := makeHTTPRequest("POST", meshcommon.ConstructionHashEndpoint, requestBody)

	service.ConstructionHash(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionHash() with invalid hex status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionHash_InvalidTransactionBytes(t *testing.T) {
	service := createMockConstructionService()

	request := types.ConstructionHashRequest{
		NetworkIdentifier: createTestNetworkIdentifier("test"),
		SignedTransaction: "0x0102030405", // Invalid transaction bytes
	}

	requestBody, _ := json.Marshal(request)
	w, req := makeHTTPRequest("POST", meshcommon.ConstructionHashEndpoint, requestBody)

	service.ConstructionHash(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionHash() with invalid transaction bytes status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionSubmit_InvalidTransactionHex(t *testing.T) {
	service := createMockConstructionService()

	request := types.ConstructionSubmitRequest{
		NetworkIdentifier: createTestNetworkIdentifier("test"),
		SignedTransaction: "0xINVALID_HEX",
	}

	requestBody, _ := json.Marshal(request)
	w, req := makeHTTPRequest("POST", meshcommon.ConstructionSubmitEndpoint, requestBody)

	service.ConstructionSubmit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionSubmit() with invalid hex status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionSubmit_InvalidTransactionBytes(t *testing.T) {
	service := createMockConstructionService()

	request := types.ConstructionSubmitRequest{
		NetworkIdentifier: createTestNetworkIdentifier("test"),
		SignedTransaction: "0x0102030405", // Invalid transaction bytes
	}

	requestBody, _ := json.Marshal(request)
	w, req := makeHTTPRequest("POST", meshcommon.ConstructionSubmitEndpoint, requestBody)

	service.ConstructionSubmit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionSubmit() with invalid transaction bytes status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}
