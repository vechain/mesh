package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
)

// WriteJSONResponse writes a JSON response with proper error handling
func WriteJSONResponse(w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GenerateNonce generates a random nonce
func GenerateNonce() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return fmt.Sprintf("0x%016x", bytes), nil
}

// GetStringFromOptions gets a string value from options map
func GetStringFromOptions(options map[string]any, key string, defaultValue string) string {
	if value, ok := options[key].(string); ok {
		return value
	}
	return defaultValue
}

// HasOperationType checks if operations contain a specific type
func HasOperationType(operations []any, operationType string) bool {
	for _, op := range operations {
		if opMap, ok := op.(map[string]any); ok {
			if opType, ok := opMap["type"].(string); ok && opType == operationType {
				return true
			}
		}
	}
	return false
}

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

// StringPtr creates a string pointer
func StringPtr(s string) *string {
	return &s
}
