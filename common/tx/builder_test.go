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
	tx, err := builder.BuildTransactionFromRequest(request, config)
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
			"transactionType":       meshcommon.TransactionTypeDynamic,
			"blockRef":              "0x0000000000000000",
			"chainTag":              float64(1),
			"gas":                   float64(21000),
			"nonce":                 "0x1",
			"maxFeePerGas":          "1000000000000000",
			"maxPriorityFeePerGas":  "0",
			"fee_delegator_account": meshtests.FirstSoloAddress,
		},
	}

	builder := NewTransactionBuilder()
	tx, err := builder.BuildTransactionFromRequest(request, config)
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

func TestBuildMeshTransactionFromTransactions(t *testing.T) {
	// Create test transaction
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
	operations := encoder.clauseParser.ParseOperationsFromAPIClauses(tx.Clauses, tx.Origin.String(), "", tx.Gas, &status)
	builder := NewTransactionBuilder()
	meshTx := builder.BuildMeshTransactionFromTransaction(tx, operations)

	if meshTx.TransactionIdentifier == nil {
		t.Errorf("BuildMeshTransactionFromTransactions() returned nil TransactionIdentifier")
	}
}

func TestBuildMeshTransactionFromAPI(t *testing.T) {
	// Create test transaction
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
	operations := encoder.ParseTransactionOperationsFromAPI(tx)
	builder := NewTransactionBuilder()
	meshTx := builder.BuildMeshTransactionFromAPI(tx, operations)

	if meshTx.TransactionIdentifier == nil {
		t.Errorf("BuildMeshTransactionFromAPI() returned nil TransactionIdentifier")
	}
}
