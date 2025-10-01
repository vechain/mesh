package validation

import (
	"testing"

	meshcommon "github.com/vechain/mesh/common"
)

func TestGetValidationsForEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		expected []ValidationType
	}{
		{
			name:     "network list endpoint",
			endpoint: meshcommon.NetworkListEndpoint,
			expected: NetworkListValidations,
		},
		{
			name:     "network status endpoint",
			endpoint: meshcommon.NetworkStatusEndpoint,
			expected: NetworkValidations,
		},
		{
			name:     "network options endpoint",
			endpoint: meshcommon.NetworkOptionsEndpoint,
			expected: NetworkValidations,
		},
		{
			name:     "account balance endpoint",
			endpoint: meshcommon.AccountBalanceEndpoint,
			expected: AccountBalanceValidations,
		},
		{
			name:     "construction derive endpoint",
			endpoint: meshcommon.ConstructionDeriveEndpoint,
			expected: ConstructionValidations,
		},
		{
			name:     "construction preprocess endpoint",
			endpoint: meshcommon.ConstructionPreprocessEndpoint,
			expected: ConstructionValidations,
		},
		{
			name:     "construction metadata endpoint",
			endpoint: meshcommon.ConstructionMetadataEndpoint,
			expected: ConstructionValidations,
		},
		{
			name:     "construction payloads endpoint",
			endpoint: meshcommon.ConstructionPayloadsEndpoint,
			expected: ConstructionPayloadsValidations,
		},
		{
			name:     "construction parse endpoint",
			endpoint: meshcommon.ConstructionParseEndpoint,
			expected: ConstructionValidations,
		},
		{
			name:     "construction combine endpoint",
			endpoint: meshcommon.ConstructionCombineEndpoint,
			expected: ConstructionValidations,
		},
		{
			name:     "construction hash endpoint",
			endpoint: meshcommon.ConstructionHashEndpoint,
			expected: ConstructionValidations,
		},
		{
			name:     "construction submit endpoint",
			endpoint: meshcommon.ConstructionSubmitEndpoint,
			expected: ConstructionValidations,
		},
		{
			name:     "block endpoint",
			endpoint: meshcommon.BlockEndpoint,
			expected: NetworkValidations,
		},
		{
			name:     "block transaction endpoint",
			endpoint: meshcommon.BlockTransactionEndpoint,
			expected: NetworkValidations,
		},
		{
			name:     "mempool endpoint",
			endpoint: meshcommon.MempoolEndpoint,
			expected: NetworkValidations,
		},
		{
			name:     "mempool transaction endpoint",
			endpoint: meshcommon.MempoolTransactionEndpoint,
			expected: NetworkValidations,
		},
		{
			name:     "events blocks endpoint",
			endpoint: meshcommon.EventsBlocksEndpoint,
			expected: NetworkValidations,
		},
		{
			name:     "search transactions endpoint",
			endpoint: meshcommon.SearchTransactionsEndpoint,
			expected: NetworkValidations,
		},
		{
			name:     "unknown endpoint",
			endpoint: "/unknown/endpoint",
			expected: NetworkValidations, // Default fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetValidationsForEndpoint(tt.endpoint)
			if len(result) != len(tt.expected) {
				t.Errorf("GetValidationsForEndpoint() returned %d validations, want %d", len(result), len(tt.expected))
			}
			// Check that the validations are the same
			for i, validation := range result {
				if validation != tt.expected[i] {
					t.Errorf("GetValidationsForEndpoint() validation[%d] = %v, want %v", i, validation, tt.expected[i])
				}
			}
		})
	}
}

func TestEndpointValidationSets(t *testing.T) {
	// Test that all expected endpoints are in the validation sets
	expectedEndpoints := []string{
		meshcommon.NetworkListEndpoint,
		meshcommon.NetworkStatusEndpoint,
		meshcommon.NetworkOptionsEndpoint,
		meshcommon.AccountBalanceEndpoint,
		meshcommon.ConstructionDeriveEndpoint,
		meshcommon.ConstructionPreprocessEndpoint,
		meshcommon.ConstructionMetadataEndpoint,
		meshcommon.ConstructionPayloadsEndpoint,
		meshcommon.ConstructionParseEndpoint,
		meshcommon.ConstructionCombineEndpoint,
		meshcommon.ConstructionHashEndpoint,
		meshcommon.ConstructionSubmitEndpoint,
		meshcommon.BlockEndpoint,
		meshcommon.BlockTransactionEndpoint,
		meshcommon.MempoolEndpoint,
		meshcommon.MempoolTransactionEndpoint,
		meshcommon.EventsBlocksEndpoint,
		meshcommon.SearchTransactionsEndpoint,
	}

	for _, endpoint := range expectedEndpoints {
		if _, exists := EndpointValidationSets[endpoint]; !exists {
			t.Errorf("EndpointValidationSets missing endpoint: %s", endpoint)
		}
	}
}
