package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

type SigningHandler struct {
	privateKeyBytes []byte
}

func NewSigningHandler(privateKeyHex string) *SigningHandler {
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	// Parse private key
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		panic(fmt.Errorf("error decoding private key: %v", err))
	}

	return &SigningHandler{
		privateKeyBytes: privateKeyBytes,
	}
}

// SignPayload signs a payload using secp256k1 and returns the signature in hex format
// This function is used by both the sign_payload script and e2e tests
func (h *SigningHandler) SignPayload(payloadHex string) (string, error) {
	payloadHex = strings.TrimPrefix(payloadHex, "0x")

	// Parse payload
	payloadBytes, err := hex.DecodeString(payloadHex)
	if err != nil {
		return "", fmt.Errorf("error decoding payload: %v", err)
	}

	// Sign the payload
	signature, err := secp256k1.Sign(payloadBytes, h.privateKeyBytes)
	if err != nil {
		return "", fmt.Errorf("error signing payload: %v", err)
	}

	// Extract r, s, and recovery ID
	r := signature[:32]
	s := signature[32:64]
	recoveryID := signature[64]

	// Format signature as hex string (r + s + v)
	// v is 0x00 or 0x01 based on recovery ID
	v := "00"
	if recoveryID == 1 {
		v = "01"
	}

	signatureHex := hex.EncodeToString(r) + hex.EncodeToString(s) + v

	return signatureHex, nil
}

// GetAddressFromPrivateKey derives the Ethereum address from a private key
func (h *SigningHandler) GetAddressFromPrivateKey() (string, error) {
	// Create ECDSA private key
	privateKey, err := crypto.ToECDSA(h.privateKeyBytes)
	if err != nil {
		return "", fmt.Errorf("error creating ECDSA private key: %v", err)
	}

	// Get public key and derive address
	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKey)

	return strings.ToLower(address.Hex()), nil
}

// SignPayloadWithAddress signs a payload and returns both signature and derived address
func (h *SigningHandler) SignPayloadWithAddress(payloadHex string) (signature, address string, err error) {
	signature, err = h.SignPayload(payloadHex)
	if err != nil {
		return "", "", err
	}

	address, err = h.GetAddressFromPrivateKey()
	if err != nil {
		return "", "", err
	}

	return signature, address, nil
}
