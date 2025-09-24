package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshthor "github.com/vechain/mesh/thor"
)

func TestNewAccountService(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()

	service := NewAccountService(mockClient)

	if service == nil {
		t.Fatal("NewAccountService() returned nil")
	}

	if service.vechainClient == nil {
		t.Errorf("NewAccountService() vechainClient is nil")
	}
}

func TestAccountService_AccountBalance_InvalidRequestBody(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

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
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

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

	// Should succeed with mock client
	if w.Code != http.StatusOK {
		t.Errorf("AccountBalance() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify response structure
	var response types.AccountBalanceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Balances == nil {
		t.Errorf("AccountBalance() response.Balances is nil")
	}
}

func TestAccountService_AccountBalance_WithSpecificCurrencies(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

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

	// Should succeed with mock client
	if w.Code != http.StatusOK {
		t.Errorf("AccountBalance() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify response structure
	var response types.AccountBalanceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Balances == nil {
		t.Errorf("AccountBalance() response.Balances is nil")
	}
}

func TestAccountService_AccountBalance_InvalidCurrencies(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

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

func TestAccountService_AccountBalance_WithBlockIdentifier(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	// Create request with block identifier
	request := types.AccountBalanceRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: "0xf077b491b355E64048cE21E3A6Fc4751eEeA77fa",
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: func() *int64 { i := int64(100); return &i }(),
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/account/balance", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call AccountBalance
	service.AccountBalance(w, req)

	// Should succeed with mock client
	if w.Code != http.StatusOK {
		t.Errorf("AccountBalance() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify response structure
	var response types.AccountBalanceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Balances == nil {
		t.Errorf("AccountBalance() response.Balances is nil")
	}
}

func TestAccountService_AccountBalance_WithVTHOCurrency(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	// Create request with VTHO currency
	request := types.AccountBalanceRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: "0xf077b491b355E64048cE21E3A6Fc4751eEeA77fa",
		},
		Currencies: []*types.Currency{
			{Symbol: "VTHO", Decimals: 18},
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/account/balance", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call AccountBalance
	service.AccountBalance(w, req)

	// Should succeed with mock client
	if w.Code != http.StatusOK {
		t.Errorf("AccountBalance() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify response structure
	var response types.AccountBalanceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Balances == nil {
		t.Errorf("AccountBalance() response.Balances is nil")
	}
}

func TestAccountService_AccountBalance_WithBothCurrencies(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	// Create request with both VET and VTHO currencies
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
			{Symbol: "VTHO", Decimals: 18},
		},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/account/balance", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call AccountBalance
	service.AccountBalance(w, req)

	// Should succeed with mock client
	if w.Code != http.StatusOK {
		t.Errorf("AccountBalance() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify response structure
	var response types.AccountBalanceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Balances == nil {
		t.Errorf("AccountBalance() response.Balances is nil")
	}
}

func TestAccountService_getVIP180TokenBalance_Success(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	// Set up mock to return a valid balance
	mockClient.SetMockCallResult("0x0000000000000000000000000000000000000000000000000de0b6b3a7640000") // 1000000000000000000 (1 token)

	currency := &types.Currency{
		Symbol:   "TOKEN",
		Decimals: 18,
		Metadata: map[string]any{
			"contractAddress": "0x1234567890123456789012345678901234567890",
		},
	}

	amount, err := service.getVIP180TokenBalance("0xf077b491b355e64048ce21e3a6fc4751eeea77fa", "0x1234567890123456789012345678901234567890", currency)

	if err != nil {
		t.Errorf("getVIP180TokenBalance() unexpected error: %v", err)
	}

	if amount.Value != "1000000000000000000" {
		t.Errorf("getVIP180TokenBalance() value = %v, want 1000000000000000000", amount.Value)
	}

	if amount.Currency != currency {
		t.Errorf("getVIP180TokenBalance() currency = %v, want %v", amount.Currency, currency)
	}
}

func TestAccountService_getVIP180TokenBalance_ContractCallError(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	// Set up mock to return an error
	mockClient.SetMockError(fmt.Errorf("contract call failed"))

	currency := &types.Currency{
		Symbol:   "TOKEN",
		Decimals: 18,
		Metadata: map[string]any{
			"contractAddress": "0x1234567890123456789012345678901234567890",
		},
	}

	amount, err := service.getVIP180TokenBalance("0xf077b491b355e64048ce21e3a6fc4751eeea77fa", "0x1234567890123456789012345678901234567890", currency)

	if err != nil {
		t.Errorf("getVIP180TokenBalance() unexpected error: %v", err)
	}

	// Should return 0 balance when contract call fails
	if amount.Value != "0" {
		t.Errorf("getVIP180TokenBalance() value = %v, want 0", amount.Value)
	}

	if amount.Currency != currency {
		t.Errorf("getVIP180TokenBalance() currency = %v, want %v", amount.Currency, currency)
	}
}

func TestAccountService_getVIP180TokenBalance_InvalidResult(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	// Set up mock to return invalid result (no 0x prefix)
	mockClient.SetMockCallResult("invalid_result")

	currency := &types.Currency{
		Symbol:   "TOKEN",
		Decimals: 18,
		Metadata: map[string]any{
			"contractAddress": "0x1234567890123456789012345678901234567890",
		},
	}

	amount, err := service.getVIP180TokenBalance("0xf077b491b355e64048ce21e3a6fc4751eeea77fa", "0x1234567890123456789012345678901234567890", currency)

	if err != nil {
		t.Errorf("getVIP180TokenBalance() unexpected error: %v", err)
	}

	// Should return 0 balance when result is invalid
	if amount.Value != "0" {
		t.Errorf("getVIP180TokenBalance() value = %v, want 0", amount.Value)
	}

	if amount.Currency != currency {
		t.Errorf("getVIP180TokenBalance() currency = %v, want %v", amount.Currency, currency)
	}
}

func TestAccountService_getVIP180TokenBalance_EmptyResult(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	// Set up mock to return empty result
	mockClient.SetMockCallResult("")

	currency := &types.Currency{
		Symbol:   "TOKEN",
		Decimals: 18,
		Metadata: map[string]any{
			"contractAddress": "0x1234567890123456789012345678901234567890",
		},
	}

	amount, err := service.getVIP180TokenBalance("0xf077b491b355e64048ce21e3a6fc4751eeea77fa", "0x1234567890123456789012345678901234567890", currency)

	if err != nil {
		t.Errorf("getVIP180TokenBalance() unexpected error: %v", err)
	}

	// Should return 0 balance when result is empty
	if amount.Value != "0" {
		t.Errorf("getVIP180TokenBalance() value = %v, want 0", amount.Value)
	}

	if amount.Currency != currency {
		t.Errorf("getVIP180TokenBalance() currency = %v, want %v", amount.Currency, currency)
	}
}

func TestAccountService_getVIP180TokenBalance_HexConversionError(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	// Set up mock to return invalid hex
	mockClient.SetMockCallResult("0xinvalid_hex")

	currency := &types.Currency{
		Symbol:   "TOKEN",
		Decimals: 18,
		Metadata: map[string]any{
			"contractAddress": "0x1234567890123456789012345678901234567890",
		},
	}

	amount, err := service.getVIP180TokenBalance("0xf077b491b355e64048ce21e3a6fc4751eeea77fa", "0x1234567890123456789012345678901234567890", currency)

	if err != nil {
		t.Errorf("getVIP180TokenBalance() unexpected error: %v", err)
	}

	// Should return 0 balance when hex conversion fails
	if amount.Value != "0" {
		t.Errorf("getVIP180TokenBalance() value = %v, want 0", amount.Value)
	}

	if amount.Currency != currency {
		t.Errorf("getVIP180TokenBalance() currency = %v, want %v", amount.Currency, currency)
	}
}
