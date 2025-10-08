package e2e

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshcrypto "github.com/vechain/mesh/common/crypto"
)

// TestOfflineMode tests that offline endpoints work correctly without Thor node connection
func TestOfflineMode(t *testing.T) {
	// This test should be run with MODE=offline environment variable

	config := GetTestConfig()
	client := NewHTTPClient(config.BaseURL, config.TimeoutSeconds)
	networkIdentifier := CreateTestNetworkIdentifier(config.Network)

	t.Log("Testing offline mode endpoints...")

	t.Run("NetworkList_Offline", func(t *testing.T) {
		t.Log("Testing /network/list in offline mode")
		resp, err := testNetworkList(client)
		if err != nil {
			t.Fatalf("Network list should work in offline mode: %v", err)
		}
		if err := ValidateNetworkListResponse(resp); err != nil {
			t.Fatalf("Invalid network list response: %v", err)
		}
		t.Log("✅ /network/list works in offline mode")
	})

	t.Run("NetworkOptions_Offline", func(t *testing.T) {
		t.Log("Testing /network/options in offline mode")
		resp, err := testNetworkOptions(client, networkIdentifier)
		if err != nil {
			t.Fatalf("Network options should work in offline mode: %v", err)
		}
		if err := ValidateNetworkOptionsResponse(resp); err != nil {
			t.Fatalf("Invalid network options response: %v", err)
		}
		t.Log("✅ /network/options works in offline mode")
	})

	t.Run("ConstructionDerive_Offline", func(t *testing.T) {
		t.Log("Testing /construction/derive in offline mode")
		publicKey := CreateTestPublicKey()
		resp, err := testConstructionDerive(client, networkIdentifier, publicKey)
		if err != nil {
			t.Fatalf("Construction derive should work in offline mode: %v", err)
		}
		if resp.AccountIdentifier == nil || resp.AccountIdentifier.Address == "" {
			t.Fatalf("Invalid derive response: missing account identifier")
		}
		t.Logf("✅ /construction/derive works in offline mode: derived address %s", resp.AccountIdentifier.Address)
	})

	t.Run("OfflineTransactionFlow", func(t *testing.T) {
		testOfflineTransactionFlow(t, client, networkIdentifier, config)
	})

	t.Log("✅ All offline mode tests completed successfully!")
}

// testOfflineTransactionFlow tests the complete offline transaction construction flow
// This includes preprocess, payloads (with pre-provided metadata), parse, combine, and hash
func testOfflineTransactionFlow(t *testing.T, client *HTTPClient, networkIdentifier *types.NetworkIdentifier, config *TestConfig) {
	t.Log("Testing complete offline transaction construction flow...")

	t.Log("Step 1: Testing /construction/preprocess")
	operations := CreateTransferOperations(config.SenderAddress, config.RecipientAddress, config.TransferAmount)
	_, err := testConstructionPreprocess(client, networkIdentifier, nil, config, meshcommon.TransactionTypeLegacy)
	if err != nil {
		t.Fatalf("Construction preprocess failed: %v", err)
	}
	t.Log("✅ Preprocess successful")

	// metadata provided offline
	t.Log("Step 2: Testing /construction/payloads with offline metadata")
	offlineMetadata := map[string]any{
		"blockRef":        "0x0000000000000000", // Mock block ref
		"chainTag":        float64(0xf6),        // Chain tag for the network  (solo)
		"gas":             float64(21000),       // Standard transfer gas
		"nonce":           "0x12345678",         // Mock nonce
		"transactionType": meshcommon.TransactionTypeLegacy,
		"gasPriceCoef":    float64(0), // Legacy transaction field
	}

	publicKey := CreateTestPublicKey()
	payloadsResp, err := testConstructionPayloadsWithMetadata(
		client,
		networkIdentifier,
		operations,
		[]*types.PublicKey{publicKey},
		offlineMetadata,
	)
	if err != nil {
		t.Fatalf("Construction payloads failed: %v", err)
	}
	if err := ValidateConstructionPayloadsResponse(payloadsResp); err != nil {
		t.Fatalf("Invalid payloads response: %v", err)
	}
	t.Log("✅ Payloads successful")

	t.Log("Step 3: Testing /construction/parse (unsigned)")
	parseUnsignedResp, err := testConstructionParse(client, networkIdentifier, []byte(payloadsResp.UnsignedTransaction), false)
	if err != nil {
		t.Fatalf("Construction parse (unsigned) failed: %v", err)
	}
	if err := ValidateConstructionParseResponse(parseUnsignedResp); err != nil {
		t.Fatalf("Invalid parse response: %v", err)
	}
	t.Logf("✅ Parse (unsigned) successful: %d operations", len(parseUnsignedResp.Operations))

	t.Log("Step 4: Signing payload offline")
	payloadHex := fmt.Sprintf("%x", payloadsResp.Payloads[0].Bytes)
	signature, err := meshcrypto.NewSigningHandler(config.SenderPrivateKey).SignPayload(payloadHex)
	if err != nil {
		t.Fatalf("Failed to sign payload: %v", err)
	}
	t.Log("✅ Signature generated")

	t.Log("Step 5: Testing /construction/combine")
	combineResp, err := testConstructionCombine(client, networkIdentifier, payloadsResp, signature)
	if err != nil {
		t.Fatalf("Construction combine failed: %v", err)
	}
	if err := ValidateConstructionCombineResponse(combineResp); err != nil {
		t.Fatalf("Invalid combine response: %v", err)
	}
	t.Log("✅ Combine successful")

	t.Log("Step 6: Testing /construction/parse (signed)")
	parseSignedResp, err := testConstructionParse(client, networkIdentifier, []byte(combineResp.SignedTransaction), true)
	if err != nil {
		t.Fatalf("Construction parse (signed) failed: %v", err)
	}
	if err := ValidateConstructionParseResponse(parseSignedResp); err != nil {
		t.Fatalf("Invalid parse response: %v", err)
	}
	t.Logf("✅ Parse (signed) successful: %d operations, %d signers",
		len(parseSignedResp.Operations), len(parseSignedResp.AccountIdentifierSigners))

	t.Log("Step 7: Testing /construction/hash")
	hashResp, err := testConstructionHash(client, networkIdentifier, combineResp)
	if err != nil {
		t.Fatalf("Construction hash failed: %v", err)
	}
	if err := ValidateTransactionIdentifierResponse(hashResp); err != nil {
		t.Fatalf("Invalid hash response: %v", err)
	}
	t.Logf("✅ Hash successful: %s", hashResp.TransactionIdentifier.Hash)

	t.Log("✅ Complete offline transaction flow successful!")
}

