package services

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshhttp "github.com/vechain/mesh/common/http"
	"github.com/vechain/mesh/common/vip180"
	meshthor "github.com/vechain/mesh/thor"
)

// AccountService handles account-related endpoints
type AccountService struct {
	requestHandler  *meshhttp.RequestHandler
	responseHandler *meshhttp.ResponseHandler
	vechainClient   meshthor.VeChainClientInterface
}

// NewAccountService creates a new account service
func NewAccountService(vechainClient meshthor.VeChainClientInterface) *AccountService {
	return &AccountService{
		requestHandler:  meshhttp.NewRequestHandler(),
		responseHandler: meshhttp.NewResponseHandler(),
		vechainClient:   vechainClient,
	}
}

// AccountBalance returns the balance of an account
func (a *AccountService) AccountBalance(w http.ResponseWriter, r *http.Request) {
	var request types.AccountBalanceRequest
	err := a.requestHandler.ParseJSONFromContext(r, &request)
	if err != nil {
		a.responseHandler.WriteErrorResponse(w, meshcommon.GetError(meshcommon.ErrInvalidRequestBody), http.StatusBadRequest)
		return
	}

	// Validate currencies if provided
	if request.Currencies != nil {
		if err := a.validateCurrencies(request.Currencies); err != nil {
			a.responseHandler.WriteErrorResponse(w, meshcommon.GetErrorWithMetadata(meshcommon.ErrInvalidCurrency, map[string]any{
				"error": err.Error(),
			}), http.StatusBadRequest)
			return
		}
	}

	// Determine revision from request or use "best"
	revision := "best"
	if request.BlockIdentifier != nil {
		if request.BlockIdentifier.Hash != nil && *request.BlockIdentifier.Hash != "" {
			revision = *request.BlockIdentifier.Hash
		} else if request.BlockIdentifier.Index != nil {
			revision = fmt.Sprintf("%d", *request.BlockIdentifier.Index)
		} else {
			a.responseHandler.WriteErrorResponse(w, meshcommon.GetError(meshcommon.ErrInvalidBlockIdentifierParameter), http.StatusBadRequest)
			return
		}
	}

	// Determine which currencies to query
	currenciesToQuery := a.getCurrenciesToQuery(request.Currencies)

	// Get balances for each currency at the specified block
	var balances []*types.Amount
	for _, currency := range currenciesToQuery {
		balance, err := a.getBalanceForCurrency(request.AccountIdentifier.Address, currency, revision)
		if err != nil {
			a.responseHandler.WriteErrorResponse(w, meshcommon.GetErrorWithMetadata(meshcommon.ErrFailedToGetAccount, map[string]any{
				"error":    err.Error(),
				"currency": currency.Symbol,
			}), http.StatusInternalServerError)
			return
		}
		if balance != nil {
			balances = append(balances, balance)
		}
	}

	// Get block information for the response
	block, err := a.vechainClient.GetBlock(revision)
	if err != nil {
		a.responseHandler.WriteErrorResponse(w, meshcommon.GetErrorWithMetadata(meshcommon.ErrBlockNotFound, map[string]any{
			"error": err.Error(),
		}), http.StatusBadRequest)
		return
	}

	response := &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(block.Number),
			Hash:  block.ID.String(),
		},
		Balances: balances,
	}

	a.responseHandler.WriteJSONResponse(w, response)
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
