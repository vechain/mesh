package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
)

func TestNewConfig(t *testing.T) {
	// Change to project root directory to find config.json
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Go up one directory to project root
	if err := os.Chdir(".."); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	// Test config loading with existing config.json
	config, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() failed: %v", err)
	}

	// Verify basic fields from config.json
	if config.MeshVersion == "" {
		t.Errorf("NewConfig() MeshVersion is empty")
	}
	if config.Port <= 0 {
		t.Errorf("NewConfig() Port = %v, want > 0", config.Port)
	}
	if config.Mode == "" {
		t.Errorf("NewConfig() Mode is empty")
	}
	if config.Network == "" {
		t.Errorf("NewConfig() Network is empty")
	}

	// Verify derived fields
	if config.NetworkIdentifier == nil {
		t.Errorf("NewConfig() NetworkIdentifier is nil")
	} else {
		if config.NetworkIdentifier.Blockchain != "vechainthor" {
			t.Errorf("NewConfig() NetworkIdentifier.Blockchain = %v, want vechainthor", config.NetworkIdentifier.Blockchain)
		}
		if config.NetworkIdentifier.Network == "" {
			t.Errorf("NewConfig() NetworkIdentifier.Network is empty")
		}
	}

	// Verify chain tag was set based on network
	if config.ChainTag == 0 {
		t.Errorf("NewConfig() ChainTag = %v, want non-zero", config.ChainTag)
	}
}

func TestNewConfigWithEnvironmentVariables(t *testing.T) {
	// Change to project root directory to find config.json
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Go up one directory to project root
	if err := os.Chdir(".."); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	// Set environment variables
	os.Setenv("MODE", "offline")
	os.Setenv("NETWORK", "test")
	os.Setenv("PORT", "9090")
	defer func() {
		os.Unsetenv("MODE")
		os.Unsetenv("NETWORK")
		os.Unsetenv("PORT")
	}()

	// Test config loading with environment variables
	config, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() failed: %v", err)
	}

	// Verify environment variables override
	if config.Mode != "offline" {
		t.Errorf("NewConfig() Mode = %v, want offline", config.Mode)
	}
	if config.Network != "test" {
		t.Errorf("NewConfig() Network = %v, want test", config.Network)
	}
	if config.Port != 9090 {
		t.Errorf("NewConfig() Port = %v, want 9090", config.Port)
	}

	// Verify derived fields reflect environment changes
	if config.NetworkIdentifier == nil {
		t.Errorf("NewConfig() NetworkIdentifier is nil")
	} else {
		if config.NetworkIdentifier.Network != "test" {
			t.Errorf("NewConfig() NetworkIdentifier.Network = %v, want test", config.NetworkIdentifier.Network)
		}
	}

	// Verify chain tag was set for test network
	if config.ChainTag != 0x27 {
		t.Errorf("NewConfig() ChainTag = %v, want 0x27", config.ChainTag)
	}
}

func TestSetDerivedFields(t *testing.T) {
	tests := []struct {
		name            string
		network         string
		chainTag        int
		expectedTag     int
		expectedNetwork string
	}{
		{
			name:            "main network",
			network:         "main",
			chainTag:        0,
			expectedTag:     0x4a,
			expectedNetwork: "main",
		},
		{
			name:            "mainnet network",
			network:         "mainnet",
			chainTag:        0,
			expectedTag:     0x4a,
			expectedNetwork: "main",
		},
		{
			name:            "test network",
			network:         "test",
			chainTag:        0,
			expectedTag:     0x27,
			expectedNetwork: "test",
		},
		{
			name:            "testnet network",
			network:         "testnet",
			chainTag:        0,
			expectedTag:     0x27,
			expectedNetwork: "test",
		},
		{
			name:            "solo network",
			network:         "solo",
			chainTag:        0,
			expectedTag:     0xf6,
			expectedNetwork: "solo",
		},
		{
			name:            "custom network",
			network:         "custom",
			chainTag:        0,
			expectedTag:     0,
			expectedNetwork: "custom",
		},
		{
			name:            "network with existing chain tag",
			network:         "solo",
			chainTag:        0x123,
			expectedTag:     0x123,
			expectedNetwork: "solo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Network:  tt.network,
				ChainTag: tt.chainTag,
			}

			config.setDerivedFields()

			if config.ChainTag != tt.expectedTag {
				t.Errorf("setDerivedFields() ChainTag = %v, want %v", config.ChainTag, tt.expectedTag)
			}

			if config.NetworkIdentifier == nil {
				t.Errorf("setDerivedFields() NetworkIdentifier is nil")
			} else {
				if config.NetworkIdentifier.Blockchain != "vechainthor" {
					t.Errorf("setDerivedFields() NetworkIdentifier.Blockchain = %v, want vechainthor", config.NetworkIdentifier.Blockchain)
				}
				if config.NetworkIdentifier.Network != tt.expectedNetwork {
					t.Errorf("setDerivedFields() NetworkIdentifier.Network = %v, want %v", config.NetworkIdentifier.Network, tt.expectedNetwork)
				}
			}
		})
	}
}

