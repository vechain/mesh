package validation

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	meshcommon "github.com/vechain/mesh/common"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshtests "github.com/vechain/mesh/tests"
)

func TestNewValidationMiddleware(t *testing.T) {
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: meshcommon.BlockchainName,
		Network:    "test",
	}
	runMode := meshcommon.OnlineMode

	middleware := NewValidationMiddleware(networkIdentifier, runMode)

	if middleware == nil {
		t.Errorf("NewValidationMiddleware() returned nil")
	} else {
		if middleware.networkIdentifier != networkIdentifier {
			t.Errorf("NewValidationMiddleware() networkIdentifier = %v, want %v", middleware.networkIdentifier, networkIdentifier)
		}
		if middleware.runMode != runMode {
			t.Errorf("NewValidationMiddleware() runMode = %v, want %v", middleware.runMode, runMode)
		}
	}
}

func TestValidationMiddleware_CheckNetwork(t *testing.T) {
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: meshcommon.BlockchainName,
		Network:    "test",
		SubNetworkIdentifier: &types.SubNetworkIdentifier{
			Network: "subnet1",
		},
	}
	runMode := meshcommon.OnlineMode
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
					Blockchain: meshcommon.BlockchainName,
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
					Blockchain: meshcommon.BlockchainName,
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
		{
			name: "valid sub network identifier",
			request: types.NetworkRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
					SubNetworkIdentifier: &types.SubNetworkIdentifier{
						Network: "subnet1",
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid sub network identifier",
			request: types.NetworkRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
					SubNetworkIdentifier: &types.SubNetworkIdentifier{
						Network: "invalid-subnet",
					},
				},
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

func TestValidationMiddleware_CheckNetworkOfflineMode(t *testing.T) {
	// Test CheckNetwork behavior in offline mode
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: meshcommon.BlockchainName,
		Network:    "test",
	}
	runMode := meshcommon.OfflineMode
	middleware := NewValidationMiddleware(networkIdentifier, runMode)

	tests := []struct {
		name        string
		request     types.NetworkRequest
		expectError bool
	}{
		{
			name: "offline mode - matching network",
			request: types.NetworkRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
			},
			expectError: false,
		},
		{
			name: "offline mode - different network but same blockchain (should pass)",
			request: types.NetworkRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "solo", // Different network but should pass in offline mode
				},
			},
			expectError: false,
		},
		{
			name: "offline mode - another different network (should pass)",
			request: types.NetworkRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "main", // Different network but should pass in offline mode
				},
			},
			expectError: false,
		},
		{
			name: "offline mode - invalid blockchain (should fail)",
			request: types.NetworkRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: "ethereum",
					Network:    "test",
				},
			},
			expectError: true,
		},
		{
			name: "offline mode - nil network identifier (should fail)",
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
				t.Errorf("CheckNetwork() expected success but got error: %s", w.Body.String())
			}
		})
	}
}

func TestValidationMiddleware_CheckRunMode(t *testing.T) {
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: meshcommon.BlockchainName,
		Network:    "test",
	}

	tests := []struct {
		name        string
		runMode     string
		expectError bool
	}{
		{
			name:        "valid online mode",
			runMode:     meshcommon.OnlineMode,
			expectError: false,
		},
		{
			name:        "invalid offline mode",
			runMode:     meshcommon.OfflineMode,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewValidationMiddleware(networkIdentifier, tt.runMode)
			req := httptest.NewRequest("POST", "/test", nil)
			w := httptest.NewRecorder()

			result := middleware.CheckRunMode(w, req)
			if tt.expectError && result {
				t.Errorf("CheckRunMode() expected error but got success")
			}
			if !tt.expectError && !result {
				t.Errorf("CheckRunMode() expected success but got error")
			}
		})
	}
}

