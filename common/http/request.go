package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type RequestBodyKeyType string

const RequestBodyKey RequestBodyKeyType = "request_body"

type RequestHandler struct{}

func NewRequestHandler() *RequestHandler {
	return &RequestHandler{}
}

// ParseJSONFromContext parses JSON from the request context into the target struct
func (h *RequestHandler) ParseJSONFromContext(r *http.Request, target any) error {
	body, ok := r.Context().Value(RequestBodyKey).([]byte)
	if !ok {
		return fmt.Errorf("request body not found in context - middleware may not be properly configured")
	}
	return json.Unmarshal(body, target)
}
