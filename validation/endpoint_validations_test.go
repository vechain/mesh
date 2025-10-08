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
			runMode:  meshcommon.OnlineMode,
			expected: NetworkListValidations,
		},
		{
			name:     "network status endpoint online",
			endpoint: meshcommon.NetworkStatusEndpoint,
			runMode:  meshcommon.OnlineMode,
			expected: NetworkValidations,
		},
		{
			name:     "network options endpoint online",
			endpoint: meshcommon.NetworkOptionsEndpoint,
			runMode:  meshcommon.OnlineMode,
			expected: NetworkValidations,
		},
		{
			name:     "account balance endpoint online",
			endpoint: meshcommon.AccountBalanceEndpoint,
			runMode:  meshcommon.OnlineMode,
			expected: AccountBalanceValidations,
		},
		{
			name:     "construction derive endpoint online",
			endpoint: meshcommon.ConstructionDeriveEndpoint,
			runMode:  meshcommon.OnlineMode,
			expected: ConstructionValidations,
		},
		{
			name:     "construction preprocess endpoint online",
			endpoint: meshcommon.ConstructionPreprocessEndpoint,
			runMode:  meshcommon.OnlineMode,
			expected: ConstructionValidations,
		},
		{
			name:     "construction metadata endpoint online",
			endpoint: meshcommon.ConstructionMetadataEndpoint,
			runMode:  meshcommon.OnlineMode,
			expected: ConstructionValidations,
		},
		{
			name:     "construction payloads endpoint online",
			endpoint: meshcommon.ConstructionPayloadsEndpoint,
			runMode:  meshcommon.OnlineMode,
			expected: ConstructionPayloadsValidations,
		},
		{
			name:     "construction parse endpoint online",
			endpoint: meshcommon.ConstructionParseEndpoint,
			runMode:  meshcommon.OnlineMode,
			expected: ConstructionValidations,
		},
		{
			name:     "construction combine endpoint online",
			endpoint: meshcommon.ConstructionCombineEndpoint,
			runMode:  meshcommon.OnlineMode,
			expected: ConstructionValidations,
		},
		{
			name:     "construction hash endpoint online",
			endpoint: meshcommon.ConstructionHashEndpoint,
			runMode:  meshcommon.OnlineMode,
			expected: ConstructionValidations,
		},
		{
			name:     "construction submit endpoint online",
			endpoint: meshcommon.ConstructionSubmitEndpoint,
			runMode:  meshcommon.OnlineMode,
			expected: ConstructionValidations,
		},
		{
			name:     "block endpoint online",
			endpoint: meshcommon.BlockEndpoint,
			runMode:  meshcommon.OnlineMode,
			expected: NetworkValidations,
		},
		{
			name:     "unknown endpoint online",
			endpoint: "/unknown/endpoint",
			runMode:  meshcommon.OnlineMode,
			expected: NetworkValidations, // Default fallback
		},
		// Offline mode tests
		{
			name:     "network list endpoint offline",
			endpoint: meshcommon.NetworkListEndpoint,
			runMode:  meshcommon.OfflineMode,
			expected: NetworkListOfflineValidations,
		},
		{
			name:     "network options endpoint offline",
			endpoint: meshcommon.NetworkOptionsEndpoint,
			runMode:  meshcommon.OfflineMode,
			expected: NetworkOfflineValidations,
		},
		{
			name:     "construction derive endpoint offline",
			endpoint: meshcommon.ConstructionDeriveEndpoint,
			runMode:  meshcommon.OfflineMode,
			expected: ConstructionOfflineValidations,
		},
		{
			name:     "construction preprocess endpoint offline",
			endpoint: meshcommon.ConstructionPreprocessEndpoint,
			runMode:  meshcommon.OfflineMode,
			expected: ConstructionOfflineValidations,
		},
		{
			name:     "construction payloads endpoint offline",
			endpoint: meshcommon.ConstructionPayloadsEndpoint,
			runMode:  meshcommon.OfflineMode,
			expected: ConstructionPayloadsOfflineValidations,
		},
		{
			name:     "construction parse endpoint offline",
			endpoint: meshcommon.ConstructionParseEndpoint,
			runMode:  meshcommon.OfflineMode,
			expected: ConstructionOfflineValidations,
		},
		{
			name:     "construction combine endpoint offline",
			endpoint: meshcommon.ConstructionCombineEndpoint,
			runMode:  meshcommon.OfflineMode,
			expected: ConstructionOfflineValidations,
		},
		{
			name:     "construction hash endpoint offline",
			endpoint: meshcommon.ConstructionHashEndpoint,
			runMode:  meshcommon.OfflineMode,
			expected: ConstructionOfflineValidations,
		},
		{
			name:     "unsupported endpoint offline (should return online validations)",
			endpoint: meshcommon.AccountBalanceEndpoint,
			runMode:  meshcommon.OfflineMode,
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
