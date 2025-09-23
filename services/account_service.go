package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshthor "github.com/vechain/mesh/thor"
	meshutils "github.com/vechain/mesh/utils"
	"github.com/vechain/thor/v2/api"
)

// AccountService handles account-related endpoints
type AccountService struct {
	vechainClient *meshthor.VeChainClient
}

// NewAccountService creates a new account service
func NewAccountService(vechainClient *meshthor.VeChainClient) *AccountService {
	return &AccountService{
		vechainClient: vechainClient,
	}
}

// AccountBalance returns the balance of an account
func (a *AccountService) AccountBalance(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrInvalidRequestBody), http.StatusBadRequest)
		return
	}

	var request types.AccountBalanceRequest
	if err := json.Unmarshal(body, &request); err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrInvalidRequestBody), http.StatusBadRequest)
		return
	}

	// Validate currencies if provided
	if request.Currencies != nil {
		if err := a.validateCurrencies(request.Currencies); err != nil {
			meshutils.WriteErrorResponse(w, meshutils.GetErrorWithMetadata(meshutils.ErrInvalidCurrency, map[string]any{
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
			meshutils.WriteErrorResponse(w, meshutils.GetErrorWithMetadata(meshutils.ErrFailedToGetAccount, map[string]any{
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
			meshutils.WriteErrorResponse(w, meshutils.GetErrorWithMetadata(meshutils.ErrBlockNotFound, map[string]any{
				"error": err.Error(),
			}), http.StatusBadRequest)
			return
		}
	} else {
		block, err = a.vechainClient.GetBestBlock()
		if err != nil {
			meshutils.WriteErrorResponse(w, meshutils.GetErrorWithMetadata(meshutils.ErrFailedToGetBestBlock, map[string]any{
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

	meshutils.WriteJSONResponse(w, response)
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
		return a.vechainClient.GetBlockByHash(*blockIdentifier.Hash)
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
			meshutils.VETCurrency,
			meshutils.VTHOCurrency,
		}
	}
	return requestCurrencies
}

// getBalanceForCurrency gets the balance for a specific currency
func (a *AccountService) getBalanceForCurrency(address string, currency *types.Currency) (*types.Amount, error) {
	// Handle VET currency
	if currency.Symbol == meshutils.VETCurrency.Symbol {
		return a.getVETBalance(address)
	}

	// Handle VTHO currency
	if currency.Symbol == meshutils.VTHOCurrency.Symbol {
		return a.getVTHOBalance(address)
	}

	// Handle VIP180 tokens
	if contractAddr, exists := currency.Metadata["contractAddress"]; exists {
		if addr, ok := contractAddr.(string); ok {
			return a.getVIP180TokenBalance(address, addr, currency)
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
	vetBalance, err := meshutils.HexToDecimal(string(balanceBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to convert VET balance: %w", err)
	}

	return &types.Amount{
		Value:    vetBalance,
		Currency: meshutils.VETCurrency,
	}, nil
}

// getVTHOBalance gets the VTHO balance for an account
func (a *AccountService) getVTHOBalance(address string) (*types.Amount, error) {
	account, err := a.vechainClient.GetAccount(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	energyBytes, _ := account.Energy.MarshalText()
	vthoBalance, err := meshutils.HexToDecimal(string(energyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to convert VTHO balance: %w", err)
	}

	return &types.Amount{
		Value:    vthoBalance,
		Currency: meshutils.VTHOCurrency,
	}, nil
}

// getVIP180TokenBalance gets the balance for a VIP180 token
func (a *AccountService) getVIP180TokenBalance(address, contractAddress string, currency *types.Currency) (*types.Amount, error) {
	// Prepare the function call data
	// balanceOf(address) function selector: 0x70a08231
	functionSelector := "0x70a08231"

	// Pad the address to 32 bytes (64 hex characters)
	paddedAddress := fmt.Sprintf("%064s", address[2:]) // Remove 0x prefix and pad

	// Combine function selector with padded address
	callData := functionSelector + paddedAddress

	// Make the contract call
	result, err := a.vechainClient.CallContract(contractAddress, callData)
	if err != nil {
		// If contract call fails, return 0 balance (token might not exist or be deployed)
		return &types.Amount{
			Value:    "0",
			Currency: currency,
		}, nil
	}

	// Parse the result (32-byte hex string representing uint256)
	if len(result) < 2 || result[:2] != "0x" {
		return &types.Amount{
			Value:    "0",
			Currency: currency,
		}, nil
	}

	// Remove 0x prefix and convert to decimal
	hexValue := result[2:]
	balance, err := meshutils.HexToDecimal("0x" + hexValue)
	if err != nil {
		return &types.Amount{
			Value:    "0",
			Currency: currency,
		}, nil
	}

	return &types.Amount{
		Value:    balance,
		Currency: currency,
	}, nil
}
