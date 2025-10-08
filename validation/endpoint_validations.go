package validation

import meshcommon "github.com/vechain/mesh/common"

// EndpointValidationSets defines the validation sets for different Mesh endpoints (online mode)
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

// EndpointValidationSetsOffline defines the validation sets for different Mesh endpoints (offline mode)
var EndpointValidationSetsOffline = map[string][]ValidationType{
	// Network API endpoints (offline support)
	meshcommon.NetworkListEndpoint:    NetworkListOfflineValidations,
	meshcommon.NetworkOptionsEndpoint: NetworkOfflineValidations,

	// Construction API endpoints (offline support)
	meshcommon.ConstructionDeriveEndpoint:     ConstructionOfflineValidations,
	meshcommon.ConstructionPreprocessEndpoint: ConstructionOfflineValidations,
	meshcommon.ConstructionPayloadsEndpoint:   ConstructionPayloadsOfflineValidations,
	meshcommon.ConstructionParseEndpoint:      ConstructionOfflineValidations,
	meshcommon.ConstructionCombineEndpoint:    ConstructionOfflineValidations,
	meshcommon.ConstructionHashEndpoint:       ConstructionOfflineValidations,
}

// GetValidationsForEndpoint returns the validation set for a specific endpoint based on run mode
func GetValidationsForEndpoint(endpoint string, runMode string) []ValidationType {
	// If in offline mode, check if endpoint supports offline
	if runMode != "online" {
		if validations, exists := EndpointValidationSetsOffline[endpoint]; exists {
			return validations
		}
		// If endpoint doesn't support offline, return online validations (which will fail with proper error)
		if validations, exists := EndpointValidationSets[endpoint]; exists {
			return validations
		}
		return NetworkValidations
	}

	// Online mode - use regular validations
	if validations, exists := EndpointValidationSets[endpoint]; exists {
		return validations
	}
	return NetworkValidations
}
