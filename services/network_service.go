package services

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
)

// NetworkService handles network-related endpoints
type NetworkService struct{}

// NewNetworkService creates a new network service
func NewNetworkService() *NetworkService {
	return &NetworkService{}
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

	// TODO: Implement real logic to get VeChain status
	status := &types.NetworkStatusResponse{
		CurrentBlockIdentifier: &types.BlockIdentifier{
			Index: 12345678,
			Hash:  "0x1234567890abcdef...",
		},
		CurrentBlockTimestamp: time.Now().UnixMilli(),
		GenesisBlockIdentifier: &types.BlockIdentifier{
			Index: 0,
			Hash:  "0x0000000000000000...",
		},
		OldestBlockIdentifier: &types.BlockIdentifier{
			Index: 1,
			Hash:  "0x1111111111111111...",
		},
		SyncStatus: &types.SyncStatus{
			CurrentIndex: int64Ptr(12345678),
			TargetIndex:  int64Ptr(12345678),
			Synced:       boolPtr(true),
		},
		Peers: []*types.Peer{
			{
				PeerID: "peer-1",
				Metadata: map[string]interface{}{
					"address": "127.0.0.1:8080",
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
