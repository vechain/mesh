package services

import (
	"context"
	"math"
	"strconv"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshconfig "github.com/vechain/mesh/config"
	meshthor "github.com/vechain/mesh/thor"
)

// NetworkService handles network-related endpoints
type NetworkService struct {
	vechainClient meshthor.VeChainClientInterface
	config        *meshconfig.Config
}

// Peer represents a connected peer
type Peer struct {
	PeerID      string
	BestBlockID string
}

// NewNetworkService creates a new network service
func NewNetworkService(vechainClient meshthor.VeChainClientInterface, config *meshconfig.Config) *NetworkService {
	return &NetworkService{
		vechainClient: vechainClient,
		config:        config,
	}
}

// NetworkList returns the list of supported networks
func (n *NetworkService) NetworkList(
	ctx context.Context,
	req *types.MetadataRequest,
) (*types.NetworkListResponse, *types.Error) {
	return &types.NetworkListResponse{
		NetworkIdentifiers: []*types.NetworkIdentifier{
			{
				Blockchain: meshcommon.BlockchainName,
				Network:    n.config.Network,
			},
		},
	}, nil
}

// NetworkStatus returns the current network status
func (n *NetworkService) NetworkStatus(
	ctx context.Context,
	req *types.NetworkRequest,
) (*types.NetworkStatusResponse, *types.Error) {
	// Get real VeChain data
	bestBlock, err := n.vechainClient.GetBlock("best")
	if err != nil {
		return nil, meshcommon.GetError(meshcommon.ErrFailedToGetBestBlock)
	}

	// Get genesis block
	genesisBlock, err := n.vechainClient.GetBlock("0")
	if err != nil {
		return nil, meshcommon.GetError(meshcommon.ErrFailedToGetGenesisBlock)
	}

	// Get sync progress
	progress, err := n.vechainClient.GetSyncProgress()
	if err != nil {
		return nil, meshcommon.GetError(meshcommon.ErrFailedToGetSyncProgress)
	}

	// Get peers
	peers, err := n.vechainClient.GetPeers()
	if err != nil {
		return nil, meshcommon.GetError(meshcommon.ErrFailedToGetPeers)
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

	bestBlockTimestamp := bestBlock.Timestamp * 1000 // Convert to milliseconds
	if bestBlockTimestamp > math.MaxInt64 {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInternalServerError, map[string]any{
			"error": "Best block timestamp is too large",
		})
	}
	safeBestBlockTimestamp := int64(bestBlockTimestamp)

	return &types.NetworkStatusResponse{
		CurrentBlockIdentifier: &types.BlockIdentifier{
			Index: int64(bestBlock.Number),
			Hash:  bestBlock.ID.String(),
		},
		CurrentBlockTimestamp: safeBestBlockTimestamp,
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
	}, nil
}

// NetworkOptions returns network options and capabilities
func (n *NetworkService) NetworkOptions(
	ctx context.Context,
	req *types.NetworkRequest,
) (*types.NetworkOptionsResponse, *types.Error) {
	// Define operation statuses
	operationStatuses := []*types.OperationStatus{
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
		meshcommon.OperationTypeContractCall,
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
		RosettaVersion: n.config.MeshVersion,
		NodeVersion:    n.config.NodeVersion,
	}

	// Create response
	return &types.NetworkOptionsResponse{
		Version: version,
		Allow:   allow,
	}, nil
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
