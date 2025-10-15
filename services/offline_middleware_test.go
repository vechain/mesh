package services

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	meshcommon "github.com/vechain/mesh/common"
	meshconfig "github.com/vechain/mesh/config"
)

func TestOfflineModeMiddleware_BlocksOnlineEndpoints(t *testing.T) {
	// Test that online-only endpoints are blocked in offline mode
	offlineConfig := &meshconfig.Config{
		Mode: meshcommon.OfflineMode,
	}

	// Create a simple handler that always returns 200 OK
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	// Wrap with offline middleware
	middleware := OfflineModeMiddleware(offlineConfig)
	wrappedHandler := middleware(handler)

	// Test each online-only endpoint
	for _, endpoint := range onlineOnlyEndpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest("POST", endpoint, nil)
			rec := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rec, req)

			// Should return error (500)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf("Expected status %d for %s in offline mode, got %d", http.StatusInternalServerError, endpoint, rec.Code)
			}

			// Should contain offline error message
			body := rec.Body.String()
			if !strings.Contains(body, "offline") && !strings.Contains(body, "API does not support offline mode") {
				t.Errorf("Expected offline error message for %s, got: %s", endpoint, body)
			}
		})
	}
}

func TestOfflineModeMiddleware_AllowsOfflineEndpoints(t *testing.T) {
	// Test that offline-compatible endpoints work in offline mode
	offlineConfig := &meshconfig.Config{
		Mode: meshcommon.OfflineMode,
	}

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	middleware := OfflineModeMiddleware(offlineConfig)
	wrappedHandler := middleware(handler)

	offlineEndpoints := []string{
		meshcommon.NetworkListEndpoint,
		meshcommon.NetworkOptionsEndpoint,
		meshcommon.ConstructionDeriveEndpoint,
		meshcommon.ConstructionPreprocessEndpoint,
		meshcommon.ConstructionPayloadsEndpoint,
		meshcommon.ConstructionParseEndpoint,
		meshcommon.ConstructionCombineEndpoint,
		meshcommon.ConstructionHashEndpoint,
	}

	for _, endpoint := range offlineEndpoints {
		t.Run(endpoint, func(t *testing.T) {
			called = false
			req := httptest.NewRequest("POST", endpoint, nil)
			rec := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rec, req)

			// Should call the underlying handler
			if !called {
				t.Errorf("Handler should be called for %s in offline mode", endpoint)
			}

			// Should return success (200)
			if rec.Code != http.StatusOK {
				t.Errorf("Expected status %d for %s in offline mode, got %d", http.StatusOK, endpoint, rec.Code)
			}
		})
	}
}

func TestOfflineModeMiddleware_AllowsAllInOnlineMode(t *testing.T) {
	// Test that all endpoints work in online mode
	onlineConfig := &meshconfig.Config{
		Mode: meshcommon.OnlineMode,
	}

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	middleware := OfflineModeMiddleware(onlineConfig)
	wrappedHandler := middleware(handler)

	// Test all endpoints, including online-only ones
	allEndpoints := append(onlineOnlyEndpoints,
		meshcommon.NetworkListEndpoint,
		meshcommon.NetworkOptionsEndpoint,
		meshcommon.ConstructionDeriveEndpoint,
		meshcommon.ConstructionPreprocessEndpoint,
		meshcommon.ConstructionPayloadsEndpoint,
		meshcommon.ConstructionParseEndpoint,
		meshcommon.ConstructionCombineEndpoint,
		meshcommon.ConstructionHashEndpoint,
	)

	for _, endpoint := range allEndpoints {
		t.Run(endpoint, func(t *testing.T) {
			called = false
			req := httptest.NewRequest("POST", endpoint, nil)
			rec := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rec, req)

			// Should call the underlying handler
			if !called {
				t.Errorf("Handler should be called for %s in online mode", endpoint)
			}

			// Should return success (200)
			if rec.Code != http.StatusOK {
				t.Errorf("Expected status %d for %s in online mode, got %d", http.StatusOK, endpoint, rec.Code)
			}
		})
	}
}
