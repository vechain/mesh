package services

import (
	"context"
	"fmt"
	"math/big"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshoperations "github.com/vechain/mesh/common/operations"
	meshconfig "github.com/vechain/mesh/config"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/thorclient"
)

// CallService handles the /call endpoint for network-specific procedure calls
type CallService struct {
	vechainClient meshthor.VeChainClientInterface
	config        *meshconfig.Config
	clauseParser  *meshoperations.ClauseParser
}

// NewCallService creates a new call service
func NewCallService(vechainClient meshthor.VeChainClientInterface, config *meshconfig.Config) *CallService {
	return &CallService{
		vechainClient: vechainClient,
		config:        config,
		clauseParser:  meshoperations.NewClauseParser(vechainClient, meshoperations.NewOperationsExtractor()),
	}
}

// Call invokes a network-specific procedure call
// For VeChain, this implements the InspectClauses functionality to simulate transactions
func (c *CallService) Call(
	ctx context.Context,
	req *types.CallRequest,
) (*types.CallResponse, *types.Error) {
	// Validate method
	if req.Method != meshcommon.CallMethodInspectClauses {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInvalidRequestBody, map[string]any{
			"error":             "unsupported method",
			"method":            req.Method,
			"supported_methods": []string{meshcommon.CallMethodInspectClauses},
		})
	}

	// Parse parameters into BatchCallData
	batchCallData, err := c.parseBatchCallDataFromParameters(req.Parameters)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInvalidRequestBody, map[string]any{
			"error": fmt.Sprintf("failed to parse parameters: %v", err),
		})
	}

	// Parse revision from parameters (optional)
	revision := "best" // default
	if revisionRaw, ok := req.Parameters["revision"]; ok {
		if revisionStr, ok := revisionRaw.(string); ok {
			revision = revisionStr
		}
	}

	// Call InspectClauses with revision
	results, err := c.vechainClient.InspectClauses(batchCallData, thorclient.Option(thorclient.Revision(revision)))
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInternalServerError, map[string]any{
			"error": fmt.Sprintf("failed to inspect clauses: %v", err),
		})
	}

	// Convert results to Mesh format
	return &types.CallResponse{
		Result: map[string]any{
			"results": convertCallResultsToMap(results),
		},
		Idempotent: true, // InspectClauses is deterministic for a given block state
	}, nil
}

// parseBatchCallDataFromParameters converts request parameters to api.BatchCallData
func (c *CallService) parseBatchCallDataFromParameters(params map[string]any) (*api.BatchCallData, error) {
	batchCallData := &api.BatchCallData{}

	// Parse clauses (required)
	clausesRaw, ok := params["clauses"]
	if !ok {
		return nil, fmt.Errorf("clauses field is required")
	}

	clausesList, ok := clausesRaw.([]any)
	if !ok {
		return nil, fmt.Errorf("clauses must be an array")
	}

	clauses := make([]*api.Clause, len(clausesList))
	for i, clauseRaw := range clausesList {
		clauseMap, ok := clauseRaw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("clause at index %d must be an object", i)
		}

		clause, err := c.clauseParser.ParseAPIClause(clauseMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse clause at index %d: %v", i, err)
		}
		clauses[i] = clause
	}
	batchCallData.Clauses = clauses

	// Parse optional fields
	if gas, ok := params["gas"].(float64); ok {
		batchCallData.Gas = uint64(gas)
	} else if gasStr, ok := params["gas"].(string); ok {
		gasInt := new(big.Int)
		if _, ok := gasInt.SetString(gasStr, 0); !ok {
			return nil, fmt.Errorf("invalid gas value: %s", gasStr)
		}
		batchCallData.Gas = gasInt.Uint64()
	}

	if gasPriceStr, ok := params["gasPrice"].(string); ok {
		gasPrice, err := c.clauseParser.ParseHexOrDecimal256(gasPriceStr)
		if err != nil {
			return nil, fmt.Errorf("invalid gasPrice: %v", err)
		}
		batchCallData.GasPrice = gasPrice
	}

	if provedWorkStr, ok := params["provedWork"].(string); ok {
		provedWork, err := c.clauseParser.ParseHexOrDecimal256(provedWorkStr)
		if err != nil {
			return nil, fmt.Errorf("invalid provedWork: %v", err)
		}
		batchCallData.ProvedWork = provedWork
	}

	if callerStr, ok := params["caller"].(string); ok {
		caller, err := thor.ParseAddress(callerStr)
		if err != nil {
			return nil, fmt.Errorf("invalid caller address: %v", err)
		}
		batchCallData.Caller = &caller
	}

	if gasPayerStr, ok := params["gasPayer"].(string); ok {
		gasPayer, err := thor.ParseAddress(gasPayerStr)
		if err != nil {
			return nil, fmt.Errorf("invalid gasPayer address: %v", err)
		}
		batchCallData.GasPayer = &gasPayer
	}

	if expiration, ok := params["expiration"].(float64); ok {
		batchCallData.Expiration = uint32(expiration)
	}

	if blockRef, ok := params["blockRef"].(string); ok {
		batchCallData.BlockRef = blockRef
	}

	return batchCallData, nil
}

// convertCallResultsToMap converts Thor API CallResults to a map for Rosetta response
func convertCallResultsToMap(results []*api.CallResult) []map[string]any {
	output := make([]map[string]any, len(results))
	for i, result := range results {
		events := make([]map[string]any, len(result.Events))
		for j, event := range result.Events {
			topics := make([]string, len(event.Topics))
			for k, topic := range event.Topics {
				topics[k] = topic.String()
			}
			events[j] = map[string]any{
				"address": event.Address.String(),
				"topics":  topics,
				"data":    event.Data,
			}
		}

		transfers := make([]map[string]any, len(result.Transfers))
		for j, transfer := range result.Transfers {
			transfers[j] = map[string]any{
				"sender":    transfer.Sender.String(),
				"recipient": transfer.Recipient.String(),
				"amount":    (*big.Int)(transfer.Amount).String(),
			}
		}

		output[i] = map[string]any{
			"data":      result.Data,
			"events":    events,
			"transfers": transfers,
			"gasUsed":   result.GasUsed,
			"reverted":  result.Reverted,
			"vmError":   result.VMError,
		}
	}
	return output
}
