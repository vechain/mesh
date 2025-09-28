package vip180

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	meshthor "github.com/vechain/mesh/thor"
)

func TestNewVIP180Contract(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()

	contract, err := NewVIP180Contract("0x1234567890123456789012345678901234567890", mockClient)
	if err != nil {
		t.Errorf("NewVIP180Contract() error = %v, want nil", err)
	}

	if contract == nil || contract.address != "0x1234567890123456789012345678901234567890" {
		t.Errorf("NewVIP180Contract() address = %v, want 0x1234567890123456789012345678901234567890", contract.address)
	}
}

func TestVIP180Contract_Symbol(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	contract, _ := NewVIP180Contract("0x1234567890123456789012345678901234567890", mockClient)

	// Test successful symbol call - return properly encoded ABI string for "USDT"
	// Format: offset(32) + length(32) + data(padded to 32)
	mockClient.SetMockCallResult("0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000045553445400000000000000000000000000000000000000000000000000000000")

	symbol, err := contract.Symbol()
	if err != nil {
		t.Errorf("Symbol() error = %v, want nil", err)
	}

	if symbol != "USDT" {
		t.Errorf("Symbol() = %v, want USDT", symbol)
	}
}

func TestVIP180Contract_Symbol_Error(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	contract, _ := NewVIP180Contract("0x1234567890123456789012345678901234567890", mockClient)

	// Test error case
	mockClient.SetMockError(errors.New("contract call failed"))

	_, err := contract.Symbol()
	if err == nil {
		t.Errorf("Symbol() error = nil, want error")
	}
}

func TestVIP180Contract_Decimals(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	contract, _ := NewVIP180Contract("0x1234567890123456789012345678901234567890", mockClient)

	// Test successful decimals call (18 decimals) - properly encoded uint8
	mockClient.SetMockCallResult("0x0000000000000000000000000000000000000000000000000000000000000012")

	decimals, err := contract.Decimals()
	if err != nil {
		t.Errorf("Decimals() error = %v, want nil", err)
	}

	if decimals != 18 {
		t.Errorf("Decimals() = %v, want 18", decimals)
	}
}

func TestVIP180Contract_Decimals_Error(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	contract, _ := NewVIP180Contract("0x1234567890123456789012345678901234567890", mockClient)

	// Test error case
	mockClient.SetMockError(errors.New("contract call failed"))

	_, err := contract.Decimals()
	if err == nil {
		t.Errorf("Decimals() error = nil, want error")
	}
}

func TestVIP180Contract_BalanceOf(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	contract, _ := NewVIP180Contract("0x1234567890123456789012345678901234567890", mockClient)

	// Test successful balance call - properly encoded uint256 (32 bytes)
	mockClient.SetMockCallResult("0x0000000000000000000000000000000000000000000000000de0b6b3a7640000")

	balance, err := contract.BalanceOf("0xf077b491b355e64048ce21e3a6fc4751eeea77fa")
	if err != nil {
		t.Errorf("BalanceOf() error = %v, want nil", err)
	}

	expected := "1000000000000000000"
	if balance.String() != expected {
		t.Errorf("BalanceOf() = %v, want %v", balance.String(), expected)
	}
}

func TestVIP180Contract_BalanceOf_Error(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	contract, _ := NewVIP180Contract("0x1234567890123456789012345678901234567890", mockClient)

	// Test error case
	mockClient.SetMockError(errors.New("contract call failed"))

	_, err := contract.BalanceOf("0xf077b491b355e64048ce21e3a6fc4751eeea77fa")
	if err == nil {
		t.Errorf("BalanceOf() error = nil, want error")
	}
}

func TestVIP180Contract_Name(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	contract, _ := NewVIP180Contract("0x1234567890123456789012345678901234567890", mockClient)

	// Test successful name call - return properly encoded ABI string for "Test Token"
	mockClient.SetMockCallResult("0x0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000a5465737420546f6b656e00000000000000000000000000000000000000000000")

	name, err := contract.Name()
	if err != nil {
		t.Errorf("Name() error = %v, want nil", err)
	}

	if name != "Test Token" {
		t.Errorf("Name() = %v, want Test Token", name)
	}
}

