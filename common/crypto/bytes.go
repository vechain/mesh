package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type BytesHandler struct{}

func NewBytesHandler() *BytesHandler {
	return &BytesHandler{}
}

// ComputeAddress computes address from public key
func (h *BytesHandler) ComputeAddress(publicKey *types.PublicKey) (string, error) {
	pubKey, err := crypto.DecompressPubkey(publicKey.Bytes)
	if err != nil {
		return "", err
	}
	address := crypto.PubkeyToAddress(*pubKey)
	return strings.ToLower(address.Hex()), nil
}

// GenerateNonce generates a random nonce
func (h *BytesHandler) GenerateNonce() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return fmt.Sprintf("0x%016x", bytes), nil
}

// DecodeHexStringWithPrefix removes the "0x" prefix (if present) and decodes the hex string to bytes
func (h *BytesHandler) DecodeHexStringWithPrefix(hexStr string) ([]byte, error) {
	cleanHex := strings.TrimPrefix(hexStr, "0x")

	return hex.DecodeString(cleanHex)
}

// HexToDecimal converts hex string to decimal string
func (h *BytesHandler) HexToDecimal(hexStr string) (string, error) {
	cleanHex := strings.TrimPrefix(hexStr, "0x")

	bigInt := new(big.Int)
	bigInt, ok := bigInt.SetString(cleanHex, 16)
	if !ok {
		return "", fmt.Errorf("invalid hex string: %s", hexStr)
	}

	return bigInt.String(), nil
}
