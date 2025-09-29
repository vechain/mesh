package common

import "github.com/coinbase/rosetta-sdk-go/types"

// Operation types for VeChain
const (
	OperationTypeTransfer      = "Transfer"
	OperationTypeFee           = "Fee"
	OperationTypeFeeDelegation = "FeeDelegation"
	OperationTypeContractCall  = "ContractCall"
)

// Operation statuses for VeChain
const (
	OperationStatusNone      = "None"
	OperationStatusPending   = "Pending"
	OperationStatusSucceeded = "Succeeded"
	OperationStatusReverted  = "Reverted"
)

var (
	// VETCurrency represents the native VeChain token
	VETCurrency = &types.Currency{
		Symbol:   "VET",
		Decimals: 18,
	}

	// VTHOCurrency represents the VeChain Thor Energy token
	VTHOCurrency = &types.Currency{
		Symbol:   "VTHO",
		Decimals: 18,
		Metadata: map[string]any{
			"contractAddress": "0x0000000000000000000000000000456E65726779",
		},
	}
)
