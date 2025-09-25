package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	meshutils "github.com/vechain/mesh/utils"
)

// CreateRequestWithContext simulates the middleware by adding request body to context
// This is needed for tests because they don't go through the actual middleware
func CreateRequestWithContext(method, url string, body any) *http.Request {
	requestBody, _ := json.Marshal(body)
	req := httptest.NewRequest(method, url, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Simulate middleware by adding request body to context
	ctx := context.WithValue(req.Context(), meshutils.RequestBodyKey, requestBody)
	req = req.WithContext(ctx)

	return req
}

// CreateInvalidJSONRequest creates a request with invalid JSON body
func CreateInvalidJSONRequest(method, url string) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewBufferString(InvalidJSON))
	req.Header.Set("Content-Type", JSONContentType)
	return req
}

// CreateResponseRecorder creates a new ResponseRecorder
func CreateResponseRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// UnmarshalResponse unmarshals the response body into the target struct
func UnmarshalResponse(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(recorder.Body.Bytes(), target); err != nil {
		t.Errorf("%s: %v", FailedToUnmarshalResponse, err)
	}
}

// AssertStatusCode asserts that the response has the expected status code
func AssertStatusCode(t *testing.T, recorder *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if recorder.Code != expected {
		t.Errorf("%s %d, got %d", ExpectedStatus, expected, recorder.Code)
	}
}
