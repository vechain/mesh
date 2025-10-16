package tx

import (
	"math/big"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/math"
	meshcommon "github.com/vechain/mesh/common"
	meshtests "github.com/vechain/mesh/tests"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/api/transactions"
	"github.com/vechain/thor/v2/thor"
	thorTx "github.com/vechain/thor/v2/tx"
)

func TestCreateTransactionBuilder(t *testing.T) {
	builder := NewTransactionBuilder()
	t.Run("Legacy transaction with valid gasPriceCoef", func(t *testing.T) {
		metadata := map[string]any{
			"gasPriceCoef": float64(100),
		}

		builder, err := builder.createTransactionBuilder(meshcommon.TransactionTypeLegacy, metadata)
		if err != nil {
			t.Errorf("createTransactionBuilder() error = %v, want nil", err)
		}
		if builder == nil {
			t.Errorf("createTransactionBuilder() returned nil builder")
		}
	})

	t.Run("Legacy transaction with uint8 gasPriceCoef", func(t *testing.T) {
		metadata := map[string]any{
			"gasPriceCoef": uint8(100),
		}

		builder, err := builder.createTransactionBuilder(meshcommon.TransactionTypeLegacy, metadata)
		if err != nil {
			t.Errorf("createTransactionBuilder() error = %v, want nil", err)
		}
		if builder == nil {
			t.Errorf("createTransactionBuilder() returned nil builder")
		}
	})

	t.Run("Legacy transaction with invalid gasPriceCoef type", func(t *testing.T) {
		metadata := map[string]any{
			"gasPriceCoef": "invalid",
		}

		builder, err := builder.createTransactionBuilder(meshcommon.TransactionTypeLegacy, metadata)
		if err == nil {
			t.Errorf("createTransactionBuilder() should return error for invalid gasPriceCoef type")
		}
		if builder != nil {
			t.Errorf("createTransactionBuilder() should return nil builder when error occurs")
		}
	})

	t.Run("Dynamic fee transaction", func(t *testing.T) {
		metadata := map[string]any{
			"maxFeePerGas":         "1000000000",
			"maxPriorityFeePerGas": "1000000000",
		}

		builder, err := builder.createTransactionBuilder(meshcommon.TransactionTypeDynamic, metadata)
		if err != nil {
			t.Errorf("createTransactionBuilder() error = %v, want nil", err)
		}
		if builder == nil {
			t.Errorf("createTransactionBuilder() returned nil builder")
		}
	})
}

func TestBuildTransactionFromRequest(t *testing.T) {
	config := createTestConfig()

	request := types.ConstructionPayloadsRequest{
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
		},
		Metadata: map[string]any{
			"transactionType": meshcommon.TransactionTypeLegacy,
			"blockRef":        "0x0000000000000000",
			"chainTag":        float64(1),
			"gas":             float64(21000),
			"nonce":           "0x1",
			"gasPriceCoef":    uint8(128),
		},
	}

	builder := NewTransactionBuilder()
	tx, err := builder.BuildTransactionFromRequest(request, config.Expiration)
	if err != nil {
		t.Errorf("BuildTransactionFromRequest() error = %v", err)
	}
	if tx == nil {
		t.Errorf("BuildTransactionFromRequest() returned nil transaction")
	}
}

func TestBuildTransactionFromRequest_WithFeeDelegation(t *testing.T) {
	config := createTestConfig()

	request := types.ConstructionPayloadsRequest{
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 0},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: meshtests.TestAddress1,
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
			{
				OperationIdentifier: &types.OperationIdentifier{Index: 1},
				Type:                meshcommon.OperationTypeTransfer,
				Account: &types.AccountIdentifier{
					Address: meshtests.FirstSoloAddress,
				},
				Amount: &types.Amount{
					Value:    "1000000000000000000",
					Currency: meshcommon.VETCurrency,
				},
			},
		},
		Metadata: map[string]any{
			"transactionType":                      meshcommon.TransactionTypeDynamic,
			"blockRef":                             "0x0000000000000000",
			"chainTag":                             float64(1),
			"gas":                                  float64(21000),
			"nonce":                                "0x1",
			"maxFeePerGas":                         "1000000000000000",
			"maxPriorityFeePerGas":                 "0",
			meshcommon.DelegatorAccountMetadataKey: meshtests.FirstSoloAddress,
		},
	}

	builder := NewTransactionBuilder()
	tx, err := builder.BuildTransactionFromRequest(request, config.Expiration)
	if err != nil {
		t.Errorf("BuildTransactionFromRequest() error = %v", err)
	}
	if tx == nil {
		t.Errorf("BuildTransactionFromRequest() returned nil transaction")
	}

	// Verify delegation feature is set
	if !tx.Features().IsDelegated() {
		t.Errorf("BuildTransactionFromRequest() expected delegation feature to be set")
	}
}

