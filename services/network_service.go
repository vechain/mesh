package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshconfig "github.com/vechain/mesh/config"
	meshthor "github.com/vechain/mesh/thor"
	meshutils "github.com/vechain/mesh/utils"
)

// NetworkService handles network-related endpoints
type NetworkService struct {
	vechainClient *meshthor.VeChainClient
	config        *meshconfig.Config
}

// NewNetworkService creates a new network service
func NewNetworkService(vechainClient *meshthor.VeChainClient, config *meshconfig.Config) *NetworkService {
	return &NetworkService{
		vechainClient: vechainClient,
		config:        config,
	}
}

// NetworkList returns the list of supported networks
func (n *NetworkService) NetworkList(w http.ResponseWriter, r *http.Request) {
	networks := &types.NetworkListResponse{
		NetworkIdentifiers: []*types.NetworkIdentifier{
			{
				Blockchain: "vechainthor",
				Network:    n.config.GetNetwork(),
			},
		},
	}

	meshutils.WriteJSONResponse(w, networks)
}

// NetworkStatus returns the current network status
func (n *NetworkService) NetworkStatus(w http.ResponseWriter, r *http.Request) {
	var request types.NetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get real VeChain data
	bestBlock, err := n.vechainClient.GetBestBlock()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get best block: %v", err), http.StatusInternalServerError)
		return
	}

	// Get genesis block (block 0)
	genesisBlock, err := n.vechainClient.GetBlockByNumber(0)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get genesis block: %v", err), http.StatusInternalServerError)
		return
	}

	status := &types.NetworkStatusResponse{
		CurrentBlockIdentifier: &types.BlockIdentifier{
			Index: bestBlock.Number,
			Hash:  bestBlock.ID,
		},
		CurrentBlockTimestamp: bestBlock.Timestamp * 1000, // Convert to milliseconds
		GenesisBlockIdentifier: &types.BlockIdentifier{
			Index: genesisBlock.Number,
			Hash:  genesisBlock.ID,
		},
		OldestBlockIdentifier: &types.BlockIdentifier{
			Index: 1,
			Hash:  "0x0000000000000000000000000000000000000000000000000000000000000001",
		},
		SyncStatus: &types.SyncStatus{
			CurrentIndex: meshutils.Int64Ptr(bestBlock.Number),
			TargetIndex:  meshutils.Int64Ptr(bestBlock.Number),
			Synced:       meshutils.BoolPtr(true),
		},
		Peers: []*types.Peer{
			{
				PeerID: "vechain-node",
				Metadata: map[string]any{
					"address": "vechain-node",
				},
			},
		},
	}

	meshutils.WriteJSONResponse(w, status)
}

// NetworkOptions returns network options and capabilities
func (n *NetworkService) NetworkOptions(w http.ResponseWriter, r *http.Request) {
	var request types.NetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Define operation statuses
	operationStatuses := []*types.OperationStatus{
		{
			Status:     "None",
			Successful: true,
		},
		{
			Status:     "Succeeded",
			Successful: true,
		},
		{
			Status:     "Reverted",
			Successful: false,
		},
	}

	// Define operation types
	operationTypes := []string{
		meshutils.OperationTypeNone,
		meshutils.OperationTypeTransfer,
		meshutils.OperationTypeFee,
		meshutils.OperationTypeFeeDelegation,
	}

	// Define balance exemptions for VTHO (dynamic exemption)
	balanceExemptions := []*types.BalanceExemption{
		{
			Currency:      meshutils.VTHOCurrency,
			ExemptionType: types.BalanceDynamic,
		},
	}

	// TODO: There should be an Error field below

	// Create allow object
	allow := &types.Allow{
		OperationStatuses:       operationStatuses,
		OperationTypes:          operationTypes,
		HistoricalBalanceLookup: true,
		CallMethods:             []string{},
		BalanceExemptions:       balanceExemptions,
		MempoolCoins:            false,
	}

	// Create version object
	version := &types.Version{
		RosettaVersion: n.config.GetMeshVersion(),
		NodeVersion:    n.config.NodeVersion,
	}

	// Create response
	response := &types.NetworkOptionsResponse{
		Version: version,
		Allow:   allow,
	}

	meshutils.WriteJSONResponse(w, response)
}
