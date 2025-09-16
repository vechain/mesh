package utils

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
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

// WriteJSONResponse writes a JSON response with proper error handling
func WriteJSONResponse(w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
