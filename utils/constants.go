package utils

import "github.com/coinbase/rosetta-sdk-go/types"


// Context key for request body
type RequestBodyKeyType string
const RequestBodyKey RequestBodyKeyType = "request_body"

// Operation types for VeChain
const (
	OperationTypeNone          = "None"
	OperationTypeTransfer      = "Transfer"
	OperationTypeFee           = "Fee"
	OperationTypeFeeDelegation = "FeeDelegation"
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