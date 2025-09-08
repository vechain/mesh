package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/vechain/mesh/services"
)

// VeChainMeshServer implements the Mesh API for VeChain
type VeChainMeshServer struct {
	router              *mux.Router
	server              *http.Server
	networkService      *services.NetworkService
	accountService      *services.AccountService
	constructionService *services.ConstructionService
}

// NewVeChainMeshServer creates a new server instance
func NewVeChainMeshServer(port string, vechainRPCURL string) *VeChainMeshServer {
	router := mux.NewRouter()

	// Initialize VeChain client
	vechainClient := services.NewVeChainClient(vechainRPCURL)

	// Initialize services with VeChain client
	networkService := services.NewNetworkService(vechainClient)
	accountService := services.NewAccountService(vechainClient)
	constructionService := services.NewConstructionService()

	meshServer := &VeChainMeshServer{
		router: router,
		server: &http.Server{
			Addr:    ":" + port,
			Handler: router,
		},
		networkService:      networkService,
		accountService:      accountService,
		constructionService: constructionService,
	}

	meshServer.setupRoutes()
	return meshServer
}

// setupRoutes configures the API routes
func (v *VeChainMeshServer) setupRoutes() {
	// Health check endpoint
	v.router.HandleFunc("/health", v.healthCheck).Methods("GET")

	// Network API endpoints
	v.router.HandleFunc("/network/list", v.networkService.NetworkList).Methods("POST")
	v.router.HandleFunc("/network/status", v.networkService.NetworkStatus).Methods("POST")

	// Account API endpoints
	v.router.HandleFunc("/account/balance", v.accountService.AccountBalance).Methods("POST")

	// Construction API endpoints
	v.router.HandleFunc("/construction/derive", v.constructionService.ConstructionDerive).Methods("POST")
	v.router.HandleFunc("/construction/preprocess", v.constructionService.ConstructionPreprocess).Methods("POST")
	v.router.HandleFunc("/construction/metadata", v.constructionService.ConstructionMetadata).Methods("POST")
	v.router.HandleFunc("/construction/payloads", v.constructionService.ConstructionPayloads).Methods("POST")
	v.router.HandleFunc("/construction/parse", v.constructionService.ConstructionParse).Methods("POST")
	v.router.HandleFunc("/construction/combine", v.constructionService.ConstructionCombine).Methods("POST")
	v.router.HandleFunc("/construction/hash", v.constructionService.ConstructionHash).Methods("POST")
	v.router.HandleFunc("/construction/submit", v.constructionService.ConstructionSubmit).Methods("POST")
}

// healthCheck endpoint to verify server status
func (v *VeChainMeshServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "VeChain Mesh API",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
