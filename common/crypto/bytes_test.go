package crypto

import (
	"bytes"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
)

func TestGenerateNonce(t *testing.T) {
	handler := NewBytesHandler()
	nonce, err := handler.GenerateNonce()
	if err != nil {
		t.Errorf("GenerateNonce() unexpected error: %v", err)
	}

	if len(nonce) != 18 { // "0x" + 16 hex chars = 18
		t.Errorf("Expected nonce length 18, got %d", len(nonce))
	}

	// Test that nonces are different
	nonce2, err := handler.GenerateNonce()
	if err != nil {
		t.Errorf("GenerateNonce() unexpected error: %v", err)
	}
	if nonce == nonce2 {
		t.Error("Expected different nonces, but got the same")
	}
}

func TestComputeAddress(t *testing.T) {
	handler := NewBytesHandler()
	// Test with valid public key
	publicKey := &types.PublicKey{
		Bytes:     []byte{0x02, 0xd9, 0x92, 0xbd, 0x20, 0x3d, 0x2b, 0xf8, 0x88, 0x38, 0x90, 0x89, 0xdb, 0x13, 0xd2, 0xd0, 0x80, 0x7c, 0x16, 0x97, 0x09, 0x1d, 0xe3, 0x77, 0x99, 0x8e, 0xfe, 0x6c, 0xf6, 0x0d, 0x66, 0xfb, 0xb3},
		CurveType: "secp256k1",
	}

	address, err := handler.ComputeAddress(publicKey)
	if err != nil {
		t.Errorf("ComputeAddress() unexpected error: %v", err)
	}

	if address == "" {
		t.Error("Expected non-empty address")
	}

	// Test with invalid public key
	invalidPublicKey := &types.PublicKey{
		Bytes:     []byte("invalid"),
		CurveType: "secp256k1",
	}

	_, err = handler.ComputeAddress(invalidPublicKey)
	if err == nil {
		t.Error("Expected error for invalid public key")
	}
}

func TestDecodeHexStringWithPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected []byte
		hasError bool
	}{
		{"0x1234", []byte{0x12, 0x34}, false},
		{"0xff", []byte{0xff}, false},
		{"invalid", nil, true},
		{"", []byte{}, false}, // Empty string is valid hex (returns empty byte slice)
	}

	handler := NewBytesHandler()
	for _, tt := range tests {
		result, err := handler.DecodeHexStringWithPrefix(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("DecodeHexStringWithPrefix(%s) expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("DecodeHexStringWithPrefix(%s) unexpected error: %v", tt.input, err)
			}
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("DecodeHexStringWithPrefix(%s) = %v, expected %v", tt.input, result, tt.expected)
			}
		}
	}
}
