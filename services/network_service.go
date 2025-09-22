package services

import (
	"encoding/json"
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
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrInvalidRequestBody), http.StatusBadRequest)
		return
	}

	// Get real VeChain data
	bestBlock, err := n.vechainClient.GetBestBlock()
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrFailedToGetBestBlock), http.StatusInternalServerError)
		return
	}

	// Get genesis block (block 0)
	genesisBlock, err := n.vechainClient.GetBlockByNumber(0)
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrFailedToGetGenesisBlock), http.StatusInternalServerError)
		return
	}

	// Get sync progress
	progress, err := n.vechainClient.GetSyncProgress()
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrFailedToGetSyncProgress), http.StatusInternalServerError)
		return
	}

	// Get peers
	peers, err := n.vechainClient.GetPeers()
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrFailedToGetPeers), http.StatusInternalServerError)
		return
	}

	// Convert peers to utils.Peer type
	utilsPeers := make([]meshutils.Peer, len(peers))
	for i, peer := range peers {
		utilsPeers[i] = meshutils.Peer{
			PeerID:      peer.PeerID,
			BestBlockID: peer.BestBlockID,
		}
	}

	// Calculate target index
	targetIndex := meshutils.GetTargetIndex(bestBlock.Number, utilsPeers)

	// Convert peers to types.Peer
	meshPeers := make([]*types.Peer, len(peers))
	for i, peer := range peers {
		meshPeers[i] = &types.Peer{
			PeerID: peer.PeerID,
		}
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
		SyncStatus: &types.SyncStatus{
			CurrentIndex: meshutils.Int64Ptr(bestBlock.Number),
			TargetIndex:  meshutils.Int64Ptr(targetIndex),
			Stage:        meshutils.StringPtr("block sync"),
			Synced:       meshutils.BoolPtr(progress == 1.0),
		},
		Peers: meshPeers,
	}

	meshutils.WriteJSONResponse(w, status)
}

// NetworkOptions returns network options and capabilities
func (n *NetworkService) NetworkOptions(w http.ResponseWriter, r *http.Request) {
	var request types.NetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrInvalidRequestBody), http.StatusBadRequest)
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

	// Create allow object
	allow := &types.Allow{
		OperationStatuses:       operationStatuses,
		OperationTypes:          operationTypes,
		Errors:                  meshutils.GetAllErrors(),
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
