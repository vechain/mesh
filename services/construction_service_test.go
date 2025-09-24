package services

import (
	"bytes"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshconfig "github.com/vechain/mesh/config"
	meshthor "github.com/vechain/mesh/thor"
	meshutils "github.com/vechain/mesh/utils"
	"github.com/vechain/thor/v2/thor"
	thorTx "github.com/vechain/thor/v2/tx"
)

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
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/construction/derive", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call ConstructionDerive
	service.ConstructionDerive(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionDerive() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionDerive_ValidRequest(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	// Create valid request
	request := types.ConstructionDeriveRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		PublicKey: &types.PublicKey{
			Bytes:     []byte{0x03, 0xe3, 0x2e, 0x59, 0x60, 0x78, 0x1c, 0xe0, 0xb4, 0x3d, 0x8c, 0x29, 0x52, 0xee, 0xea, 0x4b, 0x95, 0xe2, 0x86, 0xb1, 0xbb, 0x5f, 0x8c, 0x1f, 0x0c, 0x9f, 0x09, 0x98, 0x3b, 0xa7, 0x14, 0x1d, 0x2f},
			CurveType: "secp256k1",
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/construction/derive", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

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
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/construction/preprocess", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call ConstructionPreprocess
	service.ConstructionPreprocess(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionPreprocess() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionPreprocess_ValidRequest(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	// Create valid request
	request := types.ConstructionPreprocessRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshutils.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: "0x16277a1ff38678291c41d1820957c78bb5da59ce",
				},
				Amount: &types.Amount{
					Value:    "1000000000000000000",
					Currency: meshutils.VETCurrency,
				},
			},
		},
		Metadata: map[string]any{
			"transactionType": "legacy",
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/construction/preprocess", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

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

func TestConstructionService_ConstructionMetadata_InvalidRequestBody(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/construction/metadata", bytes.NewBufferString("invalid json"))
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
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	// Create valid request for legacy
	request := types.ConstructionMetadataRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		Options: map[string]any{
			"transactionType": "legacy",
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/construction/metadata", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
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
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	// Create valid request for dynamic
	request := types.ConstructionMetadataRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		Options: map[string]any{
			"transactionType": "dynamic",
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/construction/metadata", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
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
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/construction/payloads", bytes.NewBufferString("invalid json"))
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
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	// Create valid request
	request := types.ConstructionPayloadsRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshutils.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshutils.VETCurrency,
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
			"transactionType": "legacy",
			"blockRef":        "0x0000000000000000",
			"chainTag":        float64(1),
			"gas":             float64(21000),
			"nonce":           "0x1",
			"gasPriceCoef":    uint8(128),
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/construction/payloads", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
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
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000",
	}
	service := NewConstructionService(mockClient, config)

	// Create request with mismatched origin address
	request := types.ConstructionPayloadsRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshutils.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: "0x1234567890123456789012345678901234567890", // Different address
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshutils.VETCurrency,
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
			"transactionType": "legacy",
			"blockRef":        "0x0000000000000000",
			"chainTag":        float64(1),
			"gas":             float64(21000),
			"nonce":           "0x1",
			"gasPriceCoef":    uint8(128),
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/construction/payloads", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call ConstructionPayloads
	service.ConstructionPayloads(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionPayloads() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionPayloads_InvalidPublicKey(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000",
	}
	service := NewConstructionService(mockClient, config)

	// Create request with invalid public key
	request := types.ConstructionPayloadsRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshutils.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshutils.VETCurrency,
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
			"transactionType": "legacy",
			"blockRef":        "0x0000000000000000",
			"chainTag":        float64(1),
			"gas":             float64(21000),
			"nonce":           "0x1",
			"gasPriceCoef":    uint8(128),
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/construction/payloads", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call ConstructionPayloads
	service.ConstructionPayloads(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionPayloads() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionPayloads_DelegatorAddressMismatch(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000",
	}
	service := NewConstructionService(mockClient, config)

	// Create request with fee delegation but mismatched delegator address
	request := types.ConstructionPayloadsRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshutils.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshutils.VETCurrency,
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
			"transactionType":       "legacy",
			"blockRef":              "0x0000000000000000",
			"chainTag":              float64(1),
			"gas":                   float64(21000),
			"nonce":                 "0x1",
			"gasPriceCoef":          uint8(128),
			"fee_delegator_account": "0x1234567890123456789012345678901234567890", // Different from public key
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/construction/payloads", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call ConstructionPayloads
	service.ConstructionPayloads(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionPayloads() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionParse_InvalidRequestBody(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/construction/parse", bytes.NewBufferString("invalid json"))
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
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	request := types.ConstructionParseRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		Signed:      false,
		Transaction: "0xf85081e486039791786ecd81b4dad99416277a1ff38678291c41d1820957c78bb5da59ce82271080828ca088e6e47234f992efad94c05c334533c673582616ac2bf404b6c55efa1087808609184e72a00080",
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/construction/parse", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	service.ConstructionParse(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ConstructionParse() unexpected status code = %v", w.Code)
	}
}

func TestConstructionService_ConstructionCombine_InvalidRequestBody(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	req := httptest.NewRequest("POST", "/construction/combine", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	service.ConstructionCombine(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionCombine() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestConstructionService_ConstructionCombine_ValidRequest(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "solo",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	// Create valid request with the provided values
	request := types.ConstructionCombineRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "solo",
		},
		UnsignedTransaction: "0xf85281f68800000005e6911c7481b4dad99416277a1ff38678291c41d1820957c78bb5da59ce8227108082bb808864d53d1260b9a69f94f077b491b355e64048ce21e3a6fc4751eeea77fa808609184e72a00080",
		Signatures: []*types.Signature{
			{
				SigningPayload: &types.SigningPayload{
					AccountIdentifier: &types.AccountIdentifier{
						Address: "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
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
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/construction/combine", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
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
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "solo",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000",
	}
	service := NewConstructionService(mockClient, config)

	// Create request with invalid unsigned transaction
	request := types.ConstructionCombineRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "solo",
		},
		UnsignedTransaction: "invalid_hex",
		Signatures: []*types.Signature{
			{
				SigningPayload: &types.SigningPayload{
					AccountIdentifier: &types.AccountIdentifier{
						Address: "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
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
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/construction/combine", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	service.ConstructionCombine(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionCombine() expected status code 400, got %v. Response: %s", w.Code, w.Body.String())
	}
}

func TestConstructionService_ConstructionCombine_InvalidNumberOfSignatures(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "solo",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000",
	}
	service := NewConstructionService(mockClient, config)

	// Create request with no signatures
	request := types.ConstructionCombineRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "solo",
		},
		UnsignedTransaction: "0xf85281f68800000005e6911c7481b4dad99416277a1ff38678291c41d1820957c78bb5da59ce8227108082bb808864d53d1260b9a69f94f077b491b355e64048ce21e3a6fc4751eeea77fa808609184e72a00080",
		Signatures:          []*types.Signature{}, // No signatures
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/construction/combine", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
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
					Address: "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
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
					Address: "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
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
					Address: "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
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

	requestBody, _ = json.Marshal(request)
	req = httptest.NewRequest("POST", "/construction/combine", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	service.ConstructionCombine(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionCombine() expected status code 400, got %v. Response: %s", w.Code, w.Body.String())
	}
}

func TestConstructionService_ConstructionHash_InvalidRequestBody(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/construction/hash", bytes.NewBufferString("invalid json"))
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
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	request := types.ConstructionHashRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		SignedTransaction: "0xf89581f68800000005e6911c7481b4dad99416277a1ff38678291c41d1820957c78bb5da59ce8227108082bb808864d53d1260b9a69f94f077b491b355e64048ce21e3a6fc4751eeea77fa808609184e72a00080b84103f8a1ca4dd39d99ab5497f94b8b7911340ceac7182019b7be9d81f043c743f95a69431d715ade0c9b741f7c83d9572ad84271b4f2ecb62c8f49ddfa3e8c3aea01",
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/construction/hash", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	service.ConstructionHash(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ConstructionHash() unexpected status code = %v", w.Code)
	}
}

func TestConstructionService_ConstructionSubmit_InvalidRequestBody(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/construction/submit", bytes.NewBufferString("invalid json"))
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
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "solo",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	// Create valid request
	request := types.ConstructionSubmitRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "solo",
		},
		SignedTransaction: "0xf89581f68800000005e6911c7481b4dad99416277a1ff38678291c41d1820957c78bb5da59ce8227108082bb808864d53d1260b9a69f94f077b491b355e64048ce21e3a6fc4751eeea77fa808609184e72a00080b84103f8a1ca4dd39d99ab5497f94b8b7911340ceac7182019b7be9d81f043c743f95a69431d715ade0c9b741f7c83d9572ad84271b4f2ecb62c8f49ddfa3e8c3aea01",
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/construction/submit", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call ConstructionSubmit
	service.ConstructionSubmit(w, req)

	// Check response - this will likely fail but covers more code paths
	if w.Code != http.StatusOK {
		t.Errorf("ConstructionSubmit() unexpected status code = %v", w.Code)
	}
}

func TestConstructionService_createDelegatorPayload(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000",
	}
	service := NewConstructionService(mockClient, config)

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
	toAddr, _ := thor.ParseAddress("0x16277a1ff38678291c41d1820957c78bb5da59ce")
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
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000",
	}
	service := NewConstructionService(mockClient, config)

	// Test with valid fee delegator account
	metadata := map[string]any{
		"fee_delegator_account": "0x16277a1ff38678291c41d1820957c78bb5da59ce",
	}

	account := service.getFeeDelegatorAccount(metadata)
	if account != "0x16277a1ff38678291c41d1820957c78bb5da59ce" {
		t.Errorf("getFeeDelegatorAccount() = %v, want 0x16277a1ff38678291c41d1820957c78bb5da59ce", account)
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
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000",
	}
	service := NewConstructionService(mockClient, config)

	// Test with no clauses (base gas only)
	options := map[string]any{}
	gas := service.calculateGas(options)
	expected := int64(20000 * 1.2) // Base gas + 20% buffer
	if gas != expected {
		t.Errorf("calculateGas() with no clauses = %d, want %d", gas, expected)
	}

	// Test with VTHO contract clause
	options = map[string]any{
		"clauses": []any{
			map[string]any{
				"to": meshutils.VTHOCurrency.Metadata["contractAddress"].(string),
			},
		},
	}
	gas = service.calculateGas(options)
	expected = int64((20000 + 50000) * 1.2) // Base + VTHO gas + 20% buffer
	if gas != expected {
		t.Errorf("calculateGas() with VTHO clause = %d, want %d", gas, expected)
	}

	// Test with regular contract clause
	options = map[string]any{
		"clauses": []any{
			map[string]any{
				"to": "0x1234567890123456789012345678901234567890",
			},
		},
	}
	gas = service.calculateGas(options)
	expected = int64((20000 + 10000) * 1.2) // Base + regular gas + 20% buffer
	if gas != expected {
		t.Errorf("calculateGas() with regular clause = %d, want %d", gas, expected)
	}

	// Test with multiple clauses
	options = map[string]any{
		"clauses": []any{
			map[string]any{
				"to": meshutils.VTHOCurrency.Metadata["contractAddress"].(string),
			},
			map[string]any{
				"to": "0x1234567890123456789012345678901234567890",
			},
		},
	}
	gas = service.calculateGas(options)
	expected = int64((20000 + 50000 + 10000) * 1.2) // Base + VTHO + regular + 20% buffer
	if gas != expected {
		t.Errorf("calculateGas() with multiple clauses = %d, want %d", gas, expected)
	}

	// Test with invalid clauses format
	options = map[string]any{
		"clauses": "invalid_format",
	}
	gas = service.calculateGas(options)
	expected = int64(20000 * 1.2) // Base gas only + 20% buffer
	if gas != expected {
		t.Errorf("calculateGas() with invalid clauses = %d, want %d", gas, expected)
	}

	// Test with clause missing 'to' field
	options = map[string]any{
		"clauses": []any{
			map[string]any{
				"value": "1000000000000000000",
			},
		},
	}
	gas = service.calculateGas(options)
	expected = int64(20000 * 1.2) // Base gas only + 20% buffer
	if gas != expected {
		t.Errorf("calculateGas() with clause missing 'to' = %d, want %d", gas, expected)
	}
}
