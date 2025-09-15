package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshclient "github.com/vechain/mesh/client"
)

// NetworkService handles network-related endpoints
type NetworkService struct {
	vechainClient *meshclient.VeChainClient
}

// NewNetworkService creates a new network service
func NewNetworkService(vechainClient *meshclient.VeChainClient) *NetworkService {
	return &NetworkService{
		vechainClient: vechainClient,
	}
}

// NetworkList returns the list of supported networks
func (n *NetworkService) NetworkList(w http.ResponseWriter, r *http.Request) {
	networks := &types.NetworkListResponse{
		NetworkIdentifiers: []*types.NetworkIdentifier{
			{
				Blockchain: "VeChain",
				Network:    "mainnet",
			},
			{
				Blockchain: "VeChain",
				Network:    "testnet",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(networks)
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
			CurrentIndex: int64Ptr(bestBlock.Number),
			TargetIndex:  int64Ptr(bestBlock.Number),
			Synced:       boolPtr(true),
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Helper functions to create pointers
func int64Ptr(i int64) *int64 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
