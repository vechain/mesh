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
	"github.com/vechain/networkhub/environments/local"
)

func main() {
	// Load configuration
	cfg, err := meshconfig.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if cfg.Mode == "online" {
		// Start Thor node
		log.Println("Starting VeChain Thor node...")
		thorEnv := local.NewEnv()

		// Configure Thor node based on config.json
		thorConfig := local.PublicNetworkConfig{
			NodeID:      "thor-node-1",
			NetworkType: cfg.GetNetwork(), // "test" or "main" from config.json
			APIAddr:     "0.0.0.0:8669",   // API address as specified
			P2PPort:     11235,            // P2P port as specified
		}

		if err := thorEnv.AttachToPublicNetworkAndStart(thorConfig); err != nil {
			log.Fatalf("Failed to start Thor node: %v", err)
		}

		defer func() {
			// Stop Thor node
			log.Println("Stopping Thor node...")
			if err := thorEnv.StopNetwork(); err != nil {
				log.Printf("Error stopping Thor node: %v", err)
			} else {
					log.Println("Thor node stopped successfully")
				}
		}()

		log.Println("Thor node started successfully")
	}

	// Create server
	meshServer, err := NewVeChainMeshServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Configure signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		if err := meshServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

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

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop server
	if err := meshServer.Stop(ctx); err != nil {
		log.Printf("Error stopping server: %v", err)
	} else {
		log.Println("Server stopped successfully")
	}
}