func TestAddClausesToBuilder(t *testing.T) {
	builder := thorTx.NewBuilder(thorTx.TypeLegacy)

	operations := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			Type:                meshcommon.OperationTypeTransfer,
			Account: &types.AccountIdentifier{
				Address: meshtests.TestAddress1,
			},
			Amount: &types.Amount{
				Value:    "1000000000000000000",
				Currency: meshcommon.VETCurrency,
			},
		},
	}

	meshTxBuilder := NewTransactionBuilder()
	err := meshTxBuilder.addClausesToBuilder(builder, operations)
	if err != nil {
		t.Errorf("addClausesToBuilder() error = %v", err)
	}
}

func TestAddClausesToBuilder_VIP180Transfer(t *testing.T) {
	builder := thorTx.NewBuilder(thorTx.TypeLegacy)

	// VIP180 token transfer operations (sender with negative value, recipient with positive)
	operations := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			Type:                meshcommon.OperationTypeTransfer,
			Account: &types.AccountIdentifier{
				Address: meshtests.FirstSoloAddress,
			},
			Amount: &types.Amount{
				Value: "-1000000000000000000",
				Currency: &types.Currency{
					Symbol:   "TVIP",
					Decimals: 18,
					Metadata: map[string]any{
						"contractAddress": "0x0000000000000000000000000000456e65726779",
					},
				},
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 1},
			Type:                meshcommon.OperationTypeTransfer,
			Account: &types.AccountIdentifier{
				Address: meshtests.TestAddress1,
			},
			Amount: &types.Amount{
				Value: "1000000000000000000",
				Currency: &types.Currency{
					Symbol:   "TVIP",
					Decimals: 18,
					Metadata: map[string]any{
						"contractAddress": "0x0000000000000000000000000000456e65726779",
					},
				},
			},
		},
	}

	meshTxBuilder := NewTransactionBuilder()
	err := meshTxBuilder.addClausesToBuilder(builder, operations)
	if err != nil {
		t.Errorf("addClausesToBuilder() VIP180 error = %v", err)
	}

	// Build transaction and verify clause was added
	tx := builder.Build()
	clauses := tx.Clauses()
	if len(clauses) != 1 {
		t.Errorf("Expected 1 clause for VIP180 transfer, got %d", len(clauses))
	}

	// Verify the clause points to the contract address
	if clauses[0].To() == nil {
		t.Error("Clause 'to' address should not be nil")
	} else {
		expectedAddr, _ := thor.ParseAddress("0x0000000000000000000000000000456e65726779")
		if *clauses[0].To() != expectedAddr {
			t.Errorf("Clause 'to' = %v, want %v", clauses[0].To(), expectedAddr)
		}
	}

	// Verify the clause has data (VIP180 transfer function call)
	if len(clauses[0].Data()) == 0 {
		t.Error("Clause data should not be empty for VIP180 transfer")
	}

	// Verify value is 0 for token transfer
	if clauses[0].Value().Sign() != 0 {
		t.Errorf("Clause value should be 0 for VIP180 transfer, got %v", clauses[0].Value())
	}
}

