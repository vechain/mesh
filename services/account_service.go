package services

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshthor "github.com/vechain/mesh/thor"
	meshutils "github.com/vechain/mesh/utils"
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

	// Get current block for block identifier
	bestBlock, err := a.vechainClient.GetBestBlock()
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetErrorWithMetadata(meshutils.ErrFailedToGetBestBlock, map[string]any{
			"error": err.Error(),
		}), http.StatusInternalServerError)
		return
	}

	// Get account information
	account, err := a.vechainClient.GetAccount(request.AccountIdentifier.Address)
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetErrorWithMetadata(meshutils.ErrFailedToGetAccount, map[string]any{
			"error": err.Error(),
		}), http.StatusInternalServerError)
		return
	}

	// Convert HexOrDecimal256 to string
	balanceBytes, _ := account.Balance.MarshalText()
	energyBytes, _ := account.Energy.MarshalText()

	// Convert hex balance to decimal
	vetBalance, err := meshutils.HexToDecimal(string(balanceBytes))
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetErrorWithMetadata(meshutils.ErrFailedToConvertVETBalance, map[string]any{
			"error": err.Error(),
		}), http.StatusInternalServerError)
		return
	}

	vthoBalance, err := meshutils.HexToDecimal(string(energyBytes))
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetErrorWithMetadata(meshutils.ErrFailedToConvertVTHOBalance, map[string]any{
			"error": err.Error(),
		}), http.StatusInternalServerError)
		return
	}

	balance := &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(bestBlock.Number),
			Hash:  bestBlock.ID.String(),
		},
		Balances: []*types.Amount{
			{
				Value:    vetBalance,
				Currency: meshutils.VETCurrency,
			},
			{
				Value:    vthoBalance,
				Currency: meshutils.VTHOCurrency,
			},
		},
		Metadata: map[string]any{
			"sequence_number": 1,
		},
	}

	meshutils.WriteJSONResponse(w, balance)
}
