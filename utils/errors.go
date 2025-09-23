package utils

import "github.com/coinbase/rosetta-sdk-go/types"

// Error codes for Mesh errors
const (
	ErrInternalServerError                  = 500
	ErrContractAddressNotFound              = 1
	ErrBlockIdentifierNotFound              = 3
	ErrTransactionIdentifierNotFound        = 4
	ErrInvalidPublicKeyParameter            = 5
	ErrTransactionMultipleOrigins           = 6
	ErrTransactionOriginNotExist            = 7
	ErrTransactionMultipleDelegators        = 8
	ErrNoTransferOperation                  = 9
	ErrUnregisteredTokenOperations          = 10
	ErrGettingBlockchainMetadata            = 11
	ErrInvalidSignedTransactionParameter    = 12
	ErrSubmittingRawTransaction             = 13
	ErrInvalidPreprocessRequest             = 14
	ErrInvalidOptionsArrayParameter         = 15
	ErrInvalidMetadataObjectParameter       = 16
	ErrUnableToDecodeTransactionParameter   = 17
	ErrInvalidRequestParameters             = 18
	ErrInvalidUnsignedTransactionParameter  = 19
	ErrInvalidCombineRequestParameters      = 20
	ErrInvalidBlocksRequestParameters       = 21
	ErrInvalidNetworkIdentifierParameter    = 22
	ErrInvalidAccountIdentifierParameter    = 23
	ErrInvalidBlockIdentifierParameter      = 24
	ErrAPIDoesNotSupportOfflineMode         = 26
	ErrContractNotCreatedAtBlockIdentifier  = 27
	ErrDelegatorPublicKeyNotSet             = 28
	ErrOperationAccountAndPublicKeyMismatch = 29
	ErrOriginPublicKeyNotSet                = 30
	ErrInvalidCurrenciesParameter           = 31
	ErrInvalidRequestBody                   = 32
	ErrFailedToGetBestBlock                 = 33
	ErrFailedToGetGenesisBlock              = 34
	ErrFailedToGetSyncProgress              = 35
	ErrFailedToGetPeers                     = 36
	ErrFailedToGetAccount                   = 37
	ErrFailedToEncodeResponse               = 38
	ErrPublicKeyRequired                    = 39
	ErrInvalidPublicKeyFormat               = 40
	ErrOriginAddressMismatch                = 41
	ErrDelegatorAddressMismatch             = 42
	ErrInvalidTransactionHex                = 43
	ErrFailedToDecodeTransaction            = 44
	ErrFailedToDecodeUnsignedTransaction    = 45
	ErrInvalidNumberOfSignatures            = 46
	ErrFailedToEncodeSignedTransaction      = 47
	ErrFailedToDecodeMeshTransaction        = 48
	ErrFailedToBuildThorTransaction         = 49
	ErrFailedToEncodeTransaction            = 50
	ErrFailedToSubmitTransaction            = 51
	ErrBlockNotFound                        = 52
	ErrTransactionNotFound                  = 53
	ErrFailedToConvertVETBalance            = 54
	ErrFailedToConvertVTHOBalance           = 55
	ErrTransactionNotFoundInMempool         = 56
	ErrFailedToGetMempool                   = 57
	ErrInvalidTransactionIdentifier         = 58
	ErrInvalidTransactionHash               = 59
	ErrInvalidCurrency                      = 60
)

