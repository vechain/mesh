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
	// Load configuration
	cfg, err := meshconfig.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if cfg.Mode == "online" {
		// Start Thor node
		log.Println("Starting VeChain Thor node...")

		// Configure Thor node based on config.json
		thorConfig := thor.Config{
			NodeID:      "thor-node-1",
			NetworkType: cfg.GetNetwork(), // "test", "main", or "solo" from config.json
			APIAddr:     "0.0.0.0:8669",
			P2PPort:     11235,
		}

		if cfg.GetNetwork() == "solo" {
			thorConfig.OnDemand = true
			thorConfig.Persist = true
			thorConfig.APICORS = "*"
		}

		thorServer := thor.NewServer(thorConfig)

		var err error
		if cfg.GetNetwork() == "solo" {
			// Start in solo mode
			log.Println("Starting Thor node in solo mode...")
			err = thorServer.StartSoloNode()
		} else {
			// Start connected to public network (test/main)
			log.Printf("Starting Thor node connected to %s network...", thorConfig.NetworkType)
			err = thorServer.AttachToPublicNetworkAndStart()
		}

		if err != nil {
			log.Fatalf("Failed to start Thor node: %v", err)
		}

		defer func() {
			// Stop Thor node
			log.Println("Stopping Thor node...")
			if err := thorServer.Stop(); err != nil {
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
