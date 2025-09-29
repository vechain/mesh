package crypto

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestSignPayload(t *testing.T) {
	// Test with known private key and payload
	// Private key: 99f0500549792796c14fed62011a51081dc5b5e68fe8bd8a13b86be829c4fd36
	// Expected address: 0xf077b491b355E64048cE21E3A6Fc4751eEeA77fa
	privateKeyHex := "99f0500549792796c14fed62011a51081dc5b5e68fe8bd8a13b86be829c4fd36"
	payloadHex := "c7c260e16e3c32a6176759a3556ff5618d7d6e7e2c9c9602d40461fcaa34cbec"

	tests := []struct {
		name         string
		privateKey   string
		payload      string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid private key and payload",
			privateKey:  privateKeyHex,
			payload:     payloadHex,
			expectError: false,
		},
		{
			name:        "private key with 0x prefix",
			privateKey:  "0x" + privateKeyHex,
			payload:     payloadHex,
			expectError: false,
		},
		{
			name:        "payload with 0x prefix",
			privateKey:  privateKeyHex,
			payload:     "0x" + payloadHex,
			expectError: false,
		},
		{
			name:        "both with 0x prefix",
			privateKey:  "0x" + privateKeyHex,
			payload:     "0x" + payloadHex,
			expectError: false,
		},
		{
			name:         "invalid private key hex",
			privateKey:   "invalid_hex",
			payload:      payloadHex,
			expectError:  true,
			errorMessage: "error decoding private key",
		},
		{
			name:         "invalid payload hex",
			privateKey:   privateKeyHex,
			payload:      "invalid_hex",
			expectError:  true,
			errorMessage: "error decoding payload",
		},
		{
			name:         "empty private key",
			privateKey:   "",
			payload:      payloadHex,
			expectError:  true,
			errorMessage: "error creating ECDSA private key",
		},
		{
			name:         "empty payload",
			privateKey:   privateKeyHex,
			payload:      "",
			expectError:  true,
			errorMessage: "error signing payload",
		},
		{
			name:         "invalid private key length",
			privateKey:   "1234", // Too short
			payload:      payloadHex,
			expectError:  true,
			errorMessage: "error creating ECDSA private key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature, err := NewSigningHandler(tt.privateKey).SignPayload(tt.payload)

			if tt.expectError {
				if err == nil {
					t.Errorf("SignPayload() expected error, got nil")
					return
				}
				if tt.errorMessage != "" && !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("SignPayload() error = %v, want error containing %v", err, tt.errorMessage)
				}
			} else {
				if err != nil {
					t.Errorf("SignPayload() unexpected error: %v", err)
					return
				}

				// Validate signature format
				if len(signature) != 130 { // 64 bytes (r+s) + 2 chars (v) = 130 hex chars
					t.Errorf("SignPayload() signature length = %d, want 130", len(signature))
				}

				// Validate that signature is valid hex
				if _, err := hex.DecodeString(signature); err != nil {
					t.Errorf("SignPayload() signature is not valid hex: %v", err)
				}

				// Validate v component (last 2 characters should be 00 or 01)
				v := signature[128:]
				if v != "00" && v != "01" {
					t.Errorf("SignPayload() v component = %v, want 00 or 01", v)
				}
			}
		})
	}
}

