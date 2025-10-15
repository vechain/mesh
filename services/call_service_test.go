package services

import (
	"context"
	"math/big"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/math"
	meshcommon "github.com/vechain/mesh/common"
	meshconfig "github.com/vechain/mesh/config"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/thor"
)

func createMockCallService() *CallService {
	return createMockCallServiceWithClient(meshthor.NewMockVeChainClient())
}

func createMockCallServiceWithClient(client *meshthor.MockVeChainClient) *CallService {
	config := &meshconfig.Config{}
	config.Network = "solo"
	return NewCallService(client, config)
}

func createTestCallRequest(method string, params map[string]any) *types.CallRequest {
	return &types.CallRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: meshcommon.BlockchainName,
			Network:    "solo",
		},
		Method:     method,
		Parameters: params,
	}
}

func TestCallService_Call_UnsupportedMethod(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest("unsupported_method", map[string]any{
		"test": "value",
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err == nil {
		t.Error("Call() expected error for unsupported method")
	}
}

func TestCallService_Call_ValidInspectClauses(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	mockClient.SetInspectClausesResult([]*api.CallResult{
		{
			Data:      "0x0000000000000000000000000000000000000000000000000000000000000001",
			Events:    []*api.Event{},
			Transfers: []*api.Transfer{},
			GasUsed:   21000,
			Reverted:  false,
			VMError:   "",
		},
	})

	service := createMockCallServiceWithClient(mockClient)

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    meshcommon.VTHOContractAddress,
				"value": "0x0",
				"data":  "0xa9059cbb0000000000000000000000000f872421dc479f3c11edd89512731814d0598db50000000000000000000000000000000000000000000000013f306a2409fc0000",
			},
		},
		"gas":        float64(50000),
		"gasPrice":   "1000000000000000",
		"caller":     "0x6d95e6dca01d109882fe1726a2fb9865fa41e7aa",
		"expiration": float64(1000),
		"blockRef":   "0x00000000851caf3c",
	})

	ctx := context.Background()
	response, err := service.Call(ctx, request)

	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}

	if response.Result == nil {
		t.Error("Call() result is nil")
		return
	}

	// Verify idempotent flag
	if !response.Idempotent {
		t.Errorf("Call() idempotent = %v, want true", response.Idempotent)
	}

	// Verify results
	results, ok := response.Result["results"].([]map[string]any)
	if !ok {
		t.Error("Call() result['results'] is not an array")
		return
	}

	if len(results) != 1 {
		t.Errorf("Call() results length = %v, want 1", len(results))
		return
	}

	// Verify first result
	result := results[0]
	if gasUsed, ok := result["gasUsed"].(uint64); !ok || gasUsed != 21000 {
		t.Errorf("Call() result gasUsed = %v, want 21000", gasUsed)
	}

	if reverted, ok := result["reverted"].(bool); !ok || reverted {
		t.Errorf("Call() result reverted = %v, want false", reverted)
	}
}

func TestCallService_Call_InvalidParameters(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		// Missing clauses field
		"gas": float64(50000),
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err == nil {
		t.Error("Call() expected error for missing clauses")
	}
}

func TestCallService_Call_InvalidClauseFormat(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to": meshcommon.VTHOContractAddress,
				// Missing value and data fields
			},
		},
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err == nil {
		t.Error("Call() expected error for invalid clause format")
	}
}

func TestCallService_Call_WithRevision(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	mockClient.SetInspectClausesResult([]*api.CallResult{
		{
			Data:      "0x0000000000000000000000000000000000000000000000000000000000000002",
			Events:    []*api.Event{},
			Transfers: []*api.Transfer{},
			GasUsed:   25000,
			Reverted:  false,
			VMError:   "",
		},
	})

	service := createMockCallServiceWithClient(mockClient)

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    meshcommon.VTHOContractAddress,
				"value": "0x0",
				"data":  "0xa9059cbb",
			},
		},
		"revision": "finalized",
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
}

func TestCallService_Call_WithNilToAddress(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	mockClient.SetInspectClausesResult([]*api.CallResult{
		{
			Data:      "0x1234",
			Events:    []*api.Event{},
			Transfers: []*api.Transfer{},
			GasUsed:   100000,
			Reverted:  false,
			VMError:   "",
		},
	})

	service := createMockCallServiceWithClient(mockClient)

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    nil, // Contract deployment
				"value": "0x0",
				"data":  "0x6080604052",
			},
		},
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
}

