package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshthor "github.com/vechain/mesh/thor"
)

func TestNewAccountService(t *testing.T) {
	client := meshthor.NewVeChainClient("http://localhost:8669")

	service := NewAccountService(client)

	if service == nil {
		t.Fatal("NewAccountService() returned nil")
	}

	if service.vechainClient == nil {
		t.Errorf("NewAccountService() vechainClient is nil")
	}
}

func TestAccountService_AccountBalance_InvalidRequestBody(t *testing.T) {
	client := meshthor.NewVeChainClient("http://localhost:8669")
	service := NewAccountService(client)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/account/balance", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call AccountBalance
	service.AccountBalance(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("AccountBalance() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestAccountService_AccountBalance_ValidRequest(t *testing.T) {
	client := meshthor.NewVeChainClient("http://localhost:8669")
	service := NewAccountService(client)

	// Create request
	request := types.AccountBalanceRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: "0xf077b491b355E64048cE21E3A6Fc4751eEeA77fa",
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/account/balance", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call AccountBalance
	service.AccountBalance(w, req)

	// Note: This test will fail if the VeChain node is not running
	// but it tests the request parsing and basic flow
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("AccountBalance() status code = %v, want %v or %v", w.Code, http.StatusOK, http.StatusInternalServerError)
	}
}

func TestAccountService_AccountBalance_WithSpecificCurrencies(t *testing.T) {
	client := meshthor.NewVeChainClient("http://localhost:8669")
	service := NewAccountService(client)

	// Create request with specific currencies
	request := types.AccountBalanceRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: "0xf077b491b355E64048cE21E3A6Fc4751eEeA77fa",
		},
		Currencies: []*types.Currency{
			{Symbol: "VET", Decimals: 18},
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/account/balance", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call AccountBalance
	service.AccountBalance(w, req)

	// Note: This test will fail if the VeChain node is not running
	// but it tests the request parsing and basic flow
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("AccountBalance() status code = %v, want %v or %v", w.Code, http.StatusOK, http.StatusInternalServerError)
	}
}

func TestAccountService_AccountBalance_InvalidCurrencies(t *testing.T) {
	client := meshthor.NewVeChainClient("http://localhost:8669")
	service := NewAccountService(client)

	// Create request with invalid currencies (empty symbol)
	request := types.AccountBalanceRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: "0xf077b491b355E64048cE21E3A6Fc4751eEeA77fa",
		},
		Currencies: []*types.Currency{
			{Symbol: "", Decimals: 18}, // Empty symbol should fail validation
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/account/balance", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call AccountBalance
	service.AccountBalance(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("AccountBalance() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestAccountService_validateCurrencies(t *testing.T) {
	service := &AccountService{}

	tests := []struct {
		name       string
		currencies []*types.Currency
		wantError  bool
	}{
		{
			name: "valid VET currency",
			currencies: []*types.Currency{
				{Symbol: "VET", Decimals: 18},
			},
			wantError: false,
		},
		{
			name: "valid VTHO currency",
			currencies: []*types.Currency{
				{Symbol: "VTHO", Decimals: 18},
			},
			wantError: false,
		},
		{
			name: "valid both currencies",
			currencies: []*types.Currency{
				{Symbol: "VET", Decimals: 18},
				{Symbol: "VTHO", Decimals: 18},
			},
			wantError: false,
		},
		{
			name: "empty currency symbol",
			currencies: []*types.Currency{
				{Symbol: "", Decimals: 18},
			},
			wantError: true,
		},
		{
			name: "negative decimals",
			currencies: []*types.Currency{
				{Symbol: "VET", Decimals: -1},
			},
			wantError: true,
		},
		{
			name: "invalid contract address format",
			currencies: []*types.Currency{
				{
					Symbol:   "TOKEN",
					Decimals: 18,
					Metadata: map[string]any{
						"contractAddress": "invalid_address",
					},
				},
			},
			wantError: true,
		},
		{
			name: "valid contract address",
			currencies: []*types.Currency{
				{
					Symbol:   "TOKEN",
					Decimals: 18,
					Metadata: map[string]any{
						"contractAddress": "0x1234567890123456789012345678901234567890",
					},
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateCurrencies(tt.currencies)

			if tt.wantError {
				if err == nil {
					t.Errorf("validateCurrencies() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("validateCurrencies() unexpected error: %v", err)
				}
			}
		})
	}
}
