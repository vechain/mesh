package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <private_key_hex> <payload_hex>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s 99f0500549792796c14fed62011a51081dc5b5e68fe8bd8a13b86be829c4fd36 c7c260e16e3c32a6176759a3556ff5618d7d6e7e2c9c9602d40461fcaa34cbec\n", os.Args[0])
		os.Exit(1)
	}

	privateKeyHex := os.Args[1]
	payloadHex := os.Args[2]

	// Remove 0x prefix if present
	if len(privateKeyHex) > 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}
	if len(payloadHex) > 2 && payloadHex[:2] == "0x" {
		payloadHex = payloadHex[2:]
	}

	// Parse private key
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding private key: %v\n", err)
		os.Exit(1)
	}

	// Parse payload
	payloadBytes, err := hex.DecodeString(payloadHex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding payload: %v\n", err)
		os.Exit(1)
	}

	// Create ECDSA private key
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating ECDSA private key: %v\n", err)
		os.Exit(1)
	}

	// Sign the payload (same as elliptic.js signPayload function)
	signature, err := secp256k1.Sign(payloadBytes, privateKeyBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error signing payload: %v\n", err)
		os.Exit(1)
	}

	// Extract r, s, and recovery ID
	r := signature[:32]
	s := signature[32:64]
	recoveryID := signature[64]

	// Format signature as hex string (r + s + v)
	// v is 0x00 or 0x01 based on recovery ID (same as elliptic.js)
	v := "00"
	if recoveryID == 1 {
		v = "01"
	}

	signatureHex := hex.EncodeToString(r) + hex.EncodeToString(s) + v

	// Output the signature
	fmt.Println(signatureHex)

	// Optional: Verify the signature and show derived address
	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKey)
	fmt.Fprintf(os.Stderr, "Derived address: %s\n", address.Hex())
}
