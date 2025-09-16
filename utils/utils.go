package utils

import (
	"fmt"
	"math/big"
)

// HexToDecimal converts hex string to decimal string
func HexToDecimal(hexStr string) (string, error) {
	if len(hexStr) > 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}

	bigInt := new(big.Int)
	bigInt, ok := bigInt.SetString(hexStr, 16)
	if !ok {
		return "", fmt.Errorf("invalid hex string: %s", hexStr)
	}

	return bigInt.String(), nil
}

// StringToBigInt converts a string to big.Int
func StringToBigInt(value string, base int) (*big.Int, error) {
	bigInt := new(big.Int)
	bigInt, ok := bigInt.SetString(value, base)
	if !ok {
		return nil, fmt.Errorf("invalid string for big.Int conversion: %s (base %d)", value, base)
	}
	return bigInt, nil
}

// StringPtr creates a string pointer
func StringPtr(s string) *string {
	return &s
}
