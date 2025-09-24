package thor

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	config := Config{
		NodeID:      "test-node",
		NetworkType: "solo",
		APIAddr:     "localhost:8669",
		P2PPort:     11235,
		OnDemand:    true,
		Persist:     false,
		APICORS:     "*",
	}

	// Test that config fields are set correctly
	if config.NodeID != "test-node" {
		t.Errorf("Config.NodeID = %v, want test-node", config.NodeID)
	}
	if config.NetworkType != "solo" {
		t.Errorf("Config.NetworkType = %v, want solo", config.NetworkType)
	}
	if config.APIAddr != "localhost:8669" {
		t.Errorf("Config.APIAddr = %v, want localhost:8669", config.APIAddr)
	}
	if config.P2PPort != 11235 {
		t.Errorf("Config.P2PPort = %v, want 11235", config.P2PPort)
	}
	if !config.OnDemand {
		t.Errorf("Config.OnDemand = %v, want true", config.OnDemand)
	}
	if config.Persist {
		t.Errorf("Config.Persist = %v, want false", config.Persist)
	}
	if config.APICORS != "*" {
		t.Errorf("Config.APICORS = %v, want *", config.APICORS)
	}
}

func TestNewServer_WithMockThorBinary(t *testing.T) {
	config := Config{
		NodeID:      "test-node",
		NetworkType: "solo",
		APIAddr:     "localhost:8669",
		P2PPort:     11235,
		OnDemand:    true,
		Persist:     false,
		APICORS:     "*",
	}

	// Test that we can create a Server struct manually
	ctx, cancel := context.WithCancel(context.Background())
	server := &Server{
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
		thorPath: "/mock/thor/path",
	}

	if server.config != config {
		t.Errorf("Server config = %v, want %v", server.config, config)
	}
	if server.ctx == nil {
		t.Errorf("Server ctx is nil")
	}
	if server.cancel == nil {
		t.Errorf("Server cancel is nil")
	}
	if server.thorPath != "/mock/thor/path" {
		t.Errorf("Server thorPath = %v, want /mock/thor/path", server.thorPath)
	}
}

func TestServer_Stop_NoProcess(t *testing.T) {
	// Create a server without starting a process
	server := &Server{
		config: Config{
			NodeID:      "test-node",
			NetworkType: "solo",
			APIAddr:     "localhost:8669",
			P2PPort:     11235,
		},
		ctx:    context.Background(),
		cancel: func() {},
	}

	// Stop should not return an error when no process is running
	err := server.Stop()
	if err != nil {
		t.Errorf("Stop() with no process should not return error, got: %v", err)
	}
}

func TestServer_Stop_WithMockProcess(t *testing.T) {
	// Create a server with a mock process
	ctx, cancel := context.WithCancel(context.Background())
	server := &Server{
		config: Config{
			NodeID:      "test-node",
			NetworkType: "solo",
			APIAddr:     "localhost:8669",
			P2PPort:     11235,
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Create a mock process that will exit immediately
	server.process = &exec.Cmd{
		Path: "/bin/echo", // Use a simple command that exits quickly
		Args: []string{"echo", "test"},
	}

	// Start the mock process
	if err := server.process.Start(); err != nil {
		t.Fatalf("Failed to start mock process: %v", err)
	}

	// Wait a bit for the process to start
	time.Sleep(100 * time.Millisecond)

	// Stop should work without error
	err := server.Stop()
	if err != nil {
		t.Errorf("Stop() should not return error, got: %v", err)
	}
}

func TestServer_AttachToPublicNetworkAndStart_NoThorBinary(t *testing.T) {
	// Create a server with a non-existent Thor binary
	server := &Server{
		config: Config{
			NodeID:      "test-node",
			NetworkType: "test",
			APIAddr:     "localhost:8669",
			P2PPort:     11235,
		},
		ctx:      context.Background(),
		cancel:   func() {},
		thorPath: "/non/existent/thor",
	}

	// This should fail because Thor binary doesn't exist
	err := server.AttachToPublicNetworkAndStart()
	if err == nil {
		t.Errorf("AttachToPublicNetworkAndStart() should return error when Thor binary doesn't exist")
	}
}

func TestServer_StartSoloNode_NoThorBinary(t *testing.T) {
	// Create a server with a non-existent Thor binary
	server := &Server{
		config: Config{
			NodeID:      "test-node",
			NetworkType: "solo",
			APIAddr:     "localhost:8669",
			P2PPort:     11235,
			OnDemand:    true,
			Persist:     false,
			APICORS:     "*",
		},
		ctx:      context.Background(),
		cancel:   func() {},
		thorPath: "/non/existent/thor",
	}

	// This should fail because Thor binary doesn't exist
	err := server.StartSoloNode()
	if err == nil {
		t.Errorf("StartSoloNode() should return error when Thor binary doesn't exist")
	}
}

func TestServer_StartSoloNode_WithMockThorBinary(t *testing.T) {
	// Create a temporary directory and a mock Thor binary
	tempDir := t.TempDir()
	mockThorPath := filepath.Join(tempDir, "thor")

	// Create a mock Thor binary that will exit immediately
	file, err := os.Create(mockThorPath)
	if err != nil {
		t.Fatalf("Failed to create mock Thor binary: %v", err)
	}
	err = file.Close()
	if err != nil {
		t.Fatalf("Failed to close mock Thor binary: %v", err)
	}

	// Make it executable
	if err := os.Chmod(mockThorPath, 0755); err != nil {
		t.Fatalf("Failed to make mock Thor binary executable: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	server := &Server{
		config: Config{
			NodeID:      "test-node",
			NetworkType: "solo",
			APIAddr:     "localhost:8669",
			P2PPort:     11235,
			OnDemand:    true,
			Persist:     false,
			APICORS:     "*",
		},
		ctx:      ctx,
		cancel:   cancel,
		thorPath: mockThorPath,
	}

	// This will fail because the mock binary exits immediately
	// but we can test that the function attempts to start the process
	err = server.StartSoloNode()
	if err == nil {
		t.Errorf("StartSoloNode() should return error when mock Thor binary exits immediately")
	}
}
