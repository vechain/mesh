package services

import (
	"fmt"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshcrypto "github.com/vechain/mesh/common/crypto"
	meshhttp "github.com/vechain/mesh/common/http"
	"github.com/vechain/mesh/common/vip180"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/api"
)

// AccountService handles account-related endpoints
type AccountService struct {
	requestHandler  *meshhttp.RequestHandler
	responseHandler *meshhttp.ResponseHandler
	vechainClient   meshthor.VeChainClientInterface
	bytesHandler    *meshcrypto.BytesHandler
}

// NewAccountService creates a new account service
func NewAccountService(vechainClient meshthor.VeChainClientInterface) *AccountService {
	return &AccountService{
		requestHandler:  meshhttp.NewRequestHandler(),
		responseHandler: meshhttp.NewResponseHandler(),
		vechainClient:   vechainClient,
		bytesHandler:    meshcrypto.NewBytesHandler(),
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

	// Determine which currencies to query
	currenciesToQuery := a.getCurrenciesToQuery(request.Currencies)

	// Get balances for each currency
	var balances []*types.Amount
	for _, currency := range currenciesToQuery {
		balance, err := a.getBalanceForCurrency(request.AccountIdentifier.Address, currency)
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

	// Get block identifier (use request block_identifier or default to best block)
	var block *api.JSONExpandedBlock
	if request.BlockIdentifier != nil {
		block, err = a.getBlockFromIdentifier(*request.BlockIdentifier)
		if err != nil {
			a.responseHandler.WriteErrorResponse(w, meshcommon.GetErrorWithMetadata(meshcommon.ErrBlockNotFound, map[string]any{
				"error": err.Error(),
			}), http.StatusBadRequest)
			return
		}
	} else {
		block, err = a.vechainClient.GetBlock("best")
		if err != nil {
			a.responseHandler.WriteErrorResponse(w, meshcommon.GetErrorWithMetadata(meshcommon.ErrFailedToGetBestBlock, map[string]any{
				"error": err.Error(),
			}), http.StatusInternalServerError)
			return
		}
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

// getBlockFromIdentifier gets a block by its identifier
func (a *AccountService) getBlockFromIdentifier(blockIdentifier types.PartialBlockIdentifier) (*api.JSONExpandedBlock, error) {
	if blockIdentifier.Hash != nil && *blockIdentifier.Hash != "" {
		return a.vechainClient.GetBlock(*blockIdentifier.Hash)
	} else if blockIdentifier.Index != nil {
		return a.vechainClient.GetBlockByNumber(*blockIdentifier.Index)
	}
	return nil, fmt.Errorf("invalid block identifier")
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

// getBalanceForCurrency gets the balance for a specific currency
func (a *AccountService) getBalanceForCurrency(address string, currency *types.Currency) (*types.Amount, error) {
	// Handle VET currency
	if currency.Symbol == meshcommon.VETCurrency.Symbol {
		return a.getVETBalance(address)
	}

	// Handle VTHO currency
	if currency.Symbol == meshcommon.VTHOCurrency.Symbol {
		return a.getVTHOBalance(address)
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

// getVETBalance gets the VET balance for an account
func (a *AccountService) getVETBalance(address string) (*types.Amount, error) {
	account, err := a.vechainClient.GetAccount(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	balanceBytes, err := account.Balance.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal VET balance: %w", err)
	}
	vetBalance, err := a.bytesHandler.HexToDecimal(string(balanceBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to convert VET balance: %w", err)
	}

	return &types.Amount{
		Value:    vetBalance,
		Currency: meshcommon.VETCurrency,
	}, nil
}

// getVTHOBalance gets the VTHO balance for an account
func (a *AccountService) getVTHOBalance(address string) (*types.Amount, error) {
	account, err := a.vechainClient.GetAccount(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	energyBytes, _ := account.Energy.MarshalText()
	vthoBalance, err := a.bytesHandler.HexToDecimal(string(energyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to convert VTHO balance: %w", err)
	}

	return &types.Amount{
		Value:    vthoBalance,
		Currency: meshcommon.VTHOCurrency,
	}, nil
}