func TestGetAddressFromPrivateKey(t *testing.T) {
	// Test with known private key
	privateKeyHex := "99f0500549792796c14fed62011a51081dc5b5e68fe8bd8a13b86be829c4fd36"
	expectedAddress := "0xf077b491b355E64048cE21E3A6Fc4751eEeA77fa"

	tests := []struct {
		name         string
		privateKey   string
		expectedAddr string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "valid private key",
			privateKey:   privateKeyHex,
			expectedAddr: expectedAddress,
			expectError:  false,
		},
		{
			name:         "private key with 0x prefix",
			privateKey:   "0x" + privateKeyHex,
			expectedAddr: expectedAddress,
			expectError:  false,
		},
		{
			name:         "invalid private key hex",
			privateKey:   "invalid_hex",
			expectedAddr: "",
			expectError:  true,
			errorMessage: "error decoding private key",
		},
		{
			name:         "empty private key",
			privateKey:   "",
			expectedAddr: "",
			expectError:  true,
			errorMessage: "error creating ECDSA private key",
		},
		{
			name:         "invalid private key length",
			privateKey:   "1234", // Too short
			expectedAddr: "",
			expectError:  true,
			errorMessage: "error creating ECDSA private key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address, err := NewSigningHandler(tt.privateKey).GetAddressFromPrivateKey()

			if tt.expectError {
				if err == nil {
					t.Errorf("GetAddressFromPrivateKey() expected error, got nil")
					return
				}
				if tt.errorMessage != "" && !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("GetAddressFromPrivateKey() error = %v, want error containing %v", err, tt.errorMessage)
				}
			} else {
				if err != nil {
					t.Errorf("GetAddressFromPrivateKey() unexpected error: %v", err)
					return
				}

				if address != tt.expectedAddr {
					t.Errorf("GetAddressFromPrivateKey() = %v, want %v", address, tt.expectedAddr)
				}

				// Validate address format
				if len(address) != 42 {
					t.Errorf("GetAddressFromPrivateKey() address length = %d, want 42", len(address))
				}
				if address[:2] != "0x" {
					t.Errorf("GetAddressFromPrivateKey() address should start with 0x, got %v", address[:2])
				}
			}
		})
	}
}

func TestSignPayloadWithAddress(t *testing.T) {
	privateKeyHex := "99f0500549792796c14fed62011a51081dc5b5e68fe8bd8a13b86be829c4fd36"
	payloadHex := "c7c260e16e3c32a6176759a3556ff5618d7d6e7e2c9c9602d40461fcaa34cbec"
	expectedAddress := "0xf077b491b355E64048cE21E3A6Fc4751eEeA77fa"

	tests := []struct {
		name         string
		privateKey   string
		payload      string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid private key and payload",
			privateKey:  privateKeyHex,
			payload:     payloadHex,
			expectError: false,
		},
		{
			name:         "invalid private key",
			privateKey:   "invalid_hex",
			payload:      payloadHex,
			expectError:  true,
			errorMessage: "error decoding private key",
		},
		{
			name:         "invalid payload",
			privateKey:   privateKeyHex,
			payload:      "invalid_hex",
			expectError:  true,
			errorMessage: "error decoding payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature, address, err := NewSigningHandler(tt.privateKey).SignPayloadWithAddress(tt.payload)

			if tt.expectError {
				if err == nil {
					t.Errorf("SignPayloadWithAddress() expected error, got nil")
					return
				}
				if tt.errorMessage != "" && !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("SignPayloadWithAddress() error = %v, want error containing %v", err, tt.errorMessage)
				}
			} else {
				if err != nil {
					t.Errorf("SignPayloadWithAddress() unexpected error: %v", err)
					return
				}

				// Validate signature
				if len(signature) != 130 {
					t.Errorf("SignPayloadWithAddress() signature length = %d, want 130", len(signature))
				}

				// Validate address
				if address != expectedAddress {
					t.Errorf("SignPayloadWithAddress() address = %v, want %v", address, expectedAddress)
				}

				// Validate that signature is valid hex
				if _, err := hex.DecodeString(signature); err != nil {
					t.Errorf("SignPayloadWithAddress() signature is not valid hex: %v", err)
				}
			}
		})
	}
}

func TestSignPayloadConsistency(t *testing.T) {
	// Test that signing the same payload with the same private key produces the same result
	privateKeyHex := "99f0500549792796c14fed62011a51081dc5b5e68fe8bd8a13b86be829c4fd36"
	payloadHex := "c7c260e16e3c32a6176759a3556ff5618d7d6e7e2c9c9602d40461fcaa34cbec"

	handler := NewSigningHandler(privateKeyHex)
	signature1, err1 := handler.SignPayload(payloadHex)
	if err1 != nil {
		t.Fatalf("First SignPayload() call failed: %v", err1)
	}

	signature2, err2 := handler.SignPayload(payloadHex)
	if err2 != nil {
		t.Fatalf("Second SignPayload() call failed: %v", err2)
	}

	if signature1 != signature2 {
		t.Errorf("SignPayload() inconsistent results: %v != %v", signature1, signature2)
	}
}

