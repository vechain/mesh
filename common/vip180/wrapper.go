package vip180

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/vechain/mesh/common/vip180/contracts"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/abi"
	"github.com/vechain/thor/v2/thor"
)

// VIP180Contract represents a VIP180 token contract
type VIP180Contract struct {
	address string
	abi     *abi.ABI
	client  meshthor.VeChainClientInterface
}

// NewVIP180Contract creates a new VIP180 contract wrapper
func NewVIP180Contract(address string, client meshthor.VeChainClientInterface) (*VIP180Contract, error) {
	vip180ABIPath := "compiled/VIP180Token.abi"
	vip180ABI := contracts.MustABI(vip180ABIPath)
	contractABI, err := abi.New(vip180ABI)
	if err != nil {
		return nil, fmt.Errorf("failed to create ABI: %w", err)
	}

	return &VIP180Contract{
		address: address,
		abi:     contractABI,
		client:  client,
	}, nil
}

// Symbol returns the symbol of the token
func (c *VIP180Contract) Symbol() (string, error) {
	return c.callStringMethod("symbol")
}

// Name returns the name of the token
func (c *VIP180Contract) Name() (string, error) {
	return c.callStringMethod("name")
}

// Decimals returns the decimals of the token
func (c *VIP180Contract) Decimals() (int32, error) {
	return c.callInt32Method("decimals")
}

// TotalSupply returns the total supply of the token
func (c *VIP180Contract) TotalSupply() (*big.Int, error) {
	return c.callBigIntMethod("totalSupply")
}

// BalanceOf returns the balance of an account
func (c *VIP180Contract) BalanceOf(owner string) (*big.Int, error) {
	method, exists := c.abi.MethodByName("balanceOf")
	if !exists {
		return nil, fmt.Errorf("method balanceOf not found")
	}

	// Convert string address to thor.Address
	ownerAddr, err := thor.ParseAddress(owner)
	if err != nil {
		return nil, fmt.Errorf("invalid owner address: %w", err)
	}

	callData, err := method.EncodeInput(ownerAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to encode balanceOf call: %w", err)
	}

	result, err := c.client.CallContract(c.address, fmt.Sprintf("0x%x", callData))
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf: %w", err)
	}

	return c.decodeBigIntResult(method, result)
}

// callStringMethod calls a method that returns a string
func (c *VIP180Contract) callStringMethod(methodName string) (string, error) {
	method, exists := c.abi.MethodByName(methodName)
	if !exists {
		return "", fmt.Errorf("method %s not found", methodName)
	}

	callData, err := method.EncodeInput()
	if err != nil {
		return "", fmt.Errorf("failed to encode %s call: %w", methodName, err)
	}

	result, err := c.client.CallContract(c.address, fmt.Sprintf("0x%x", callData))
	if err != nil {
		return "", fmt.Errorf("failed to call %s: %w", methodName, err)
	}

	return c.decodeStringResult(method, result)
}

// callInt32Method calls a method that returns an int32
func (c *VIP180Contract) callInt32Method(methodName string) (int32, error) {
	method, exists := c.abi.MethodByName(methodName)
	if !exists {
		return 0, fmt.Errorf("method %s not found", methodName)
	}

	callData, err := method.EncodeInput()
	if err != nil {
		return 0, fmt.Errorf("failed to encode %s call: %w", methodName, err)
	}

	result, err := c.client.CallContract(c.address, fmt.Sprintf("0x%x", callData))
	if err != nil {
		return 0, fmt.Errorf("failed to call %s: %w", methodName, err)
	}

	return c.decodeInt32Result(method, result)
}

// callBigIntMethod calls a method that returns a big.Int
func (c *VIP180Contract) callBigIntMethod(methodName string) (*big.Int, error) {
	method, exists := c.abi.MethodByName(methodName)
	if !exists {
		return nil, fmt.Errorf("method %s not found", methodName)
	}

	callData, err := method.EncodeInput()
	if err != nil {
		return nil, fmt.Errorf("failed to encode %s call: %w", methodName, err)
	}

	result, err := c.client.CallContract(c.address, fmt.Sprintf("0x%x", callData))
	if err != nil {
		return nil, fmt.Errorf("failed to call %s: %w", methodName, err)
	}

	return c.decodeBigIntResult(method, result)
}

// decodeStringResult decodes a string result from a contract call
func (c *VIP180Contract) decodeStringResult(method *abi.Method, result string) (string, error) {

	hexStr := strings.TrimPrefix(result, "0x")
	resultBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex result for %s: %w, %s", method.Name(), err, result)
	}

	var val string
	err = method.DecodeOutput(resultBytes, &val)
	if err != nil {
		return "", fmt.Errorf("failed to decode result for %s: %w", method.Name(), err)
	}
	return val, nil
}

// decodeInt32Result decodes an int32 result from a contract call
func (c *VIP180Contract) decodeInt32Result(method *abi.Method, result string) (int32, error) {

	hexStr := strings.TrimPrefix(result, "0x")
	resultBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0, fmt.Errorf("failed to decode hex result: %w", err)
	}

	var val uint8
	err = method.DecodeOutput(resultBytes, &val)
	if err != nil {
		return 0, fmt.Errorf("failed to decode result: %w", err)
	}

	return int32(val), nil
}

// decodeBigIntResult decodes a big.Int result from a contract call
func (c *VIP180Contract) decodeBigIntResult(method *abi.Method, result string) (*big.Int, error) {

	hexStr := strings.TrimPrefix(result, "0x")
	resultBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex result: %w", err)
	}

	val := new(big.Int)
	err = method.DecodeOutput(resultBytes, &val)
	if err != nil {
		return nil, fmt.Errorf("failed to decode result: %w", err)
	}

	return val, nil
}