// Errors contains all the predefined Mesh errors for VeChain
var Errors = map[int]*types.Error{
	ErrInternalServerError:                  {Code: ErrInternalServerError, Message: "Internal server error.", Retriable: true},
	ErrContractAddressNotFound:              {Code: ErrContractAddressNotFound, Message: "Contract address not found in token list.", Retriable: false},
	ErrBlockIdentifierNotFound:              {Code: ErrBlockIdentifierNotFound, Message: "Block identifier not found.", Retriable: true},
	ErrTransactionIdentifierNotFound:        {Code: ErrTransactionIdentifierNotFound, Message: "Transaction identifier not found.", Retriable: true},
	ErrInvalidPublicKeyParameter:            {Code: ErrInvalidPublicKeyParameter, Message: "Invalid public key parameter.", Retriable: false},
	ErrTransactionMultipleOrigins:           {Code: ErrTransactionMultipleOrigins, Message: "Transaction has multiple origins.", Retriable: false},
	ErrTransactionOriginNotExist:            {Code: ErrTransactionOriginNotExist, Message: "Transaction origin does not exist.", Retriable: false},
	ErrTransactionMultipleDelegators:        {Code: ErrTransactionMultipleDelegators, Message: "Transaction has multiple delegators.", Retriable: false},
	ErrNoTransferOperation:                  {Code: ErrNoTransferOperation, Message: "No transfer operation involved.", Retriable: false},
	ErrUnregisteredTokenOperations:          {Code: ErrUnregisteredTokenOperations, Message: "Unregistered token operations found.", Retriable: false},
	ErrGettingBlockchainMetadata:            {Code: ErrGettingBlockchainMetadata, Message: "Error getting blockchain metadata.", Retriable: true},
	ErrInvalidSignedTransactionParameter:    {Code: ErrInvalidSignedTransactionParameter, Message: "Invalid signed transaction parameter.", Retriable: false},
	ErrSubmittingRawTransaction:             {Code: ErrSubmittingRawTransaction, Message: "Error submitting raw transaction.", Retriable: false},
	ErrInvalidPreprocessRequest:             {Code: ErrInvalidPreprocessRequest, Message: "Invalid preprocess request.", Retriable: false},
	ErrInvalidOptionsArrayParameter:         {Code: ErrInvalidOptionsArrayParameter, Message: "Invalid options array parameter.", Retriable: false},
	ErrInvalidMetadataObjectParameter:       {Code: ErrInvalidMetadataObjectParameter, Message: "Invalid metadata object parameter.", Retriable: false},
	ErrUnableToDecodeTransactionParameter:   {Code: ErrUnableToDecodeTransactionParameter, Message: "Unable to decode transaction parameter.", Retriable: false},
	ErrInvalidRequestParameters:             {Code: ErrInvalidRequestParameters, Message: "Invalid request parameters.", Retriable: false},
	ErrInvalidUnsignedTransactionParameter:  {Code: ErrInvalidUnsignedTransactionParameter, Message: "Invalid unsigned transaction parameter.", Retriable: false},
	ErrInvalidCombineRequestParameters:      {Code: ErrInvalidCombineRequestParameters, Message: "Invalid combine request parameters.", Retriable: false},
	ErrInvalidBlocksRequestParameters:       {Code: ErrInvalidBlocksRequestParameters, Message: "Invalid blocks request parameters.", Retriable: false},
	ErrInvalidNetworkIdentifierParameter:    {Code: ErrInvalidNetworkIdentifierParameter, Message: "Invalid network identifier parameter.", Retriable: false},
	ErrInvalidAccountIdentifierParameter:    {Code: ErrInvalidAccountIdentifierParameter, Message: "Invalid account identifier parameter.", Retriable: false},
	ErrInvalidBlockIdentifierParameter:      {Code: ErrInvalidBlockIdentifierParameter, Message: "Invalid block identifier parameter.", Retriable: false},
	ErrAPIDoesNotSupportOfflineMode:         {Code: ErrAPIDoesNotSupportOfflineMode, Message: "API does not support offline mode.", Retriable: false},
	ErrContractNotCreatedAtBlockIdentifier:  {Code: ErrContractNotCreatedAtBlockIdentifier, Message: "Contract was not created at the specified block identifier.", Retriable: false},
	ErrDelegatorPublicKeyNotSet:             {Code: ErrDelegatorPublicKeyNotSet, Message: "Delegator public key not set.", Retriable: false},
	ErrOperationAccountAndPublicKeyMismatch: {Code: ErrOperationAccountAndPublicKeyMismatch, Message: "Operation account and public key do not match.", Retriable: false},
	ErrOriginPublicKeyNotSet:                {Code: ErrOriginPublicKeyNotSet, Message: "Origin public key not set.", Retriable: false},
	ErrInvalidCurrenciesParameter:           {Code: ErrInvalidCurrenciesParameter, Message: "Invalid currencies parameter.", Retriable: false},
	ErrInvalidRequestBody:                   {Code: ErrInvalidRequestBody, Message: "Invalid request body.", Retriable: false},
	ErrFailedToGetBestBlock:                 {Code: ErrFailedToGetBestBlock, Message: "Failed to get best block.", Retriable: true},
	ErrFailedToGetGenesisBlock:              {Code: ErrFailedToGetGenesisBlock, Message: "Failed to get genesis block.", Retriable: true},
	ErrFailedToGetSyncProgress:              {Code: ErrFailedToGetSyncProgress, Message: "Failed to get sync progress.", Retriable: true},
	ErrFailedToGetPeers:                     {Code: ErrFailedToGetPeers, Message: "Failed to get peers.", Retriable: true},
	ErrFailedToGetAccount:                   {Code: ErrFailedToGetAccount, Message: "Failed to get account.", Retriable: true},
	ErrFailedToEncodeResponse:               {Code: ErrFailedToEncodeResponse, Message: "Failed to encode response.", Retriable: false},
	ErrPublicKeyRequired:                    {Code: ErrPublicKeyRequired, Message: "Public key is required.", Retriable: false},
	ErrInvalidPublicKeyFormat:               {Code: ErrInvalidPublicKeyFormat, Message: "Invalid public key format.", Retriable: false},
	ErrOriginAddressMismatch:                {Code: ErrOriginAddressMismatch, Message: "Origin address mismatch.", Retriable: false},
	ErrDelegatorAddressMismatch:             {Code: ErrDelegatorAddressMismatch, Message: "Delegator address mismatch.", Retriable: false},
	ErrInvalidTransactionHex:                {Code: ErrInvalidTransactionHex, Message: "Invalid transaction hex.", Retriable: false},
	ErrFailedToDecodeTransaction:            {Code: ErrFailedToDecodeTransaction, Message: "Failed to decode transaction.", Retriable: false},
	ErrFailedToDecodeUnsignedTransaction:    {Code: ErrFailedToDecodeUnsignedTransaction, Message: "Failed to decode unsigned transaction.", Retriable: false},
	ErrInvalidNumberOfSignatures:            {Code: ErrInvalidNumberOfSignatures, Message: "Invalid number of signatures.", Retriable: false},
	ErrFailedToEncodeSignedTransaction:      {Code: ErrFailedToEncodeSignedTransaction, Message: "Failed to encode signed transaction.", Retriable: false},
	ErrFailedToDecodeMeshTransaction:        {Code: ErrFailedToDecodeMeshTransaction, Message: "Failed to decode Mesh transaction.", Retriable: false},
	ErrFailedToBuildThorTransaction:         {Code: ErrFailedToBuildThorTransaction, Message: "Failed to build Thor transaction.", Retriable: false},
	ErrFailedToEncodeTransaction:            {Code: ErrFailedToEncodeTransaction, Message: "Failed to encode transaction.", Retriable: false},
	ErrFailedToSubmitTransaction:            {Code: ErrFailedToSubmitTransaction, Message: "Failed to submit transaction.", Retriable: false},
	ErrBlockNotFound:                        {Code: ErrBlockNotFound, Message: "Block not found.", Retriable: true},
	ErrTransactionNotFound:                  {Code: ErrTransactionNotFound, Message: "Transaction not found.", Retriable: true},
	ErrFailedToConvertVETBalance:            {Code: ErrFailedToConvertVETBalance, Message: "Failed to convert VET balance.", Retriable: false},
	ErrFailedToConvertVTHOBalance:           {Code: ErrFailedToConvertVTHOBalance, Message: "Failed to convert VTHO balance.", Retriable: false},
	ErrTransactionNotFoundInMempool:         {Code: ErrTransactionNotFoundInMempool, Message: "Transaction not found in mempool.", Retriable: true},
	ErrFailedToGetMempool:                   {Code: ErrFailedToGetMempool, Message: "Failed to get mempool data.", Retriable: true},
	ErrInvalidTransactionIdentifier:         {Code: ErrInvalidTransactionIdentifier, Message: "Invalid transaction identifier.", Retriable: false},
	ErrInvalidTransactionHash:               {Code: ErrInvalidTransactionHash, Message: "Invalid transaction hash.", Retriable: false},
	ErrInvalidCurrency:                      {Code: ErrInvalidCurrency, Message: "Invalid currency format.", Retriable: false},
}

// GetError returns an error by code, or nil if not found
func GetError(code int) *types.Error {
	if err, exists := Errors[code]; exists {
		return err
	}
	return nil
}

// AllErrors is a pre-computed slice of all errors for efficiency
var AllErrors []*types.Error

// init initializes the AllErrors slice once at package load time
func init() {
	AllErrors = make([]*types.Error, 0, len(Errors))
	for _, err := range Errors {
		AllErrors = append(AllErrors, err)
	}
}

// GetAllErrors returns all errors as a slice (now just returns the pre-computed slice)
func GetAllErrors() []*types.Error {
	return AllErrors
}

// GetErrorWithMetadata returns an error by code with optional metadata, similar to the reference implementation
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
