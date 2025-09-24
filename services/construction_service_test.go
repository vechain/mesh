package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshconfig "github.com/vechain/mesh/config"
	meshthor "github.com/vechain/mesh/thor"
	meshutils "github.com/vechain/mesh/utils"
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
	}
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
		Metadata: map[string]interface{}{
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

	// Create valid request
	request := types.ConstructionMetadataRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		Options: map[string]interface{}{
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
		Metadata: map[string]interface{}{
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

func TestConstructionService_ConstructionCombine_InvalidRequestBody(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	config := &meshconfig.Config{
		NodeAPI:      "http://localhost:8669",
		Network:      "test",
		Mode:         "online",
		BaseGasPrice: "1000000000000000000", // 1 VTHO
	}
	service := NewConstructionService(mockClient, config)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/construction/combine", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call ConstructionCombine
	service.ConstructionCombine(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionCombine() status code = %v, want %v", w.Code, http.StatusBadRequest)
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