func TestAddressDerivationConsistency(t *testing.T) {
	// Test that the same private key always derives the same address
	privateKeyHex := "99f0500549792796c14fed62011a51081dc5b5e68fe8bd8a13b86be829c4fd36"
	expectedAddress := "0xf077b491b355E64048cE21E3A6Fc4751eEeA77fa"

	handler := NewSigningHandler(privateKeyHex)

	address1, err1 := handler.GetAddressFromPrivateKey()
	if err1 != nil {
		t.Fatalf("First GetAddressFromPrivateKey() call failed: %v", err1)
	}

	address2, err2 := handler.GetAddressFromPrivateKey()
	if err2 != nil {
		t.Fatalf("Second GetAddressFromPrivateKey() call failed: %v", err2)
	}

	if address1 != address2 {
		t.Errorf("GetAddressFromPrivateKey() inconsistent results: %v != %v", address1, address2)
	}

	if address1 != expectedAddress {
		t.Errorf("GetAddressFromPrivateKey() = %v, want %v", address1, expectedAddress)
	}
}

func TestSignPayloadWithAddressConsistency(t *testing.T) {
	// Test that SignPayloadWithAddress produces consistent results
	privateKeyHex := "99f0500549792796c14fed62011a51081dc5b5e68fe8bd8a13b86be829c4fd36"
	payloadHex := "c7c260e16e3c32a6176759a3556ff5618d7d6e7e2c9c9602d40461fcaa34cbec"

	handler := NewSigningHandler(privateKeyHex)
	sig1, addr1, err1 := handler.SignPayloadWithAddress(payloadHex)
	if err1 != nil {
		t.Fatalf("First SignPayloadWithAddress() call failed: %v", err1)
	}

	sig2, addr2, err2 := handler.SignPayloadWithAddress(payloadHex)
	if err2 != nil {
		t.Fatalf("Second SignPayloadWithAddress() call failed: %v", err2)
	}

	if sig1 != sig2 {
		t.Errorf("SignPayloadWithAddress() inconsistent signatures: %v != %v", sig1, sig2)
	}

	if addr1 != addr2 {
		t.Errorf("SignPayloadWithAddress() inconsistent addresses: %v != %v", addr1, addr2)
	}
}

func TestSignPayloadWithDifferentKeys(t *testing.T) {
	// Test that different private keys produce different signatures
	privateKey1 := "99f0500549792796c14fed62011a51081dc5b5e68fe8bd8a13b86be829c4fd36"
	privateKey2 := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	payloadHex := "c7c260e16e3c32a6176759a3556ff5618d7d6e7e2c9c9602d40461fcaa34cbec"

	signature1, err1 := NewSigningHandler(privateKey1).SignPayload(payloadHex)
	if err1 != nil {
		t.Fatalf("SignPayload() with first key failed: %v", err1)
	}

	signature2, err2 := NewSigningHandler(privateKey2).SignPayload(payloadHex)
	if err2 != nil {
		t.Fatalf("SignPayload() with second key failed: %v", err2)
	}

	if signature1 == signature2 {
		t.Errorf("SignPayload() with different keys produced same signature")
	}
}

func TestSignPayloadWithDifferentPayloads(t *testing.T) {
	// Test that different payloads produce different signatures
	privateKeyHex := "99f0500549792796c14fed62011a51081dc5b5e68fe8bd8a13b86be829c4fd36"
	payload1 := "c7c260e16e3c32a6176759a3556ff5618d7d6e7e2c9c9602d40461fcaa34cbec"
	payload2 := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	handler := NewSigningHandler(privateKeyHex)
	signature1, err1 := handler.SignPayload(payload1)
	if err1 != nil {
		t.Fatalf("SignPayload() with first payload failed: %v", err1)
	}

	signature2, err2 := handler.SignPayload(payload2)
	if err2 != nil {
		t.Fatalf("SignPayload() with second payload failed: %v", err2)
	}

	if signature1 == signature2 {
		t.Errorf("SignPayload() with different payloads produced same signature")
	}
}