func TestValidationMiddleware_CheckModeNetwork(t *testing.T) {
	tests := []struct {
		name        string
		network     string
		runMode     string
		expectError bool
	}{
		{
			name:        "valid test network",
			network:     "test",
			runMode:     meshcommon.OnlineMode,
			expectError: false,
		},
		{
			name:        "valid main network",
			network:     "main",
			runMode:     meshcommon.OnlineMode,
			expectError: false,
		},
		{
			name:        "valid solo network",
			network:     "solo",
			runMode:     meshcommon.OnlineMode,
			expectError: false,
		},
		{
			name:        "invalid network",
			network:     "invalid",
			runMode:     meshcommon.OnlineMode,
			expectError: true,
		},
		{
			name:        "invalid network with offline mode",
			network:     "invalid",
			runMode:     meshcommon.OfflineMode,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			networkIdentifier := &types.NetworkIdentifier{
				Blockchain: meshcommon.BlockchainName,
				Network:    tt.network,
			}
			middleware := NewValidationMiddleware(networkIdentifier, tt.runMode)
			req := httptest.NewRequest("POST", "/test", nil)
			w := httptest.NewRecorder()

			result := middleware.CheckModeNetwork(w, req)
			if tt.expectError && result {
				t.Errorf("CheckModeNetwork() expected error but got success")
			}
			if !tt.expectError && !result {
				t.Errorf("CheckModeNetwork() expected success but got error")
			}
		})
	}
}

