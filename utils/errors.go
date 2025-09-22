package utils

import "github.com/coinbase/rosetta-sdk-go/types"

// Errors contains all the predefined Mesh errors for VeChain
var Errors = map[int]*types.Error{
	500: {Code: 500, Message: "Internal server error.", Retriable: true},
	1:   {Code: 1, Message: "Contract address not found in token list.", Retriable: false},
	3:   {Code: 3, Message: "Block identifier not found.", Retriable: true},
	4:   {Code: 4, Message: "Transaction identifier not found.", Retriable: true},
	5:   {Code: 5, Message: "Invalid public key parameter.", Retriable: false},
	6:   {Code: 6, Message: "Transaction has multiple origins.", Retriable: false},
	7:   {Code: 7, Message: "Transaction origin does not exist.", Retriable: false},
	8:   {Code: 8, Message: "Transaction has multiple delegators.", Retriable: false},
	9:   {Code: 9, Message: "No transfer operation involved.", Retriable: false},
	10:  {Code: 10, Message: "Unregistered token operations found.", Retriable: false},
	11:  {Code: 11, Message: "Error getting blockchain metadata.", Retriable: true},
	12:  {Code: 12, Message: "Invalid signed transaction parameter.", Retriable: false},
	13:  {Code: 13, Message: "Error submitting raw transaction.", Retriable: false},
	14:  {Code: 14, Message: "Invalid preprocess request.", Retriable: false},
	15:  {Code: 15, Message: "Invalid options array parameter.", Retriable: false},
	16:  {Code: 16, Message: "Invalid metadata object parameter.", Retriable: false},
	17:  {Code: 17, Message: "Unable to decode transaction parameter.", Retriable: false},
	18:  {Code: 18, Message: "Invalid request parameters.", Retriable: false},
	19:  {Code: 19, Message: "Invalid unsigned transaction parameter.", Retriable: false},
	20:  {Code: 20, Message: "Invalid combine request parameters.", Retriable: false},
	21:  {Code: 21, Message: "Invalid blocks request parameters.", Retriable: false},
	22:  {Code: 22, Message: "Invalid network identifier parameter.", Retriable: false},
	23:  {Code: 23, Message: "Invalid account identifier parameter.", Retriable: false},
	24:  {Code: 24, Message: "Invalid block identifier parameter.", Retriable: false},
	25:  {Code: 25, Message: "Invalid transaction identifier parameter.", Retriable: false},
	26:  {Code: 26, Message: "API does not support offline mode.", Retriable: false},
	27:  {Code: 27, Message: "Contract was not created at the specified block identifier.", Retriable: false},
	28:  {Code: 28, Message: "Delegator public key not set.", Retriable: false},
	29:  {Code: 29, Message: "Operation account and public key do not match.", Retriable: false},
	30:  {Code: 30, Message: "Origin public key not set.", Retriable: false},
	31:  {Code: 31, Message: "Invalid currencies parameter.", Retriable: false},
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
