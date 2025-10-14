package e2e

import (
	"encoding/hex"
	"fmt"
	"maps"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshcrypto "github.com/vechain/mesh/common/crypto"
	meshtests "github.com/vechain/mesh/tests"
)

// TestDelegation tests the complete fee delegation flow using VIP-191
// The delegator (FirstSoloAddress) pays fees for transactions originated by TestAddress1
func TestDelegation(t *testing.T) {
	t.Log("Starting fee delegation test with dynamic transaction...")

	// Get test configuration
	config := GetTestConfig()
	client := NewHTTPClient(config.BaseURL, config.TimeoutSeconds)
	networkIdentifier := CreateTestNetworkIdentifier(config.Network)

	// Use TestAddress1 as origin (sender) and FirstSoloAddress as delegator
	originAddress := meshtests.TestAddress1
	originPrivateKey := meshtests.TestAddress1PrivateKey
	delegatorAddress := meshtests.FirstSoloAddress
	delegatorPrivateKey := config.SenderPrivateKey // FirstSoloAddress private key
	recipientAddress := config.RecipientAddress
	transferAmount := config.TransferAmount

	testDelegationFlow(t, client, networkIdentifier, originAddress, originPrivateKey, delegatorAddress, delegatorPrivateKey, recipientAddress, transferAmount)
}