func TestBuildTransactionMetadata_Legacy(t *testing.T) {
	builder := NewTransactionBuilder()
	gasPriceCoef := uint8(128)

	metadata := builder.buildTransactionMetadata(
		byte(1),              // chainTag
		"0x0000000000000000", // blockRef
		720,                  // expiration
		21000,                // gas
		200,                  // size
		&gasPriceCoef,        // gasPriceCoef
		nil,                  // maxFeePerGas
		nil,                  // maxPriorityFeePerGas
	)

	// Verify common fields
	if metadata["chainTag"] != byte(1) {
		t.Errorf("Expected chainTag = 1, got %v", metadata["chainTag"])
	}
	if metadata["blockRef"] != "0x0000000000000000" {
		t.Errorf("Expected blockRef = 0x0000000000000000, got %v", metadata["blockRef"])
	}
	if metadata["expiration"] != uint32(720) {
		t.Errorf("Expected expiration = 720, got %v", metadata["expiration"])
	}
	if metadata["gas"] != uint64(21000) {
		t.Errorf("Expected gas = 21000, got %v", metadata["gas"])
	}
	if metadata["size"] != uint32(200) {
		t.Errorf("Expected size = 200, got %v", metadata["size"])
	}

	// Verify legacy-specific fields
	if metadata["transactionType"] != meshcommon.TransactionTypeLegacy {
		t.Errorf("Expected transactionType = %s, got %v", meshcommon.TransactionTypeLegacy, metadata["transactionType"])
	}
	if metadata["gasPriceCoef"] != uint8(128) {
		t.Errorf("Expected gasPriceCoef = 128, got %v", metadata["gasPriceCoef"])
	}

	// Verify dynamic fields are not present
	if _, exists := metadata["maxFeePerGas"]; exists {
		t.Error("maxFeePerGas should not exist in legacy transaction metadata")
	}
	if _, exists := metadata["maxPriorityFeePerGas"]; exists {
		t.Error("maxPriorityFeePerGas should not exist in legacy transaction metadata")
	}
}

func TestBuildTransactionMetadata_Dynamic_BothFees(t *testing.T) {
	builder := NewTransactionBuilder()
	maxFeePerGas := math.HexOrDecimal256(*big.NewInt(1000000000))
	maxPriorityFeePerGas := math.HexOrDecimal256(*big.NewInt(500000000))

	metadata := builder.buildTransactionMetadata(
		byte(1),               // chainTag
		"0x0000000000000000",  // blockRef
		720,                   // expiration
		21000,                 // gas
		200,                   // size
		nil,                   // gasPriceCoef
		&maxFeePerGas,         // maxFeePerGas
		&maxPriorityFeePerGas, // maxPriorityFeePerGas
	)

	// Verify transaction type
	if metadata["transactionType"] != meshcommon.TransactionTypeDynamic {
		t.Errorf("Expected transactionType = %s, got %v", meshcommon.TransactionTypeDynamic, metadata["transactionType"])
	}

	// Verify dynamic-specific fields
	if metadata["maxFeePerGas"] != "1000000000" {
		t.Errorf("Expected maxFeePerGas = 1000000000, got %v", metadata["maxFeePerGas"])
	}
	if metadata["maxPriorityFeePerGas"] != "500000000" {
		t.Errorf("Expected maxPriorityFeePerGas = 500000000, got %v", metadata["maxPriorityFeePerGas"])
	}

	// Verify legacy field is not present
	if _, exists := metadata["gasPriceCoef"]; exists {
		t.Error("gasPriceCoef should not exist in dynamic transaction metadata")
	}
}

func TestBuildTransactionMetadata_Dynamic_OnlyMaxFee(t *testing.T) {
	builder := NewTransactionBuilder()
	maxFeePerGas := math.HexOrDecimal256(*big.NewInt(2000000000))

	metadata := builder.buildTransactionMetadata(
		byte(1),              // chainTag
		"0x0000000000000000", // blockRef
		720,                  // expiration
		21000,                // gas
		200,                  // size
		nil,                  // gasPriceCoef
		&maxFeePerGas,        // maxFeePerGas
		nil,                  // maxPriorityFeePerGas
	)

	// Verify transaction type
	if metadata["transactionType"] != meshcommon.TransactionTypeDynamic {
		t.Errorf("Expected transactionType = %s, got %v", meshcommon.TransactionTypeDynamic, metadata["transactionType"])
	}

	// Verify maxFeePerGas is present
	if metadata["maxFeePerGas"] != "2000000000" {
		t.Errorf("Expected maxFeePerGas = 2000000000, got %v", metadata["maxFeePerGas"])
	}

	// Verify maxPriorityFeePerGas is not present
	if _, exists := metadata["maxPriorityFeePerGas"]; exists {
		t.Error("maxPriorityFeePerGas should not exist when nil")
	}
}

