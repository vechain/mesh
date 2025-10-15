package services

import (
	"context"
	"testing"

	meshcommon "github.com/vechain/mesh/common"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshtests "github.com/vechain/mesh/tests"
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

func TestAccountService_AccountBalance_ValidRequest(t *testing.T) {
	tests := []struct {
		name         string
		currencies   []*types.Currency
		wantError    bool
		errorMessage string
	}{
		{
			name:       "no specific currencies",
			currencies: nil,
			wantError:  false,
		},
		{
			name: "VET currency",
			currencies: []*types.Currency{
				{Symbol: "VET", Decimals: 18},
			},
			wantError: false,
		},
		{
			name: "invalid currency - empty symbol",
			currencies: []*types.Currency{
				{Symbol: "", Decimals: 18},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := meshthor.NewMockVeChainClient()
			service := NewAccountService(mockClient)

			request := &types.AccountBalanceRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				AccountIdentifier: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Currencies: tt.currencies,
			}

			ctx := context.Background()
			response, err := service.AccountBalance(ctx, request)

			if tt.wantError {
				if err == nil {
					t.Errorf("AccountBalance() expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("AccountBalance() returned error: %v", err)
				}

				if response.Balances == nil {
					t.Errorf("AccountBalance() response.Balances is nil")
				}
			}
		})
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
						"contractAddress": "invalid",
					},
				},
			},
			wantError: true,
		},
		{
			name: "valid contract address format",
			currencies: []*types.Currency{
				{
					Symbol:   "TOKEN",
					Decimals: 18,
					Metadata: map[string]any{
						"contractAddress": "0x0000000000000000000000000000456E65726779",
					},
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateCurrencies(tt.currencies)
			if (err != nil) != tt.wantError {
				t.Errorf("validateCurrencies() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAccountService_getCurrenciesToQuery(t *testing.T) {
	service := &AccountService{}

	tests := []struct {
		name               string
		requestCurrencies  []*types.Currency
		expectedCurrencies int
	}{
		{
			name:               "no currencies requested - should return default VET and VTHO",
			requestCurrencies:  nil,
			expectedCurrencies: 2, // VET and VTHO
		},
		{
			name:               "empty currencies array - should return default VET and VTHO",
			requestCurrencies:  []*types.Currency{},
			expectedCurrencies: 2, // VET and VTHO
		},
		{
			name: "specific currency requested - should return that currency",
			requestCurrencies: []*types.Currency{
				{Symbol: "VET", Decimals: 18},
			},
			expectedCurrencies: 1,
		},
		{
			name: "multiple currencies requested - should return those currencies",
			requestCurrencies: []*types.Currency{
				{Symbol: "VET", Decimals: 18},
				{Symbol: "VTHO", Decimals: 18},
			},
			expectedCurrencies: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getCurrenciesToQuery(tt.requestCurrencies)
			if len(result) != tt.expectedCurrencies {
				t.Errorf("getCurrenciesToQuery() returned %v currencies, want %v", len(result), tt.expectedCurrencies)
			}
		})
	}
}

func TestAccountService_AccountBalance_WithBlockIdentifier(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	tests := []struct {
		name            string
		blockIdentifier *types.PartialBlockIdentifier
		wantError       bool
	}{
		{
			name:            "nil block identifier - uses best block",
			blockIdentifier: nil,
			wantError:       false,
		},
		{
			name: "block by hash",
			blockIdentifier: &types.PartialBlockIdentifier{
				Hash: stringPtr("0x00000001c458949db492fb211c05c4f05f770648fc58db33d05c9a94cb3ece8e"),
			},
			wantError: false,
		},
		{
			name: "block by index",
			blockIdentifier: &types.PartialBlockIdentifier{
				Index: int64Ptr(1),
			},
			wantError: false,
		},
		{
			name:            "no hash and no index - should fail",
			blockIdentifier: &types.PartialBlockIdentifier{},
			wantError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &types.AccountBalanceRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				AccountIdentifier: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				BlockIdentifier: tt.blockIdentifier,
			}

			ctx := context.Background()
			response, err := service.AccountBalance(ctx, request)

			if tt.wantError {
				if err == nil {
					t.Errorf("AccountBalance() expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("AccountBalance() returned error: %v", err)
				}
				if response == nil {
					t.Errorf("AccountBalance() returned nil response")
				}
			}
		})
	}
}

func TestAccountService_AccountBalance_VIP180Token(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	// Request balance for VTHO (which is a VIP180 token)
	request := &types.AccountBalanceRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: meshtests.FirstSoloAddress,
		},
		Currencies: []*types.Currency{
			{
				Symbol:   "VTHO",
				Decimals: 18,
			},
		},
	}

	ctx := context.Background()
	response, err := service.AccountBalance(ctx, request)

	if err != nil {
		t.Fatalf("AccountBalance() returned error: %v", err)
	}

	if response == nil {
		t.Fatal("AccountBalance() returned nil response")
	}

	if len(response.Balances) == 0 {
		t.Error("AccountBalance() returned no balances")
	}

	// Check that balance has VTHO currency
	foundVTHO := false
	for _, balance := range response.Balances {
		if balance.Currency.Symbol == "VTHO" {
			foundVTHO = true
			break
		}
	}

	if !foundVTHO {
		t.Error("AccountBalance() did not return VTHO balance")
	}
}

func TestAccountService_AccountCoins(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	request := &types.AccountCoinsRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "test",
		},
		AccountIdentifier: &types.AccountIdentifier{
			Address: meshtests.FirstSoloAddress,
		},
	}

	ctx := context.Background()
	response, err := service.AccountCoins(ctx, request)

	// Should return error as VeChain doesn't support UTXO model
	if err == nil {
		t.Error("AccountCoins() expected error for unsupported operation")
	}

	if response != nil {
		t.Error("AccountCoins() should return nil response for unsupported operation")
	}
}
