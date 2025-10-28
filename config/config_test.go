package config

import (
	"bytes"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	meshcommon "github.com/vechain/mesh/common"
)

func TestNewConfig(t *testing.T) {
	// Change to project root directory to find config.json
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to change to original directory: %v", err)
		}
	}()

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
		if config.NetworkIdentifier.Blockchain != meshcommon.BlockchainName {
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
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to change to original directory: %v", err)
		}
	}()

	// Go up one directory to project root
	if err := os.Chdir(".."); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	// Set environment variables
	err = os.Setenv("MODE", meshcommon.OfflineMode)
	if err != nil {
		t.Fatalf("Failed to set MODE environment variable: %v", err)
	}
	err = os.Setenv("NETWORK", meshcommon.TestNetwork)
	if err != nil {
		t.Fatalf("Failed to set NETWORK environment variable: %v", err)
	}
	err = os.Setenv("PORT", "9090")
	if err != nil {
		t.Fatalf("Failed to set PORT environment variable: %v", err)
	}
	defer func() {
		err = os.Unsetenv("MODE")
		if err != nil {
			t.Fatalf("Failed to unset MODE environment variable: %v", err)
		}
		err = os.Unsetenv("NETWORK")
		if err != nil {
			t.Fatalf("Failed to unset NETWORK environment variable: %v", err)
		}
		err = os.Unsetenv("PORT")
		if err != nil {
			t.Fatalf("Failed to unset PORT environment variable: %v", err)
		}
	}()

	// Test config loading with environment variables
	config, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() failed: %v", err)
	}

	// Verify environment variables override
	if config.Mode != meshcommon.OfflineMode {
		t.Errorf("NewConfig() Mode = %v, want offline", config.Mode)
	}
	if config.Network != meshcommon.TestNetwork {
		t.Errorf("NewConfig() Network = %v, want test", config.Network)
	}
	if config.Port != 9090 {
		t.Errorf("NewConfig() Port = %v, want 9090", config.Port)
	}

	// Verify derived fields reflect environment changes
	if config.NetworkIdentifier == nil {
		t.Errorf("NewConfig() NetworkIdentifier is nil")
	} else {
		if config.NetworkIdentifier.Network != meshcommon.TestNetwork {
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
		chainTag        byte
		expectedTag     byte
		expectedNetwork string
	}{
		{
			name:            "main network",
			network:         meshcommon.MainNetwork,
			chainTag:        0,
			expectedTag:     0x4a,
			expectedNetwork: meshcommon.MainNetwork,
		},
		{
			name:            "mainnet network",
			network:         "mainnet",
			chainTag:        0,
			expectedTag:     0x4a,
			expectedNetwork: meshcommon.MainNetwork,
		},
		{
			name:            "test network",
			network:         meshcommon.TestNetwork,
			chainTag:        0,
			expectedTag:     0x27,
			expectedNetwork: meshcommon.TestNetwork,
		},
		{
			name:            "testnet network",
			network:         "testnet",
			chainTag:        0,
			expectedTag:     0x27,
			expectedNetwork: meshcommon.TestNetwork,
		},
		{
			name:            "solo network",
			network:         meshcommon.SoloNetwork,
			chainTag:        0,
			expectedTag:     0xf6,
			expectedNetwork: meshcommon.SoloNetwork,
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
			network:         meshcommon.SoloNetwork,
			chainTag:        0xf6,
			expectedTag:     0xf6,
			expectedNetwork: meshcommon.SoloNetwork,
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
				if config.NetworkIdentifier.Blockchain != meshcommon.BlockchainName {
					t.Errorf("setDerivedFields() NetworkIdentifier.Blockchain = %v, want vechainthor", config.NetworkIdentifier.Blockchain)
				}
				if config.NetworkIdentifier.Network != tt.expectedNetwork {
					t.Errorf("setDerivedFields() NetworkIdentifier.Network = %v, want %v", config.NetworkIdentifier.Network, tt.expectedNetwork)
				}
			}
		})
	}
}

func TestIsOnlineMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		expected bool
	}{
		{"online mode", meshcommon.OnlineMode, true},
		{"offline mode", meshcommon.OfflineMode, false},
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
	err := os.Setenv("MODE", meshcommon.OfflineMode)
	if err != nil {
		t.Fatalf("Failed to set MODE environment variable: %v", err)
	}
	err = os.Setenv("NETWORK", meshcommon.TestNetwork)
	if err != nil {
		t.Fatalf("Failed to set NETWORK environment variable: %v", err)
	}
	err = os.Setenv("PORT", "9090")
	if err != nil {
		t.Fatalf("Failed to set PORT environment variable: %v", err)
	}
	defer func() {
		err = os.Unsetenv("MODE")
		if err != nil {
			t.Fatalf("Failed to unset MODE environment variable: %v", err)
		}
		err = os.Unsetenv("NETWORK")
		if err != nil {
			t.Fatalf("Failed to unset NETWORK environment variable: %v", err)
		}
		err = os.Unsetenv("PORT")
		if err != nil {
			t.Fatalf("Failed to unset PORT environment variable: %v", err)
		}
	}()

	config := &Config{
		Mode:    meshcommon.OnlineMode,
		Network: meshcommon.SoloNetwork,
		Port:    8080,
	}

	config.loadFromEnv()

	// Verify environment variables were loaded
	if config.Mode != meshcommon.OfflineMode {
		t.Errorf("loadFromEnv() Mode = %v, want offline", config.Mode)
	}
	if config.Network != meshcommon.TestNetwork {
		t.Errorf("loadFromEnv() Network = %v, want test", config.Network)
	}
	if config.Port != 9090 {
		t.Errorf("loadFromEnv() Port = %v, want 9090", config.Port)
	}
}

func TestLoadFromEnvWithInvalidPort(t *testing.T) {
	// Set invalid port environment variable
	err := os.Setenv("PORT", "invalid")
	if err != nil {
		t.Fatalf("Failed to set PORT environment variable: %v", err)
	}
	defer func() {
		err = os.Unsetenv("PORT")
		if err != nil {
			t.Fatalf("Failed to unset PORT environment variable: %v", err)
		}
	}()

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
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to change to original directory: %v", err)
		}
	}()

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
	// Create a temporary config dir with invalid JSON file
	tempDir := t.TempDir()
	cfgDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	configPath := filepath.Join(cfgDir, "config.json")

	invalidJSON := `{"invalid": json}`
	if err := os.WriteFile(configPath, []byte(invalidJSON), 0644); err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	// Change to temp directory to test config loading
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to change to original directory: %v", err)
		}
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test config loading with invalid JSON
	_, err = NewConfig()
	if err == nil {
		t.Errorf("NewConfig() expected error for invalid JSON, got nil")
	}
}

func TestNewConfig_ConfigDirExistsButFileMissing(t *testing.T) {
	// Create a temporary config dir without config.json
	tempDir := t.TempDir()
	cfgDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to change to original directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Expect error opening config file
	_, err = NewConfig()
	if err == nil {
		t.Errorf("NewConfig() expected error when config.json is missing inside config/, got nil")
	}
}

func TestPrintConfig(t *testing.T) {
	config := &Config{
		ServiceName: "Test Service",
		Port:        8080,
		MeshVersion: "1.0.0",
		Mode:        meshcommon.OnlineMode,
		NodeAPI:     "http://localhost:8669",
		NodeVersion: "1.0.0",
		Network:     meshcommon.MainNetwork,
		ChainTag:    0x27,
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	config.PrintConfig()

	// Restore stdout
	err := w.Close()
	if err != nil {
		t.Fatalf("Failed to close pipe: %v", err)
	}
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Failed to copy from pipe: %v", err)
	}
	output := buf.String()

	// Check that the output contains expected fields
	expectedFields := []string{
		"Test Service",
		"8080",
		"1.0.0",
		meshcommon.OnlineMode,
		"http://localhost:8669",
		meshcommon.MainNetwork,
		"0x27",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("PrintConfig() output should contain %s, got: %s", field, output)
		}
	}
}

