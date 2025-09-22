package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"slices"
	"strings"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/crypto"
)

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

// RemoveHexPrefix removes the "0x" prefix from a hex string if present
func RemoveHexPrefix(hexStr string) string {
	if len(hexStr) > 2 && hexStr[:2] == "0x" {
		return hexStr[2:]
	}
	return hexStr
}

// HexToDecimal converts hex string to decimal string
func HexToDecimal(hexStr string) (string, error) {
	cleanHex := RemoveHexPrefix(hexStr)

	bigInt := new(big.Int)
	bigInt, ok := bigInt.SetString(cleanHex, 16)
	if !ok {
		return "", fmt.Errorf("invalid hex string: %s", hexStr)
	}

	return bigInt.String(), nil
}

// StringPtr creates a string pointer
func StringPtr(s string) *string {
	return &s
}

// Helper functions to create pointers
func Int64Ptr(i int64) *int64 {
	return &i
}

// BoolPtr creates a bool pointer
func BoolPtr(b bool) *bool {
	return &b
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
		if op.Type == OperationTypeFee {
			origins = append(origins, address)
			continue
		}

		// Consider Transfer operations with negative value (sending)
		if op.Type == OperationTypeTransfer && op.Amount != nil && op.Amount.Value != "" {
			// Parse amount value
			amount := new(big.Int)
			if _, ok := amount.SetString(op.Amount.Value, 10); ok && amount.Sign() < 0 {
				origins = append(origins, address)
			}
		}
	}

	return origins
}

// DecodeHexStringWithPrefix removes the "0x" prefix (if present) and decodes the hex string to bytes
func DecodeHexStringWithPrefix(hexStr string) ([]byte, error) {
	cleanHex := RemoveHexPrefix(hexStr)

	return hex.DecodeString(cleanHex)
}
