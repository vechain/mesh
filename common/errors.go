package common

import "github.com/coinbase/rosetta-sdk-go/types"

// Error codes for Mesh errors
const (
	// General errors
	ErrInternalServerError = 500

	// Request validation errors
	ErrInvalidRequestParameters            = 1
	ErrInvalidRequestBody                  = 2
	ErrInvalidBlockIdentifierParameter     = 3
	ErrInvalidTransactionIdentifier        = 4
	ErrInvalidTransactionHash              = 5
	ErrInvalidCurrency                     = 6
	ErrInvalidPublicKeyFormat              = 7
	ErrPublicKeyRequired                   = 8
	ErrInvalidUnsignedTransactionParameter = 9
	ErrInvalidTransactionHex               = 10

	// Transaction building errors
	ErrTransactionMultipleOrigins = 11
	ErrTransactionOriginNotExist  = 12
	ErrNoTransferOperation        = 13
	ErrOriginAddressMismatch      = 14
	ErrDelegatorAddressMismatch   = 15
	ErrInvalidNumberOfSignatures  = 16

	// Encoding/Decoding errors
	ErrFailedToDecodeTransaction         = 17
	ErrFailedToDecodeUnsignedTransaction = 18
	ErrFailedToDecodeMeshTransaction     = 19
	ErrFailedToEncodeSignedTransaction   = 20
	ErrFailedToEncodeTransaction         = 21

	// Blockchain query errors
	ErrFailedToGetBestBlock      = 22
	ErrFailedToGetGenesisBlock   = 23
	ErrFailedToGetSyncProgress   = 24
	ErrFailedToGetPeers          = 25
	ErrFailedToGetAccount        = 26
	ErrFailedToGetMempool        = 27
	ErrGettingBlockchainMetadata = 28

	// Not found errors
	ErrBlockNotFound                = 29
	ErrTransactionNotFound          = 30
	ErrTransactionNotFoundInMempool = 31

	// Transaction submission errors
	ErrFailedToSubmitTransaction = 32

	// Mode errors
	ErrAPIDoesNotSupportOfflineMode = 33
)

