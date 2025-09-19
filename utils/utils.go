package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"slices"
	"strings"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/crypto"
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

// ComputeAddress computes address from public key
func ComputeAddress(publicKey *types.PublicKey) (string, error) {
	pubKey, err := crypto.DecompressPubkey(publicKey.Bytes)
	if err != nil {
		return "", err
	}
	address := crypto.PubkeyToAddress(*pubKey)
	return strings.ToLower(address.Hex()), nil
}

// GetTxOrigins extracts origin addresses from operations
func GetTxOrigins(operations []*types.Operation) []string {
	var origins []string

	for _, op := range operations {
		if op.Account == nil || op.Account.Address == "" {
			continue
		}

		address := strings.ToLower(op.Account.Address)

		// Check if address already exists
		if slices.Contains(origins, address) {
			continue
		}

		// Consider Fee operations
		if op.Type == "Fee" {
			origins = append(origins, address)
			continue
		}

		// Consider Transfer operations with negative value (sending)
		if op.Type == "Transfer" && op.Amount != nil && op.Amount.Value != "" {
			// Parse amount value
			amount := new(big.Int)
			if _, ok := amount.SetString(op.Amount.Value, 10); ok && amount.Sign() < 0 {
				origins = append(origins, address)
			}
		}
	}

	return origins
}
