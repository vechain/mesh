package services

import (
	"context"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshthor "github.com/vechain/mesh/thor"
)

func TestNewEventsService(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()

	service := NewEventsService(mockClient)

	if service == nil {
		t.Fatal("NewEventsService() returned nil")
	}

	if service.vechainClient == nil {
		t.Errorf("NewEventsService() vechainClient is nil")
	}
}

func TestEventsService_EventsBlocks(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewEventsService(mockClient)

	tests := []struct {
		name      string
		offset    *int64
		limit     *int64
		wantError bool
	}{
		{
			name:      "default parameters",
			offset:    nil,
			limit:     nil,
			wantError: false,
		},
		{
			name:      "with offset",
			offset:    int64Ptr(10),
			limit:     nil,
			wantError: false,
		},
		{
			name:      "with limit",
			offset:    nil,
			limit:     int64Ptr(50),
			wantError: false,
		},
		{
			name:      "negative offset",
			offset:    int64Ptr(-1),
			limit:     nil,
			wantError: true,
		},
		{
			name:      "limit too small",
			offset:    nil,
			limit:     int64Ptr(0),
			wantError: true,
		},
		{
			name:      "limit too large",
			offset:    nil,
			limit:     int64Ptr(101),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &types.EventsBlocksRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: meshcommon.BlockchainName,
					Network:    "test",
				},
				Offset: tt.offset,
				Limit:  tt.limit,
			}

			ctx := context.Background()
			response, err := service.EventsBlocks(ctx, request)

			if tt.wantError {
				if err == nil {
					t.Error("EventsBlocks() expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("EventsBlocks() returned error: %v", err)
				}

				if response == nil {
					t.Fatal("EventsBlocks() returned nil response")
				}

				if response.Events == nil {
					t.Error("EventsBlocks() Events is nil")
				}

				if response.MaxSequence < 0 {
					t.Errorf("EventsBlocks() MaxSequence = %v, want >= 0", response.MaxSequence)
				}
			}
		})
	}
}
