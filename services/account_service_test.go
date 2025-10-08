package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	meshcommon "github.com/vechain/mesh/common"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/math"
	meshtests "github.com/vechain/mesh/tests"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/api"
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
	req := httptest.NewRequest("POST", meshcommon.AccountBalanceEndpoint, bytes.NewBufferString("invalid json"))
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
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: meshtests.FirstSoloAddress,
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.AccountBalanceEndpoint, request)
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
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: meshtests.FirstSoloAddress,
		},
		Currencies: []*types.Currency{
			{Symbol: "VET", Decimals: 18},
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.AccountBalanceEndpoint, request)
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
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: meshtests.FirstSoloAddress,
		},
		Currencies: []*types.Currency{
			{Symbol: "", Decimals: 18}, // Empty symbol should fail validation
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.AccountBalanceEndpoint, request)
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
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: meshtests.FirstSoloAddress,
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: func() *int64 { i := int64(100); return &i }(),
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.AccountBalanceEndpoint, request)
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
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: meshtests.FirstSoloAddress,
		},
		Currencies: []*types.Currency{
			{Symbol: "VTHO", Decimals: 18},
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.AccountBalanceEndpoint, request)
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
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: meshtests.FirstSoloAddress,
		},
		Currencies: []*types.Currency{
			{Symbol: "VET", Decimals: 18},
			{Symbol: "VTHO", Decimals: 18},
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.AccountBalanceEndpoint, request)
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

func TestAccountService_AccountBalance_WithVIP180Token_Success(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	// Set up mock to return a valid balance
	mockClient.SetMockCallResult("0x0000000000000000000000000000000000000000000000000de0b6b3a7640000") // 1000000000000000000 (1 token)

	// Create request with VIP180 token currency
	request := types.AccountBalanceRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: meshtests.FirstSoloAddress,
		},
		Currencies: []*types.Currency{
			{
				Symbol:   "TOKEN",
				Decimals: 18,
				Metadata: map[string]any{
					"contractAddress": "0x1234567890123456789012345678901234567890",
				},
			},
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.AccountBalanceEndpoint, request)
	w := httptest.NewRecorder()

	service.AccountBalance(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("AccountBalance() status = %v, want %v", w.Code, http.StatusOK)
	}

	var response types.AccountBalanceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("AccountBalance() failed to unmarshal response: %v", err)
	}

	if len(response.Balances) != 1 {
		t.Errorf("AccountBalance() balances count = %v, want 1", len(response.Balances))
	}

	if response.Balances[0].Value != "1000000000000000000" {
		t.Errorf("AccountBalance() value = %v, want 1000000000000000000", response.Balances[0].Value)
	}

	if response.Balances[0].Currency.Symbol != "TOKEN" {
		t.Errorf("AccountBalance() currency symbol = %v, want TOKEN", response.Balances[0].Currency.Symbol)
	}
}

func TestAccountService_getVETBalance(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	t.Run("Successful VET balance retrieval", func(t *testing.T) {
		// Set up mock account with VET balance
		balance := math.HexOrDecimal256{}
		err := balance.UnmarshalText([]byte("1000000000000000000")) // 1 VET
		if err != nil {
			t.Errorf("getVETBalance() error = %v, want nil", err)
		}
		energy := math.HexOrDecimal256{}
		err = energy.UnmarshalText([]byte("1000000"))
		if err != nil {
			t.Errorf("getVETBalance() error = %v, want nil", err)
		}

		mockAccount := &api.Account{
			Balance: &balance,
			Energy:  &energy,
		}
		mockClient.SetMockAccount(mockAccount)

		// Test getting VET balance
		amount, err := service.getVETBalance(meshtests.FirstSoloAddress, "best")
		if err != nil {
			t.Errorf("getVETBalance() error = %v, want nil", err)
		}
		if amount == nil || amount.Currency.Symbol != "VET" {
			t.Errorf("getVETBalance() currency symbol = %v, want VET", amount.Currency.Symbol)
		}
	})

	t.Run("Error case - account not found", func(t *testing.T) {
		// Set up mock error
		mockClient.SetMockError(fmt.Errorf("account not found"))

		// Test error case
		amount, err := service.getVETBalance(meshtests.FirstSoloAddress, "best")
		if err == nil {
			t.Errorf("getVETBalance() should return error when account not found")
		}
		if amount != nil {
			t.Errorf("getVETBalance() should return nil amount when error occurs")
		}
	})
}