func TestCallService_Call_WithOptionalParameters(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	mockClient.SetInspectClausesResult([]*api.CallResult{
		{
			Data:      "0x0000000000000000000000000000000000000000000000000000000000000001",
			Events:    []*api.Event{},
			Transfers: []*api.Transfer{},
			GasUsed:   50000,
			Reverted:  false,
			VMError:   "",
		},
	})

	service := createMockCallServiceWithClient(mockClient)

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    meshcommon.VTHOContractAddress,
				"value": "0x1234",
				"data":  "0xa9059cbb",
			},
		},
		"gas":        float64(100000),
		"gasPrice":   "0x16345785d8a0000",
		"provedWork": "0x3e8",
		"caller":     "0x6d95e6dca01d109882fe1726a2fb9865fa41e7aa",
		"gasPayer":   "0xd3ae78222beadb038203be21ed5ce7c9b1bff602",
		"expiration": float64(1000),
		"blockRef":   "0x00000000851caf3c",
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
}

func TestCallService_Call_InvalidAddress(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    "invalid-address",
				"value": "0x0",
				"data":  "0x",
			},
		},
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err == nil {
		t.Error("Call() expected error for invalid address")
	}
}

// Test convertCallResultsToMap with events and transfers
func TestCallService_Call_WithEventsAndTransfers(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()

	// Parse addresses and topics
	vthoAddr, _ := thor.ParseAddress(meshcommon.VTHOContractAddress)
	topic1Bytes, _ := thor.ParseBytes32("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	topic2Bytes, _ := thor.ParseBytes32("0x0000000000000000000000006d95e6dca01d109882fe1726a2fb9865fa41e7aa")
	senderAddr, _ := thor.ParseAddress("0x6d95e6dca01d109882fe1726a2fb9865fa41e7aa")
	recipientAddr, _ := thor.ParseAddress("0x0f872421dc479f3c11edd89512731814d0598db5")

	amount := new(big.Int)
	amount.SetString("1000000000000000000", 10)
	hexAmount := math.HexOrDecimal256(*amount)

	// Create mock results with events and transfers
	mockClient.SetInspectClausesResult([]*api.CallResult{
		{
			Data: "0x0000000000000000000000000000000000000000000000000000000000000001",
			Events: []*api.Event{
				{
					Address: vthoAddr,
					Topics: []thor.Bytes32{
						topic1Bytes,
						topic2Bytes,
					},
					Data: "0x0000000000000000000000000000000000000000000000000de0b6b3a7640000",
				},
			},
			Transfers: []*api.Transfer{
				{
					Sender:    senderAddr,
					Recipient: recipientAddr,
					Amount:    &hexAmount,
				},
			},
			GasUsed:  50000,
			Reverted: false,
			VMError:  "",
		},
	})

	service := createMockCallServiceWithClient(mockClient)

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    meshcommon.VTHOContractAddress,
				"value": "0x0",
				"data":  "0xa9059cbb0000000000000000000000000f872421dc479f3c11edd89512731814d0598db50000000000000000000000000000000000000000000000000de0b6b3a7640000",
			},
		},
	})

	ctx := context.Background()
	response, err := service.Call(ctx, request)

	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}

	// Verify results with events and transfers
	results, ok := response.Result["results"].([]map[string]any)
	if !ok || len(results) == 0 {
		t.Fatal("Call() result['results'] is not an array or is empty")
	}

	result := results[0]
	// Verify events
	events, ok := result["events"].([]map[string]any)
	if !ok {
		t.Error("Call() result['events'] is not an array")
	} else if len(events) != 1 {
		t.Errorf("Call() events length = %v, want 1", len(events))
	} else {
		event := events[0]
		if address, ok := event["address"].(string); !ok || address == "" {
			t.Error("Call() event address is invalid")
		}
		if topics, ok := event["topics"].([]string); !ok || len(topics) != 2 {
			t.Errorf("Call() event topics length = %v, want 2", len(topics))
		}
	}

	// Verify transfers
	transfers, ok := result["transfers"].([]map[string]any)
	if !ok {
		t.Error("Call() result['transfers'] is not an array")
	} else if len(transfers) != 1 {
		t.Errorf("Call() transfers length = %v, want 1", len(transfers))
	} else {
		transfer := transfers[0]
		if sender, ok := transfer["sender"].(string); !ok || sender == "" {
			t.Error("Call() transfer sender is invalid")
		}
		if recipient, ok := transfer["recipient"].(string); !ok || recipient == "" {
			t.Error("Call() transfer recipient is invalid")
		}
		if amount, ok := transfer["amount"].(string); !ok || amount == "" {
			t.Error("Call() transfer amount is invalid")
		}
	}
}

