package vip180

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vechain/thor/v2/abi"
)

// vip180TransferData represents decoded transfer data from a VIP180 token
type vip180TransferData struct {
	To     common.Address
	Amount *big.Int
}

// DecodeVIP180TransferCallData decodes VIP180 transfer function call data using ABI
func DecodeVIP180TransferCallData(data string) (*vip180TransferData, error) {
	cleanData := strings.TrimPrefix(data, "0x")

	dataBytes, err := hex.DecodeString(cleanData)
	if err != nil {
		return nil, fmt.Errorf("invalid hex data: %w", err)
	}

	contractABI, err := abi.New([]byte(VIP180ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to create ABI: %w", err)
	}

	method, err := contractABI.MethodByInput(dataBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to find method: %w", err)
	}

	if method.Name() != "transfer" {
		return nil, fmt.Errorf("expected transfer method, got %s", method.Name())
	}

	// Decode the input parameters
	var params vip180TransferData
	err = method.DecodeInput(dataBytes, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transfer parameters: %w", err)
	}

	return &params, nil
}

// IsVIP180TransferCallData checks if the data represents a VIP180 transfer call
func IsVIP180TransferCallData(data string) bool {
	cleanData := strings.TrimPrefix(data, "0x")

	if len(cleanData) < 8 {
		return false
	}

	dataBytes, err := hex.DecodeString(cleanData)
	if err != nil {
		return false
	}

	contractABI, err := abi.New([]byte(VIP180ABI))
	if err != nil {
		return false
	}

	_, err = contractABI.MethodByInput(dataBytes)
	return err == nil
}

// EncodeVIP180TransferCallData encodes VIP180 transfer call data
func EncodeVIP180TransferCallData(to string, amount string) (string, error) {
	contractABI, err := abi.New([]byte(VIP180ABI))
	if err != nil {
		return "", fmt.Errorf("failed to create ABI: %v", err)
	}

	// Get the transfer method
	method, exists := contractABI.MethodByName("transfer")
	if !exists {
		return "", fmt.Errorf("transfer method not found in ABI")
	}

	// Parse the amount
	amountBig, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return "", fmt.Errorf("invalid amount: %s", amount)
	}

	// Parse the 'to' address
	toAddr := common.HexToAddress(to)

	// Encode the method call
	data, err := method.EncodeInput(toAddr, amountBig)
	if err != nil {
		return "", fmt.Errorf("failed to encode transfer call: %v", err)
	}

	return "0x" + hex.EncodeToString(data), nil
}