func TestConfigGetters(t *testing.T) {
	config := &Config{
		MeshVersion: "1.0.0",
		Port:        8080,
		Mode:        "online",
		Network:     "solo",
		NodeAPI:     "http://localhost:8669",
		ChainTag:    0xf6,
		APIVersion:  "1.4.10",
		NodeVersion: "1.0.0",
		ServiceName: "vechain-mesh",
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "vechainthor",
			Network:    "solo",
		},
	}

	// Test GetNetworkIdentifier
	networkID := config.GetNetworkIdentifier()
	if networkID == nil {
		t.Errorf("GetNetworkIdentifier() returned nil")
	} else {
		if networkID.Blockchain != "vechainthor" {
			t.Errorf("GetNetworkIdentifier() Blockchain = %v, want vechainthor", networkID.Blockchain)
		}
		if networkID.Network != "solo" {
			t.Errorf("GetNetworkIdentifier() Network = %v, want solo", networkID.Network)
		}
	}

	// Test GetRunMode
	mode := config.GetRunMode()
	if mode != "online" {
		t.Errorf("GetRunMode() = %v, want online", mode)
	}

	// Test GetPort
	port := config.GetPort()
	if port != 8080 {
		t.Errorf("GetPort() = %v, want 8080", port)
	}

	// Test GetNodeAPI
	nodeAPI := config.GetNodeAPI()
	if nodeAPI != "http://localhost:8669" {
		t.Errorf("GetNodeAPI() = %v, want http://localhost:8669", nodeAPI)
	}

	// Test GetNetwork
	network := config.GetNetwork()
	if network != "solo" {
		t.Errorf("GetNetwork() = %v, want solo", network)
	}

	// Test GetChainTag
	chainTag := config.GetChainTag()
	if chainTag != 0xf6 {
		t.Errorf("GetChainTag() = %v, want 0xf6", chainTag)
	}

	// Test GetMeshVersion
	version := config.GetMeshVersion()
	if version != "1.0.0" {
		t.Errorf("GetMeshVersion() = %v, want 1.0.0", version)
	}
}

func TestIsOnlineMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		expected bool
	}{
		{"online mode", "online", true},
		{"offline mode", "offline", false},
		{"empty mode", "", false},
		{"invalid mode", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Mode: tt.mode}
			result := config.IsOnlineMode()
			if result != tt.expected {
				t.Errorf("IsOnlineMode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("MODE", "offline")
	os.Setenv("NETWORK", "test")
	os.Setenv("PORT", "9090")
	defer func() {
		os.Unsetenv("MODE")
		os.Unsetenv("NETWORK")
		os.Unsetenv("PORT")
	}()

	config := &Config{
		Mode:    "online",
		Network: "solo",
		Port:    8080,
	}

	config.loadFromEnv()

	// Verify environment variables were loaded
	if config.Mode != "offline" {
		t.Errorf("loadFromEnv() Mode = %v, want offline", config.Mode)
	}
	if config.Network != "test" {
		t.Errorf("loadFromEnv() Network = %v, want test", config.Network)
	}
	if config.Port != 9090 {
		t.Errorf("loadFromEnv() Port = %v, want 9090", config.Port)
	}
}

func TestLoadFromEnvWithInvalidPort(t *testing.T) {
	// Set invalid port environment variable
	os.Setenv("PORT", "invalid")
	defer os.Unsetenv("PORT")

	config := &Config{Port: 8080}
	config.loadFromEnv()

	// Port should remain unchanged due to invalid value
	if config.Port != 8080 {
		t.Errorf("loadFromEnv() Port = %v, want 8080 (unchanged)", config.Port)
	}
}

func TestNewConfigWithMissingFile(t *testing.T) {
	// Change to a directory without config.json
	tempDir := t.TempDir()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test config loading with missing file
	_, err = NewConfig()
	if err == nil {
		t.Errorf("NewConfig() expected error for missing config file, got nil")
	}
}

func TestNewConfigWithInvalidJSON(t *testing.T) {
	// Create a temporary config file with invalid JSON
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	invalidJSON := `{"invalid": json}`
	if err := os.WriteFile(configPath, []byte(invalidJSON), 0644); err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	// Change to temp directory to test config loading
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test config loading with invalid JSON
	_, err = NewConfig()
	if err == nil {
		t.Errorf("NewConfig() expected error for invalid JSON, got nil")
	}
}