// Errors contains all the predefined Mesh errors for VeChain
var Errors = map[int]*types.Error{
	// General errors
	ErrInternalServerError: {Code: ErrInternalServerError, Message: "Internal server error.", Retriable: true},

	// Request validation errors
	ErrInvalidRequestParameters:            {Code: ErrInvalidRequestParameters, Message: "Invalid request parameters.", Retriable: false},
	ErrInvalidRequestBody:                  {Code: ErrInvalidRequestBody, Message: "Invalid request body.", Retriable: false},
	ErrInvalidBlockIdentifierParameter:     {Code: ErrInvalidBlockIdentifierParameter, Message: "Invalid block identifier parameter.", Retriable: false},
	ErrInvalidTransactionIdentifier:        {Code: ErrInvalidTransactionIdentifier, Message: "Invalid transaction identifier.", Retriable: false},
	ErrInvalidTransactionHash:              {Code: ErrInvalidTransactionHash, Message: "Invalid transaction hash.", Retriable: false},
	ErrInvalidCurrency:                     {Code: ErrInvalidCurrency, Message: "Invalid currency format.", Retriable: false},
	ErrInvalidPublicKeyFormat:              {Code: ErrInvalidPublicKeyFormat, Message: "Invalid public key format.", Retriable: false},
	ErrPublicKeyRequired:                   {Code: ErrPublicKeyRequired, Message: "Public key is required.", Retriable: false},
	ErrInvalidUnsignedTransactionParameter: {Code: ErrInvalidUnsignedTransactionParameter, Message: "Invalid unsigned transaction parameter.", Retriable: false},
	ErrInvalidTransactionHex:               {Code: ErrInvalidTransactionHex, Message: "Invalid transaction hex.", Retriable: false},

	// Transaction building errors
	ErrTransactionMultipleOrigins: {Code: ErrTransactionMultipleOrigins, Message: "Transaction has multiple origins.", Retriable: false},
	ErrTransactionOriginNotExist:  {Code: ErrTransactionOriginNotExist, Message: "Transaction origin does not exist.", Retriable: false},
	ErrNoTransferOperation:        {Code: ErrNoTransferOperation, Message: "No transfer operation involved.", Retriable: false},
	ErrOriginAddressMismatch:      {Code: ErrOriginAddressMismatch, Message: "Origin address mismatch.", Retriable: false},
	ErrDelegatorAddressMismatch:   {Code: ErrDelegatorAddressMismatch, Message: "Delegator address mismatch.", Retriable: false},
	ErrInvalidNumberOfSignatures:  {Code: ErrInvalidNumberOfSignatures, Message: "Invalid number of signatures.", Retriable: false},

	// Encoding/Decoding errors
	ErrFailedToDecodeTransaction:         {Code: ErrFailedToDecodeTransaction, Message: "Failed to decode transaction.", Retriable: false},
	ErrFailedToDecodeUnsignedTransaction: {Code: ErrFailedToDecodeUnsignedTransaction, Message: "Failed to decode unsigned transaction.", Retriable: false},
	ErrFailedToDecodeMeshTransaction:     {Code: ErrFailedToDecodeMeshTransaction, Message: "Failed to decode Mesh transaction.", Retriable: false},
	ErrFailedToEncodeSignedTransaction:   {Code: ErrFailedToEncodeSignedTransaction, Message: "Failed to encode signed transaction.", Retriable: false},
	ErrFailedToEncodeTransaction:         {Code: ErrFailedToEncodeTransaction, Message: "Failed to encode transaction.", Retriable: false},

	// Blockchain query errors
	ErrFailedToGetBestBlock:      {Code: ErrFailedToGetBestBlock, Message: "Failed to get best block.", Retriable: true},
	ErrFailedToGetGenesisBlock:   {Code: ErrFailedToGetGenesisBlock, Message: "Failed to get genesis block.", Retriable: true},
	ErrFailedToGetSyncProgress:   {Code: ErrFailedToGetSyncProgress, Message: "Failed to get sync progress.", Retriable: true},
	ErrFailedToGetPeers:          {Code: ErrFailedToGetPeers, Message: "Failed to get peers.", Retriable: true},
	ErrFailedToGetAccount:        {Code: ErrFailedToGetAccount, Message: "Failed to get account.", Retriable: true},
	ErrFailedToGetMempool:        {Code: ErrFailedToGetMempool, Message: "Failed to get mempool data.", Retriable: true},
	ErrGettingBlockchainMetadata: {Code: ErrGettingBlockchainMetadata, Message: "Error getting blockchain metadata.", Retriable: true},

	// Not found errors
	ErrBlockNotFound:                {Code: ErrBlockNotFound, Message: "Block not found.", Retriable: true},
	ErrTransactionNotFound:          {Code: ErrTransactionNotFound, Message: "Transaction not found.", Retriable: true},
	ErrTransactionNotFoundInMempool: {Code: ErrTransactionNotFoundInMempool, Message: "Transaction not found in mempool.", Retriable: true},

	// Transaction submission errors
	ErrFailedToSubmitTransaction: {Code: ErrFailedToSubmitTransaction, Message: "Failed to submit transaction.", Retriable: false},

	// Mode errors
	ErrAPIDoesNotSupportOfflineMode: {Code: ErrAPIDoesNotSupportOfflineMode, Message: "API does not support offline mode.", Retriable: false},
}

// GetError returns an error by code, or nil if not found
func GetError(code int) *types.Error {
	if err, exists := Errors[code]; exists {
		return err
	}
	return nil
}

// allErrors is a pre-computed slice of all errors for efficiency
var allErrors []*types.Error

// init initializes the AllErrors slice once at package load time
func init() {
	allErrors = make([]*types.Error, 0, len(Errors))
	for _, err := range Errors {
		allErrors = append(allErrors, err)
	}
}

// GetAllErrors returns all errors as a slice (now just returns the pre-computed slice)
func GetAllErrors() []*types.Error {
	return allErrors
}

// GetErrorWithMetadata returns an error by code with optional metadata
func GetErrorWithMetadata(code int, metadata map[string]any) *types.Error {
	err := GetError(code)
	if err == nil {
		return nil
	}

	// Create a copy of the error to avoid modifying the original
	errorCopy := &types.Error{
		Code:      err.Code,
		Message:   err.Message,
		Retriable: err.Retriable,
		Details:   metadata,
	}

	return errorCopy
}
