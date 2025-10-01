package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshconfig "github.com/vechain/mesh/config"
	meshtests "github.com/vechain/mesh/tests"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/api"
)

func createMockCallService() *CallService {
	return createMockCallServiceWithClient(meshthor.NewMockVeChainClient())
}

func createMockCallServiceWithClient(client *meshthor.MockVeChainClient) *CallService {
	config := &meshconfig.Config{}
	config.Network = "solo"
	return NewCallService(client, config)
}

func createTestCallRequest(method string, params map[string]any) types.CallRequest {
	return types.CallRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "solo",
		},
		Method:     method,
		Parameters: params,
	}
}

func TestCallService_Call_UnsupportedMethod(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest("unsupported_method", map[string]any{
		"test": "value",
	})

	req := meshtests.CreateRequestWithContext("POST", meshcommon.CallEndpoint, request)
	w := httptest.NewRecorder()

	service.Call(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Call() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestCallService_Call_ValidInspectClauses(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	mockClient.SetInspectClausesResult([]*api.CallResult{
		{
			Data:      "0x0000000000000000000000000000000000000000000000000000000000000001",
			Events:    []*api.Event{},
			Transfers: []*api.Transfer{},
			GasUsed:   21000,
			Reverted:  false,
			VMError:   "",
		},
	})

	service := createMockCallServiceWithClient(mockClient)

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    "0x0000000000000000000000000000456E65726779",
				"value": "0x0",
				"data":  "0xa9059cbb0000000000000000000000000f872421dc479f3c11edd89512731814d0598db50000000000000000000000000000000000000000000000013f306a2409fc0000",
			},
		},
		"gas":        float64(50000),
		"gasPrice":   "1000000000000000",
		"caller":     "0x6d95e6dca01d109882fe1726a2fb9865fa41e7aa",
		"expiration": float64(1000),
		"blockRef":   "0x00000000851caf3c",
	})

	req := meshtests.CreateRequestWithContext("POST", meshcommon.CallEndpoint, request)
	w := httptest.NewRecorder()

	service.Call(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Call() status code = %v, want %v. Body: %s", w.Code, http.StatusOK, w.Body.String())
		return
	}

	var response types.CallResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result == nil {
		t.Errorf("Call() result is nil")
		return
	}

	// Verify idempotent flag
	if !response.Idempotent {
		t.Errorf("Call() idempotent = %v, want true", response.Idempotent)
	}

	// Verify results
	results, ok := response.Result["results"].([]any)
	if !ok {
		t.Errorf("Call() result['results'] is not an array")
		return
	}

	if len(results) != 1 {
		t.Errorf("Call() results length = %v, want 1", len(results))
		return
	}

	// Verify first result
	result := results[0].(map[string]any)
	if gasUsed, ok := result["gasUsed"].(float64); !ok || gasUsed != 21000 {
		t.Errorf("Call() result gasUsed = %v, want 21000", gasUsed)
	}

	if reverted, ok := result["reverted"].(bool); !ok || reverted {
		t.Errorf("Call() result reverted = %v, want false", reverted)
	}
}

func TestCallService_Call_InvalidParameters(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		// Missing clauses field
		"gas": float64(50000),
	})

	req := meshtests.CreateRequestWithContext("POST", meshcommon.CallEndpoint, request)
	w := httptest.NewRecorder()

	service.Call(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Call() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestCallService_Call_InvalidClauseFormat(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to": "0x0000000000000000000000000000456E65726779",
				// Missing value and data fields
			},
		},
	})

	req := meshtests.CreateRequestWithContext("POST", meshcommon.CallEndpoint, request)
	w := httptest.NewRecorder()

	service.Call(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Call() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestCallService_Call_WithRevision(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	mockClient.SetInspectClausesResult([]*api.CallResult{
		{
			Data:      "0x0000000000000000000000000000000000000000000000000000000000000002",
			Events:    []*api.Event{},
			Transfers: []*api.Transfer{},
			GasUsed:   25000,
			Reverted:  false,
			VMError:   "",
		},
	})

	service := createMockCallServiceWithClient(mockClient)

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    "0x0000000000000000000000000000456E65726779",
				"value": "0x0",
				"data":  "0xa9059cbb",
			},
		},
		"revision": "finalized",
	})

	req := meshtests.CreateRequestWithContext("POST", meshcommon.CallEndpoint, request)
	w := httptest.NewRecorder()

	service.Call(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Call() status code = %v, want %v. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestCallService_Call_WithNilToAddress(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	mockClient.SetInspectClausesResult([]*api.CallResult{
		{
			Data:      "0x1234",
			Events:    []*api.Event{},
			Transfers: []*api.Transfer{},
			GasUsed:   100000,
			Reverted:  false,
			VMError:   "",
		},
	})

	service := createMockCallServiceWithClient(mockClient)

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    nil, // Contract deployment
				"value": "0x0",
				"data":  "0x6080604052",
			},
		},
	})

	req := meshtests.CreateRequestWithContext("POST", meshcommon.CallEndpoint, request)
	w := httptest.NewRecorder()

	service.Call(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Call() status code = %v, want %v. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestCallService_Call_InvalidRequestBody(t *testing.T) {
	service := createMockCallService()

	req := httptest.NewRequest("POST", meshcommon.CallEndpoint, nil)
	w := httptest.NewRecorder()

	service.Call(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Call() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestCallService_Call_WithOptionalParameters(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	mockClient.SetInspectClausesResult([]*api.CallResult{
		{
			Data:      "0x0000000000000000000000000000000000000000000000000000000000000001",
			Events:    []*api.Event{},
			Transfers: []*api.Transfer{},
			GasUsed:   50000,
			Reverted:  false,
			VMError:   "",
		},
	})

	service := createMockCallServiceWithClient(mockClient)

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    "0x0000000000000000000000000000456E65726779",
				"value": "0x1234",
				"data":  "0xa9059cbb",
			},
		},
		"gas":        float64(100000),
		"gasPrice":   "0x16345785d8a0000",
		"provedWork": "0x3e8",
		"caller":     "0x6d95e6dca01d109882fe1726a2fb9865fa41e7aa",
		"gasPayer":   "0xd3ae78222beadb038203be21ed5ce7c9b1bff602",
		"expiration": float64(1000),
		"blockRef":   "0x00000000851caf3c",
	})

	req := meshtests.CreateRequestWithContext("POST", meshcommon.CallEndpoint, request)
	w := httptest.NewRecorder()

	service.Call(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Call() status code = %v, want %v. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestCallService_Call_InvalidAddress(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    "invalid-address",
				"value": "0x0",
				"data":  "0x",
			},
		},
	})

	req := meshtests.CreateRequestWithContext("POST", meshcommon.CallEndpoint, request)
	w := httptest.NewRecorder()

	service.Call(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Call() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}
