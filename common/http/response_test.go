package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	meshcommon "github.com/vechain/mesh/common"
)

func TestWriteJSONResponse(t *testing.T) {
	// Test successful JSON response
	w := httptest.NewRecorder()
	response := map[string]string{"message": "test"}

	NewResponseHandler().WriteJSONResponse(w, response)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("WriteJSONResponse() status = %v, want %v", w.Code, http.StatusOK)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("WriteJSONResponse() Content-Type = %v, want application/json", contentType)
	}

	// Check response body
	var result map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &result)
	if err != nil {
		t.Errorf("WriteJSONResponse() failed to unmarshal response: %v", err)
	}

	if result["message"] != "test" {
		t.Errorf("WriteJSONResponse() message = %v, want test", result["message"])
	}

	// Test with invalid JSON (should not happen in practice, but test error handling)
	w2 := httptest.NewRecorder()
	invalidResponse := make(chan int) // channels cannot be marshaled to JSON

	NewResponseHandler().WriteJSONResponse(w2, invalidResponse)

	// Should return 500 error for invalid JSON
	if w2.Code != http.StatusInternalServerError {
		t.Errorf("WriteJSONResponse() status = %v, want %v", w2.Code, http.StatusInternalServerError)
	}
}

func TestWriteErrorResponse(t *testing.T) {
	// Test error response with valid error
	error := meshcommon.GetError(500)
	w := httptest.NewRecorder()

	NewResponseHandler().WriteErrorResponse(w, error, http.StatusBadRequest)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}

	if result["error"] == nil {
		t.Error("Expected error field in response")
	}
}

func TestWriteErrorResponse_NilError(t *testing.T) {
	// Test error response with nil error (should use default)
	w := httptest.NewRecorder()

	NewResponseHandler().WriteErrorResponse(w, nil, http.StatusInternalServerError)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}

	if result["error"] == nil {
		t.Error("Expected error field in response")
	}
}