// TestOnlineOnlyEndpointsFailInOfflineMode tests that online-only endpoints fail appropriately in offline mode
func TestOnlineOnlyEndpointsFailInOfflineMode(t *testing.T) {
	// This test should be run with MODE=offline environment variable
	t.Skip("This test requires offline mode configuration - run manually with MODE=offline")

	config := GetTestConfig()
	client := NewHTTPClient(config.BaseURL, config.TimeoutSeconds)
	networkIdentifier := CreateTestNetworkIdentifier(config.Network)

	t.Log("Testing that online-only endpoints fail in offline mode...")

	t.Run("NetworkStatus_ShouldFail", func(t *testing.T) {
		t.Log("Testing /network/status should fail in offline mode")
		request := &types.NetworkRequest{
			NetworkIdentifier: networkIdentifier,
		}

		resp, err := client.Post(meshcommon.NetworkStatusEndpoint, request)
		if err == nil {
			var statusResp types.NetworkStatusResponse
			if err := ParseResponse(resp, &statusResp); err == nil {
				t.Fatal("Network status should fail in offline mode, but succeeded")
			}
		}
		t.Log("✅ /network/status correctly fails in offline mode")
	})

	t.Run("AccountBalance_ShouldFail", func(t *testing.T) {
		t.Log("Testing /account/balance should fail in offline mode")
		request := &types.AccountBalanceRequest{
			NetworkIdentifier: networkIdentifier,
			AccountIdentifier: &types.AccountIdentifier{
				Address: config.SenderAddress,
			},
		}

		resp, err := client.Post(meshcommon.AccountBalanceEndpoint, request)
		if err == nil {
			body := make([]byte, 0)
			if _, readErr := resp.Body.Read(body); readErr == nil {
				t.Log("Response body:", string(body))
			}
			if resp.StatusCode == http.StatusOK {
				t.Fatal("Account balance should fail in offline mode, but succeeded")
			}
		}
		t.Log("✅ /account/balance correctly fails in offline mode")
	})

	t.Run("ConstructionSubmit_ShouldFail", func(t *testing.T) {
		t.Log("Testing /construction/submit should fail in offline mode")
		request := &types.ConstructionSubmitRequest{
			NetworkIdentifier: networkIdentifier,
			SignedTransaction: "0x" + strings.Repeat("00", 100), // Mock signed transaction
		}

		resp, err := client.Post(meshcommon.ConstructionSubmitEndpoint, request)
		if err == nil {
			var submitResp types.TransactionIdentifierResponse
			if err := ParseResponse(resp, &submitResp); err == nil {
				t.Fatal("Construction submit should fail in offline mode, but succeeded")
			}
		}
		t.Log("✅ /construction/submit correctly fails in offline mode")
	})

	t.Log("✅ All online-only endpoint failure tests passed!")
}
