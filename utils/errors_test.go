package utils

import (
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
)

func TestGetError(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected *types.Error
	}{
		{
			name:     "valid error code",
			code:     ErrInternalServerError,
			expected: &types.Error{Code: ErrInternalServerError, Message: "Internal server error.", Retriable: true},
		},
		{
			name:     "another valid error code",
			code:     ErrInvalidPublicKeyParameter,
			expected: &types.Error{Code: ErrInvalidPublicKeyParameter, Message: "Invalid public key parameter.", Retriable: false},
		},
		{
			name:     "non-existent error code",
			code:     9999,
			expected: nil,
		},
		{
			name:     "zero error code",
			code:     0,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetError(tt.code)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("GetError() = %v, want nil", result)
				}
			} else {
				if result == nil {
					t.Errorf("GetError() = nil, want %v", tt.expected)
					return
				}

				if result.Code != tt.expected.Code {
					t.Errorf("GetError() Code = %v, want %v", result.Code, tt.expected.Code)
				}
				if result.Message != tt.expected.Message {
					t.Errorf("GetError() Message = %v, want %v", result.Message, tt.expected.Message)
				}
				if result.Retriable != tt.expected.Retriable {
					t.Errorf("GetError() Retriable = %v, want %v", result.Retriable, tt.expected.Retriable)
				}
			}
		})
	}
}

func TestGetAllErrors(t *testing.T) {
	allErrors := GetAllErrors()

	// Check that we get all errors
	expectedCount := len(Errors)
	if len(allErrors) != expectedCount {
		t.Errorf("GetAllErrors() returned %d errors, want %d", len(allErrors), expectedCount)
	}

	// Check that all errors are present
	errorCodes := make(map[int32]bool)
	for _, err := range allErrors {
		errorCodes[err.Code] = true
	}

	for code := range Errors {
		if !errorCodes[int32(code)] {
			t.Errorf("GetAllErrors() missing error code %d", code)
		}
	}
}

func TestGetErrorWithMetadata(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		metadata map[string]any
		expected *types.Error
	}{
		{
			name: "valid error with metadata",
			code: ErrInternalServerError,
			metadata: map[string]any{
				"details": "test details",
				"field":   "test field",
			},
			expected: &types.Error{
				Code:      ErrInternalServerError,
				Message:   "Internal server error.",
				Retriable: true,
				Details: map[string]any{
					"details": "test details",
					"field":   "test field",
				},
			},
		},
		{
			name:     "valid error with nil metadata",
			code:     ErrInvalidPublicKeyParameter,
			metadata: nil,
			expected: &types.Error{
				Code:      ErrInvalidPublicKeyParameter,
				Message:   "Invalid public key parameter.",
				Retriable: false,
				Details:   nil,
			},
		},
		{
			name:     "valid error with empty metadata",
			code:     ErrInvalidPublicKeyParameter,
			metadata: map[string]any{},
			expected: &types.Error{
				Code:      ErrInvalidPublicKeyParameter,
				Message:   "Invalid public key parameter.",
				Retriable: false,
				Details:   map[string]any{},
			},
		},
		{
			name:     "non-existent error code",
			code:     9999,
			metadata: map[string]any{"test": "value"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorWithMetadata(tt.code, tt.metadata)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("GetErrorWithMetadata() = %v, want nil", result)
				}
			} else {
				if result == nil {
					t.Errorf("GetErrorWithMetadata() = nil, want %v", tt.expected)
					return
				}

				if result.Code != tt.expected.Code {
					t.Errorf("GetErrorWithMetadata() Code = %v, want %v", result.Code, tt.expected.Code)
				}
				if result.Message != tt.expected.Message {
					t.Errorf("GetErrorWithMetadata() Message = %v, want %v", result.Message, tt.expected.Message)
				}
				if result.Retriable != tt.expected.Retriable {
					t.Errorf("GetErrorWithMetadata() Retriable = %v, want %v", result.Retriable, tt.expected.Retriable)
				}

				// Check metadata
				if tt.expected.Details == nil {
					if result.Details != nil {
						t.Errorf("GetErrorWithMetadata() Details = %v, want nil", result.Details)
					}
				} else {
					if result.Details == nil {
						t.Errorf("GetErrorWithMetadata() Details = nil, want %v", tt.expected.Details)
					} else {
						// Compare metadata maps
						if len(result.Details) != len(tt.expected.Details) {
							t.Errorf("GetErrorWithMetadata() Details length = %v, want %v", len(result.Details), len(tt.expected.Details))
						}
						for key, expectedValue := range tt.expected.Details {
							if result.Details[key] != expectedValue {
								t.Errorf("GetErrorWithMetadata() Details[%s] = %v, want %v", key, result.Details[key], expectedValue)
							}
						}
					}
				}
			}
		})
	}
}