func TestAccountService_AccountBalance_WithBlockHash(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	// Create request with block hash
	blockHash := "0x00003abbf8435573e0c50fed42647160eabbe140a87efbe0ffab8ef895b7686e"
	request := types.AccountBalanceRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: meshtests.FirstSoloAddress,
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Hash: &blockHash,
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.AccountBalanceEndpoint, request)
	w := httptest.NewRecorder()

	// Call AccountBalance
	service.AccountBalance(w, req)

	// Should succeed with mock client
	if w.Code != http.StatusOK {
		t.Errorf("AccountBalance() status code = %v, want %v. Body: %s", w.Code, http.StatusOK, w.Body.String())
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

func TestAccountService_AccountBalance_WithEmptyBlockIdentifier(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	// Create request with empty block identifier (no hash, no index)
	request := types.AccountBalanceRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: meshtests.FirstSoloAddress,
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			// No Hash and no Index - should trigger error
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.AccountBalanceEndpoint, request)
	w := httptest.NewRecorder()

	// Call AccountBalance
	service.AccountBalance(w, req)

	// Should fail with bad request
	if w.Code != http.StatusBadRequest {
		t.Errorf("AccountBalance() status code = %v, want %v. Body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}

	// Verify error response contains expected error information
	bodyStr := w.Body.String()
	if bodyStr == "" {
		t.Errorf("Expected error response body, got empty")
	}
}

func TestAccountService_AccountBalance_GetBalanceError(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	// Set up mock to return error when getting balance
	mockClient.SetMockError(fmt.Errorf("failed to get account from node"))

	// Create request
	request := types.AccountBalanceRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: meshtests.FirstSoloAddress,
		},
		Currencies: []*types.Currency{
			{Symbol: "VET", Decimals: 18},
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.AccountBalanceEndpoint, request)
	w := httptest.NewRecorder()

	// Call AccountBalance
	service.AccountBalance(w, req)

	// Should fail with internal server error
	if w.Code != http.StatusInternalServerError {
		t.Errorf("AccountBalance() status code = %v, want %v. Body: %s", w.Code, http.StatusInternalServerError, w.Body.String())
	}

	// Verify error response contains expected error information
	bodyStr := w.Body.String()
	if bodyStr == "" {
		t.Errorf("Expected error response body, got empty")
	}
}

func TestAccountService_AccountBalance_GetBlockError(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	// Set up mock account first (so GetAccount succeeds)
	balance := math.HexOrDecimal256{}
	err := balance.UnmarshalText([]byte("1000000000000000000"))
	if err != nil {
		t.Fatalf("Failed to setup test: %v", err)
	}
	energy := math.HexOrDecimal256{}
	err = energy.UnmarshalText([]byte("500000000000000000"))
	if err != nil {
		t.Fatalf("Failed to setup test: %v", err)
	}
	mockAccount := &api.Account{
		Balance: &balance,
		Energy:  &energy,
	}
	mockClient.SetMockAccount(mockAccount)

	// Set up mock to return error when getting block
	mockClient.SetMockBlockError(fmt.Errorf("block not found"))

	// Create request
	blockIndex := int64(999999)
	request := types.AccountBalanceRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: meshtests.FirstSoloAddress,
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: &blockIndex,
		},
	}

	req := meshtests.CreateRequestWithContext("POST", meshcommon.AccountBalanceEndpoint, request)
	w := httptest.NewRecorder()

	// Call AccountBalance
	service.AccountBalance(w, req)

	// Should fail with bad request
	if w.Code != http.StatusBadRequest {
		t.Errorf("AccountBalance() status code = %v, want %v. Body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}

	// Verify error response contains expected error information
	bodyStr := w.Body.String()
	if bodyStr == "" {
		t.Errorf("Expected error response body, got empty")
	}
}