func TestVIP180Contract_Name_Error(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	contract, _ := NewVIP180Contract("0x1234567890123456789012345678901234567890", mockClient)

	// Test error case
	mockClient.SetMockError(fmt.Errorf("contract call failed"))

	_, err := contract.Name()
	if err == nil {
		t.Errorf("Name() error = nil, want error")
	}
}

func TestVIP180Contract_TotalSupply(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	contract, _ := NewVIP180Contract("0x1234567890123456789012345678901234567890", mockClient)

	// Test successful totalSupply call - return properly encoded uint256
	mockClient.SetMockCallResult("0x0000000000000000000000000000000000000000000000000000000000000001")

	totalSupply, err := contract.TotalSupply()
	if err != nil {
		t.Errorf("TotalSupply() error = %v, want nil", err)
	}

	expected := big.NewInt(1)
	if totalSupply.Cmp(expected) != 0 {
		t.Errorf("TotalSupply() = %v, want %v", totalSupply, expected)
	}
}

func TestVIP180Contract_TotalSupply_Error(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	contract, _ := NewVIP180Contract("0x1234567890123456789012345678901234567890", mockClient)

	// Test error case
	mockClient.SetMockError(fmt.Errorf("contract call failed"))

	_, err := contract.TotalSupply()
	if err == nil {
		t.Errorf("TotalSupply() error = nil, want error")
	}
}

func TestVIP180Contract_callStringMethod(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	contract, _ := NewVIP180Contract("0x0000000000000000000000000000456e65726779", mockClient)

	t.Run("Successful call", func(t *testing.T) {
		// Set up mock response for string method
		mockClient.SetMockCallResult("0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000045445535400000000000000000000000000000000000000000000000000000000")

		result, err := contract.callStringMethod("symbol")
		if err != nil {
			t.Errorf("callStringMethod() error = %v, want nil", err)
		}
		if result != "TEST" {
			t.Errorf("callStringMethod() result = %v, want TEST", result)
		}
	})

	t.Run("Error case", func(t *testing.T) {
		// Set up mock error
		mockClient.SetMockError(fmt.Errorf("contract call failed"))

		_, err := contract.callStringMethod("symbol")
		if err == nil {
			t.Errorf("callStringMethod() error = nil, want error")
		}
	})
}

func TestVIP180Contract_callInt32Method(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	contract, _ := NewVIP180Contract("0x0000000000000000000000000000456e65726779", mockClient)

	t.Run("Successful call", func(t *testing.T) {
		// Set up mock response for uint8 method
		mockClient.SetMockCallResult("0x0000000000000000000000000000000000000000000000000000000000000012")

		result, err := contract.callInt32Method("decimals")
		if err != nil {
			t.Errorf("callInt32Method() error = %v, want nil", err)
		}
		if result != 18 {
			t.Errorf("callInt32Method() result = %v, want 18", result)
		}
	})

	t.Run("Error case", func(t *testing.T) {
		// Set up mock error
		mockClient.SetMockError(fmt.Errorf("contract call failed"))

		_, err := contract.callInt32Method("decimals")
		if err == nil {
			t.Errorf("callInt32Method() error = nil, want error")
		}
	})
}

func TestVIP180Contract_callBigIntMethod(t *testing.T) {
	mockClient := meshthor.NewMockVeChainClient()
	contract, _ := NewVIP180Contract("0x0000000000000000000000000000456e65726779", mockClient)

	t.Run("Successful call", func(t *testing.T) {
		// Set up mock response for uint256 method
		mockClient.SetMockCallResult("0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")

		result, err := contract.callBigIntMethod("totalSupply")
		if err != nil {
			t.Errorf("callBigIntMethod() error = %v, want nil", err)
		}
		if result == nil {
			t.Errorf("callBigIntMethod() result = nil, want non-nil")
		}
	})

	t.Run("Error case", func(t *testing.T) {
		// Set up mock error
		mockClient.SetMockError(fmt.Errorf("contract call failed"))

		_, err := contract.callBigIntMethod("totalSupply")
		if err == nil {
			t.Errorf("callBigIntMethod() error = nil, want error")
		}
	})
}