func TestNewConfig_UsesRootScopedRead(t *testing.T) {
	// Prepare a temporary directory with a config/ subdirectory
	tempDir := t.TempDir()
	cfgDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Minimal valid config JSON
	jsonContent := `{
        "meshVersion": "test",
        "port": 8081,
        "mode": "online",
        "network": "test",
        "nodeApi": "http://localhost:8669",
        "apiVersion": "v1",
        "nodeVersion": "v1",
        "serviceName": "svc",
        "baseGasPrice": "0",
        "initialBaseFee": "0",
        "expiration": 720
    }`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(jsonContent), 0o644); err != nil {
		t.Fatalf("failed to write temp config.json: %v", err)
	}

	// Change to tempDir so NewConfig opens config/config.json from the secure root
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to change to original directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to chdir to temp dir: %v", err)
	}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() failed: %v", err)
	}
	if cfg.Port != 8081 {
		t.Errorf("expected port 8081, got %d", cfg.Port)
	}
	if cfg.NetworkIdentifier == nil {
		t.Errorf("expected NetworkIdentifier to be set")
	}
}

func TestGetBaseGasPrice(t *testing.T) {
	tests := []struct {
		name           string
		baseGasPrice   string
		expectedResult *big.Int
	}{
		{
			name:           "valid gas price",
			baseGasPrice:   "1000000000000000000",
			expectedResult: big.NewInt(1000000000000000000),
		},
		{
			name:           "empty gas price",
			baseGasPrice:   "",
			expectedResult: nil,
		},
		{
			name:           "invalid gas price",
			baseGasPrice:   "invalid",
			expectedResult: nil,
		},
		{
			name:           "zero gas price",
			baseGasPrice:   "0",
			expectedResult: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				BaseGasPrice: tt.baseGasPrice,
			}

			result := config.GetBaseGasPrice()

			if tt.expectedResult == nil {
				if result != nil {
					t.Errorf("GetBaseGasPrice() expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("GetBaseGasPrice() expected %v, got nil", tt.expectedResult)
				} else if result.Cmp(tt.expectedResult) != 0 {
					t.Errorf("GetBaseGasPrice() expected %v, got %v", tt.expectedResult, result)
				}
			}
		})
	}
}

// TODO: Delete the test once Thor is updated again in this regard
func TestLoadFromEnv_SoloOnDemand(t *testing.T) {
	t.Run("sets true when SOLO_ONDEMAND=true", func(t *testing.T) {
		err := os.Setenv("SOLO_ONDEMAND", "true")
		if err != nil {
			t.Fatalf("failed to set env: %v", err)
		}
		defer func() {
			err = os.Unsetenv("SOLO_ONDEMAND")
			if err != nil {
				t.Fatalf("failed to unset env: %v", err)
			}
		}()

		cfg := &Config{SoloOnDemand: false}
		cfg.loadFromEnv()
		if !cfg.SoloOnDemand {
			t.Errorf("SoloOnDemand = %v, want true", cfg.SoloOnDemand)
		}
	})

	t.Run("sets false when SOLO_ONDEMAND=false", func(t *testing.T) {
		err := os.Setenv("SOLO_ONDEMAND", "false")
		if err != nil {
			t.Fatalf("failed to set env: %v", err)
		}
		defer func() {
			err = os.Unsetenv("SOLO_ONDEMAND")
			if err != nil {
				t.Fatalf("failed to unset env: %v", err)
			}
		}()

		cfg := &Config{SoloOnDemand: true}
		cfg.loadFromEnv()
		if cfg.SoloOnDemand {
			t.Errorf("SoloOnDemand = %v, want false", cfg.SoloOnDemand)
		}
	})

	t.Run("ignores invalid values", func(t *testing.T) {
		err := os.Setenv("SOLO_ONDEMAND", "notabool")
		if err != nil {
			t.Fatalf("failed to set env: %v", err)
		}
		defer func() {
			err = os.Unsetenv("SOLO_ONDEMAND")
			if err != nil {
				t.Fatalf("failed to unset env: %v", err)
			}
		}()

		cfg := &Config{SoloOnDemand: false}
		cfg.loadFromEnv()
		if cfg.SoloOnDemand {
			t.Errorf("SoloOnDemand changed on invalid value, want unchanged false")
		}
	})
}
