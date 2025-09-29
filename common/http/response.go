package http

import (
	"encoding/json"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
)

type ResponseHandler struct{}

func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

// WriteJSONResponse writes a JSON response with proper error handling
func (h *ResponseHandler) WriteJSONResponse(w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// WriteErrorResponse writes an error response in Mesh format
func (h *ResponseHandler) WriteErrorResponse(w http.ResponseWriter, err *types.Error, statusCode int) {
	if err == nil {
		err = meshcommon.GetError(500) // Default to internal server error
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]any{
		"error": err,
	}

	if encodeErr := json.NewEncoder(w).Encode(errorResponse); encodeErr != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		return
	}
}
