package thor

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	meshcommon "github.com/vechain/mesh/common"
)

// Config represents the configuration for a Thor node
type Config struct {
	NodeID      string
	NetworkType string // "test", "main", or "solo"
	APIAddr     string
	P2PPort     int
	// Solo mode specific options
	OnDemand bool   // Create blocks on demand when there are pending transactions
	Persist  bool   // Persist blockchain data to disk instead of memory
	APICORS  string // CORS settings for API
}

// Server manages Thor node processes
type Server struct {
	config   Config
	process  *exec.Cmd
	ctx      context.Context
	cancel   context.CancelFunc
	thorPath string
}

// NewServer creates a new Thor server instance
func NewServer(config Config) *Server {
	// Get the directory where the executable is located
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	execDir := filepath.Dir(execPath)

	thorPath := filepath.Join(execDir, "thor")

	// Check if Thor binaries exist
	if _, err := os.Stat(thorPath); os.IsNotExist(err) {
		log.Fatalf("Thor binary not found at: %s", thorPath)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
		thorPath: thorPath,
	}
}

// AttachToPublicNetworkAndStart starts a Thor node and connects to the public network
func (ts *Server) AttachToPublicNetworkAndStart() error {
	log.Printf("Starting Thor node with config: %+v", ts.config)

	// Build command arguments
	args := []string{
		"--network", ts.config.NetworkType,
		"--api-addr", ts.config.APIAddr,
		"--p2p-port", strconv.Itoa(ts.config.P2PPort),
		"--data-dir", meshcommon.DataDirectory,
	}

	// Create the command
	ts.process = exec.CommandContext(ts.ctx, ts.thorPath, args...)

	// Set up process attributes
	ts.process.Stdout = os.Stdout
	ts.process.Stderr = os.Stderr
	ts.process.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Start the process
	if err := ts.process.Start(); err != nil {
		return fmt.Errorf("failed to start Thor process: %v", err)
	}

	log.Printf("Thor node started with PID: %d", ts.process.Process.Pid)

	// Wait a bit to ensure the node is starting up
	time.Sleep(2 * time.Second)

	// Check if the process is still running
	if ts.process.ProcessState != nil && ts.process.ProcessState.Exited() {
		return fmt.Errorf("thor process exited unexpectedly")
	}

	return nil
}

// StartSoloNode starts a Thor node in solo mode
func (ts *Server) StartSoloNode() error {
	log.Printf("Starting Thor node in solo mode with config: %+v", ts.config)

	// Build command arguments for solo mode
	args := []string{
		"solo", // Solo mode command
		"--api-addr", ts.config.APIAddr,
		"--data-dir", meshcommon.DataDirectory,
		"--api-enable-txpool", // Enable txpool API
	}

	// Add optional solo mode arguments
	if ts.config.OnDemand {
		args = append(args, "--on-demand")
	}

	if ts.config.Persist {
		args = append(args, "--persist")
	}

	if ts.config.APICORS != "" {
		args = append(args, "--api-cors", ts.config.APICORS)
	} else {
		args = append(args, "--api-cors", "*") // Default to allow all CORS
	}

	// Create the command
	ts.process = exec.CommandContext(ts.ctx, ts.thorPath, args...)

	// Set up process attributes
	ts.process.Stdout = os.Stdout
	ts.process.Stderr = os.Stderr
	ts.process.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Start the process
	if err := ts.process.Start(); err != nil {
		return fmt.Errorf("failed to start Thor solo process: %v", err)
	}

	log.Printf("Thor solo node started with PID: %d", ts.process.Process.Pid)

	// Wait a bit to ensure the node is starting up
	time.Sleep(2 * time.Second)

	// Check if the process is still running
	if ts.process.ProcessState != nil && ts.process.ProcessState.Exited() {
		return fmt.Errorf("thor solo process exited unexpectedly")
	}

	return nil
}

// Stop stops the Thor node
func (ts *Server) Stop() error {
	if ts.process == nil {
		log.Println("No Thor process to stop")
		return nil
	}

	log.Println("Stopping Thor node...")

	// Cancel the context to signal the process to stop
	ts.cancel()

	// Wait for the process to finish with a timeout
	done := make(chan error, 1)
	go func() {
		done <- ts.process.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			log.Printf("Thor process finished with error: %v", err)
		} else {
			log.Println("Thor process finished successfully")
		}
		return err
	case <-time.After(10 * time.Second):
		log.Println("Thor process did not stop gracefully, forcing termination...")

		// Force kill the process group
		if ts.process.Process != nil {
			if err := syscall.Kill(-ts.process.Process.Pid, syscall.SIGKILL); err != nil {
				log.Printf("Failed to kill Thor process: %v", err)
				return err
			}
		}

		// Wait a bit more for cleanup
		select {
		case <-done:
			log.Println("Thor process terminated after force kill")
		case <-time.After(5 * time.Second):
			log.Println("Thor process still running after force kill")
		}

		return nil
	}
}
