package services

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshtests "github.com/vechain/mesh/tests"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/thor"
)

func TestEventsService_EventsBlocks_Success(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewEventsService(mockClient)

	// Mock best block response (for GetBlock("best"))
	mockClient.SetMockBlock(&api.JSONExpandedBlock{
		JSONBlockSummary: &api.JSONBlockSummary{
			Number: 100,
			ID: func() thor.Bytes32 {
				hash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
				return hash
			}(),
		},
	})

	// Mock block response for specific numbers
	mockClient.SetBlockByNumber(&api.JSONExpandedBlock{
		JSONBlockSummary: &api.JSONBlockSummary{
			Number: 10,
			ID: func() thor.Bytes32 {
				hash, _ := thor.ParseBytes32("0x4567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef123")
				return hash
			}(),
		},
	})

	request := types.EventsBlocksRequest{
		Offset: int64Ptr(10),
		Limit:  int64Ptr(5),
	}

	req := meshtests.CreateRequestWithContext(meshtests.POSTMethod, "/events/blocks", request)
	w := meshtests.CreateResponseRecorder()

	service.EventsBlocks(w, req)

	meshtests.AssertStatusCode(t, w, http.StatusOK)

	var response types.EventsBlocksResponse
	meshtests.UnmarshalResponse(t, w, &response)

	if response.MaxSequence != 100 {
		t.Errorf("Expected max_sequence 100, got %d", response.MaxSequence)
	}

	if len(response.Events) != 5 {
		t.Errorf("Expected 5 events, got %d", len(response.Events))
	}

	// Check first event
	if response.Events[0].Sequence != 10 {
		t.Errorf("Expected first event sequence 10, got %d", response.Events[0].Sequence)
	}

	if response.Events[0].Type != "block_added" {
		t.Errorf("Expected event type 'block_added', got %s", response.Events[0].Type)
	}
}

func TestEventsService_EventsBlocks_DefaultValues(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewEventsService(mockClient)

	// Mock best block response (for GetBlock("best"))
	mockClient.SetMockBlock(&api.JSONExpandedBlock{
		JSONBlockSummary: &api.JSONBlockSummary{
			Number: 49, // Blocks 0-49 = 50 blocks total
			ID: func() thor.Bytes32 {
				hash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
				return hash
			}(),
		},
	})

	// Mock block response for specific numbers
	mockClient.SetBlockByNumber(&api.JSONExpandedBlock{
		JSONBlockSummary: &api.JSONBlockSummary{
			Number: 0,
			ID: func() thor.Bytes32 {
				hash, _ := thor.ParseBytes32("0x4567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef123")
				return hash
			}(),
		},
	})

	// Empty request to test default values
	request := types.EventsBlocksRequest{}

	req := meshtests.CreateRequestWithContext(meshtests.POSTMethod, "/events/blocks", request)
	w := meshtests.CreateResponseRecorder()

	service.EventsBlocks(w, req)

	meshtests.AssertStatusCode(t, w, http.StatusOK)

	var response types.EventsBlocksResponse
	meshtests.UnmarshalResponse(t, w, &response)

	if response.MaxSequence != 49 {
		t.Errorf("Expected max_sequence 49, got %d", response.MaxSequence)
	}

	if len(response.Events) != 50 {
		t.Errorf("Expected 50 events (limited by best block), got %d", len(response.Events))
	}
}

func TestEventsService_EventsBlocks_OffsetBeyondBestBlock(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewEventsService(mockClient)

	// Mock best block response with low block number
	mockClient.SetMockBlock(&api.JSONExpandedBlock{
		JSONBlockSummary: &api.JSONBlockSummary{
			Number: 10,
			ID: func() thor.Bytes32 {
				hash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
				return hash
			}(),
		},
	})

	request := types.EventsBlocksRequest{
		Offset: int64Ptr(20), // Beyond best block
		Limit:  int64Ptr(5),
	}

	req := meshtests.CreateRequestWithContext(meshtests.POSTMethod, "/events/blocks", request)
	w := meshtests.CreateResponseRecorder()

	service.EventsBlocks(w, req)

	meshtests.AssertStatusCode(t, w, http.StatusOK)

	var response types.EventsBlocksResponse
	meshtests.UnmarshalResponse(t, w, &response)

	if response.MaxSequence != 10 {
		t.Errorf("Expected max_sequence 10, got %d", response.MaxSequence)
	}

	if len(response.Events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(response.Events))
	}
}

func TestEventsService_EventsBlocks_InvalidOffset(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewEventsService(mockClient)

	request := types.EventsBlocksRequest{
		Offset: int64Ptr(-1), // Invalid negative offset
		Limit:  int64Ptr(5),
	}

	req := meshtests.CreateRequestWithContext(meshtests.POSTMethod, "/events/blocks", request)
	w := meshtests.CreateResponseRecorder()

	service.EventsBlocks(w, req)

	meshtests.AssertStatusCode(t, w, http.StatusBadRequest)
}

func TestEventsService_EventsBlocks_InvalidLimit(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewEventsService(mockClient)

	request := types.EventsBlocksRequest{
		Offset: int64Ptr(0),
		Limit:  int64Ptr(0), // Invalid limit
	}

	req := meshtests.CreateRequestWithContext(meshtests.POSTMethod, "/events/blocks", request)
	w := meshtests.CreateResponseRecorder()

	service.EventsBlocks(w, req)

	meshtests.AssertStatusCode(t, w, http.StatusBadRequest)
}

func TestEventsService_EventsBlocks_InvalidLimitTooHigh(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewEventsService(mockClient)

	request := types.EventsBlocksRequest{
		Offset: int64Ptr(0),
		Limit:  int64Ptr(101), // Invalid limit too high
	}

	req := meshtests.CreateRequestWithContext(meshtests.POSTMethod, "/events/blocks", request)
	w := meshtests.CreateResponseRecorder()

	service.EventsBlocks(w, req)

	meshtests.AssertStatusCode(t, w, http.StatusBadRequest)
}

func TestEventsService_EventsBlocks_InvalidRequestBody(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewEventsService(mockClient)

	req := meshtests.CreateInvalidJSONRequest(meshtests.POSTMethod, "/events/blocks")
	w := meshtests.CreateResponseRecorder()

	service.EventsBlocks(w, req)

	meshtests.AssertStatusCode(t, w, http.StatusBadRequest)
}

func TestEventsService_EventsBlocks_ThorClientError(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewEventsService(mockClient)

	// Set mock error to simulate client error
	mockClient.SetMockError(fmt.Errorf("client error"))

	request := types.EventsBlocksRequest{
		Offset: int64Ptr(0),
		Limit:  int64Ptr(5),
	}

	req := meshtests.CreateRequestWithContext(meshtests.POSTMethod, "/events/blocks", request)
	w := meshtests.CreateResponseRecorder()

	service.EventsBlocks(w, req)

	meshtests.AssertStatusCode(t, w, http.StatusInternalServerError)
}

// Helper function to create int64 pointer
func int64Ptr(v int64) *int64 {
	return &v
}
