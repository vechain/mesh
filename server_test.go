package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/types"

	meshcommon "github.com/vechain/mesh/common"
	meshconfig "github.com/vechain/mesh/config"
)

func createTestAsserter() (*asserter.Asserter, error) {
	supportedOperationTypes := []string{
		meshcommon.OperationTypeTransfer,
		meshcommon.OperationTypeFee,
		meshcommon.OperationTypeFeeDelegation,
		meshcommon.OperationTypeContractCall,
	}

	return asserter.NewServer(
		supportedOperationTypes,
		true, // historical balance lookup
		[]*types.NetworkIdentifier{
			{
				Blockchain: meshcommon.BlockchainName,
				Network:    meshcommon.TestNetwork,
			},
		},
		nil,
		false,
		"",
	)
}

func TestNewVeChainMeshServer(t *testing.T) {
	config := &meshconfig.Config{
		NodeAPI: "http://localhost:8669",
		Network: meshcommon.TestNetwork,
		Mode:    meshcommon.OnlineMode,
	}

	asrt, err := createTestAsserter()
	if err != nil {
		t.Fatalf("Failed to create asserter: %v", err)
	}

	server, err := NewVeChainMeshServer(config, asrt)
	if err != nil {
		t.Fatalf("NewVeChainMeshServer() error = %v", err)
	}

	if server == nil {
		t.Fatal("NewVeChainMeshServer() returned nil")
	}

	if server.server == nil {
		t.Error("NewVeChainMeshServer() server is nil")
	}

	if server.asserter == nil {
		t.Error("NewVeChainMeshServer() asserter is nil")
	}

	if server.config != config {
		t.Error("NewVeChainMeshServer() config mismatch")
	}
}

