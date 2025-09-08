package services

import (
	"encoding/json"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
)

// AccountService handles account-related endpoints
type AccountService struct{}

// NewAccountService creates a new account service
func NewAccountService() *AccountService {
	return &AccountService{}
}

// AccountBalance returns the balance of an account
func (a *AccountService) AccountBalance(w http.ResponseWriter, r *http.Request) {
	var request types.AccountBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement real logic to get VeChain balance
	balance := &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: 12345678,
			Hash:  "0x1234567890abcdef...",
		},
		Balances: []*types.Amount{
			{
				Value: "1000000000000000000", // 1 VET in wei
				Currency: &types.Currency{
					Symbol:   "VET",
					Decimals: 18,
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
