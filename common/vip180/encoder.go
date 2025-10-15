package vip180

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	meshcrypto "github.com/vechain/mesh/common/crypto"
	"github.com/vechain/mesh/common/vip180/contracts"
	"github.com/vechain/thor/v2/abi"
)

// vip180TransferData represents decoded transfer data from a VIP180 token
type vip180TransferData struct {
	To    common.Address
	Value *big.Int
}

type VIP180Encoder struct {
	abi          *abi.ABI
	bytesHandler *meshcrypto.BytesHandler
}

func NewVIP180Encoder() *VIP180Encoder {
	vip180ABIPath := "compiled/IVIP180.abi"
	vip180ABI := contracts.MustABI(vip180ABIPath)
	contractABI, err := abi.New(vip180ABI)
	if err != nil {
		panic(fmt.Errorf("failed to create ABI: %w", err))
	}
	return &VIP180Encoder{
		abi:          contractABI,
		bytesHandler: meshcrypto.NewBytesHandler(),
	}
}

// DecodeVIP180TransferCallData decodes VIP180 transfer function call data using ABI
func (e *VIP180Encoder) DecodeVIP180TransferCallData(data string) (*vip180TransferData, error) {
	dataBytes, err := e.bytesHandler.DecodeHexStringWithPrefix(data)
	if err != nil {
		return nil, fmt.Errorf("invalid hex data: %w", err)
	}

	method, err := e.abi.MethodByInput(dataBytes)
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
func (e *VIP180Encoder) IsVIP180TransferCallData(data string) bool {
	if len(strings.TrimPrefix(data, "0x")) < 8 {
		return false
	}

	dataBytes, err := e.bytesHandler.DecodeHexStringWithPrefix(data)
	if err != nil {
		return false
	}

	_, err = e.abi.MethodByInput(dataBytes)
	return err == nil
}

// EncodeVIP180TransferCallData encodes VIP180 transfer call data
func (e *VIP180Encoder) EncodeVIP180TransferCallData(to string, amount string) (string, error) {
	// Get the transfer method
	method, exists := e.abi.MethodByName("transfer")
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
