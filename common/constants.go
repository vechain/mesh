package common

import "github.com/coinbase/rosetta-sdk-go/types"

const (
	OfflineMode = "offline"
	OnlineMode  = "online"
)

// Endpoints
const (
	AccountBalanceEndpoint         = "/account/balance"
	BlockEndpoint                  = "/block"
	BlockTransactionEndpoint       = "/block/transaction"
	ConstructionCombineEndpoint    = "/construction/combine"
	ConstructionDeriveEndpoint     = "/construction/derive"
	ConstructionHashEndpoint       = "/construction/hash"
	ConstructionPreprocessEndpoint = "/construction/preprocess"
	ConstructionMetadataEndpoint   = "/construction/metadata"
	ConstructionParseEndpoint      = "/construction/parse"
	ConstructionPayloadsEndpoint   = "/construction/payloads"
	ConstructionSubmitEndpoint     = "/construction/submit"
	MempoolEndpoint                = "/mempool"
	MempoolTransactionEndpoint     = "/mempool/transaction"
	NetworkListEndpoint            = "/network/list"
	NetworkOptionsEndpoint         = "/network/options"
	NetworkStatusEndpoint          = "/network/status"
	HealthEndpoint                 = "/health"
	EventsBlocksEndpoint           = "/events/blocks"
	SearchTransactionsEndpoint     = "/search/transactions"
	CallEndpoint                   = "/call"
)

// Operation types for VeChain
const (
	OperationTypeTransfer      = "Transfer"
	OperationTypeFee           = "Fee"
	OperationTypeFeeDelegation = "FeeDelegation"
	OperationTypeContractCall  = "ContractCall"
)

// Operation statuses for VeChain
const (
	OperationStatusNone      = "None"
	OperationStatusPending   = "Pending"
	OperationStatusSucceeded = "Succeeded"
	OperationStatusReverted  = "Reverted"
)

// Blockchain identifier
const (
	BlockchainName = "vechainthor"
)

// Data directory
const (
	DataDirectory = "/tmp/thor_data"
)

// Transaction types
const (
	TransactionTypeLegacy  = "legacy"
	TransactionTypeDynamic = "dynamic"
)

// Call methods for VeChain
const (
	CallMethodInspectClauses = "inspect_clauses"
)

const (
	VTHOContractAddress = "0x0000000000000000000000000000456e65726779"
)

var (
	// VETCurrency represents the native VeChain token
	VETCurrency = &types.Currency{
		Symbol:   "VET",
		Decimals: 18,
	}

	// VTHOCurrency represents the VeChain Thor Energy token
	VTHOCurrency = &types.Currency{
		Symbol:   "VTHO",
		Decimals: 18,
		Metadata: map[string]any{
			"contractAddress": VTHOContractAddress,
		},
	}
)