// testDelegationFlow tests the complete delegation transaction flow
func testDelegationFlow(
	t *testing.T,
	client *HTTPClient,
	networkIdentifier *types.NetworkIdentifier,
	originAddress string,
	originPrivateKey string,
	delegatorAddress string,
	delegatorPrivateKey string,
	recipientAddress string,
	transferAmount string,
) {
	// Step 1: Construction Preprocess with delegation metadata
	t.Log("Step 1: Testing /construction/preprocess with fee delegation")
	operations := createDelegationTransferOperations(originAddress, recipientAddress, transferAmount)

	preprocessReq := &types.ConstructionPreprocessRequest{
		NetworkIdentifier: networkIdentifier,
		Operations:        operations,
		Metadata: map[string]any{
			"fee_delegator_account": delegatorAddress,
		},
	}

	resp, err := client.Post(meshcommon.ConstructionPreprocessEndpoint, preprocessReq)
	if err != nil {
		t.Fatalf("Preprocess request failed: %v", err)
	}

	var preprocessResp types.ConstructionPreprocessResponse
	if err := ParseResponse(resp, &preprocessResp); err != nil {
		t.Fatalf("Failed to parse preprocess response: %v", err)
	}

	// Validate that we have 2 required public keys (origin + delegator)
	if len(preprocessResp.RequiredPublicKeys) != 2 {
		t.Fatalf("Expected 2 required public keys for delegation, got %d", len(preprocessResp.RequiredPublicKeys))
	}

	if preprocessResp.RequiredPublicKeys[0].Address != originAddress {
		t.Fatalf("Expected first required public key to be origin address %s, got %s",
			originAddress, preprocessResp.RequiredPublicKeys[0].Address)
	}

	if preprocessResp.RequiredPublicKeys[1].Address != delegatorAddress {
		t.Fatalf("Expected second required public key to be delegator address %s, got %s",
			delegatorAddress, preprocessResp.RequiredPublicKeys[1].Address)
	}

	t.Logf("✅ Preprocess successful: 2 required public keys (origin: %s, delegator: %s)",
		originAddress, delegatorAddress)

	// Step 2: Construction Metadata
	t.Log("Step 2: Testing /construction/metadata with dynamic fee transaction")
	metadataResp, err := testConstructionMetadata(client, networkIdentifier, &preprocessResp, meshcommon.TransactionTypeDynamic)
	if err != nil {
		t.Fatalf("Construction metadata test failed: %v", err)
	}
	t.Logf("Metadata response: %+v", metadataResp)

	// Step 3: Construction Payloads with 2 public keys
	t.Log("Step 3: Testing /construction/payloads with delegation (2 public keys)")
	publicKeys := []*types.PublicKey{
		CreateTestAddress1PublicKey(), // Origin
		CreateTestPublicKey(),         // Delegator (FirstSoloAddress)
	}

	// Add fee_delegator_account to metadata for payloads endpoint
	payloadsMetadata := make(map[string]any)
	maps.Copy(payloadsMetadata, metadataResp.Metadata)
	payloadsMetadata["fee_delegator_account"] = delegatorAddress

	payloadsResp, err := testConstructionPayloadsWithMetadata(
		client,
		networkIdentifier,
		operations,
		publicKeys,
		payloadsMetadata,
	)
	if err != nil {
		t.Fatalf("Construction payloads test failed: %v", err)
	}

	// Validate that we have 2 payloads (one for origin, one for delegator)
	if len(payloadsResp.Payloads) != 2 {
		t.Fatalf("Expected 2 signing payloads for delegation, got %d", len(payloadsResp.Payloads))
	}

	t.Logf("✅ Payloads successful: 2 signing payloads generated")
	t.Logf("  Origin payload account: %s", payloadsResp.Payloads[0].AccountIdentifier.Address)
	t.Logf("  Delegator payload account: %s", payloadsResp.Payloads[1].AccountIdentifier.Address)

	// Step 4: Construction Parse (unsigned)
	t.Log("Step 4: Testing /construction/parse (unsigned)")
	parseUnsignedResp, err := testConstructionParse(client, networkIdentifier, []byte(payloadsResp.UnsignedTransaction), false)
	if err != nil {
		t.Fatalf("Construction parse (unsigned) test failed: %v", err)
	}
	t.Logf("Parse (unsigned) response: operations count = %d, signers count = %d",
		len(parseUnsignedResp.Operations), len(parseUnsignedResp.AccountIdentifierSigners))

	// Step 5: Sign both payloads
	t.Log("Step 5: Signing both payloads (origin and delegator)")

	// Sign origin payload
	originPayloadHex := fmt.Sprintf("%x", payloadsResp.Payloads[0].Bytes)
	originSignature, err := meshcrypto.NewSigningHandler(originPrivateKey).SignPayload(originPayloadHex)
	if err != nil {
		t.Fatalf("Failed to sign origin payload: %v", err)
	}
	t.Logf("  Origin signature generated: %s", originSignature[:20]+"...")

	// Sign delegator payload
	delegatorPayloadHex := fmt.Sprintf("%x", payloadsResp.Payloads[1].Bytes)
	delegatorSignature, err := meshcrypto.NewSigningHandler(delegatorPrivateKey).SignPayload(delegatorPayloadHex)
	if err != nil {
		t.Fatalf("Failed to sign delegator payload: %v", err)
	}
	t.Logf("  Delegator signature generated: %s", delegatorSignature[:20]+"...")

	// Step 6: Construction Combine with 2 signatures
	t.Log("Step 6: Testing /construction/combine with 2 signatures")
	signatures := []*types.Signature{
		{
			SigningPayload: payloadsResp.Payloads[0],
			PublicKey:      publicKeys[0],
			SignatureType:  "ecdsa_recovery",
			Bytes:          func() []byte { b, _ := hex.DecodeString(originSignature); return b }(),
		},
		{
			SigningPayload: payloadsResp.Payloads[1],
			PublicKey:      publicKeys[1],
			SignatureType:  "ecdsa_recovery",
			Bytes:          func() []byte { b, _ := hex.DecodeString(delegatorSignature); return b }(),
		},
	}

	combineResp, err := testConstructionCombineWithSignatures(client, networkIdentifier, payloadsResp, signatures)
	if err != nil {
		t.Fatalf("Construction combine test failed: %v", err)
	}
	t.Logf("✅ Combine response: signed transaction = %s", combineResp.SignedTransaction[:20]+"...")

	// Step 7: Construction Parse (signed) - should show fee delegation operation
	t.Log("Step 7: Testing /construction/parse (signed) - verifying fee delegation operation")
	parseSignedResp, err := testConstructionParse(client, networkIdentifier, []byte(combineResp.SignedTransaction), true)
	if err != nil {
		t.Fatalf("Construction parse (signed) test failed: %v", err)
	}
	t.Logf("Parse (signed) response: operations count = %d, signers count = %d",
		len(parseSignedResp.Operations), len(parseSignedResp.AccountIdentifierSigners))

	// Validate that we have 2 signers
	if len(parseSignedResp.AccountIdentifierSigners) != 2 {
		t.Fatalf("Expected 2 signers for delegation transaction, got %d", len(parseSignedResp.AccountIdentifierSigners))
	}
	t.Logf("✅ Parse (signed) has 2 signers:")
	for i, signer := range parseSignedResp.AccountIdentifierSigners {
		t.Logf("  Signer %d: %s", i+1, signer.Address)
	}

	// Validate that there's a fee delegation operation
	hasFeeDelegation := false
	for _, op := range parseSignedResp.Operations {
		if op.Type == meshcommon.OperationTypeFeeDelegation {
			hasFeeDelegation = true
			t.Logf("✅ Found fee delegation operation: account=%s, amount=%s %s",
				op.Account.Address, op.Amount.Value, op.Amount.Currency.Symbol)

			// Verify it's the delegator paying
			if op.Account.Address != delegatorAddress {
				t.Fatalf("Expected fee delegation operation to be paid by delegator %s, got %s",
					delegatorAddress, op.Account.Address)
			}
			break
		}
	}

	if !hasFeeDelegation {
		t.Fatal("Expected to find a fee delegation operation in parsed signed transaction")
	}

	// Step 8: Construction Hash
	t.Log("Step 8: Testing /construction/hash")
	hashResp, err := testConstructionHash(client, networkIdentifier, combineResp)
	if err != nil {
		t.Fatalf("Construction hash test failed: %v", err)
	}
	t.Logf("Hash response: transaction hash = %s", hashResp.TransactionIdentifier.Hash)

	// Step 9: Construction Submit
	t.Log("Step 9: Testing /construction/submit")
	submitResp, err := testConstructionSubmit(client, networkIdentifier, combineResp)
	if err != nil {
		t.Fatalf("Construction submit test failed: %v", err)
	}
	t.Logf("✅ Submit response: transaction hash = %s", submitResp.TransactionIdentifier.Hash)

	// Step 10: Search for transaction to verify it was included
	t.Log("Step 10: Searching for transaction to verify inclusion")
	retries := 3
	delayInSeconds := 1
	searchResp, err := testSearchTransactionsWithRetry(client, networkIdentifier, submitResp.TransactionIdentifier, retries, delayInSeconds)
	if err != nil {
		t.Fatalf("Search transactions test failed: %v", err)
	}

	if len(searchResp.Transactions) == 0 {
		t.Fatal("Transaction not found after submission")
	}

	t.Logf("✅ Transaction found in blockchain")

	// Verify the transaction has fee delegation operation
	foundTx := searchResp.Transactions[0].Transaction
	hasFeeDelegationInBlock := false
	for _, op := range foundTx.Operations {
		if op.Type == meshcommon.OperationTypeFeeDelegation {
			hasFeeDelegationInBlock = true
			t.Logf("✅ Confirmed: Fee delegation operation in block - delegator %s paid %s %s",
				op.Account.Address, op.Amount.Value, op.Amount.Currency.Symbol)
			break
		}
	}

	if !hasFeeDelegationInBlock {
		t.Fatal("Expected to find fee delegation operation in the transaction in the block")
	}

	t.Log("✅ ========================================")
	t.Log("✅ Fee delegation test completed successfully!")
	t.Log("✅ Origin (sender):", originAddress)
	t.Log("✅ Delegator (fee payer):", delegatorAddress)
	t.Log("✅ Transaction hash:", submitResp.TransactionIdentifier.Hash)
	t.Log("✅ ========================================")
}

// createDelegationTransferOperations creates transfer operations for delegation test
func createDelegationTransferOperations(senderAddress, recipientAddress, amount string) []*types.Operation {
	status := meshcommon.OperationStatusNone
	return []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 0,
			},
			Type:   meshcommon.OperationTypeTransfer,
			Status: &status,
			Account: &types.AccountIdentifier{
				Address: recipientAddress,
			},
			Amount: &types.Amount{
				Value:    amount,
				Currency: CreateVETCurrency(),
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{
				Index: 1,
			},
			Type:   meshcommon.OperationTypeTransfer,
			Status: &status,
			Account: &types.AccountIdentifier{
				Address: senderAddress,
			},
			Amount: &types.Amount{
				Value:    "-" + amount,
				Currency: CreateVETCurrency(),
			},
		},
	}
}
