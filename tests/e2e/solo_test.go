package e2e

import (
	"fmt"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshutils "github.com/vechain/mesh/utils"
)

// TestSolo tests the complete construction flow and other endpoints in solo mode
// It tests both legacy and dynamic transaction types in the same flow
func TestSolo(t *testing.T) {
	t.Log("Starting construction endpoints and then mempool test sequence...")

	// Get test configuration
	config := GetTestConfig()
	client := NewHTTPClient(config.BaseURL, config.TimeoutSeconds)
	networkIdentifier := CreateTestNetworkIdentifier(config.Network)

	// Test both transaction types in the same flow
	transactionTypes := []string{TransactionTypeLegacy, TransactionTypeDynamic}

	for _, transactionType := range transactionTypes {
		t.Run(transactionType+"Transaction", func(t *testing.T) {
			t.Logf("Testing %s transaction flow...", transactionType)
			testTransactionFlow(t, client, networkIdentifier, config, transactionType)
		})
	}

	t.Log("✅ All test steps completed successfully!")
}

// testTransactionFlow tests the complete transaction flow for a specific transaction type
func testTransactionFlow(t *testing.T, client *HTTPClient, networkIdentifier *types.NetworkIdentifier, config *TestConfig, transactionType string) {
	// Step 1: Network List
	t.Log("Step 1: Testing /network/list")
	networkListResp, err := testNetworkList(client)
	if err != nil {
		t.Fatalf("Network list test failed: %v", err)
	}
	t.Logf("Network list response: %+v", networkListResp)

	// Step 2: Network Options
	t.Log("Step 2: Testing /network/options")
	networkOptionsResp, err := testNetworkOptions(client, networkIdentifier)
	if err != nil {
		t.Fatalf("Network options test failed: %v", err)
	}
	t.Logf("Network options response: %+v", networkOptionsResp)

	// Step 3: Network Status
	t.Log("Step 3: Testing /network/status")
	networkStatusResp, err := testNetworkStatus(client, networkIdentifier)
	if err != nil {
		t.Fatalf("Network status test failed: %v", err)
	}
	t.Logf("Network status response: %+v", networkStatusResp)

	// Step 4: Construction Preprocess
	t.Logf("Step 4: Testing /construction/preprocess for %s transaction", transactionType)
	preprocessResp, err := testConstructionPreprocess(client, networkIdentifier, config, transactionType)
	if err != nil {
		t.Fatalf("Construction preprocess test failed: %v", err)
	}
	t.Logf("Preprocess response: %+v", preprocessResp)

	// Step 5: Construction Metadata
	t.Logf("Step 5: Testing /construction/metadata for %s transaction", transactionType)
	metadataResp, err := testConstructionMetadata(client, networkIdentifier, preprocessResp, transactionType)
	if err != nil {
		t.Fatalf("Construction metadata test failed: %v", err)
	}
	t.Logf("Metadata response: %+v", metadataResp)

	// Validate transaction type in metadata
	if err := ValidateTransactionTypeInMetadata(metadataResp.Metadata, transactionType); err != nil {
		t.Fatalf("Metadata validation failed: %v", err)
	}

	// Step 6: Construction Payloads
	t.Logf("Step 6: Testing /construction/payloads for %s transaction", transactionType)
	payloadsResp, err := testConstructionPayloads(client, networkIdentifier, metadataResp, config, transactionType)
	if err != nil {
		t.Fatalf("Construction payloads test failed: %v", err)
	}
	t.Logf("Payloads response: %+v", payloadsResp)

	// Step 7: Sign the payload
	t.Log("Step 7: Signing payload")
	payloadHex := fmt.Sprintf("%x", payloadsResp.Payloads[0].Bytes)
	signature, err := meshutils.SignPayload(config.SenderPrivateKey, payloadHex)
	if err != nil {
		t.Fatalf("Failed to sign payload: %v", err)
	}
	t.Logf("Generated signature: %s", signature)

	// Step 8: Construction Combine
	t.Log("Step 8: Testing /construction/combine")
	combineResp, err := testConstructionCombine(client, networkIdentifier, payloadsResp, signature)
	if err != nil {
		t.Fatalf("Construction combine test failed: %v", err)
	}
	t.Logf("Combine response: signed transaction = %s", combineResp.SignedTransaction)

	// Step 9: Construction Hash
	t.Log("Step 9: Testing /construction/hash")
	hashResp, err := testConstructionHash(client, networkIdentifier, combineResp)
	if err != nil {
		t.Fatalf("Construction hash test failed: %v", err)
	}
	t.Logf("Hash response: transaction hash = %s", hashResp.TransactionIdentifier.Hash)

	// Step 10: Construction Submit
	t.Log("Step 10: Testing /construction/submit")
	submitResp, err := testConstructionSubmit(client, networkIdentifier, combineResp)
	if err != nil {
		t.Fatalf("Construction submit test failed: %v", err)
	}
	t.Logf("Submit response: transaction hash = %s", submitResp.TransactionIdentifier.Hash)

	// Step 11: Mempool
	t.Log("Step 11: Testing /mempool")
	mempoolResp, err := testMempool(client, networkIdentifier)
	if err != nil {
		t.Fatalf("Mempool test failed: %v", err)
	}
	t.Logf("Mempool response: %d transactions", len(mempoolResp.TransactionIdentifiers))
	for i, txID := range mempoolResp.TransactionIdentifiers {
		t.Logf("  Transaction [%d]: hash = %s", i, txID.Hash)
	}

	// Step 12: Mempool Transaction
	t.Log("Step 12: Testing /mempool/transaction")
	mempoolTxResp, err := testMempoolTransaction(client, networkIdentifier, submitResp.TransactionIdentifier)
	if err != nil {
		t.Fatalf("Mempool transaction test failed: %v", err)
	}
	t.Logf("Mempool transaction response: transaction hash = %s, operations count = %d",
		mempoolTxResp.Transaction.TransactionIdentifier.Hash,
		len(mempoolTxResp.Transaction.Operations))

	for i, op := range mempoolTxResp.Transaction.Operations {
		t.Logf("  Operation [%d]: type=%s, account=%s, amount=%s %s",
			i, op.Type, op.Account.Address, op.Amount.Value, op.Amount.Currency.Symbol)
	}

	// Step 13: Block (get latest block)
	t.Log("Step 13: Testing /block (latest block)")
	latestBlockResp, err := testBlock(client, networkIdentifier, &types.PartialBlockIdentifier{
		Hash: func() *string { h := "best"; return &h }(),
	})
	if err != nil {
		t.Fatalf("Block test failed: %v", err)
	}
	if err := ValidateBlockResponse(latestBlockResp); err != nil {
		t.Fatalf("Block response validation failed: %v", err)
	}
	t.Logf("Block response: block hash = %s, block index = %d, transactions count = %d",
		latestBlockResp.Block.BlockIdentifier.Hash,
		latestBlockResp.Block.BlockIdentifier.Index,
		len(latestBlockResp.Block.Transactions))
	for i, tx := range latestBlockResp.Block.Transactions {
		t.Logf("  Block Transaction [%d]: hash = %s, operations count = %d",
			i, tx.TransactionIdentifier.Hash, len(tx.Operations))
	}

	// Step 14: Events Blocks (get recent block events)
	t.Log("Step 14: Testing /events/blocks")
	offset := int64(0)
	limit := int64(10)
	eventsResp, err := testEventsBlocks(client, networkIdentifier, &offset, &limit)
	if err != nil {
		t.Fatalf("Events blocks test failed: %v", err)
	}
	if err := ValidateEventsBlocksResponse(eventsResp); err != nil {
		t.Fatalf("Events blocks response validation failed: %v", err)
	}
	t.Logf("Events blocks response: max sequence = %d, events count = %d",
		eventsResp.MaxSequence, len(eventsResp.Events))
	for i, event := range eventsResp.Events {
		t.Logf("  Event [%d]: sequence = %d, block hash = %s, block index = %d, type = %s",
			i, event.Sequence, event.BlockIdentifier.Hash, event.BlockIdentifier.Index, event.Type)
	}

	// Step 15: Search Transactions (search for our submitted transaction)
	t.Log("Step 15: Testing /search/transactions")
	retries := 3
	delayInSeconds := 1
	searchResp, err := testSearchTransactionsWithRetry(client, networkIdentifier, submitResp.TransactionIdentifier, retries, delayInSeconds)
	if err != nil {
		t.Fatalf("Search transactions test failed: %v", err)
	}
	if err := ValidateSearchTransactionsResponse(searchResp); err != nil {
		t.Fatalf("Search transactions response validation failed: %v", err)
	}
	t.Logf("Search transactions response: total count = %d, transactions count = %d",
		searchResp.TotalCount, len(searchResp.Transactions))
	for i, tx := range searchResp.Transactions {
		t.Logf("  Found Transaction [%d]: hash = %s, block hash = %s, block index = %d, operations count = %d",
			i, tx.Transaction.TransactionIdentifier.Hash,
			tx.BlockIdentifier.Hash, tx.BlockIdentifier.Index, len(tx.Transaction.Operations))
	}

	// Step 16: Block Transaction (get our found transaction from the block)
	t.Log("Step 16: Testing /block/transaction")
	blockTxResp, err := testBlockTransaction(client, networkIdentifier, searchResp.Transactions[0].BlockIdentifier, searchResp.Transactions[0].Transaction.TransactionIdentifier)
	if err != nil {
		t.Fatalf("Block transaction test failed: %v", err)
	}
	if err := ValidateBlockTransactionResponse(blockTxResp); err != nil {
		t.Fatalf("Block transaction response validation failed: %v", err)
	}
	t.Logf("Block transaction response: transaction hash = %s, operations count = %d",
		blockTxResp.Transaction.TransactionIdentifier.Hash,
		len(blockTxResp.Transaction.Operations))

	t.Logf("✅ All steps completed successfully for %s transaction!", transactionType)
}
