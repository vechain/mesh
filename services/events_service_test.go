package services

import (
	"context"
	"errors"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
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

	offset := int64(10)
	limit := int64(5)
	request := &types.EventsBlocksRequest{
		Offset: &offset,
		Limit:  &limit,
	}

	ctx := context.Background()
	response, err := service.EventsBlocks(ctx, request)

	if err != nil {
		t.Fatalf("EventsBlocks() error = %v", err)
	}

	if response.MaxSequence < 0 {
		t.Errorf("Expected max_sequence >= 0, got %d", response.MaxSequence)
	}

	if len(response.Events) == 0 {
		t.Error("Expected events, got none")
	}
}

func TestEventsService_EventsBlocks_DefaultValues(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewEventsService(mockClient)

	// Mock best block response
	mockClient.SetMockBlock(&api.JSONExpandedBlock{
		JSONBlockSummary: &api.JSONBlockSummary{
			Number: 100,
			ID: func() thor.Bytes32 {
				hash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
				return hash
			}(),
		},
	})

	// Create request with nil offset and limit (should use defaults)
	request := &types.EventsBlocksRequest{
		Offset: nil,
		Limit:  nil,
	}

	ctx := context.Background()
	response, err := service.EventsBlocks(ctx, request)

	if err != nil {
		t.Fatalf("EventsBlocks() error = %v", err)
	}

	if response == nil {
		t.Error("EventsBlocks() returned nil response")
	}
}

func TestEventsService_EventsBlocks_OffsetBeyondBestBlock(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewEventsService(mockClient)

	// Mock best block response
	mockClient.SetMockBlock(&api.JSONExpandedBlock{
		JSONBlockSummary: &api.JSONBlockSummary{
			Number: 100,
			ID: func() thor.Bytes32 {
				hash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
				return hash
			}(),
		},
	})

	// Request with offset beyond best block
	offset := int64(200)
	limit := int64(5)
	request := &types.EventsBlocksRequest{
		Offset: &offset,
		Limit:  &limit,
	}

	ctx := context.Background()
	response, err := service.EventsBlocks(ctx, request)

	if err != nil {
		t.Fatalf("EventsBlocks() error = %v", err)
	}

	// Should return empty events array
	if len(response.Events) != 0 {
		t.Errorf("Expected empty events array, got %d events", len(response.Events))
	}
}

func TestEventsService_EventsBlocks_InvalidOffset(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewEventsService(mockClient)

	// Request with negative offset
	offset := int64(-1)
	request := &types.EventsBlocksRequest{
		Offset: &offset,
	}

	ctx := context.Background()
	_, err := service.EventsBlocks(ctx, request)

	if err == nil {
		t.Error("EventsBlocks() expected error for negative offset")
	}
}

func TestEventsService_EventsBlocks_InvalidLimit(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewEventsService(mockClient)

	// Request with negative limit
	limit := int64(-1)
	request := &types.EventsBlocksRequest{
		Limit: &limit,
	}

	ctx := context.Background()
	_, err := service.EventsBlocks(ctx, request)

	if err == nil {
		t.Error("EventsBlocks() expected error for negative limit")
	}
}

func TestEventsService_EventsBlocks_InvalidLimitTooHigh(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewEventsService(mockClient)

	// Request with limit > 1000
	limit := int64(2000)
	request := &types.EventsBlocksRequest{
		Limit: &limit,
	}

	ctx := context.Background()
	_, err := service.EventsBlocks(ctx, request)

	if err == nil {
		t.Error("EventsBlocks() expected error for limit > 1000")
	}
}

func TestEventsService_EventsBlocks_ThorClientError(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	service := NewEventsService(mockClient)

	// Set error on mock client
	mockClient.SetMockError(errors.New("thor client error"))

	request := &types.EventsBlocksRequest{}

	ctx := context.Background()
	_, err := service.EventsBlocks(ctx, request)

	if err == nil {
		t.Error("EventsBlocks() expected error when thor client returns error")
	}
}
