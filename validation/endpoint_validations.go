package validation

import meshcommon "github.com/vechain/mesh/common"

// EndpointValidationSets defines the validation sets for different Mesh endpoints
var EndpointValidationSets = map[string][]ValidationType{
	// Network API endpoints
	meshcommon.NetworkListEndpoint:    NetworkListValidations,
	meshcommon.NetworkStatusEndpoint:  NetworkValidations,
	meshcommon.NetworkOptionsEndpoint: NetworkValidations,

	// Account API endpoints
	meshcommon.AccountBalanceEndpoint: AccountBalanceValidations,

	// Construction API endpoints
	meshcommon.ConstructionDeriveEndpoint:     ConstructionValidations,
	meshcommon.ConstructionPreprocessEndpoint: ConstructionValidations,
	meshcommon.ConstructionMetadataEndpoint:   ConstructionValidations,
	meshcommon.ConstructionPayloadsEndpoint:   ConstructionPayloadsValidations,
	meshcommon.ConstructionParseEndpoint:      ConstructionValidations,
	meshcommon.ConstructionCombineEndpoint:    ConstructionValidations,
	meshcommon.ConstructionHashEndpoint:       ConstructionValidations,
	meshcommon.ConstructionSubmitEndpoint:     ConstructionValidations,

	// Block API endpoints
	meshcommon.BlockEndpoint:            NetworkValidations,
	meshcommon.BlockTransactionEndpoint: NetworkValidations,

	// Mempool API endpoints
	meshcommon.MempoolEndpoint:            NetworkValidations,
	meshcommon.MempoolTransactionEndpoint: NetworkValidations,

	// Events API endpoints
	meshcommon.EventsBlocksEndpoint: NetworkValidations,

	// Search API endpoints
	meshcommon.SearchTransactionsEndpoint: NetworkValidations,

	// Call API endpoints
	meshcommon.CallEndpoint: NetworkValidations,
}

// GetValidationsForEndpoint returns the validation set for a specific endpoint
func GetValidationsForEndpoint(endpoint string) []ValidationType {
	if validations, exists := EndpointValidationSets[endpoint]; exists {
		return validations
	}
	return NetworkValidations
}
