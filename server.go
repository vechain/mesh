package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	meshconfig "github.com/vechain/mesh/config"
	"github.com/vechain/mesh/services"
	meshthor "github.com/vechain/mesh/thor"
	meshutils "github.com/vechain/mesh/utils"
	meshvalidation "github.com/vechain/mesh/validation"
)

// VeChainMeshServer implements the Mesh API for VeChain
type VeChainMeshServer struct {
	router               *mux.Router
	server               *http.Server
	validationMiddleware func(http.Handler) http.Handler
	networkService       *services.NetworkService
	accountService       *services.AccountService
	constructionService  *services.ConstructionService
	blockService         *services.BlockService
	mempoolService       *services.MempoolService
	eventsService        *services.EventsService
}

// NewVeChainMeshServer creates a new server instance
func NewVeChainMeshServer(cfg *meshconfig.Config) (*VeChainMeshServer, error) {
	router := mux.NewRouter()

	vechainClient := meshthor.NewVeChainClient(cfg.GetNodeAPI())

	validationMiddleware := meshvalidation.NewValidationMiddleware(cfg.GetNetworkIdentifier(), cfg.GetRunMode())

	// Create validation middleware function
	validationMiddlewareFunc := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}

			endpoint := r.URL.Path

			if !validationMiddleware.ValidateEndpoint(w, r, body, endpoint) {
				return
			}

			// Store the body in context for services to use
			ctx := context.WithValue(r.Context(), meshutils.RequestBodyKey, body)
			r = r.WithContext(ctx)

			// Restore the body for the next handler (in case some service still needs it)
			r.Body = io.NopCloser(strings.NewReader(string(body)))

			next.ServeHTTP(w, r)
		})
	}

	// Initialize services
	networkService := services.NewNetworkService(vechainClient, cfg)
	accountService := services.NewAccountService(vechainClient)

	constructionService := services.NewConstructionService(vechainClient, cfg)
	blockService := services.NewBlockService(vechainClient)
	mempoolService := services.NewMempoolService(vechainClient)
	eventsService := services.NewEventsService(vechainClient)

	meshServer := &VeChainMeshServer{
		router: router,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.GetPort()),
			Handler: router,
		},
		validationMiddleware: validationMiddlewareFunc,
		networkService:       networkService,
		accountService:       accountService,
		constructionService:  constructionService,
		blockService:         blockService,
		mempoolService:       mempoolService,
		eventsService:        eventsService,
	}

	meshServer.setupRoutes()
	cfg.PrintConfig()

	return meshServer, nil
}

// setupRoutes configures the API routes
func (v *VeChainMeshServer) setupRoutes() {
	v.router.HandleFunc("/health", v.healthCheck).Methods("GET")

	// Create a subrouter for API endpoints that need validation
	apiRouter := v.router.PathPrefix("/").Subrouter()
	apiRouter.Use(v.validationMiddleware)

	// Network API endpoints
	apiRouter.HandleFunc("/network/list", v.networkService.NetworkList).Methods("POST")
	apiRouter.HandleFunc("/network/options", v.networkService.NetworkOptions).Methods("POST")
	apiRouter.HandleFunc("/network/status", v.networkService.NetworkStatus).Methods("POST")

	// Account API endpoints
	apiRouter.HandleFunc("/account/balance", v.accountService.AccountBalance).Methods("POST")

	// Construction API endpoints
	apiRouter.HandleFunc("/construction/derive", v.constructionService.ConstructionDerive).Methods("POST")
	apiRouter.HandleFunc("/construction/preprocess", v.constructionService.ConstructionPreprocess).Methods("POST")
	apiRouter.HandleFunc("/construction/metadata", v.constructionService.ConstructionMetadata).Methods("POST")
	apiRouter.HandleFunc("/construction/payloads", v.constructionService.ConstructionPayloads).Methods("POST")
	apiRouter.HandleFunc("/construction/parse", v.constructionService.ConstructionParse).Methods("POST")
	apiRouter.HandleFunc("/construction/combine", v.constructionService.ConstructionCombine).Methods("POST")
	apiRouter.HandleFunc("/construction/hash", v.constructionService.ConstructionHash).Methods("POST")
	apiRouter.HandleFunc("/construction/submit", v.constructionService.ConstructionSubmit).Methods("POST")

	// Block API endpoints
	apiRouter.HandleFunc("/block", v.blockService.Block).Methods("POST")
	apiRouter.HandleFunc("/block/transaction", v.blockService.BlockTransaction).Methods("POST")

	// Mempool API endpoints
	apiRouter.HandleFunc("/mempool", v.mempoolService.Mempool).Methods("POST")
	apiRouter.HandleFunc("/mempool/transaction", v.mempoolService.MempoolTransaction).Methods("POST")

	// Events API endpoints
	apiRouter.HandleFunc("/events/blocks", v.eventsService.EventsBlocks).Methods("POST")
}

// healthCheck endpoint to verify server status
func (v *VeChainMeshServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "VeChain Mesh API",
	}

	meshutils.WriteJSONResponse(w, response)
}

// Start starts the server
func (v *VeChainMeshServer) Start() error {
	log.Printf("Starting VeChain Mesh API server on port %s", v.server.Addr)
	return v.server.ListenAndServe()
}

// Stop stops the server
func (v *VeChainMeshServer) Stop(ctx context.Context) error {
	log.Println("Stopping VeChain Mesh API server...")
	return v.server.Shutdown(ctx)
}
