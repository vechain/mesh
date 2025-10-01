package services

import (
	"net/http"
	"strconv"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshhttp "github.com/vechain/mesh/common/http"
	meshconfig "github.com/vechain/mesh/config"
	meshthor "github.com/vechain/mesh/thor"
)

// NetworkService handles network-related endpoints
type NetworkService struct {
	requestHandler  *meshhttp.RequestHandler
	responseHandler *meshhttp.ResponseHandler
	vechainClient   meshthor.VeChainClientInterface
	config          *meshconfig.Config
}

// Peer represents a connected peer
type Peer struct {
	PeerID      string
	BestBlockID string
}

// NewNetworkService creates a new network service
func NewNetworkService(vechainClient meshthor.VeChainClientInterface, config *meshconfig.Config) *NetworkService {
	return &NetworkService{
		requestHandler:  meshhttp.NewRequestHandler(),
		responseHandler: meshhttp.NewResponseHandler(),
		vechainClient:   vechainClient,
		config:          config,
	}
}

// NetworkList returns the list of supported networks
func (n *NetworkService) NetworkList(w http.ResponseWriter, r *http.Request) {
	networks := &types.NetworkListResponse{
		NetworkIdentifiers: []*types.NetworkIdentifier{
			{
				Blockchain: meshcommon.BlockchainName,
				Network:    n.config.GetNetwork(),
			},
		},
	}

	n.responseHandler.WriteJSONResponse(w, networks)
}

// NetworkStatus returns the current network status
func (n *NetworkService) NetworkStatus(w http.ResponseWriter, r *http.Request) {
	var request types.NetworkRequest
	if err := n.requestHandler.ParseJSONFromContext(r, &request); err != nil {
		n.responseHandler.WriteErrorResponse(w, meshcommon.GetError(meshcommon.ErrInvalidRequestBody), http.StatusBadRequest)
		return
	}

	// Get real VeChain data
	bestBlock, err := n.vechainClient.GetBlock("best")
	if err != nil {
		n.responseHandler.WriteErrorResponse(w, meshcommon.GetError(meshcommon.ErrFailedToGetBestBlock), http.StatusInternalServerError)
		return
	}

	// Get genesis block
	genesisBlock, err := n.vechainClient.GetBlock("0")
	if err != nil {
		n.responseHandler.WriteErrorResponse(w, meshcommon.GetError(meshcommon.ErrFailedToGetGenesisBlock), http.StatusInternalServerError)
		return
	}

	// Get sync progress
	progress, err := n.vechainClient.GetSyncProgress()
	if err != nil {
		n.responseHandler.WriteErrorResponse(w, meshcommon.GetError(meshcommon.ErrFailedToGetSyncProgress), http.StatusInternalServerError)
		return
	}

	// Get peers
	peers, err := n.vechainClient.GetPeers()
	if err != nil {
		n.responseHandler.WriteErrorResponse(w, meshcommon.GetError(meshcommon.ErrFailedToGetPeers), http.StatusInternalServerError)
		return
	}

	// Convert peers to utils.Peer type
	utilsPeers := make([]Peer, len(peers))
	for i, peer := range peers {
		utilsPeers[i] = Peer{
			PeerID:      peer.PeerID,
			BestBlockID: peer.BestBlockID,
		}
	}

	// Calculate target index
	targetIndex := getTargetIndex(int64(bestBlock.Number), utilsPeers)

	// Convert peers to types.Peer
	meshPeers := make([]*types.Peer, len(peers))
	for i, peer := range peers {
		meshPeers[i] = &types.Peer{
			PeerID: peer.PeerID,
		}
	}

	currentIndex := int64(bestBlock.Number)
	stage := "block sync"
	synced := progress == 1.0

	status := &types.NetworkStatusResponse{
		CurrentBlockIdentifier: &types.BlockIdentifier{
			Index: int64(bestBlock.Number),
			Hash:  bestBlock.ID.String(),
		},
		CurrentBlockTimestamp: int64(bestBlock.Timestamp) * 1000, // Convert to milliseconds
		GenesisBlockIdentifier: &types.BlockIdentifier{
			Index: int64(genesisBlock.Number),
			Hash:  genesisBlock.ID.String(),
		},
		SyncStatus: &types.SyncStatus{
			CurrentIndex: &currentIndex,
			TargetIndex:  &targetIndex,
			Stage:        &stage,
			Synced:       &synced,
		},
		Peers: meshPeers,
	}

	n.responseHandler.WriteJSONResponse(w, status)
}

// NetworkOptions returns network options and capabilities
func (n *NetworkService) NetworkOptions(w http.ResponseWriter, r *http.Request) {
	var request types.NetworkRequest
	if err := n.requestHandler.ParseJSONFromContext(r, &request); err != nil {
		n.responseHandler.WriteErrorResponse(w, meshcommon.GetError(meshcommon.ErrInvalidRequestBody), http.StatusBadRequest)
		return
	}

	// Define operation statuses
	operationStatuses := []*types.OperationStatus{
		{
			Status:     meshcommon.OperationStatusNone,
			Successful: true,
		},
		{
			Status:     meshcommon.OperationStatusSucceeded,
			Successful: true,
		},
		{
			Status:     meshcommon.OperationStatusReverted,
			Successful: false,
		},
	}

	// Define operation types
	operationTypes := []string{
		meshcommon.OperationTypeTransfer,
		meshcommon.OperationTypeFee,
		meshcommon.OperationTypeFeeDelegation,
	}

	// Define balance exemptions for VTHO (dynamic exemption)
	balanceExemptions := []*types.BalanceExemption{
		{
			Currency:      meshcommon.VTHOCurrency,
			ExemptionType: types.BalanceDynamic,
		},
	}

	// Create allow object
	allow := &types.Allow{
		OperationStatuses:       operationStatuses,
		OperationTypes:          operationTypes,
		Errors:                  meshcommon.GetAllErrors(),
		HistoricalBalanceLookup: true,
		CallMethods:             []string{meshcommon.CallMethodInspectClauses},
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

	n.responseHandler.WriteJSONResponse(w, response)
}

// getTargetIndex calculates the target index based on local index and peers
func getTargetIndex(localIndex int64, peers []Peer) int64 {
	result := localIndex
	for _, peer := range peers {
		// Extract block number from bestBlockID (first 8 bytes = 16 hex characters)
		if len(peer.BestBlockID) >= 16 {
			blockNumHex := peer.BestBlockID[:16]
			if blockNum, err := strconv.ParseInt(blockNumHex, 16, 64); err == nil {
				if result < blockNum {
					result = blockNum
				}
			}
		}
	}
	return result
}
