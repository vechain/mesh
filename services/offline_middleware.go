package services

import (
	"encoding/json"
	"net/http"
	"strings"

	meshcommon "github.com/vechain/mesh/common"
	meshconfig "github.com/vechain/mesh/config"
)

// onlineOnlyEndpoints are the API endpoints that require online mode
var onlineOnlyEndpoints = []string{
	meshcommon.NetworkStatusEndpoint,
	meshcommon.AccountBalanceEndpoint,
	meshcommon.BlockEndpoint,
	meshcommon.BlockTransactionEndpoint,
	meshcommon.ConstructionMetadataEndpoint,
	meshcommon.ConstructionSubmitEndpoint,
	meshcommon.MempoolEndpoint,
	meshcommon.MempoolTransactionEndpoint,
	meshcommon.EventsBlocksEndpoint,
	meshcommon.SearchTransactionsEndpoint,
	meshcommon.CallEndpoint,
}

// OfflineModeMiddleware returns a middleware that blocks online-only endpoints in offline mode
func OfflineModeMiddleware(config *meshconfig.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If in offline mode, check if this endpoint requires online mode
			if config.Mode == meshcommon.OfflineMode {
				for _, endpoint := range onlineOnlyEndpoints {
					if strings.HasSuffix(r.URL.Path, endpoint) {
						err := meshcommon.GetError(meshcommon.ErrAPIDoesNotSupportOfflineMode)
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusInternalServerError)
						_ = json.NewEncoder(w).Encode(err)
						return
					}
				}
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}
