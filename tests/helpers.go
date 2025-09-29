package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	meshhttp "github.com/vechain/mesh/common/http"
)

// CreateRequestWithContext simulates the middleware by adding request body to context
// This is needed for tests because they don't go through the actual middleware
func CreateRequestWithContext(method, url string, body any) *http.Request {
	requestBody, _ := json.Marshal(body)
	req := httptest.NewRequest(method, url, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Simulate middleware by adding request body to context
	ctx := context.WithValue(req.Context(), meshhttp.RequestBodyKey, requestBody)
	req = req.WithContext(ctx)

	return req
}