func TestVeChainMeshServer_Start(t *testing.T) {
	config := &meshconfig.Config{
		NodeAPI: "http://localhost:8669",
		Network: meshcommon.TestNetwork,
		Mode:    meshcommon.OnlineMode,
		Port:    8080,
	}

	asrt, err := createTestAsserter()
	if err != nil {
		t.Fatalf("Failed to create asserter: %v", err)
	}

	server, err := NewVeChainMeshServer(config, asrt)
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
		Network: meshcommon.TestNetwork,
		Mode:    meshcommon.OnlineMode,
		Port:    8080,
	}

	asrt, err := createTestAsserter()
	if err != nil {
		t.Fatalf("Failed to create asserter: %v", err)
	}

	server, err := NewVeChainMeshServer(config, asrt)
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
		Network: meshcommon.TestNetwork,
		Mode:    meshcommon.OnlineMode,
	}

	asrt, err := createTestAsserter()
	if err != nil {
		t.Fatalf("Failed to create asserter: %v", err)
	}

	server, err := NewVeChainMeshServer(config, asrt)
	if err != nil {
		t.Fatalf("NewVeChainMeshServer() error = %v", err)
	}

	// Test network/list endpoint
	req := httptest.NewRequest("POST", meshcommon.NetworkListEndpoint, bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.server.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("network/list status code = %v, want %v, body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestVeChainMeshServer_AccountEndpoints(t *testing.T) {
	config := &meshconfig.Config{
		NodeAPI: "http://localhost:8669",
		Network: meshcommon.TestNetwork,
		Mode:    meshcommon.OnlineMode,
	}

	asrt, err := createTestAsserter()
	if err != nil {
		t.Fatalf("Failed to create asserter: %v", err)
	}

	server, err := NewVeChainMeshServer(config, asrt)
	if err != nil {
		t.Fatalf("NewVeChainMeshServer() error = %v", err)
	}

	// Test account/balance endpoint with invalid request
	req := httptest.NewRequest("POST", meshcommon.AccountBalanceEndpoint, bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.server.Handler.ServeHTTP(w, req)

	// Should return 400 or 500 for invalid JSON (asserter validates)
	if w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("AccountEndpoints() status code = %v, want 400 or 500", w.Code)
	}
}

func TestVeChainMeshServer_ConstructionEndpoints(t *testing.T) {
	config := &meshconfig.Config{
		NodeAPI: "http://localhost:8669",
		Network: meshcommon.TestNetwork,
		Mode:    meshcommon.OnlineMode,
	}

	asrt, err := createTestAsserter()
	if err != nil {
		t.Fatalf("Failed to create asserter: %v", err)
	}

	server, err := NewVeChainMeshServer(config, asrt)
	if err != nil {
		t.Fatalf("NewVeChainMeshServer() error = %v", err)
	}

	// Test construction/derive endpoint with invalid request
	req := httptest.NewRequest("POST", meshcommon.ConstructionDeriveEndpoint, bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.server.Handler.ServeHTTP(w, req)

	// Should return 400 or 500 for invalid JSON (asserter validates)
	if w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("ConstructionEndpoints() status code = %v, want 400 or 500", w.Code)
	}
}

func TestVeChainMeshServer_GetEndpoints(t *testing.T) {
	config := &meshconfig.Config{
		NodeAPI: "http://localhost:8669",
		Network: meshcommon.TestNetwork,
		Mode:    meshcommon.OnlineMode,
	}

	asrt, err := createTestAsserter()
	if err != nil {
		t.Fatalf("Failed to create asserter: %v", err)
	}

	server, err := NewVeChainMeshServer(config, asrt)
	if err != nil {
		t.Fatalf("NewVeChainMeshServer() error = %v", err)
	}

	endpoints, err := server.GetEndpoints()
	if err != nil {
		t.Fatalf("GetEndpoints() error = %v", err)
	}

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
	missingEndpoints := []string{}
	for endpoint, found := range expectedEndpoints {
		if !found {
			missingEndpoints = append(missingEndpoints, endpoint)
		}
	}

	if len(missingEndpoints) > 0 {
		t.Errorf("Expected endpoints not found: %v", missingEndpoints)
	}

	// Check that we have the expected number of endpoints
	expectedCount := len(expectedEndpoints)
	if len(endpoints) != expectedCount {
		t.Errorf("GetEndpoints() returned %d endpoints, expected %d. Endpoints: %v", len(endpoints), expectedCount, endpoints)
	}
}

func TestVeChainMeshServer_NetworkListWithValidRequest(t *testing.T) {
	config := &meshconfig.Config{
		NodeAPI: "http://localhost:8669",
		Network: meshcommon.TestNetwork,
		Mode:    meshcommon.OnlineMode,
	}

	asrt, err := createTestAsserter()
	if err != nil {
		t.Fatalf("Failed to create asserter: %v", err)
	}

	server, err := NewVeChainMeshServer(config, asrt)
	if err != nil {
		t.Fatalf("NewVeChainMeshServer() error = %v", err)
	}

	// Create valid request body
	requestBody := map[string]any{}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", meshcommon.NetworkListEndpoint, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.server.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("network/list status code = %v, want %v, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	// Parse response
	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check that network_identifiers exists
	if _, ok := response["network_identifiers"]; !ok {
		t.Error("Response missing network_identifiers field")
	}
}

func TestVeChainMeshServer_HealthEndpoint(t *testing.T) {
	config := &meshconfig.Config{
		NodeAPI: "http://localhost:8669",
		Network: meshcommon.TestNetwork,
		Mode:    meshcommon.OfflineMode,
		Port:    8080,
	}

	asrt, err := createTestAsserter()
	if err != nil {
		t.Fatalf("Failed to create asserter: %v", err)
	}

	server, err := NewVeChainMeshServer(config, asrt)
	if err != nil {
		t.Fatalf("NewVeChainMeshServer() error = %v", err)
	}

	// Test health endpoint
	req := httptest.NewRequest(http.MethodGet, meshcommon.HealthEndpoint, nil)
	w := httptest.NewRecorder()

	server.server.Handler.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("health endpoint status code = %v, want %v", w.Code, http.StatusOK)
	}

	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("health endpoint content-type = %v, want application/json", contentType)
	}

	// Parse and verify response body
	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal health response: %v", err)
	}

	if status, ok := response["status"]; !ok || status != "ok" {
		t.Errorf("health endpoint response = %v, want {\"status\":\"ok\"}", response)
	}
}
