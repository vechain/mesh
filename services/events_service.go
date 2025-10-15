package services

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshthor "github.com/vechain/mesh/thor"
)

// EventsService handles events API endpoints
type EventsService struct {
	vechainClient meshthor.VeChainClientInterface
}

// NewEventsService creates a new events service
func NewEventsService(vechainClient meshthor.VeChainClientInterface) *EventsService {
	return &EventsService{
		vechainClient: vechainClient,
	}
}

// EventsBlocks handles the /events/blocks endpoint
func (e *EventsService) EventsBlocks(
	ctx context.Context,
	req *types.EventsBlocksRequest,
) (*types.EventsBlocksResponse, *types.Error) {
	// Validate request parameters
	offset := int64(0)
	if req.Offset != nil {
		offset = *req.Offset
		if offset < 0 {
			return nil, meshcommon.GetError(meshcommon.ErrInvalidRequestParameters)
		}
	}

	limit := int64(100)
	if req.Limit != nil {
		limit = *req.Limit
		if limit < 1 || limit > 100 {
			return nil, meshcommon.GetError(meshcommon.ErrInvalidRequestParameters)
		}
	}

	// Get the best block number
	bestBlock, err := e.vechainClient.GetBlock("best")
	if err != nil {
		return nil, meshcommon.GetError(meshcommon.ErrFailedToGetBestBlock)
	}

	bestBlockNum := int64(bestBlock.Number)

	// If offset is beyond the best block, return empty events
	if bestBlockNum < offset {
		return &types.EventsBlocksResponse{
			MaxSequence: bestBlockNum,
			Events:      []*types.BlockEvent{},
		}, nil
	}

	// Collect events
	var events []*types.BlockEvent
	for index := offset; index < offset+limit; index++ {
		// Don't go beyond the best block
		if index > bestBlockNum {
			break
		}

		// Get block information
		block, err := e.vechainClient.GetBlockByNumber(index)
		if err != nil {
			// If block doesn't exist, skip it
			continue
		}

		// Create event
		event := types.BlockEvent{
			Sequence: index,
			BlockIdentifier: &types.BlockIdentifier{
				Index: index,
				Hash:  block.ID.String(),
			},
			Type: "block_added",
		}
		events = append(events, &event)
	}

	return &types.EventsBlocksResponse{
		MaxSequence: bestBlockNum,
		Events:      events,
	}, nil
}
