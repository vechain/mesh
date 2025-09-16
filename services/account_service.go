package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshclient "github.com/vechain/mesh/client"
	meshmodels "github.com/vechain/mesh/models"
	meshutils "github.com/vechain/mesh/utils"
)

// AccountService handles account-related endpoints
type AccountService struct {
	vechainClient *meshclient.VeChainClient
}

// NewAccountService creates a new account service
func NewAccountService(vechainClient *meshclient.VeChainClient) *AccountService {
	return &AccountService{
		vechainClient: vechainClient,
	}
}

// AccountBalance returns the balance of an account
func (a *AccountService) AccountBalance(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body for account balance", http.StatusBadRequest)
		return
	}

	var request types.AccountBalanceRequest
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Invalid request body for account balance", http.StatusBadRequest)
		return
	}

	// Get current block for block identifier
	bestBlock, err := a.vechainClient.GetBestBlock()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get best block for account balance: %v", err), http.StatusInternalServerError)
		return
	}

	// Get account information
	account, err := a.vechainClient.GetAccount(request.AccountIdentifier.Address)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get account for account balance: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert hex balance to decimal
	vetBalance, err := meshutils.HexToDecimal(account.Balance)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to convert VET balance for account balance: %v", err), http.StatusInternalServerError)
		return
	}

	vthoBalance, err := meshutils.HexToDecimal(account.Energy)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to convert VTHO balance for account balance: %v", err), http.StatusInternalServerError)
		return
	}

	balance := &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: bestBlock.Number,
			Hash:  bestBlock.ID,
		},
		Balances: []*types.Amount{
			{
				Value:    vetBalance,
				Currency: meshmodels.VETCurrency,
			},
			{
				Value:    vthoBalance,
				Currency: meshmodels.VTHOCurrency,
			},
		},
		Metadata: map[string]any{
			"sequence_number": 1,
		},
	}

	meshutils.WriteJSONResponse(w, balance)
}
