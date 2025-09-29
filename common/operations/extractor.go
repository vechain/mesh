package operations

import (
	"math/big"
	"slices"
	"strings"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	"github.com/vechain/mesh/common/vip180"
	"github.com/vechain/mesh/thor"
)

type OperationsExtractor struct{}

func NewOperationsExtractor() *OperationsExtractor {
	return &OperationsExtractor{}
}

// GetStringFromOptions gets a string value from options map
func (e *OperationsExtractor) GetStringFromOptions(options map[string]any, key string) string {
	if value, ok := options[key].(string); ok {
		return value
	}
	return "dynamic"
}

// GetTxOrigins extracts origin addresses from operations
func (e *OperationsExtractor) GetTxOrigins(operations []*types.Operation) []string {
	var origins []string

	for _, op := range operations {
		if op.Account == nil || op.Account.Address == "" {
			continue
		}

		address := strings.ToLower(op.Account.Address)

		// Check if address already exists
		if slices.Contains(origins, address) {
			continue
		}

		// Consider Fee operations
		if op.Type == meshcommon.OperationTypeFee {
			origins = append(origins, address)
			continue
		}

		// Consider Transfer operations with negative value (sending)
		if op.Type == meshcommon.OperationTypeTransfer && op.Amount != nil && op.Amount.Value != "" {
			// Parse amount value
			amount := new(big.Int)
			if _, ok := amount.SetString(op.Amount.Value, 10); ok && amount.Sign() < 0 {
				origins = append(origins, address)
			}
		}
	}

	return origins
}

// GetTokenCurrencyFromContractAddress returns the currency definition for a token contract
func (e *OperationsExtractor) GetTokenCurrencyFromContractAddress(contractAddress string, client thor.VeChainClientInterface) (*types.Currency, error) {
	if strings.EqualFold(contractAddress, meshcommon.VTHOCurrency.Metadata["contractAddress"].(string)) {
		return meshcommon.VTHOCurrency, nil
	}

	// For other tokens, fetch metadata from the contract using the wrapper
	contract, err := vip180.NewVIP180Contract(contractAddress, client)
	if err != nil {
		return nil, err
	}

	// Get symbol
	symbol, err := contract.Symbol()
	if err != nil {
		return nil, err
	}

	// Get decimals
	decimals, err := contract.Decimals()
	if err != nil {
		return nil, err
	}

	return &types.Currency{
		Symbol:   symbol,
		Decimals: decimals,
		Metadata: map[string]any{
			"contractAddress": strings.ToLower(contractAddress),
		},
	}, nil
}

// GetVETOperations extracts VET transfer operations from a list of operations
func (e *OperationsExtractor) GetVETOperations(operations []*types.Operation) []map[string]string {
	var result []map[string]string
	for _, op := range operations {
		if op.Type == meshcommon.OperationTypeTransfer && op.Amount != nil && op.Amount.Currency != nil {
			// Check if it's a VET operation (positive amount and VET currency)
			if op.Amount.Currency.Symbol == meshcommon.VETCurrency.Symbol {
				amount, ok := new(big.Int).SetString(op.Amount.Value, 10)
				if ok && amount.Cmp(big.NewInt(0)) > 0 {
					result = append(result, map[string]string{
						"value": op.Amount.Value,
						"to":    op.Account.Address,
					})
				}
			}
		}
	}
	return result
}

// GetTokensOperations extracts VIP180 token operations from a list of operations
func (e *OperationsExtractor) GetTokensOperations(operations []*types.Operation) (registered []map[string]string) {
	for _, op := range operations {
		if op.Type == meshcommon.OperationTypeTransfer && op.Amount != nil && op.Amount.Currency != nil {
			// Check if it's a token operation (positive amount and has contract address)
			if op.Amount.Currency.Metadata != nil {
				if contractAddr, exists := op.Amount.Currency.Metadata["contractAddress"]; exists {
					if addr, ok := contractAddr.(string); ok {
						amount, ok := new(big.Int).SetString(op.Amount.Value, 10)
						if ok && amount.Cmp(big.NewInt(0)) > 0 {
							registered = append(registered, map[string]string{
								"token": addr,
								"value": op.Amount.Value,
								"to":    op.Account.Address,
							})
						}
					}
				}
			}
		}
	}
	return registered
}
