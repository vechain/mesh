package services

import (
	"fmt"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshthor "github.com/vechain/mesh/thor"
	meshutils "github.com/vechain/mesh/utils"
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
func (e *EventsService) EventsBlocks(w http.ResponseWriter, r *http.Request) {
	var request types.EventsBlocksRequest
	if err := meshutils.ParseJSONFromRequestContext(r, &request); err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrInvalidRequestBody), http.StatusBadRequest)
		return
	}

	// Validate request parameters
	offset := int64(0)
	if request.Offset != nil {
		offset = *request.Offset
		if offset < 0 {
			meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrInvalidRequestParameters), http.StatusBadRequest)
			return
		}
	}

	limit := int64(100)
	if request.Limit != nil {
		limit = *request.Limit
		if limit < 1 || limit > 100 {
			meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrInvalidRequestParameters), http.StatusBadRequest)
			return
		}
	}

	// Get the best block number
	bestBlock, err := e.vechainClient.GetBlock("best")
	if err != nil {
		meshutils.WriteErrorResponse(w, meshutils.GetError(meshutils.ErrFailedToGetBestBlock), http.StatusInternalServerError)
		return
	}

	bestBlockNum := int64(bestBlock.Number)

	// If offset is beyond the best block, return empty events
	if bestBlockNum < offset {
		response := types.EventsBlocksResponse{
			MaxSequence: bestBlockNum,
			Events:      []*types.BlockEvent{},
		}
		meshutils.WriteJSONResponse(w, response)
		return
	}

	// Collect events
	var events []*types.BlockEvent
	for index := offset; index < offset+limit; index++ {
		// Don't go beyond the best block
		if index > bestBlockNum {
			break
		}

		// Get block information
		block, err := e.vechainClient.GetBlock(fmt.Sprintf("%d", index))
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

	response := types.EventsBlocksResponse{
		MaxSequence: bestBlockNum,
		Events:      events,
	}

	meshutils.WriteJSONResponse(w, response)
}
