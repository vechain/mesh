package utils

import "github.com/coinbase/rosetta-sdk-go/types"

// Error codes for Mesh errors
const (
	ErrInternalServerError                   = 500
	ErrContractAddressNotFound               = 1
	ErrBlockIdentifierNotFound               = 3
	ErrTransactionIdentifierNotFound         = 4
	ErrInvalidPublicKeyParameter             = 5
	ErrTransactionMultipleOrigins            = 6
	ErrTransactionOriginNotExist             = 7
	ErrTransactionMultipleDelegators         = 8
	ErrNoTransferOperation                   = 9
	ErrUnregisteredTokenOperations           = 10
	ErrGettingBlockchainMetadata             = 11
	ErrInvalidSignedTransactionParameter     = 12
	ErrSubmittingRawTransaction              = 13
	ErrInvalidPreprocessRequest              = 14
	ErrInvalidOptionsArrayParameter          = 15
	ErrInvalidMetadataObjectParameter        = 16
	ErrUnableToDecodeTransactionParameter    = 17
	ErrInvalidRequestParameters              = 18
	ErrInvalidUnsignedTransactionParameter   = 19
	ErrInvalidCombineRequestParameters       = 20
	ErrInvalidBlocksRequestParameters        = 21
	ErrInvalidNetworkIdentifierParameter     = 22
	ErrInvalidAccountIdentifierParameter     = 23
	ErrInvalidBlockIdentifierParameter       = 24
	ErrInvalidTransactionIdentifierParameter = 25
	ErrAPIDoesNotSupportOfflineMode          = 26
	ErrContractNotCreatedAtBlockIdentifier   = 27
	ErrDelegatorPublicKeyNotSet              = 28
	ErrOperationAccountAndPublicKeyMismatch  = 29
	ErrOriginPublicKeyNotSet                 = 30
	ErrInvalidCurrenciesParameter            = 31
	ErrInvalidRequestBody                    = 32
	ErrFailedToGetBestBlock                  = 33
	ErrFailedToGetGenesisBlock               = 34
	ErrFailedToGetSyncProgress               = 35
	ErrFailedToGetPeers                      = 36
	ErrFailedToGetAccount                    = 37
	ErrFailedToEncodeResponse                = 38
)

