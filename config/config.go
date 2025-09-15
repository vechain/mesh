package services

import "github.com/coinbase/rosetta-sdk-go/types"

// Config holds the service configuration
type Config struct {
	NetworkIdentifier *types.NetworkIdentifier
	RunMode           string
}

// NewConfig creates a new configuration
func NewConfig() *Config {
	return &Config{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "test", // Change to "mainnet" for production
		},
		RunMode: "online",
	}
}

// GetNetworkIdentifier returns the network identifier
func (c *Config) GetNetworkIdentifier() *types.NetworkIdentifier {
	return c.NetworkIdentifier
}

// GetRunMode returns the run mode
func (c *Config) GetRunMode() string {
	return c.RunMode
}
