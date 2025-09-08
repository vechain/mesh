package services

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
)

// AccountService handles account-related endpoints
type AccountService struct {
	vechainClient *VeChainClient
}

// NewAccountService creates a new account service
func NewAccountService(vechainClient *VeChainClient) *AccountService {
	return &AccountService{
		vechainClient: vechainClient,
	}
}

// AccountBalance returns the balance of an account
func (a *AccountService) AccountBalance(w http.ResponseWriter, r *http.Request) {
	var request types.AccountBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get current block for block identifier
	bestBlock, err := a.vechainClient.GetBestBlock()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get best block: %v", err), http.StatusInternalServerError)
		return
	}

	// Get account information
	account, err := a.vechainClient.GetAccount(request.AccountIdentifier.Address)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get account: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert hex balance to decimal
	vetBalance, err := hexToDecimal(account.Balance)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to convert VET balance: %v", err), http.StatusInternalServerError)
		return
	}

	vthoBalance, err := hexToDecimal(account.Energy)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to convert VTHO balance: %v", err), http.StatusInternalServerError)
		return
	}

	balance := &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: bestBlock.Number,
			Hash:  bestBlock.ID,
		},
		Balances: []*types.Amount{
			{
				Value: vetBalance,
				Currency: &types.Currency{
					Symbol:   "VET",
					Decimals: 18,
				},
			},
			{
				Value: vthoBalance,
				Currency: &types.Currency{
					Symbol:   "VTHO",
					Decimals: 18,
					Metadata: map[string]interface{}{
						"contractAddress": "0x0000000000000000000000000000456E65726779",
					},
				},
			},
		},
		Metadata: map[string]interface{}{
			"sequence_number": 1,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

// hexToDecimal converts hex string to decimal string
func hexToDecimal(hexStr string) (string, error) {
	// Remove 0x prefix if present
	if len(hexStr) > 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}

	// Convert hex to big.Int
	bigInt := new(big.Int)
	bigInt, ok := bigInt.SetString(hexStr, 16)
	if !ok {
		return "", fmt.Errorf("invalid hex string: %s", hexStr)
	}

	return bigInt.String(), nil
}
