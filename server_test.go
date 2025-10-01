package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	meshcommon "github.com/vechain/mesh/common"
	meshconfig "github.com/vechain/mesh/config"
)

func TestNewVeChainMeshServer(t *testing.T) {
	config := &meshconfig.Config{
		NodeAPI: "http://localhost:8669",
		Network: "test",
		Mode:    "online",
	}

	server, err := NewVeChainMeshServer(config)
	if err != nil {
		t.Fatalf("NewVeChainMeshServer() error = %v", err)
	}

	if server == nil {
		t.Errorf("NewVeChainMeshServer() returned nil")
	} else {
		if server.router == nil {
			t.Errorf("NewVeChainMeshServer() router is nil")
		}
		if server.networkService == nil {
			t.Errorf("NewVeChainMeshServer() networkService is nil")
		}
		if server.accountService == nil {
			t.Errorf("NewVeChainMeshServer() accountService is nil")
		}
		if server.constructionService == nil {
			t.Errorf("NewVeChainMeshServer() constructionService is nil")
		}
		if server.blockService == nil {
			t.Errorf("NewVeChainMeshServer() blockService is nil")
		}
		if server.mempoolService == nil {
			t.Errorf("NewVeChainMeshServer() mempoolService is nil")
		}
	}
}

func TestVeChainMeshServer_HealthCheck(t *testing.T) {
	config := &meshconfig.Config{
		NodeAPI: "http://localhost:8669",
		Network: "test",
		Mode:    "online",
	}

	server, err := NewVeChainMeshServer(config)
	if err != nil {
		t.Fatalf("NewVeChainMeshServer() error = %v", err)
	}

	req := httptest.NewRequest("GET", meshcommon.HealthEndpoint, nil)
	w := httptest.NewRecorder()

	server.healthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("healthCheck() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Check response body
	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("healthCheck() response status = %v, want healthy", response["status"])
	}
}

func TestVeChainMeshServer_Start(t *testing.T) {
	config := &meshconfig.Config{
		NodeAPI: "http://localhost:8669",
		Network: "test",
		Mode:    "online",
		Port:    8080,
	}

	server, err := NewVeChainMeshServer(config)
	if err != nil {
		t.Fatalf("NewVeChainMeshServer() error = %v", err)
	}

	// Start server in a goroutine
	serverStarted := make(chan bool)
	go func() {
		serverStarted <- true
		if err := server.Start(); err != nil && err.Error() != "http: Server closed" {
			t.Errorf("Start() error = %v", err)
		}
	}()

	// Wait for server to start
	<-serverStarted
	time.Sleep(50 * time.Millisecond)

	// Stop the server
	ctx := context.Background()
	if err := server.Stop(ctx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestVeChainMeshServer_Stop(t *testing.T) {
	config := &meshconfig.Config{
		NodeAPI: "http://localhost:8669",
		Network: "test",
		Mode:    "online",
		Port:    8080,
	}

	server, err := NewVeChainMeshServer(config)
	if err != nil {
		t.Fatalf("NewVeChainMeshServer() error = %v", err)
	}

	// Stop should not return an error even if the server wasn't started
	ctx := context.Background()
	if err := server.Stop(ctx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestVeChainMeshServer_NetworkEndpoints(t *testing.T) {
	config := &meshconfig.Config{
		NodeAPI: "http://localhost:8669",
		Network: "test",
		Mode:    "online",
	}

	server, err := NewVeChainMeshServer(config)
	if err != nil {
		t.Fatalf("NewVeChainMeshServer() error = %v", err)
	}

	// Test network/list endpoint
	req := httptest.NewRequest("POST", meshcommon.NetworkListEndpoint, bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("network/list status code = %v, want %v", w.Code, http.StatusOK)
	}
}

func TestVeChainMeshServer_AccountEndpoints(t *testing.T) {
	config := &meshconfig.Config{
		NodeAPI: "http://localhost:8669",
		Network: "test",
		Mode:    "online",
	}

	server, err := NewVeChainMeshServer(config)
	if err != nil {
		t.Fatalf("NewVeChainMeshServer() error = %v", err)
	}

	// Test account/balance endpoint with invalid request (to avoid validation issues)
	req := httptest.NewRequest("POST", meshcommon.AccountBalanceEndpoint, bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	// Should return 400 for invalid JSON
	if w.Code != http.StatusBadRequest {
		t.Errorf("AccountEndpoints() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestVeChainMeshServer_ConstructionEndpoints(t *testing.T) {
	config := &meshconfig.Config{
		NodeAPI: "http://localhost:8669",
		Network: "test",
		Mode:    "online",
	}

	server, err := NewVeChainMeshServer(config)
	if err != nil {
		t.Fatalf("NewVeChainMeshServer() error = %v", err)
	}

	// Test construction/derive endpoint with invalid request (to avoid validation issues)
	req := httptest.NewRequest("POST", meshcommon.ConstructionDeriveEndpoint, bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	// Should return 400 for invalid JSON
	if w.Code != http.StatusBadRequest {
		t.Errorf("ConstructionEndpoints() status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestVeChainMeshServer_GetEndpoints(t *testing.T) {
	config := &meshconfig.Config{
		NodeAPI: "http://localhost:8669",
		Network: "test",
		Mode:    "online",
	}

	server, err := NewVeChainMeshServer(config)
	if err != nil {
		t.Fatalf("NewVeChainMeshServer() error = %v", err)
	}

	endpoints, err := server.GetEndpoints()
	if err != nil {
		t.Fatalf("GetEndpoints() error = %v", err)
	}

	fmt.Println("endpoints", endpoints)

	// Check that we have endpoints
	if len(endpoints) == 0 {
		t.Errorf("GetEndpoints() returned empty slice")
	}

	// Check for specific expected endpoints
	expectedEndpoints := map[string]bool{
		"GET /health":                   false,
		"POST /network/list":            false,
		"POST /network/status":          false,
		"POST /network/options":         false,
		"POST /account/balance":         false,
		"POST /construction/derive":     false,
		"POST /construction/preprocess": false,
		"POST /construction/metadata":   false,
		"POST /construction/payloads":   false,
		"POST /construction/parse":      false,
		"POST /construction/combine":    false,
		"POST /construction/hash":       false,
		"POST /construction/submit":     false,
		"POST /block":                   false,
		"POST /block/transaction":       false,
		"POST /mempool":                 false,
		"POST /mempool/transaction":     false,
		"POST /events/blocks":           false,
		"POST /search/transactions":     false,
		"POST /call":                    false,
	}

	// Mark found endpoints
	for _, endpoint := range endpoints {
		if _, exists := expectedEndpoints[endpoint]; exists {
			expectedEndpoints[endpoint] = true
		}
	}

	// Check that all expected endpoints are present
	for endpoint, found := range expectedEndpoints {
		if !found {
			t.Errorf("Expected endpoint not found: %s", endpoint)
		}
	}

	// Check that we have the expected number of endpoints
	expectedCount := len(expectedEndpoints)
	if len(endpoints) != expectedCount {
		t.Errorf("GetEndpoints() returned %d endpoints, expected %d", len(endpoints), expectedCount)
	}
}