func TestErrorConstants(t *testing.T) {
	// Test that all error constants are unique
	errorCodes := make(map[int]bool)

	allCodes := []int{
		ErrInternalServerError,
		ErrContractAddressNotFound,
		ErrBlockIdentifierNotFound,
		ErrTransactionIdentifierNotFound,
		ErrInvalidPublicKeyParameter,
		ErrTransactionMultipleOrigins,
		ErrTransactionOriginNotExist,
		ErrTransactionMultipleDelegators,
		ErrNoTransferOperation,
		ErrUnregisteredTokenOperations,
		ErrGettingBlockchainMetadata,
		ErrInvalidSignedTransactionParameter,
		ErrSubmittingRawTransaction,
		ErrInvalidPreprocessRequest,
		ErrInvalidOptionsArrayParameter,
		ErrInvalidMetadataObjectParameter,
		ErrUnableToDecodeTransactionParameter,
		ErrInvalidRequestParameters,
		ErrInvalidUnsignedTransactionParameter,
		ErrInvalidCombineRequestParameters,
		ErrInvalidBlocksRequestParameters,
		ErrInvalidNetworkIdentifierParameter,
		ErrInvalidAccountIdentifierParameter,
		ErrInvalidBlockIdentifierParameter,
		ErrAPIDoesNotSupportOfflineMode,
		ErrContractNotCreatedAtBlockIdentifier,
		ErrDelegatorPublicKeyNotSet,
		ErrOperationAccountAndPublicKeyMismatch,
		ErrOriginPublicKeyNotSet,
		ErrInvalidCurrenciesParameter,
		ErrInvalidRequestBody,
		ErrFailedToGetBestBlock,
		ErrFailedToGetGenesisBlock,
		ErrFailedToGetSyncProgress,
		ErrFailedToGetPeers,
		ErrFailedToGetAccount,
		ErrFailedToEncodeResponse,
		ErrPublicKeyRequired,
		ErrInvalidPublicKeyFormat,
		ErrOriginAddressMismatch,
		ErrDelegatorAddressMismatch,
		ErrInvalidTransactionHex,
		ErrFailedToDecodeTransaction,
		ErrFailedToDecodeUnsignedTransaction,
		ErrInvalidNumberOfSignatures,
		ErrFailedToEncodeSignedTransaction,
		ErrFailedToDecodeMeshTransaction,
		ErrFailedToBuildThorTransaction,
		ErrFailedToEncodeTransaction,
		ErrFailedToSubmitTransaction,
		ErrBlockNotFound,
		ErrTransactionNotFound,
		ErrFailedToConvertVETBalance,
		ErrFailedToConvertVTHOBalance,
		ErrTransactionNotFoundInMempool,
		ErrFailedToGetMempool,
		ErrInvalidTransactionIdentifier,
		ErrInvalidTransactionHash,
		ErrInvalidCurrency,
	}

	for _, code := range allCodes {
		if errorCodes[code] {
			t.Errorf("Duplicate error code found: %d", code)
		}
		errorCodes[code] = true
	}
}

func TestErrorsMap(t *testing.T) {
	// Test that all errors in the Errors map have the correct structure
	for code, err := range Errors {
		if err == nil {
			t.Errorf("Error with code %d is nil", code)
			continue
		}

		if err.Code != int32(code) {
			t.Errorf("Error code mismatch: expected %d, got %d", code, err.Code)
		}

		if err.Message == "" {
			t.Errorf("Error with code %d has empty message", code)
		}
	}
}

func TestAllErrorsSlice(t *testing.T) {
	// Test that AllErrors slice is properly initialized
	if AllErrors == nil {
		t.Errorf("AllErrors slice is nil")
		return
	}

	if len(AllErrors) != len(Errors) {
		t.Errorf("AllErrors length = %d, want %d", len(AllErrors), len(Errors))
	}

	// Test that all errors in AllErrors are also in Errors map
	errorCodes := make(map[int32]bool)
	for _, err := range AllErrors {
		errorCodes[err.Code] = true
	}

	for code := range Errors {
		if !errorCodes[int32(code)] {
			t.Errorf("AllErrors missing error code %d", code)
		}
	}
}
