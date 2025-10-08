package validation

import (
	"testing"

	meshcommon "github.com/vechain/mesh/common"
)

func TestGetValidationsForEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		runMode  string
		expected []ValidationType
	}{
		// Online mode tests
		{
			name:     "network list endpoint online",
			endpoint: meshcommon.NetworkListEndpoint,
			runMode:  "online",
			expected: NetworkListValidations,
		},
		{
			name:     "network status endpoint online",
			endpoint: meshcommon.NetworkStatusEndpoint,
			runMode:  "online",
			expected: NetworkValidations,
		},
		{
			name:     "network options endpoint online",
			endpoint: meshcommon.NetworkOptionsEndpoint,
			runMode:  "online",
			expected: NetworkValidations,
		},
		{
			name:     "account balance endpoint online",
			endpoint: meshcommon.AccountBalanceEndpoint,
			runMode:  "online",
			expected: AccountBalanceValidations,
		},
		{
			name:     "construction derive endpoint online",
			endpoint: meshcommon.ConstructionDeriveEndpoint,
			runMode:  "online",
			expected: ConstructionValidations,
		},
		{
			name:     "construction preprocess endpoint online",
			endpoint: meshcommon.ConstructionPreprocessEndpoint,
			runMode:  "online",
			expected: ConstructionValidations,
		},
		{
			name:     "construction metadata endpoint online",
			endpoint: meshcommon.ConstructionMetadataEndpoint,
			runMode:  "online",
			expected: ConstructionValidations,
		},
		{
			name:     "construction payloads endpoint online",
			endpoint: meshcommon.ConstructionPayloadsEndpoint,
			runMode:  "online",
			expected: ConstructionPayloadsValidations,
		},
		{
			name:     "construction parse endpoint online",
			endpoint: meshcommon.ConstructionParseEndpoint,
			runMode:  "online",
			expected: ConstructionValidations,
		},
		{
			name:     "construction combine endpoint online",
			endpoint: meshcommon.ConstructionCombineEndpoint,
			runMode:  "online",
			expected: ConstructionValidations,
		},
		{
			name:     "construction hash endpoint online",
			endpoint: meshcommon.ConstructionHashEndpoint,
			runMode:  "online",
			expected: ConstructionValidations,
		},
		{
			name:     "construction submit endpoint online",
			endpoint: meshcommon.ConstructionSubmitEndpoint,
			runMode:  "online",
			expected: ConstructionValidations,
		},
		{
			name:     "block endpoint online",
			endpoint: meshcommon.BlockEndpoint,
			runMode:  "online",
			expected: NetworkValidations,
		},
		{
			name:     "unknown endpoint online",
			endpoint: "/unknown/endpoint",
			runMode:  "online",
			expected: NetworkValidations, // Default fallback
		},
		// Offline mode tests
		{
			name:     "network list endpoint offline",
			endpoint: meshcommon.NetworkListEndpoint,
			runMode:  "offline",
			expected: NetworkListOfflineValidations,
		},
		{
			name:     "network options endpoint offline",
			endpoint: meshcommon.NetworkOptionsEndpoint,
			runMode:  "offline",
			expected: NetworkOfflineValidations,
		},
		{
			name:     "construction derive endpoint offline",
			endpoint: meshcommon.ConstructionDeriveEndpoint,
			runMode:  "offline",
			expected: ConstructionOfflineValidations,
		},
		{
			name:     "construction preprocess endpoint offline",
			endpoint: meshcommon.ConstructionPreprocessEndpoint,
			runMode:  "offline",
			expected: ConstructionOfflineValidations,
		},
		{
			name:     "construction payloads endpoint offline",
			endpoint: meshcommon.ConstructionPayloadsEndpoint,
			runMode:  "offline",
			expected: ConstructionPayloadsOfflineValidations,
		},
		{
			name:     "construction parse endpoint offline",
			endpoint: meshcommon.ConstructionParseEndpoint,
			runMode:  "offline",
			expected: ConstructionOfflineValidations,
		},
		{
			name:     "construction combine endpoint offline",
			endpoint: meshcommon.ConstructionCombineEndpoint,
			runMode:  "offline",
			expected: ConstructionOfflineValidations,
		},
		{
			name:     "construction hash endpoint offline",
			endpoint: meshcommon.ConstructionHashEndpoint,
			runMode:  "offline",
			expected: ConstructionOfflineValidations,
		},
		{
			name:     "unsupported endpoint offline (should return online validations)",
			endpoint: meshcommon.AccountBalanceEndpoint,
			runMode:  "offline",
			expected: AccountBalanceValidations, // Falls back to online validations
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetValidationsForEndpoint(tt.endpoint, tt.runMode)
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
