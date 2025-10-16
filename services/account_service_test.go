package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/math"
	meshcommon "github.com/vechain/mesh/common"
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

// Request body validation is handled by SDK asserter, not by individual service tests

func TestAccountService_AccountBalance_ValidRequest(t *testing.T) {
	tests := []struct {
		name       string
		currencies []*types.Currency
		wantError  bool
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
					t.Error("AccountBalance() expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("AccountBalance() error = %v", err)
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

func TestAccountService_AccountBalance_WithCurrenciesAndBlocks(t *testing.T) {
	tests := []struct {
		name            string
		currencies      []*types.Currency
		blockIdentifier *types.PartialBlockIdentifier
	}{
		{
			name: "VTHO currency",
			currencies: []*types.Currency{
				{Symbol: "VTHO", Decimals: 18},
			},
			blockIdentifier: nil,
		},
		{
			name: "both VET and VTHO currencies",
			currencies: []*types.Currency{
				{Symbol: "VET", Decimals: 18},
				{Symbol: "VTHO", Decimals: 18},
			},
			blockIdentifier: nil,
		},
		{
			name:       "with block index",
			currencies: nil,
			blockIdentifier: &types.PartialBlockIdentifier{
				Index: func() *int64 { i := int64(100); return &i }(),
			},
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
				Currencies:      tt.currencies,
				BlockIdentifier: tt.blockIdentifier,
			}

			ctx := context.Background()
			response, err := service.AccountBalance(ctx, request)

			if err != nil {
				t.Fatalf("AccountBalance() error = %v", err)
			}

			if response.Balances == nil {
				t.Errorf("AccountBalance() response.Balances is nil")
			}
		})
	}
}

func TestAccountService_AccountBalance_WithVIP180Token_Success(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewAccountService(mockClient)

	// Set up mock to return a valid balance
	mockClient.SetMockCallResult("0x0000000000000000000000000000000000000000000000000de0b6b3a7640000") // 1000000000000000000 (1 token)

	// Create request with VIP180 token currency
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
				Symbol:   "TOKEN",
				Decimals: 18,
				Metadata: map[string]any{
					"contractAddress": "0x1234567890123456789012345678901234567890",
				},
			},
		},
	}

	ctx := context.Background()
	response, err := service.AccountBalance(ctx, request)

	if err != nil {
		t.Fatalf("AccountBalance() error = %v", err)
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

func TestAccountService_AccountBalance_WithBlockIdentifiers(t *testing.T) {
	tests := []struct {
		name            string
		blockIdentifier *types.PartialBlockIdentifier
		wantError       bool
	}{
		{
			name: "with block hash",
			blockIdentifier: &types.PartialBlockIdentifier{
				Hash: func() *string { s := "0x00003abbf8435573e0c50fed42647160eabbe140a87efbe0ffab8ef895b7686e"; return &s }(),
			},
			wantError: false,
		},
		{
			name:            "with empty block identifier - should error",
			blockIdentifier: &types.PartialBlockIdentifier{
				// No Hash and no Index - should trigger error
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
				BlockIdentifier: tt.blockIdentifier,
			}

			ctx := context.Background()
			response, err := service.AccountBalance(ctx, request)

			if tt.wantError {
				if err == nil {
					t.Error("AccountBalance() expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("AccountBalance() error = %v", err)
				}

				if response.Balances == nil {
					t.Errorf("AccountBalance() response.Balances is nil")
				}
			}
		})
	}
}

func TestAccountService_AccountBalance_ErrorCases(t *testing.T) {
	tests := []struct {
		name            string
		setupMock       func(*meshthor.MockVeChainClient)
		currencies      []*types.Currency
		blockIdentifier *types.PartialBlockIdentifier
	}{
		{
			name: "get balance error",
			setupMock: func(m *meshthor.MockVeChainClient) {
				// Set up mock block to succeed, but account to fail
				m.SetMockAccountError(fmt.Errorf("failed to get account from node"))
			},
			currencies: []*types.Currency{
				{Symbol: "VET", Decimals: 18},
			},
			blockIdentifier: nil,
		},
		{
			name: "get block error",
			setupMock: func(m *meshthor.MockVeChainClient) {
				// Set up mock account first (so GetAccount succeeds)
				balance := math.HexOrDecimal256{}
				_ = balance.UnmarshalText([]byte("1000000000000000000"))
				energy := math.HexOrDecimal256{}
				_ = energy.UnmarshalText([]byte("500000000000000000"))
				mockAccount := &api.Account{
					Balance: &balance,
					Energy:  &energy,
				}
				m.SetMockAccount(mockAccount)
				// Set up mock to return error when getting block
				m.SetMockBlockError(fmt.Errorf("block not found"))
			},
			currencies: nil,
			blockIdentifier: &types.PartialBlockIdentifier{
				Index: func() *int64 { i := int64(999999); return &i }(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := meshthor.NewMockVeChainClient()
			tt.setupMock(mockClient)
			service := NewAccountService(mockClient)

			request := &types.AccountBalanceRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				AccountIdentifier: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Currencies:      tt.currencies,
				BlockIdentifier: tt.blockIdentifier,
			}

			ctx := context.Background()
			_, err := service.AccountBalance(ctx, request)

			if err == nil {
				t.Error("AccountBalance() expected error but got none")
			}
		})
	}
}

func TestAccountService_AccountBalance_InvalidContractAddressType(t *testing.T) {
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
		Currencies: []*types.Currency{
			{
				Symbol:   "TOKEN",
				Decimals: 18,
				Metadata: map[string]any{
					"contractAddress": 123456,
				},
			},
		},
	}

	ctx := context.Background()
	_, err := service.AccountBalance(ctx, request)

	if err == nil {
		t.Error("AccountBalance() expected error for invalid contractAddress type")
	}
}
