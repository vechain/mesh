package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	meshconfig "github.com/vechain/mesh/config"
	"github.com/vechain/mesh/thor"
)

func main() {
	cfg := loadConfiguration()
	thorServer := startThorNode(cfg)

	// Ensure Thor node is stopped on exit
	if thorServer != nil {
		defer stopThorNode(thorServer)
	}

	meshServer := createMeshServer(cfg)
	startServer(meshServer)
	printEndpoints()
	waitForShutdown(meshServer)
}

// loadConfiguration loads and validates the application configuration
func loadConfiguration() *meshconfig.Config {
	cfg, err := meshconfig.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	return cfg
}

// startThorNode starts the Thor node if in online mode
func startThorNode(cfg *meshconfig.Config) *thor.Server {
	if cfg.Mode != "online" {
		return nil
	}

	log.Println("Starting VeChain Thor node...")

	thorConfig := createThorConfig(cfg)
	thorServer := thor.NewServer(thorConfig)

	if err := startThorWithConfig(thorServer, cfg); err != nil {
		log.Fatalf("Failed to start Thor node: %v", err)
	}

	log.Println("Thor node started successfully")
	return thorServer
}

// stopThorNode stops the Thor node gracefully
func stopThorNode(thorServer *thor.Server) {
	log.Println("Stopping Thor node...")
	if err := thorServer.Stop(); err != nil {
		log.Printf("Error stopping Thor node: %v", err)
	} else {
		log.Println("Thor node stopped successfully")
	}
}

// createThorConfig creates Thor configuration based on network type
func createThorConfig(cfg *meshconfig.Config) thor.Config {
	thorConfig := thor.Config{
		NodeID:      "thor-node-1",
		NetworkType: cfg.GetNetwork(),
		APIAddr:     "0.0.0.0:8669",
		P2PPort:     11235,
	}

	if cfg.GetNetwork() == "solo" {
		thorConfig.OnDemand = true
		thorConfig.Persist = true
		thorConfig.APICORS = "*"
	}

	return thorConfig
}

// startThorWithConfig starts Thor with the appropriate method based on network type
func startThorWithConfig(thorServer *thor.Server, cfg *meshconfig.Config) error {
	if cfg.GetNetwork() == "solo" {
		log.Println("Starting Thor node in solo mode...")
		return thorServer.StartSoloNode()
	}

	log.Printf("Starting Thor node connected to %s network...", cfg.GetNetwork())
	return thorServer.AttachToPublicNetworkAndStart()
}

// createMeshServer creates and configures the Mesh API server
func createMeshServer(cfg *meshconfig.Config) *VeChainMeshServer {
	meshServer, err := NewVeChainMeshServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	return meshServer
}

// startServer starts the Mesh API server in a goroutine
func startServer(meshServer *VeChainMeshServer) {
	go func() {
		if err := meshServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()
}

// printEndpoints prints available API endpoints
func printEndpoints() {
	log.Println("Available endpoints:")
	log.Println("  GET  /health")
	log.Println("  POST /network/list")
	log.Println("  POST /network/status")
	log.Println("  POST /account/balance")
	log.Println("  POST /construction/derive")
	log.Println("  POST /construction/preprocess")
	log.Println("  POST /construction/metadata")
	log.Println("  POST /construction/payloads")
	log.Println("  POST /construction/parse")
	log.Println("  POST /construction/combine")
	log.Println("  POST /construction/hash")
	log.Println("  POST /construction/submit")
}

// waitForShutdown handles graceful shutdown of the application
func waitForShutdown(meshServer *VeChainMeshServer) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutdown signal received...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := meshServer.Stop(ctx); err != nil {
		log.Printf("Error stopping server: %v", err)
	} else {
		log.Println("Server stopped successfully")
	}
}
