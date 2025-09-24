package validation

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
)

func TestNewValidationMiddleware(t *testing.T) {
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: "vechainthor",
		Network:    "test",
	}
	runMode := "online"

	middleware := NewValidationMiddleware(networkIdentifier, runMode)

	if middleware == nil {
		t.Errorf("NewValidationMiddleware() returned nil")
	}
	if middleware.networkIdentifier != networkIdentifier {
		t.Errorf("NewValidationMiddleware() networkIdentifier = %v, want %v", middleware.networkIdentifier, networkIdentifier)
	}
	if middleware.runMode != runMode {
		t.Errorf("NewValidationMiddleware() runMode = %v, want %v", middleware.runMode, runMode)
	}
}

func TestValidationMiddleware_CheckNetwork(t *testing.T) {
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: "vechainthor",
		Network:    "test",
	}
	runMode := "online"
	middleware := NewValidationMiddleware(networkIdentifier, runMode)

	tests := []struct {
		name        string
		request     types.NetworkRequest
		expectError bool
	}{
		{
			name: "valid network request",
			request: types.NetworkRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: "vechainthor",
					Network:    "test",
				},
			},
			expectError: false,
		},
		{
			name: "invalid blockchain",
			request: types.NetworkRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: "ethereum",
					Network:    "test",
				},
			},
			expectError: true,
		},
		{
			name: "invalid network",
			request: types.NetworkRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: "vechainthor",
					Network:    "mainnet",
				},
			},
			expectError: true,
		},
		{
			name: "nil network identifier",
			request: types.NetworkRequest{
				NetworkIdentifier: nil,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(requestBody))
			w := httptest.NewRecorder()

			result := middleware.CheckNetwork(w, req, requestBody)
			if tt.expectError && result {
				t.Errorf("CheckNetwork() expected error but got success")
			}
			if !tt.expectError && !result {
				t.Errorf("CheckNetwork() expected success but got error")
			}
		})
	}
}

func TestValidationMiddleware_CheckRunMode(t *testing.T) {
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: "vechainthor",
		Network:    "test",
	}
	runMode := "online"
	middleware := NewValidationMiddleware(networkIdentifier, runMode)

	// Test with a valid request
	request := types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(requestBody))
	w := httptest.NewRecorder()

	result := middleware.CheckRunMode(w, req)
	if !result {
		t.Errorf("CheckRunMode() expected success but got error")
	}
}

func TestValidationMiddleware_CheckModeNetwork(t *testing.T) {
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: "vechainthor",
		Network:    "test",
	}
	runMode := "online"
	middleware := NewValidationMiddleware(networkIdentifier, runMode)

	// Test with a valid request
	request := types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(requestBody))
	w := httptest.NewRecorder()

	result := middleware.CheckModeNetwork(w, req)
	if !result {
		t.Errorf("CheckModeNetwork() expected success but got error")
	}
}

func TestValidationMiddleware_CheckAccount(t *testing.T) {
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: "vechainthor",
		Network:    "test",
	}
	runMode := "online"
	middleware := NewValidationMiddleware(networkIdentifier, runMode)

	tests := []struct {
		name        string
		request     types.AccountBalanceRequest
		expectError bool
	}{
		{
			name: "valid account",
			request: types.AccountBalanceRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: "vechainthor",
					Network:    "test",
				},
				AccountIdentifier: &types.AccountIdentifier{
					Address: "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
				},
			},
			expectError: false,
		},
		{
			name: "invalid account address",
			request: types.AccountBalanceRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: "vechainthor",
					Network:    "test",
				},
				AccountIdentifier: &types.AccountIdentifier{
					Address: "invalid-address",
				},
			},
			expectError: true,
		},
		{
			name: "nil account identifier",
			request: types.AccountBalanceRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: "vechainthor",
					Network:    "test",
				},
				AccountIdentifier: nil,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(requestBody))
			w := httptest.NewRecorder()

			result := middleware.CheckAccount(w, req, requestBody)
			if tt.expectError && result {
				t.Errorf("CheckAccount() expected error but got success")
			}
			if !tt.expectError && !result {
				t.Errorf("CheckAccount() expected success but got error")
			}
		})
	}
}

func TestValidationMiddleware_CheckConstructionPayloads(t *testing.T) {
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: "vechainthor",
		Network:    "test",
	}
	runMode := "online"
	middleware := NewValidationMiddleware(networkIdentifier, runMode)

	tests := []struct {
		name        string
		request     types.ConstructionPayloadsRequest
		expectError bool
	}{
		{
			name: "empty operations",
			request: types.ConstructionPayloadsRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: "vechainthor",
					Network:    "test",
				},
				Operations: []*types.Operation{},
			},
			expectError: true,
		},
		{
			name: "nil operations",
			request: types.ConstructionPayloadsRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: "vechainthor",
					Network:    "test",
				},
				Operations: nil,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(requestBody))
			w := httptest.NewRecorder()

			result := middleware.CheckConstructionPayloads(w, req, requestBody)
			if tt.expectError && result {
				t.Errorf("CheckConstructionPayloads() expected error but got success")
			}
			if !tt.expectError && !result {
				t.Errorf("CheckConstructionPayloads() expected success but got error")
			}
		})
	}
}

func TestValidationMiddleware_ValidateRequest(t *testing.T) {
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: "vechainthor",
		Network:    "test",
	}
	runMode := "online"
	middleware := NewValidationMiddleware(networkIdentifier, runMode)

	// Test with a valid request
	request := types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/network/status", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	result := middleware.ValidateRequest(w, req, requestBody, NetworkValidations)
	if !result {
		t.Errorf("ValidateRequest() expected success but got error")
	}
}

func TestValidationMiddleware_ValidateEndpoint(t *testing.T) {
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: "vechainthor",
		Network:    "test",
	}
	runMode := "online"
	middleware := NewValidationMiddleware(networkIdentifier, runMode)

	// Test with a valid request
	request := types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/network/status", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	result := middleware.ValidateEndpoint(w, req, requestBody, "/network/status")
	if !result {
		t.Errorf("ValidateEndpoint() expected success but got error")
	}
}