func TestBuildTransactionMetadata_Dynamic_OnlyMaxPriorityFee(t *testing.T) {
	builder := NewTransactionBuilder()
	maxPriorityFeePerGas := math.HexOrDecimal256(*big.NewInt(300000000))

	metadata := builder.buildTransactionMetadata(
		byte(1),               // chainTag
		"0x0000000000000000",  // blockRef
		720,                   // expiration
		21000,                 // gas
		200,                   // size
		nil,                   // gasPriceCoef
		nil,                   // maxFeePerGas
		&maxPriorityFeePerGas, // maxPriorityFeePerGas
	)

	// Verify transaction type
	if metadata["transactionType"] != meshcommon.TransactionTypeDynamic {
		t.Errorf("Expected transactionType = %s, got %v", meshcommon.TransactionTypeDynamic, metadata["transactionType"])
	}

	// Verify maxPriorityFeePerGas is present
	if metadata["maxPriorityFeePerGas"] != "300000000" {
		t.Errorf("Expected maxPriorityFeePerGas = 300000000, got %v", metadata["maxPriorityFeePerGas"])
	}

	// Verify maxFeePerGas is not present
	if _, exists := metadata["maxFeePerGas"]; exists {
		t.Error("maxFeePerGas should not exist when nil")
	}
}

func TestBuildTransactionMetadata_Dynamic_NoFees(t *testing.T) {
	builder := NewTransactionBuilder()

	metadata := builder.buildTransactionMetadata(
		byte(1),              // chainTag
		"0x0000000000000000", // blockRef
		720,                  // expiration
		21000,                // gas
		200,                  // size
		nil,                  // gasPriceCoef
		nil,                  // maxFeePerGas
		nil,                  // maxPriorityFeePerGas
	)

	// Verify transaction type
	if metadata["transactionType"] != meshcommon.TransactionTypeDynamic {
		t.Errorf("Expected transactionType = %s, got %v", meshcommon.TransactionTypeDynamic, metadata["transactionType"])
	}

	// Verify no fee fields are present
	if _, exists := metadata["gasPriceCoef"]; exists {
		t.Error("gasPriceCoef should not exist in dynamic transaction metadata")
	}
	if _, exists := metadata["maxFeePerGas"]; exists {
		t.Error("maxFeePerGas should not exist when nil")
	}
	if _, exists := metadata["maxPriorityFeePerGas"]; exists {
		t.Error("maxPriorityFeePerGas should not exist when nil")
	}
}

func TestBuildMeshTransactionFromTransactions(t *testing.T) {
	tx := &transactions.Transaction{
		ID: func() thor.Bytes32 {
			hash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
			return hash
		}(),
		Clauses: api.Clauses{
			{
				To: func() *thor.Address {
					addr, _ := thor.ParseAddress(meshtests.TestAddress1)
					return &addr
				}(),
				Value: func() *math.HexOrDecimal256 {
					val, _ := new(big.Int).SetString("0xde0b6b3a7640000", 0)
					hexVal := math.HexOrDecimal256(*val)
					return &hexVal
				}(),
				Data: "0x",
			},
		},
		Origin: func() thor.Address {
			addr, _ := thor.ParseAddress(meshtests.FirstSoloAddress)
			return addr
		}(),
	}

	encoder := NewMeshTransactionEncoder(meshthor.NewMockVeChainClient())
	status := meshcommon.OperationStatusSucceeded
	operations, err := encoder.clauseParser.ParseOperationsFromAPIClauses(tx.Clauses, tx.Origin.String(), "", tx.Gas, &status)
	if err != nil {
		t.Errorf("ParseOperationsFromAPIClauses() error = %v", err)
	}
	builder := NewTransactionBuilder()
	meshTx := builder.BuildMeshTransactionFromTransaction(tx, operations)

	if meshTx.TransactionIdentifier == nil {
		t.Errorf("BuildMeshTransactionFromTransactions() returned nil TransactionIdentifier")
	}
}

