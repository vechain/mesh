package e2e

import (
	"os"
	"strconv"
)

// TestConfig holds configuration for e2e tests
type TestConfig struct {
	BaseURL          string
	Network          string
	SenderPrivateKey string
	SenderAddress    string
	RecipientAddress string
	TransferAmount   string
	GasPrice         string
	GasLimit         int
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
		GasPrice:         getEnv("GAS_PRICE", "1000000000000000000"),       // 1 VTHO in wei
		GasLimit:         getEnvInt("GAS_LIMIT", 21000),
		TimeoutSeconds:   getEnvInt("TIMEOUT_SECONDS", 30),
	}

	return config
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