// Errors contains all the predefined Mesh errors for VeChain
var Errors = map[int]*types.Error{
	ErrInternalServerError:                   {Code: ErrInternalServerError, Message: "Internal server error.", Retriable: true},
	ErrContractAddressNotFound:               {Code: ErrContractAddressNotFound, Message: "Contract address not found in token list.", Retriable: false},
	ErrBlockIdentifierNotFound:               {Code: ErrBlockIdentifierNotFound, Message: "Block identifier not found.", Retriable: true},
	ErrTransactionIdentifierNotFound:         {Code: ErrTransactionIdentifierNotFound, Message: "Transaction identifier not found.", Retriable: true},
	ErrInvalidPublicKeyParameter:             {Code: ErrInvalidPublicKeyParameter, Message: "Invalid public key parameter.", Retriable: false},
	ErrTransactionMultipleOrigins:            {Code: ErrTransactionMultipleOrigins, Message: "Transaction has multiple origins.", Retriable: false},
	ErrTransactionOriginNotExist:             {Code: ErrTransactionOriginNotExist, Message: "Transaction origin does not exist.", Retriable: false},
	ErrTransactionMultipleDelegators:         {Code: ErrTransactionMultipleDelegators, Message: "Transaction has multiple delegators.", Retriable: false},
	ErrNoTransferOperation:                   {Code: ErrNoTransferOperation, Message: "No transfer operation involved.", Retriable: false},
	ErrUnregisteredTokenOperations:           {Code: ErrUnregisteredTokenOperations, Message: "Unregistered token operations found.", Retriable: false},
	ErrGettingBlockchainMetadata:             {Code: ErrGettingBlockchainMetadata, Message: "Error getting blockchain metadata.", Retriable: true},
	ErrInvalidSignedTransactionParameter:     {Code: ErrInvalidSignedTransactionParameter, Message: "Invalid signed transaction parameter.", Retriable: false},
	ErrSubmittingRawTransaction:              {Code: ErrSubmittingRawTransaction, Message: "Error submitting raw transaction.", Retriable: false},
	ErrInvalidPreprocessRequest:              {Code: ErrInvalidPreprocessRequest, Message: "Invalid preprocess request.", Retriable: false},
	ErrInvalidOptionsArrayParameter:          {Code: ErrInvalidOptionsArrayParameter, Message: "Invalid options array parameter.", Retriable: false},
	ErrInvalidMetadataObjectParameter:        {Code: ErrInvalidMetadataObjectParameter, Message: "Invalid metadata object parameter.", Retriable: false},
	ErrUnableToDecodeTransactionParameter:    {Code: ErrUnableToDecodeTransactionParameter, Message: "Unable to decode transaction parameter.", Retriable: false},
	ErrInvalidRequestParameters:              {Code: ErrInvalidRequestParameters, Message: "Invalid request parameters.", Retriable: false},
	ErrInvalidUnsignedTransactionParameter:   {Code: ErrInvalidUnsignedTransactionParameter, Message: "Invalid unsigned transaction parameter.", Retriable: false},
	ErrInvalidCombineRequestParameters:       {Code: ErrInvalidCombineRequestParameters, Message: "Invalid combine request parameters.", Retriable: false},
	ErrInvalidBlocksRequestParameters:        {Code: ErrInvalidBlocksRequestParameters, Message: "Invalid blocks request parameters.", Retriable: false},
	ErrInvalidNetworkIdentifierParameter:     {Code: ErrInvalidNetworkIdentifierParameter, Message: "Invalid network identifier parameter.", Retriable: false},
	ErrInvalidAccountIdentifierParameter:     {Code: ErrInvalidAccountIdentifierParameter, Message: "Invalid account identifier parameter.", Retriable: false},
	ErrInvalidBlockIdentifierParameter:       {Code: ErrInvalidBlockIdentifierParameter, Message: "Invalid block identifier parameter.", Retriable: false},
	ErrInvalidTransactionIdentifierParameter: {Code: ErrInvalidTransactionIdentifierParameter, Message: "Invalid transaction identifier parameter.", Retriable: false},
	ErrAPIDoesNotSupportOfflineMode:          {Code: ErrAPIDoesNotSupportOfflineMode, Message: "API does not support offline mode.", Retriable: false},
	ErrContractNotCreatedAtBlockIdentifier:   {Code: ErrContractNotCreatedAtBlockIdentifier, Message: "Contract was not created at the specified block identifier.", Retriable: false},
	ErrDelegatorPublicKeyNotSet:              {Code: ErrDelegatorPublicKeyNotSet, Message: "Delegator public key not set.", Retriable: false},
	ErrOperationAccountAndPublicKeyMismatch:  {Code: ErrOperationAccountAndPublicKeyMismatch, Message: "Operation account and public key do not match.", Retriable: false},
	ErrOriginPublicKeyNotSet:                 {Code: ErrOriginPublicKeyNotSet, Message: "Origin public key not set.", Retriable: false},
	ErrInvalidCurrenciesParameter:            {Code: ErrInvalidCurrenciesParameter, Message: "Invalid currencies parameter.", Retriable: false},
	ErrInvalidRequestBody:                    {Code: ErrInvalidRequestBody, Message: "Invalid request body.", Retriable: false},
	ErrFailedToGetBestBlock:                  {Code: ErrFailedToGetBestBlock, Message: "Failed to get best block.", Retriable: true},
	ErrFailedToGetGenesisBlock:               {Code: ErrFailedToGetGenesisBlock, Message: "Failed to get genesis block.", Retriable: true},
	ErrFailedToGetSyncProgress:               {Code: ErrFailedToGetSyncProgress, Message: "Failed to get sync progress.", Retriable: true},
	ErrFailedToGetPeers:                      {Code: ErrFailedToGetPeers, Message: "Failed to get peers.", Retriable: true},
	ErrFailedToGetAccount:                    {Code: ErrFailedToGetAccount, Message: "Failed to get account.", Retriable: true},
	ErrFailedToEncodeResponse:                {Code: ErrFailedToEncodeResponse, Message: "Failed to encode response.", Retriable: false},
}

// GetError returns an error by code, or nil if not found
func GetError(code int) *types.Error {
	if err, exists := Errors[code]; exists {
		return err
	}
	return nil
}

// GetAllErrors returns all errors as a slice
func GetAllErrors() []*types.Error {
	errors := make([]*types.Error, 0, len(Errors))
	for _, err := range Errors {
		errors = append(errors, err)
	}
	return errors
}
