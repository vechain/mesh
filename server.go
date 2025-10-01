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

	meshcommon "github.com/vechain/mesh/common"
	meshhttp "github.com/vechain/mesh/common/http"
	meshconfig "github.com/vechain/mesh/config"
	"github.com/vechain/mesh/services"
	meshthor "github.com/vechain/mesh/thor"
	meshvalidation "github.com/vechain/mesh/validation"
)

// VeChainMeshServer implements the Mesh API for VeChain
type VeChainMeshServer struct {
	router               *mux.Router
	server               *http.Server
	responseHandler      *meshhttp.ResponseHandler
	validationMiddleware func(http.Handler) http.Handler
	networkService       *services.NetworkService
	accountService       *services.AccountService
	constructionService  *services.ConstructionService
	blockService         *services.BlockService
	mempoolService       *services.MempoolService
	eventsService        *services.EventsService
	searchService        *services.SearchService
	callService          *services.CallService
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
			ctx := context.WithValue(r.Context(), meshhttp.RequestBodyKey, body)
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
	searchService := services.NewSearchService(vechainClient)
	callService := services.NewCallService(vechainClient, cfg)

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
		searchService:        searchService,
		callService:          callService,
	}

	meshServer.setupRoutes()
	cfg.PrintConfig()

	return meshServer, nil
}

// setupRoutes configures the API routes
func (v *VeChainMeshServer) setupRoutes() {
	v.router.HandleFunc(meshcommon.HealthEndpoint, v.healthCheck).Methods("GET")

	// Create a subrouter for API endpoints that need validation
	apiRouter := v.router.PathPrefix("/").Subrouter()
	apiRouter.Use(v.validationMiddleware)

	// Network API endpoints
	apiRouter.HandleFunc(meshcommon.NetworkListEndpoint, v.networkService.NetworkList).Methods("POST")
	apiRouter.HandleFunc(meshcommon.NetworkOptionsEndpoint, v.networkService.NetworkOptions).Methods("POST")
	apiRouter.HandleFunc(meshcommon.NetworkStatusEndpoint, v.networkService.NetworkStatus).Methods("POST")

	// Account API endpoints
	apiRouter.HandleFunc(meshcommon.AccountBalanceEndpoint, v.accountService.AccountBalance).Methods("POST")

	// Construction API endpoints
	apiRouter.HandleFunc(meshcommon.ConstructionDeriveEndpoint, v.constructionService.ConstructionDerive).Methods("POST")
	apiRouter.HandleFunc(meshcommon.ConstructionPreprocessEndpoint, v.constructionService.ConstructionPreprocess).Methods("POST")
	apiRouter.HandleFunc(meshcommon.ConstructionMetadataEndpoint, v.constructionService.ConstructionMetadata).Methods("POST")
	apiRouter.HandleFunc(meshcommon.ConstructionPayloadsEndpoint, v.constructionService.ConstructionPayloads).Methods("POST")
	apiRouter.HandleFunc(meshcommon.ConstructionParseEndpoint, v.constructionService.ConstructionParse).Methods("POST")
	apiRouter.HandleFunc(meshcommon.ConstructionCombineEndpoint, v.constructionService.ConstructionCombine).Methods("POST")
	apiRouter.HandleFunc(meshcommon.ConstructionHashEndpoint, v.constructionService.ConstructionHash).Methods("POST")
	apiRouter.HandleFunc(meshcommon.ConstructionSubmitEndpoint, v.constructionService.ConstructionSubmit).Methods("POST")

	// Block API endpoints
	apiRouter.HandleFunc(meshcommon.BlockEndpoint, v.blockService.Block).Methods("POST")
	apiRouter.HandleFunc(meshcommon.BlockTransactionEndpoint, v.blockService.BlockTransaction).Methods("POST")

	// Mempool API endpoints
	apiRouter.HandleFunc(meshcommon.MempoolEndpoint, v.mempoolService.Mempool).Methods("POST")
	apiRouter.HandleFunc(meshcommon.MempoolTransactionEndpoint, v.mempoolService.MempoolTransaction).Methods("POST")

	// Events API endpoints
	apiRouter.HandleFunc(meshcommon.EventsBlocksEndpoint, v.eventsService.EventsBlocks).Methods("POST")

	// Search API endpoints
	apiRouter.HandleFunc(meshcommon.SearchTransactionsEndpoint, v.searchService.SearchTransactions).Methods("POST")

	// Call API endpoints
	apiRouter.HandleFunc(meshcommon.CallEndpoint, v.callService.Call).Methods("POST")
}

// healthCheck endpoint to verify server status
func (v *VeChainMeshServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "VeChain Mesh API",
	}

	v.responseHandler.WriteJSONResponse(w, response)
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

// GetEndpoints returns a list of all registered endpoints
func (v *VeChainMeshServer) GetEndpoints() ([]string, error) {
	var endpoints []string
	err := v.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err != nil {
			return err
		}

		methods, err := route.GetMethods()
		if err != nil {
			return nil
		}

		for _, method := range methods {
			endpoints = append(endpoints, fmt.Sprintf("%s %s", method, pathTemplate))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return endpoints, nil
}
