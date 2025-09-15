package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/coinbase/rosetta-sdk-go/types"
)

// Config holds the service configuration
type Config struct {
	RosettaVersion    string                   `json:"rosetta_version"`
	Port              int                      `json:"port"`
	Mode              string                   `json:"mode"`
	Network           string                   `json:"network"`
	NodeAPI           string                   `json:"nodeApi"`
	ChainTag          int                      `json:"chainTag"`
	APIVersion        string                   `json:"apiVersion"`
	NodeVersion       string                   `json:"nodeVersion"`
	ServiceName       string                   `json:"serviceName"`
	TokenList         []interface{}            `json:"tokenlist"`
	NetworkIdentifier *types.NetworkIdentifier `json:"-"`
}

// NewConfig creates a new configuration by loading from JSON and environment variables
func NewConfig() (*Config, error) {
	// Load base config from JSON file
	configPath := filepath.Join("config", "config.json")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Override with environment variables
	config.loadFromEnv()

	// Set derived fields
	config.setDerivedFields()

	return &config, nil
}

// loadFromEnv loads configuration from environment variables
func (c *Config) loadFromEnv() {
	if mode := os.Getenv("MODE"); mode != "" {
		c.Mode = mode
	}

	if network := os.Getenv("NETWORK"); network != "" {
		c.Network = network
	}

	if nodeAPI := os.Getenv("NODEURL"); nodeAPI != "" {
		c.NodeAPI = nodeAPI
	}

	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Port = p
		}
	}
}

// setDerivedFields sets fields that are derived from other configuration values
func (c *Config) setDerivedFields() {
	// Set network identifier based on network
	var networkName string
	switch c.Network {
	case "main", "mainnet":
		networkName = "mainnet"
		if c.ChainTag == 0 {
			c.ChainTag = 0x4a // Mainnet chain tag
		}
	case "test", "testnet":
		networkName = "test"
		if c.ChainTag == 0 {
			c.ChainTag = 0x27 // Testnet chain tag
		}
	default:
		networkName = "custom"
	}

	c.NetworkIdentifier = &types.NetworkIdentifier{
		Blockchain: "vechainthor",
		Network:    networkName,
	}
}

// GetNetworkIdentifier returns the network identifier
func (c *Config) GetNetworkIdentifier() *types.NetworkIdentifier {
	return c.NetworkIdentifier
}

// GetRunMode returns the run mode
func (c *Config) GetRunMode() string {
	return c.Mode
}

// GetPort returns the server port
func (c *Config) GetPort() int {
	return c.Port
}

// GetNodeAPI returns the VeChain node API URL
func (c *Config) GetNodeAPI() string {
	return c.NodeAPI
}

// GetNetwork returns the network name
func (c *Config) GetNetwork() string {
	return c.Network
}

// GetChainTag returns the chain tag
func (c *Config) GetChainTag() int {
	return c.ChainTag
}

// GetServiceName returns the service name
func (c *Config) GetServiceName() string {
	return c.ServiceName
}

// GetRosettaVersion returns the Rosetta version
func (c *Config) GetRosettaVersion() string {
	return c.RosettaVersion
}

// IsOnlineMode returns true if running in online mode
func (c *Config) IsOnlineMode() bool {
	return c.Mode == "online"
}

// PrintConfig prints the current configuration
func (c *Config) PrintConfig() {
	fmt.Printf(`
******************** %s ********************
|   Api               |   localhost:%d
|   Rosetta Version   |   %s
|   Mode              |   %s
|   Node URL          |   %s
|   Node Version      |   %s
|   Network           |   %s
|   Chain Tag         |   0x%x
*******************************************************************
`,
		c.ServiceName,
		c.Port,
		c.RosettaVersion,
		c.Mode,
		c.NodeAPI,
		c.NodeVersion,
		c.Network,
		c.ChainTag,
	)
}
