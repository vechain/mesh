package tests

const (
	// HTTP Methods
	GETMethod  = "GET"
	POSTMethod = "POST"

	// Content Types
	JSONContentType = "application/json"

	// Common Test Data
	InvalidJSON = "invalid json"
	EmptyJSON   = "{}"

	// Common Error Messages
	FailedToUnmarshalResponse = "Failed to unmarshal response"
	ExpectedStatus            = "Expected status"

	// Endpoints
	AccountBalanceEndpoint       = "/account/balance"
	BlockEndpoint                = "/block"
	BlockTransactionEndpoint     = "/block/transaction"
	ConstructionCombineEndpoint  = "/construction/combine"
	ConstructionDeriveEndpoint   = "/construction/derive"
	ConstructionHashEndpoint     = "/construction/hash"
	ConstructionMetadataEndpoint = "/construction/metadata"
	ConstructionParseEndpoint    = "/construction/parse"
	ConstructionPayloadsEndpoint = "/construction/payloads"
	ConstructionSubmitEndpoint   = "/construction/submit"
	MempoolEndpoint              = "/mempool"
	MempoolTransactionEndpoint   = "/mempool/transaction"
	NetworkListEndpoint          = "/network/list"
	NetworkOptionsEndpoint       = "/network/options"
	NetworkStatusEndpoint        = "/network/status"
	HealthEndpoint               = "/health"
	EventsBlocksEndpoint         = "/events/blocks"
)
