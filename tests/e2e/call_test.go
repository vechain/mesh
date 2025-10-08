package e2e

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	"github.com/vechain/mesh/common/vip180/contracts"
	meshtests "github.com/vechain/mesh/tests"
	"github.com/vechain/thor/v2/abi"
)

// TestCallService_InspectClausesWithVIP180 tests the /call endpoint with VIP180 contract simulation
// It includes three clauses:
// 1. Deploy VIP180 contract
// 2. Call name() method on the simulated deployed contract
// 3. Call symbol() method on the simulated deployed contract
func TestCallService_InspectClausesWithVIP180(t *testing.T) {
	config := GetVIP180TestConfig()

	constructorData := encodeConstructorData(config)

	// Combine bytecode with constructor data
	fullBytecode := append(config.ContractBytecode, constructorData...)

	// Load VIP180 ABI using VeChain's ABI package
	vip180ABI := contracts.MustABI("compiled/VIP180.abi")
	parsedABI, err := abi.New(vip180ABI)
	if err != nil {
		t.Fatalf("Failed to parse VIP180 ABI: %v", err)
	}

	// Get method IDs for name() and symbol()
	nameMethod, exists := parsedABI.MethodByName("name")
	if !exists {
		t.Fatal("name() method not found in ABI")
	}
	nameID := nameMethod.ID()
	nameMethodID := hex.EncodeToString(nameID[:])

	symbolMethod, exists := parsedABI.MethodByName("symbol")
	if !exists {
		t.Fatal("symbol() method not found in ABI")
	}
	symbolID := symbolMethod.ID()
	symbolMethodID := hex.EncodeToString(symbolID[:])

	// Create HTTP client
	httpClient := NewHTTPClient(config.BaseURL, config.TimeoutSeconds)
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: meshcommon.BlockchainName,
		Network:    config.Network,
	}

	// Prepare parameters for /call endpoint
	parameters := map[string]any{
		"clauses": []any{
			// Clause 1: Deploy VIP180 contract
			map[string]any{
				"to":    nil,
				"value": "0x0",
				"data":  "0x" + hex.EncodeToString(fullBytecode),
			},
			// Clause 2: Call name() method on simulated contract
			map[string]any{
				"to":    meshtests.SimulatedContractAddress,
				"value": "0x0",
				"data":  "0x" + nameMethodID,
			},
			// Clause 3: Call symbol() method on simulated contract
			map[string]any{
				"to":    meshtests.SimulatedContractAddress,
				"value": "0x0",
				"data":  "0x" + symbolMethodID,
			},
		},
		"gas":      float64(3000000),
		"caller":   config.SenderAddress,
		"gasPayer": config.SenderAddress,
		"revision": "best",
	}

	callResponse, err := testCall(httpClient, networkIdentifier, meshcommon.CallMethodInspectClauses, parameters)
	if err != nil {
		t.Fatalf("Call endpoint test failed: %v", err)
	}

	// Validate response
	if !callResponse.Idempotent {
		t.Errorf("Expected idempotent to be true, got false")
	}

	results, ok := callResponse.Result["results"].([]any)
	if !ok {
		t.Fatalf("Expected results to be an array, got %T", callResponse.Result["results"])
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Validate Clause 1 result (contract deployment)
	result1, ok := results[0].(map[string]any)
	if !ok {
		t.Fatalf("Expected result 1 to be a map, got %T", results[0])
	}

	if reverted, ok := result1["reverted"].(bool); !ok || reverted {
		t.Errorf("Clause 1 (deployment) should not be reverted. Reverted: %v, VMError: %v", result1["reverted"], result1["vmError"])
	}

	if gasUsed, ok := result1["gasUsed"].(float64); !ok || gasUsed == 0 {
		t.Errorf("Clause 1 (deployment) should have non-zero gasUsed, got %v", gasUsed)
	}

	// Validate Clause 2 result (name() call)
	result2, ok := results[1].(map[string]any)
	if !ok {
		t.Fatalf("Expected result 2 to be a map, got %T", results[1])
	}

	if reverted, ok := result2["reverted"].(bool); !ok || reverted {
		t.Errorf("Clause 2 (name call) should not be reverted. Reverted: %v, VMError: %v", result2["reverted"], result2["vmError"])
	}

	// Decode and validate name() result
	if data, ok := result2["data"].(string); ok {
		dataBytes, err := hex.DecodeString(strings.TrimPrefix(data, "0x"))
		if err != nil {
			t.Errorf("Failed to decode name() result data: %v", err)
		} else {
			var name string
			if err := nameMethod.DecodeOutput(dataBytes, &name); err != nil {
				t.Errorf("Failed to unpack name() result: %v", err)
			} else {
				expectedName := config.TokenName
				if name != expectedName {
					t.Errorf("Expected name to be '%s', got '%s'", expectedName, name)
				} else {
					t.Logf("✓ Contract name: %s", name)
				}
			}
		}
	} else {
		t.Errorf("Clause 2 (name call) should have data field, got %v", result2["data"])
	}

	// Validate Clause 3 result (symbol() call)
	result3, ok := results[2].(map[string]any)
	if !ok {
		t.Fatalf("Expected result 3 to be a map, got %T", results[2])
	}

	if reverted, ok := result3["reverted"].(bool); !ok || reverted {
		t.Errorf("Clause 3 (symbol call) should not be reverted. Reverted: %v, VMError: %v", result3["reverted"], result3["vmError"])
	}

	// Decode and validate symbol() result
	if data, ok := result3["data"].(string); ok {
		dataBytes, err := hex.DecodeString(strings.TrimPrefix(data, "0x"))
		if err != nil {
			t.Errorf("Failed to decode symbol() result data: %v", err)
		} else {
			var symbol string
			if err := symbolMethod.DecodeOutput(dataBytes, &symbol); err != nil {
				t.Errorf("Failed to unpack symbol() result: %v", err)
			} else {
				expectedSymbol := config.TokenSymbol
				if symbol != expectedSymbol {
					t.Errorf("Expected symbol to be '%s', got '%s'", expectedSymbol, symbol)
				} else {
					t.Logf("✓ Contract symbol: %s", symbol)
				}
			}
		}
	} else {
		t.Errorf("Clause 3 (symbol call) should have data field, got %v", result3["data"])
	}
}
