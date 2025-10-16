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
	vechainClient := meshthor.NewVeChainClient(cfg.NodeAPI)

	// Initialize services
	networkService := services.NewNetworkService(vechainClient, cfg)
	accountService := services.NewAccountService(vechainClient)
	constructionService := services.NewConstructionService(vechainClient, cfg)
	blockService := services.NewBlockService(vechainClient)
	mempoolService := services.NewMempoolService(vechainClient)
	eventsService := services.NewEventsService(vechainClient)
	searchService := services.NewSearchService(vechainClient)
	callService := services.NewCallService(vechainClient, cfg)

	// Create API controllers
	networkController := server.NewNetworkAPIController(networkService, asrt)
	accountController := server.NewAccountAPIController(accountService, asrt)
	constructionController := server.NewConstructionAPIController(constructionService, asrt)
	blockController := server.NewBlockAPIController(blockService, asrt)
	mempoolController := server.NewMempoolAPIController(mempoolService, asrt)
	eventsController := server.NewEventsAPIController(eventsService, asrt)
	searchController := server.NewSearchAPIController(searchService, asrt)
	callController := server.NewCallAPIController(callService, asrt)

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

	// Apply middleware stack: offline mode validation, logging, and CORS
	offlineRouter := services.OfflineModeMiddleware(cfg)(router)
	loggedRouter := server.LoggerMiddleware(offlineRouter)
	corsRouter := server.CorsMiddleware(loggedRouter)

	meshServer := &VeChainMeshServer{
		server: &http.Server{
			Addr:        fmt.Sprintf(":%d", cfg.Port),
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
	// Return standard Mesh API endpoints
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
