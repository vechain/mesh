package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/server"

	meshcommon "github.com/vechain/mesh/common"
	meshconfig "github.com/vechain/mesh/config"
	"github.com/vechain/mesh/services"
	meshthor "github.com/vechain/mesh/thor"
)

// VeChainMeshServer implements the Mesh API for VeChain
type VeChainMeshServer struct {
	server   *http.Server
	asserter *asserter.Asserter
	config   *meshconfig.Config
}

// NewVeChainMeshServer creates a new server instance
func NewVeChainMeshServer(cfg *meshconfig.Config, asrt *asserter.Asserter) (*VeChainMeshServer, error) {
	vechainClient := meshthor.NewVeChainClient(cfg.GetNodeAPI())

	// Initialize base services
	networkService := services.NewNetworkService(vechainClient, cfg)
	accountService := services.NewAccountService(vechainClient)
	constructionService := services.NewConstructionService(vechainClient, cfg)
	blockService := services.NewBlockService(vechainClient)
	mempoolService := services.NewMempoolService(vechainClient)
	eventsService := services.NewEventsService(vechainClient)
	searchService := services.NewSearchService(vechainClient)
	callService := services.NewCallService(vechainClient, cfg)

	// Wrap services with offline validators
	// These wrappers intercept calls and return appropriate errors when in offline mode
	offlineNetworkService := services.NewOfflineNetworkService(networkService, cfg)
	offlineAccountService := services.NewOfflineAccountService(accountService, cfg)
	offlineConstructionService := services.NewOfflineConstructionService(constructionService, cfg)
	offlineBlockService := services.NewOfflineBlockService(blockService, cfg)
	offlineMempoolService := services.NewOfflineMempoolService(mempoolService, cfg)
	offlineEventsService := services.NewOfflineEventsService(eventsService, cfg)
	offlineSearchService := services.NewOfflineSearchService(searchService, cfg)
	offlineCallService := services.NewOfflineCallService(callService, cfg)

	// Create API controllers with wrapped services
	networkController := server.NewNetworkAPIController(offlineNetworkService, asrt)
	accountController := server.NewAccountAPIController(offlineAccountService, asrt)
	constructionController := server.NewConstructionAPIController(offlineConstructionService, asrt)
	blockController := server.NewBlockAPIController(offlineBlockService, asrt)
	mempoolController := server.NewMempoolAPIController(offlineMempoolService, asrt)
	eventsController := server.NewEventsAPIController(offlineEventsService, asrt)
	searchController := server.NewSearchAPIController(offlineSearchService, asrt)
	callController := server.NewCallAPIController(offlineCallService, asrt)

	// Create router with all controllers
	router := server.NewRouter(
		networkController,
		accountController,
		blockController,
		constructionController,
		mempoolController,
		eventsController,
		searchController,
		callController,
	)

	// Apply logging and CORS middleware
	loggedRouter := server.LoggerMiddleware(router)
	corsRouter := server.CorsMiddleware(loggedRouter)

	meshServer := &VeChainMeshServer{
		server: &http.Server{
			Addr:        fmt.Sprintf(":%d", cfg.GetPort()),
			Handler:     corsRouter,
			ReadTimeout: 30 * time.Second,
		},
		asserter: asrt,
		config:   cfg,
	}

	cfg.PrintConfig()

	return meshServer, nil
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
	// Return standard Rosetta API endpoints
	endpoints := []string{
		fmt.Sprintf("POST %s", meshcommon.NetworkListEndpoint),
		fmt.Sprintf("POST %s", meshcommon.NetworkOptionsEndpoint),
		fmt.Sprintf("POST %s", meshcommon.NetworkStatusEndpoint),
		fmt.Sprintf("POST %s", meshcommon.AccountBalanceEndpoint),
		fmt.Sprintf("POST %s", meshcommon.BlockEndpoint),
		fmt.Sprintf("POST %s", meshcommon.BlockTransactionEndpoint),
		fmt.Sprintf("POST %s", meshcommon.ConstructionDeriveEndpoint),
		fmt.Sprintf("POST %s", meshcommon.ConstructionPreprocessEndpoint),
		fmt.Sprintf("POST %s", meshcommon.ConstructionMetadataEndpoint),
		fmt.Sprintf("POST %s", meshcommon.ConstructionPayloadsEndpoint),
		fmt.Sprintf("POST %s", meshcommon.ConstructionParseEndpoint),
		fmt.Sprintf("POST %s", meshcommon.ConstructionCombineEndpoint),
		fmt.Sprintf("POST %s", meshcommon.ConstructionHashEndpoint),
		fmt.Sprintf("POST %s", meshcommon.ConstructionSubmitEndpoint),
		fmt.Sprintf("POST %s", meshcommon.MempoolEndpoint),
		fmt.Sprintf("POST %s", meshcommon.MempoolTransactionEndpoint),
		fmt.Sprintf("POST %s", meshcommon.EventsBlocksEndpoint),
		fmt.Sprintf("POST %s", meshcommon.SearchTransactionsEndpoint),
		fmt.Sprintf("POST %s", meshcommon.CallEndpoint),
	}

	return endpoints, nil
}
