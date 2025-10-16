package services

import (
	"context"
	"fmt"
	"math/big"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	"github.com/vechain/mesh/common/vip180"
	meshthor "github.com/vechain/mesh/thor"
)

// AccountService handles account-related endpoints
type AccountService struct {
	vechainClient meshthor.VeChainClientInterface
}

// NewAccountService creates a new account service
func NewAccountService(vechainClient meshthor.VeChainClientInterface) *AccountService {
	return &AccountService{
		vechainClient: vechainClient,
	}
}

// AccountBalance returns the balance of an account
func (a *AccountService) AccountBalance(
	ctx context.Context,
	req *types.AccountBalanceRequest,
) (*types.AccountBalanceResponse, *types.Error) {
	// Validate currencies if provided
	if req.Currencies != nil {
		if err := a.validateCurrencies(req.Currencies); err != nil {
			return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInvalidCurrency, map[string]any{
				"error": err.Error(),
			})
		}
	}

	// Determine revision from request or use "best"
	revision := "best"
	if req.BlockIdentifier != nil {
		if req.BlockIdentifier.Hash != nil && *req.BlockIdentifier.Hash != "" {
			revision = *req.BlockIdentifier.Hash
		} else if req.BlockIdentifier.Index != nil {
			revision = fmt.Sprintf("%d", *req.BlockIdentifier.Index)
		} else {
			return nil, meshcommon.GetError(meshcommon.ErrInvalidBlockIdentifierParameter)
		}
	}

	// Get block information first to ensure atomicity
	block, err := a.vechainClient.GetBlock(revision)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrBlockNotFound, map[string]any{
			"error": err.Error(),
		})
	}
	blockRevision := block.ID.String()

	// Determine which currencies to query
	currenciesToQuery := a.getCurrenciesToQuery(req.Currencies)

	// Get balances for each currency at the specified block
	var balances []*types.Amount
	for _, currency := range currenciesToQuery {
		balance, err := a.getBalanceForCurrency(req.AccountIdentifier.Address, currency, blockRevision)
		if err != nil {
			return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrFailedToGetAccount, map[string]any{
				"error":    err.Error(),
				"currency": currency.Symbol,
			})
		}
		if balance != nil {
			balances = append(balances, balance)
		}
	}

	return &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(block.Number),
			Hash:  block.ID.String(),
		},
		Balances: balances,
	}, nil
}

// AccountCoins is not implemented for VeChain (account-based model, not UTXO)
func (a *AccountService) AccountCoins(
	ctx context.Context,
	req *types.AccountCoinsRequest,
) (*types.AccountCoinsResponse, *types.Error) {
	// VeChain uses an account-based model, not UTXO, so AccountCoins is not applicable
	return nil, &types.Error{
		Code:      500,
		Message:   "AccountCoins is not supported for VeChain (account-based blockchain)",
		Retriable: false,
	}
}

// validateCurrencies validates the currencies array format
func (a *AccountService) validateCurrencies(currencies []*types.Currency) error {
	for _, currency := range currencies {
		if currency.Symbol == "" {
			return fmt.Errorf("currency symbol is required")
		}
		if currency.Decimals < 0 {
			return fmt.Errorf("currency decimals must be non-negative")
		}
		// Validate contract address if present (for VIP180 tokens)
		if contractAddr, exists := currency.Metadata["contractAddress"]; exists {
			if addr, ok := contractAddr.(string); ok {
				if len(addr) != 42 || (addr[:2] != "0x" && addr[:2] != "-0x") {
					return fmt.Errorf("invalid contract address format: %s", addr)
				}
			} else {
				return fmt.Errorf("invalid contract address format: %v", contractAddr)
			}
		}
	}
	return nil
}

// getCurrenciesToQuery determines which currencies to query based on request
func (a *AccountService) getCurrenciesToQuery(requestCurrencies []*types.Currency) []*types.Currency {
	if len(requestCurrencies) == 0 {
		// Default: return VET and VTHO
		return []*types.Currency{
			meshcommon.VETCurrency,
			meshcommon.VTHOCurrency,
		}
	}
	return requestCurrencies
}

// getBalanceForCurrency gets the balance for a specific currency at a specific block revision
func (a *AccountService) getBalanceForCurrency(address string, currency *types.Currency, revision string) (*types.Amount, error) {
	// Handle VET currency
	if currency.Symbol == meshcommon.VETCurrency.Symbol {
		return a.getVETBalance(address, revision)
	}

	// Handle VTHO currency
	if currency.Symbol == meshcommon.VTHOCurrency.Symbol {
		return a.getVTHOBalance(address, revision)
	}

	// Handle VIP180 tokens
	if contractAddr, exists := currency.Metadata["contractAddress"]; exists {
		if addr, ok := contractAddr.(string); ok {
			// Create VIP180 contract wrapper
			contract, err := vip180.NewVIP180Contract(addr, a.vechainClient)
			if err != nil {
				// If contract creation fails, return 0 balance
				return &types.Amount{
					Value:    "0",
					Currency: currency,
				}, nil
			}

			balance, err := contract.BalanceOf(address)
			if err != nil {
				return nil, err
			}

			return &types.Amount{
				Value:    balance.String(),
				Currency: currency,
			}, nil
		}
	}

	return nil, fmt.Errorf("unsupported currency: %s", currency.Symbol)
}

// getVETBalance gets the VET balance for an account at a specific block revision
func (a *AccountService) getVETBalance(address string, revision string) (*types.Amount, error) {
	account, err := a.vechainClient.GetAccountAtRevision(address, revision)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	vetBalance := (*big.Int)(account.Balance)

	return &types.Amount{
		Value:    vetBalance.String(),
		Currency: meshcommon.VETCurrency,
	}, nil
}

// getVTHOBalance gets the VTHO balance for an account at a specific block revision
func (a *AccountService) getVTHOBalance(address string, revision string) (*types.Amount, error) {
	account, err := a.vechainClient.GetAccountAtRevision(address, revision)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	vthoBalance := (*big.Int)(account.Energy)

	return &types.Amount{
		Value:    vthoBalance.String(),
		Currency: meshcommon.VTHOCurrency,
	}, nil
}
