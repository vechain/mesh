package e2e

import (
	"encoding/hex"
	"os"
	"strconv"

	"github.com/vechain/mesh/common/vip180/contracts"
)

// TestConfig holds configuration for e2e tests
type TestConfig struct {
	BaseURL          string
	Network          string
	SenderPrivateKey string
	SenderAddress    string
	RecipientAddress string
	TransferAmount   string
	TimeoutSeconds   int
}

// GetTestConfig returns the test configuration from environment variables or defaults
func GetTestConfig() *TestConfig {
	config := &TestConfig{
		BaseURL:          getEnv("MESH_BASE_URL", "http://localhost:8080"),
		Network:          getEnv("MESH_NETWORK", "solo"),
		SenderPrivateKey: getEnv("SENDER_PRIVATE_KEY", "99f0500549792796c14fed62011a51081dc5b5e68fe8bd8a13b86be829c4fd36"),
		SenderAddress:    getEnv("SENDER_ADDRESS", "0xf077b491b355E64048cE21E3A6Fc4751eEeA77fa"),
		RecipientAddress: getEnv("RECIPIENT_ADDRESS", "0x16277a1ff38678291c41d1820957c78bb5da59ce"),
		TransferAmount:   getEnv("TRANSFER_AMOUNT", "1000000000000000000"), // 1 VET in wei
		TimeoutSeconds:   getEnvInt("TIMEOUT_SECONDS", 30),
	}

	return config
}

// VIP180TestConfig holds configuration for VIP180 e2e tests
type VIP180TestConfig struct {
	*TestConfig
	ThorURL          string
	ContractBytecode string
	ContractAddress  string
	TokenName        string
	TokenSymbol      string
	TokenDecimals    uint8
	TokenTotalSupply string
	BridgeAddress    string
}

// GetVIP180TestConfig returns the VIP180 test configuration
func GetVIP180TestConfig() *VIP180TestConfig {
	baseConfig := GetTestConfig()

	return &VIP180TestConfig{
		TestConfig:       baseConfig,
		ThorURL:          getEnv("THOR_URL", "http://localhost:8669"),
		ContractBytecode: hex.EncodeToString(contracts.MustBIN("compiled/VIP180.bin")),
		TokenName:        "Test VIP180 Token",
		TokenSymbol:      "TVIP",
		TokenDecimals:    18,
		TokenTotalSupply: "1000000000000000000000000",                  // 1M tokens with 18 decimals
		BridgeAddress:    "0x0000000000000000000000000000000000000000", // Zero address for bridge
	}
}

// getEnv gets an environment variable or returns the default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as integer or returns the default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