func TestBuildMeshTransactionFromAPI(t *testing.T) {
	tx := &api.JSONEmbeddedTx{
		ID: func() thor.Bytes32 {
			hash, _ := thor.ParseBytes32("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
			return hash
		}(),
		Clauses: []*api.JSONClause{
			{
				To: func() *thor.Address {
					addr, _ := thor.ParseAddress(meshtests.TestAddress1)
					return &addr
				}(),
				Value: func() math.HexOrDecimal256 {
					val, _ := new(big.Int).SetString("0xde0b6b3a7640000", 0)
					return math.HexOrDecimal256(*val)
				}(),
				Data: "0x",
			},
		},
		Origin: func() thor.Address {
			addr, _ := thor.ParseAddress(meshtests.FirstSoloAddress)
			return addr
		}(),
	}

	encoder := NewMeshTransactionEncoder(meshthor.NewMockVeChainClient())
	operations, err := encoder.ParseTransactionOperationsFromAPI(tx)
	if err != nil {
		t.Errorf("ParseTransactionOperationsFromAPI() error = %v", err)
	}
	builder := NewTransactionBuilder()
	meshTx := builder.BuildMeshTransactionFromAPI(tx, operations)

	if meshTx.TransactionIdentifier == nil {
		t.Errorf("BuildMeshTransactionFromAPI() returned nil TransactionIdentifier")
	}
}
func TestTransactionBuilder_AddClausesToBuilder_ErrorHandling(t *testing.T) {
	builder := NewTransactionBuilder()

	tests := []struct {
		name       string
		operations []*types.Operation
		expectErr  bool
	}{
		{
			name: "invalid amount format",
			operations: []*types.Operation{
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 0},
					Type:                meshcommon.OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: meshtests.TestAddress1,
					},
					Amount: &types.Amount{
						Value:    "invalid",
						Currency: meshcommon.VETCurrency,
					},
				},
			},
			expectErr: true,
		},
		{
			name: "invalid address format",
			operations: []*types.Operation{
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 0},
					Type:                meshcommon.OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: "invalid_address",
					},
					Amount: &types.Amount{
						Value:    "1000000000000000000",
						Currency: meshcommon.VETCurrency,
					},
				},
			},
			expectErr: true,
		},
		{
			name: "VIP180 with invalid contract address",
			operations: []*types.Operation{
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 0},
					Type:                meshcommon.OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: meshtests.TestAddress1,
					},
					Amount: &types.Amount{
						Value: "1000000000000000000",
						Currency: &types.Currency{
							Symbol:   "VTHO",
							Decimals: 18,
							Metadata: map[string]any{
								"contractAddress": "invalid_contract",
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "VIP180 with failed encoding - invalid amount",
			operations: []*types.Operation{
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 0},
					Type:                meshcommon.OperationTypeTransfer,
					Account: &types.AccountIdentifier{
						Address: meshtests.TestAddress1,
					},
					Amount: &types.Amount{
						Value: "invalid_value",
						Currency: &types.Currency{
							Symbol:   "TEST",
							Decimals: 18,
							Metadata: map[string]any{
								"contractAddress": "0x0000000000000000000000000000456e65726779",
							},
						},
					},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txBuilder := thorTx.NewBuilder(thorTx.TypeLegacy)
			txBuilder.ChainTag(0x27)
			txBuilder.BlockRef([8]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef})
			txBuilder.Expiration(720)
			txBuilder.Gas(21000)
			txBuilder.GasPriceCoef(0)
			txBuilder.Nonce(0x1234567890abcdef)

			err := builder.addClausesToBuilder(txBuilder, tt.operations)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}
