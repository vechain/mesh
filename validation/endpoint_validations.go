package validation

// EndpointValidationSets defines the validation sets for different Rosetta endpoints
var EndpointValidationSets = map[string][]ValidationType{
	// Network API endpoints
	"/network/list":    NetworkValidations,
	"/network/status":  NetworkValidations,
	"/network/options": NetworkValidations,

	// Account API endpoints
	"/account/balance": AccountBalanceValidations,

	// Construction API endpoints
	"/construction/derive":     ConstructionValidations,
	"/construction/preprocess": ConstructionValidations,
	"/construction/metadata":   ConstructionValidations,
	"/construction/payloads":   ConstructionValidations,
	"/construction/parse":      ConstructionValidations,
	"/construction/combine":    ConstructionValidations,
	"/construction/hash":       ConstructionValidations,
	"/construction/submit":     ConstructionValidations,

	// Block API endpoints
	"/block":             NetworkValidations, // Block endpoints typically need network validation
	"/block/transaction": NetworkValidations,

	// Mempool API endpoints
	"/mempool":             NetworkValidations,
	"/mempool/transaction": NetworkValidations,
}

// GetValidationsForEndpoint returns the validation set for a specific endpoint
func GetValidationsForEndpoint(endpoint string) []ValidationType {
	if validations, exists := EndpointValidationSets[endpoint]; exists {
		return validations
	}
	// Default to network validations for unknown endpoints
	return NetworkValidations
}
