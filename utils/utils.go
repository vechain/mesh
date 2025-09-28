package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/vechain/mesh/thor"
	"github.com/vechain/mesh/utils/vip180"
)

// WriteJSONResponse writes a JSON response with proper error handling
func WriteJSONResponse(w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// WriteErrorResponse writes an error response in Mesh format
func WriteErrorResponse(w http.ResponseWriter, err *types.Error, statusCode int) {
	if err == nil {
		err = GetError(500) // Default to internal server error
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]any{
		"error": err,
	}

	if encodeErr := json.NewEncoder(w).Encode(errorResponse); encodeErr != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
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
func GetStringFromOptions(options map[string]any, key string) string {
	if value, ok := options[key].(string); ok {
		return value
	}
	return "dynamic"
}

// HexToDecimal converts hex string to decimal string
func HexToDecimal(hexStr string) (string, error) {
	cleanHex := strings.TrimPrefix(hexStr, "0x")

	bigInt := new(big.Int)
	bigInt, ok := bigInt.SetString(cleanHex, 16)
	if !ok {
		return "", fmt.Errorf("invalid hex string: %s", hexStr)
	}

	return bigInt.String(), nil
}

// GetTargetIndex calculates the target index based on local index and peers
func GetTargetIndex(localIndex int64, peers []Peer) int64 {
	result := localIndex
	for _, peer := range peers {
		// Extract block number from bestBlockID (first 8 bytes = 16 hex characters)
		if len(peer.BestBlockID) >= 16 {
			blockNumHex := peer.BestBlockID[:16]
			if blockNum, err := strconv.ParseInt(blockNumHex, 16, 64); err == nil {
				if result < blockNum {
					result = blockNum
				}
			}
		}
	}
	return result
}

// Peer represents a connected peer
type Peer struct {
	PeerID      string
	BestBlockID string
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
	cleanHex := strings.TrimPrefix(hexStr, "0x")

	return hex.DecodeString(cleanHex)
}

// ParseJSONFromRequestContext parses JSON from the request context into the target struct
func ParseJSONFromRequestContext(r *http.Request, target any) error {
	body, ok := r.Context().Value(RequestBodyKey).([]byte)
	if !ok {
		return fmt.Errorf("request body not found in context - middleware may not be properly configured")
	}
	return json.Unmarshal(body, target)
}

// GetTokenCurrencyFromContractAddress returns the currency definition for a token contract
func GetTokenCurrencyFromContractAddress(contractAddress string, client thor.VeChainClientInterface) (*types.Currency, error) {
	if strings.EqualFold(contractAddress, VTHOCurrency.Metadata["contractAddress"].(string)) {
		return VTHOCurrency, nil
	}

	// For other tokens, fetch metadata from the contract using the wrapper
	contract, err := vip180.NewVIP180Contract(contractAddress, client)
	if err != nil {
		return nil, err
	}

	// Get symbol
	symbol, err := contract.Symbol()
	if err != nil {
		return nil, err
	}

	// Get decimals
	decimals, err := contract.Decimals()
	if err != nil {
		return nil, err
	}

	return &types.Currency{
		Symbol:   symbol,
		Decimals: decimals,
		Metadata: map[string]any{
			"contractAddress": strings.ToLower(contractAddress),
		},
	}, nil
}

// GetVETOperations extracts VET transfer operations from a list of operations
func GetVETOperations(operations []*types.Operation) []map[string]string {
	var result []map[string]string
	for _, op := range operations {
		if op.Type == OperationTypeTransfer && op.Amount != nil && op.Amount.Currency != nil {
			// Check if it's a VET operation (positive amount and VET currency)
			if op.Amount.Currency.Symbol == VETCurrency.Symbol {
				amount, ok := new(big.Int).SetString(op.Amount.Value, 10)
				if ok && amount.Cmp(big.NewInt(0)) > 0 {
					result = append(result, map[string]string{
						"value": op.Amount.Value,
						"to":    op.Account.Address,
					})
				}
			}
		}
	}
	return result
}

// GetTokensOperations extracts VIP180 token operations from a list of operations
func GetTokensOperations(operations []*types.Operation) (registered []map[string]string) {
	for _, op := range operations {
		if op.Type == OperationTypeTransfer && op.Amount != nil && op.Amount.Currency != nil {
			// Check if it's a token operation (positive amount and has contract address)
			if op.Amount.Currency.Metadata != nil {
				if contractAddr, exists := op.Amount.Currency.Metadata["contractAddress"]; exists {
					if addr, ok := contractAddr.(string); ok {
						amount, ok := new(big.Int).SetString(op.Amount.Value, 10)
						if ok && amount.Cmp(big.NewInt(0)) > 0 {
							registered = append(registered, map[string]string{
								"token": addr,
								"value": op.Amount.Value,
								"to":    op.Account.Address,
							})
						}
					}
				}
			}
		}
	}
	return registered
}
