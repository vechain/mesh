package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// VeChainMeshServer implements the Mesh API for VeChain
type VeChainMeshServer struct {
	router *mux.Router
	server *http.Server
}

// NewVeChainMeshServer creates a new server instance
func NewVeChainMeshServer(port string) *VeChainMeshServer {
	router := mux.NewRouter()

	meshServer := &VeChainMeshServer{
		router: router,
		server: &http.Server{
			Addr:    ":" + port,
			Handler: router,
		},
	}

	meshServer.setupRoutes()
	return meshServer
}

// setupRoutes configures the API routes
func (v *VeChainMeshServer) setupRoutes() {
	// Basic Mesh API endpoints
	v.router.HandleFunc("/health", v.healthCheck).Methods("GET")
	v.router.HandleFunc("/network/list", v.networkList).Methods("POST")
	v.router.HandleFunc("/network/status", v.networkStatus).Methods("POST")
	v.router.HandleFunc("/account/balance", v.accountBalance).Methods("POST")
	v.router.HandleFunc("/construction/derive", v.constructionDerive).Methods("POST")
	v.router.HandleFunc("/construction/preprocess", v.constructionPreprocess).Methods("POST")
	v.router.HandleFunc("/construction/metadata", v.constructionMetadata).Methods("POST")
	v.router.HandleFunc("/construction/payloads", v.constructionPayloads).Methods("POST")
	v.router.HandleFunc("/construction/parse", v.constructionParse).Methods("POST")
	v.router.HandleFunc("/construction/combine", v.constructionCombine).Methods("POST")
	v.router.HandleFunc("/construction/hash", v.constructionHash).Methods("POST")
	v.router.HandleFunc("/construction/submit", v.constructionSubmit).Methods("POST")
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