func TestValidationMiddleware_CheckAccount(t *testing.T) {
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: meshcommon.BlockchainName,
		Network:    "test",
	}
	runMode := meshcommon.OnlineMode
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
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				AccountIdentifier: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
			},
			expectError: false,
		},
		{
			name: "invalid account address",
			request: types.AccountBalanceRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				AccountIdentifier: &types.AccountIdentifier{
					Address: "invalid-address",
				},
			},
			expectError: true,
		},
		{
			name: "address too short",
			request: types.AccountBalanceRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				AccountIdentifier: &types.AccountIdentifier{
					Address: "0x123",
				},
			},
			expectError: true,
		},
		{
			name: "address without 0x prefix",
			request: types.AccountBalanceRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				AccountIdentifier: &types.AccountIdentifier{
					Address: "f077b491b355e64048ce21e3a6fc4751eeea77fa",
				},
			},
			expectError: true,
		},
		{
			name: "address with invalid characters",
			request: types.AccountBalanceRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				AccountIdentifier: &types.AccountIdentifier{
					Address: "0xf077b491b355e64048ce21e3a6fc4751eeea77fg",
				},
			},
			expectError: true,
		},
		{
			name: "nil account identifier",
			request: types.AccountBalanceRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
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
		Blockchain: meshcommon.BlockchainName,
		Network:    "test",
	}
	runMode := meshcommon.OnlineMode
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
					Blockchain: meshcommon.BlockchainName,
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
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				Operations: nil,
			},
			expectError: true,
		},
		{
			name: "no public keys",
			request: types.ConstructionPayloadsRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				Operations: []*types.Operation{
					{
						OperationIdentifier: &types.OperationIdentifier{Index: 0},
						Type:                "Transfer",
						Account: &types.AccountIdentifier{
							Address: meshtests.FirstSoloAddress,
						},
						Amount: &types.Amount{
							Value:    "-1000000000000000000",
							Currency: &types.Currency{Symbol: "VET", Decimals: 18},
						},
					},
				},
				PublicKeys: []*types.PublicKey{},
			},
			expectError: true,
		},
		{
			name: "too many public keys",
			request: types.ConstructionPayloadsRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				Operations: []*types.Operation{
					{
						OperationIdentifier: &types.OperationIdentifier{Index: 0},
						Type:                "Transfer",
						Account: &types.AccountIdentifier{
							Address: meshtests.FirstSoloAddress,
						},
						Amount: &types.Amount{
							Value:    "-1000000000000000000",
							Currency: &types.Currency{Symbol: "VET", Decimals: 18},
						},
					},
				},
				PublicKeys: []*types.PublicKey{
					{Bytes: []byte{1, 2, 3}, CurveType: "secp256k1"},
					{Bytes: []byte{4, 5, 6}, CurveType: "secp256k1"},
					{Bytes: []byte{7, 8, 9}, CurveType: "secp256k1"},
				},
			},
			expectError: true,
		},
		{
			name: "no metadata",
			request: types.ConstructionPayloadsRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				Operations: []*types.Operation{
					{
						OperationIdentifier: &types.OperationIdentifier{Index: 0},
						Type:                "Transfer",
						Account: &types.AccountIdentifier{
							Address: meshtests.FirstSoloAddress,
						},
						Amount: &types.Amount{
							Value:    "-1000000000000000000",
							Currency: &types.Currency{Symbol: "VET", Decimals: 18},
						},
					},
				},
				PublicKeys: []*types.PublicKey{
					{Bytes: []byte{1, 2, 3}, CurveType: "secp256k1"},
				},
				Metadata: nil,
			},
			expectError: true,
		},
		{
			name: "fee delegation with wrong number of public keys",
			request: types.ConstructionPayloadsRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				Operations: []*types.Operation{
					{
						OperationIdentifier: &types.OperationIdentifier{Index: 0},
						Type:                "Transfer",
						Account: &types.AccountIdentifier{
							Address: meshtests.FirstSoloAddress,
						},
						Amount: &types.Amount{
							Value:    "-1000000000000000000",
							Currency: &types.Currency{Symbol: "VET", Decimals: 18},
						},
					},
				},
				PublicKeys: []*types.PublicKey{
					{Bytes: []byte{1, 2, 3}, CurveType: "secp256k1"},
				},
				Metadata: map[string]any{
					"fee_delegator_account": "0x1234567890123456789012345678901234567890",
				},
			},
			expectError: true,
		},
		{
			name: "no origin in operations",
			request: types.ConstructionPayloadsRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				Operations: []*types.Operation{
					{
						OperationIdentifier: &types.OperationIdentifier{Index: 0},
						Type:                "Transfer",
						Account: &types.AccountIdentifier{
							Address: meshtests.FirstSoloAddress,
						},
						Amount: &types.Amount{
							Value:    "1000000000000000000",
							Currency: &types.Currency{Symbol: "VET", Decimals: 18},
						},
					},
				},
				PublicKeys: []*types.PublicKey{
					{Bytes: []byte{1, 2, 3}, CurveType: "secp256k1"},
				},
				Metadata: map[string]any{
					"transactionType": meshcommon.TransactionTypeLegacy,
				},
			},
			expectError: true,
		},
		{
			name: "multiple origins in operations",
			request: types.ConstructionPayloadsRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				Operations: []*types.Operation{
					{
						OperationIdentifier: &types.OperationIdentifier{Index: 0},
						Type:                "Transfer",
						Account: &types.AccountIdentifier{
							Address: meshtests.FirstSoloAddress,
						},
						Amount: &types.Amount{
							Value:    "-1000000000000000000",
							Currency: &types.Currency{Symbol: "VET", Decimals: 18},
						},
					},
					{
						OperationIdentifier: &types.OperationIdentifier{Index: 1},
						Type:                "Transfer",
						Account: &types.AccountIdentifier{
							Address: "0x1234567890123456789012345678901234567890",
						},
						Amount: &types.Amount{
							Value:    "-500000000000000000",
							Currency: &types.Currency{Symbol: "VET", Decimals: 18},
						},
					},
				},
				PublicKeys: []*types.PublicKey{
					{Bytes: []byte{1, 2, 3}, CurveType: "secp256k1"},
				},
				Metadata: map[string]any{
					"transactionType": meshcommon.TransactionTypeLegacy,
				},
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
		Blockchain: meshcommon.BlockchainName,
		Network:    "test",
	}
	runMode := meshcommon.OnlineMode
	middleware := NewValidationMiddleware(networkIdentifier, runMode)

	tests := []struct {
		name        string
		request     types.NetworkRequest
		validations []ValidationType
		expectError bool
	}{
		{
			name: "valid network request",
			request: types.NetworkRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
			},
			validations: NetworkValidations,
			expectError: false,
		},
		{
			name: "invalid network request",
			request: types.NetworkRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: "ethereum",
					Network:    "test",
				},
			},
			validations: NetworkValidations,
			expectError: true,
		},
		{
			name: "invalid run mode",
			request: types.NetworkRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
			},
			validations: []ValidationType{ValidationRunMode},
			expectError: false, // This will pass because we're using meshcommon.OnlineMode mode
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", meshcommon.NetworkStatusEndpoint, bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			result := middleware.ValidateRequest(w, req, requestBody, tt.validations)
			if tt.expectError && result {
				t.Errorf("ValidateRequest() expected error but got success")
			}
			if !tt.expectError && !result {
				t.Errorf("ValidateRequest() expected success but got error")
			}
		})
	}
}

func TestValidationMiddleware_ValidateEndpoint(t *testing.T) {
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: meshcommon.BlockchainName,
		Network:    "test",
	}
	runMode := meshcommon.OnlineMode
	middleware := NewValidationMiddleware(networkIdentifier, runMode)

	// Test with a valid request
	request := types.NetworkRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", meshcommon.NetworkStatusEndpoint, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	result := middleware.ValidateEndpoint(w, req, requestBody, meshcommon.NetworkStatusEndpoint)
	if !result {
		t.Errorf("ValidateEndpoint() expected success but got error")
	}
}
