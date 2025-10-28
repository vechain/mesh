package config

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
)

// Config holds the service configuration
type Config struct {
	MeshVersion       string                   `json:"meshVersion"`
	Port              int                      `json:"port"`
	Mode              string                   `json:"mode"`
	Network           string                   `json:"network"`
	NodeAPI           string                   `json:"nodeApi"`
	ChainTag          byte                     `json:"chainTag"`
	APIVersion        string                   `json:"apiVersion"`
	NodeVersion       string                   `json:"nodeVersion"`
	ServiceName       string                   `json:"serviceName"`
	BaseGasPrice      string                   `json:"baseGasPrice"`
	InitialBaseFee    string                   `json:"initialBaseFee"`
	Expiration        uint32                   `json:"expiration"`
	NetworkIdentifier *types.NetworkIdentifier `json:"-"`
	SoloOnDemand      bool                     `json:"soloOnDemand"`
}

// NewConfig creates a new configuration by loading from JSON and environment variables
func NewConfig() (*Config, error) {
	// Load base config from JSON file
	rootDir, err := os.OpenRoot("config")
	if err != nil {
		return nil, fmt.Errorf("failed to open config root: %v", err)
	}
	defer func() {
		if err := rootDir.Close(); err != nil {
			fmt.Printf("failed to close config root: %v", err)
		}
	}()

	f, err := rootDir.Open("config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("failed to close config file: %v", err)
		}
	}()

	configData, err := io.ReadAll(f)
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

	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Port = p
		}
	}

	// TODO: Delete the snippet (will always be true) once Thor is updated again in this regard
	if soloOnDemand := os.Getenv("SOLO_ONDEMAND"); soloOnDemand != "" {
		if soloOnDemandBool, err := strconv.ParseBool(soloOnDemand); err == nil {
			c.SoloOnDemand = soloOnDemandBool
		}
	}
}

// setDerivedFields sets fields that are derived from other configuration values
func (c *Config) setDerivedFields() {
	// Set network identifier based on network
	var networkName string
	switch c.Network {
	case meshcommon.MainNetwork, "mainnet":
		networkName = meshcommon.MainNetwork
		c.ChainTag = 0x4a // Mainnet chain tag
	case meshcommon.TestNetwork, "testnet":
		networkName = meshcommon.TestNetwork
		c.ChainTag = 0x27 // Testnet chain tag
	case meshcommon.SoloNetwork:
		networkName = meshcommon.SoloNetwork
		c.ChainTag = 0xf6 // Solo chain tag
	default:
		networkName = "custom"
	}

	c.NetworkIdentifier = &types.NetworkIdentifier{
		Blockchain: meshcommon.BlockchainName,
		Network:    networkName,
	}
}

// IsOnlineMode returns true if running in online mode
func (c *Config) IsOnlineMode() bool {
	return c.Mode == meshcommon.OnlineMode
}

// PrintConfig prints the current configuration
func (c *Config) PrintConfig() {
	fmt.Printf(`
******************** %s ********************
|   Api               |   localhost:%d
|   Mesh Version      |   %s
|   Mode              |   %s
|   Node URL          |   %s
|   Node Version      |   %s
|   Network           |   %s
|   Chain Tag         |   0x%x
*******************************************************************
`,
		c.ServiceName,
		c.Port,
		c.MeshVersion,
		c.Mode,
		c.NodeAPI,
		c.NodeVersion,
		c.Network,
		c.ChainTag,
	)
}

// GetBaseGasPrice returns the base gas price as a big.Int
func (c *Config) GetBaseGasPrice() *big.Int {
	if c.BaseGasPrice == "" {
		return nil
	}

	baseGasPrice, ok := new(big.Int).SetString(c.BaseGasPrice, 10)
	if !ok {
		return nil
	}

	return baseGasPrice
}
