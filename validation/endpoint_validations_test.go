package validation

import (
	"testing"
)

func TestGetValidationsForEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		expected []ValidationType
	}{
		{
			name:     "network list endpoint",
			endpoint: "/network/list",
			expected: NetworkListValidations,
		},
		{
			name:     "network status endpoint",
			endpoint: "/network/status",
			expected: NetworkValidations,
		},
		{
			name:     "network options endpoint",
			endpoint: "/network/options",
			expected: NetworkValidations,
		},
		{
			name:     "account balance endpoint",
			endpoint: "/account/balance",
			expected: AccountBalanceValidations,
		},
		{
			name:     "construction derive endpoint",
			endpoint: "/construction/derive",
			expected: ConstructionValidations,
		},
		{
			name:     "construction preprocess endpoint",
			endpoint: "/construction/preprocess",
			expected: ConstructionValidations,
		},
		{
			name:     "construction metadata endpoint",
			endpoint: "/construction/metadata",
			expected: ConstructionValidations,
		},
		{
			name:     "construction payloads endpoint",
			endpoint: "/construction/payloads",
			expected: ConstructionPayloadsValidations,
		},
		{
			name:     "construction parse endpoint",
			endpoint: "/construction/parse",
			expected: ConstructionValidations,
		},
		{
			name:     "construction combine endpoint",
			endpoint: "/construction/combine",
			expected: ConstructionValidations,
		},
		{
			name:     "construction hash endpoint",
			endpoint: "/construction/hash",
			expected: ConstructionValidations,
		},
		{
			name:     "construction submit endpoint",
			endpoint: "/construction/submit",
			expected: ConstructionValidations,
		},
		{
			name:     "block endpoint",
			endpoint: "/block",
			expected: NetworkValidations,
		},
		{
			name:     "block transaction endpoint",
			endpoint: "/block/transaction",
			expected: NetworkValidations,
		},
		{
			name:     "mempool endpoint",
			endpoint: "/mempool",
			expected: NetworkValidations,
		},
		{
			name:     "mempool transaction endpoint",
			endpoint: "/mempool/transaction",
			expected: NetworkValidations,
		},
		{
			name:     "events blocks endpoint",
			endpoint: "/events/blocks",
			expected: NetworkValidations,
		},
		{
			name:     "search transactions endpoint",
			endpoint: "/search/transactions",
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
		"/network/list",
		"/network/status",
		"/network/options",
		"/account/balance",
		"/construction/derive",
		"/construction/preprocess",
		"/construction/metadata",
		"/construction/payloads",
		"/construction/parse",
		"/construction/combine",
		"/construction/hash",
		"/construction/submit",
		"/block",
		"/block/transaction",
		"/mempool",
		"/mempool/transaction",
		"/events/blocks",
		"/search/transactions",
	}

	for _, endpoint := range expectedEndpoints {
		if _, exists := EndpointValidationSets[endpoint]; !exists {
			t.Errorf("EndpointValidationSets missing endpoint: %s", endpoint)
		}
	}
}