// Test error cases for parseBatchCallDataFromParameters
func TestCallService_Call_InvalidClausesNotArray(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": "not-an-array",
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err == nil {
		t.Error("Call() expected error for clauses not being an array")
	}
}

func TestCallService_Call_InvalidClauseNotObject(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			"not-an-object",
		},
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err == nil {
		t.Error("Call() expected error for clause not being an object")
	}
}

func TestCallService_Call_InvalidToAddress(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    123, // Not a string
				"value": "0x0",
				"data":  "0x",
			},
		},
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err == nil {
		t.Error("Call() expected error for invalid to address type")
	}
}

func TestCallService_Call_MissingValue(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":   meshcommon.VTHOContractAddress,
				"data": "0x",
			},
		},
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err == nil {
		t.Error("Call() expected error for missing value")
	}
}

func TestCallService_Call_InvalidValue(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    meshcommon.VTHOContractAddress,
				"value": "invalid-hex",
				"data":  "0x",
			},
		},
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err == nil {
		t.Error("Call() expected error for invalid value")
	}
}

func TestCallService_Call_MissingData(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    meshcommon.VTHOContractAddress,
				"value": "0x0",
			},
		},
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err == nil {
		t.Error("Call() expected error for missing data")
	}
}

func TestCallService_Call_InvalidGasString(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    meshcommon.VTHOContractAddress,
				"value": "0x0",
				"data":  "0x",
			},
		},
		"gas": "invalid-number",
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err == nil {
		t.Error("Call() expected error for invalid gas")
	}
}

func TestCallService_Call_InvalidGasPrice(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    meshcommon.VTHOContractAddress,
				"value": "0x0",
				"data":  "0x",
			},
		},
		"gasPrice": "invalid-hex",
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err == nil {
		t.Error("Call() expected error for invalid gas price")
	}
}

func TestCallService_Call_InvalidProvedWork(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    meshcommon.VTHOContractAddress,
				"value": "0x0",
				"data":  "0x",
			},
		},
		"provedWork": "invalid-hex",
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err == nil {
		t.Error("Call() expected error for invalid proved work")
	}
}

func TestCallService_Call_InvalidCallerAddress(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    meshcommon.VTHOContractAddress,
				"value": "0x0",
				"data":  "0x",
			},
		},
		"caller": "invalid-address",
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err == nil {
		t.Error("Call() expected error for invalid caller address")
	}
}

func TestCallService_Call_InvalidGasPayerAddress(t *testing.T) {
	service := createMockCallService()

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    meshcommon.VTHOContractAddress,
				"value": "0x0",
				"data":  "0x",
			},
		},
		"gasPayer": "short",
	})

	ctx := context.Background()
	_, err := service.Call(ctx, request)

	if err == nil {
		t.Error("Call() expected error for invalid gas payer address")
	}
}

func TestCallService_Call_WithRevertedTransaction(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()

	mockClient.SetInspectClausesResult([]*api.CallResult{
		{
			Data:      "0x",
			Events:    []*api.Event{},
			Transfers: []*api.Transfer{},
			GasUsed:   21000,
			Reverted:  true,
			VMError:   "execution reverted",
		},
	})

	service := createMockCallServiceWithClient(mockClient)

	request := createTestCallRequest(meshcommon.CallMethodInspectClauses, map[string]any{
		"clauses": []any{
			map[string]any{
				"to":    meshcommon.VTHOContractAddress,
				"value": "0x0",
				"data":  "0x",
			},
		},
	})

	ctx := context.Background()
	response, err := service.Call(ctx, request)

	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}

	results, ok := response.Result["results"].([]map[string]any)
	if !ok || len(results) == 0 {
		t.Fatal("Call() result['results'] is not an array or is empty")
	}

	result := results[0]
	if reverted, ok := result["reverted"].(bool); !ok || !reverted {
		t.Errorf("Call() result reverted = %v, want true", reverted)
	}
	if vmError, ok := result["vmError"].(string); !ok || vmError != "execution reverted" {
		t.Errorf("Call() result vmError = %v, want 'execution reverted'", vmError)
	}
}
