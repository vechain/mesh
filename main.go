package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Create server
	meshServer, err := NewVeChainMeshServer()
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
