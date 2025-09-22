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
)

// ThorConfig represents the configuration for a Thor node
type ThorConfig struct {
	NodeID      string
	NetworkType string
	APIAddr     string
	P2PPort     int
}

// ThorService manages Thor node processes
type ThorServer struct {
	config   ThorConfig
	process  *exec.Cmd
	ctx      context.Context
	cancel   context.CancelFunc
	thorPath string
}

// NewThorService creates a new Thor service instance
func NewThorServer(config ThorConfig) *ThorServer {
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

	return &ThorServer{
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
		thorPath: thorPath,
	}
}

// AttachToPublicNetworkAndStart starts a Thor node and connects to the public network
func (ts *ThorServer) AttachToPublicNetworkAndStart() error {
	log.Printf("Starting Thor node with config: %+v", ts.config)

	// Build command arguments
	args := []string{
		"--network", ts.config.NetworkType,
		"--api-addr", ts.config.APIAddr,
		"--p2p-port", strconv.Itoa(ts.config.P2PPort),
		"--data-dir", "/tmp/thor_data",
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
		return fmt.Errorf("Thor process exited unexpectedly")
	}

	return nil
}

// StopNetwork stops the Thor node
func (ts *ThorServer) StopNetwork() error {
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

// IsRunning checks if the Thor node is currently running
func (ts *ThorServer) IsRunning() bool {
	if ts.process == nil {
		return false
	}

	if ts.process.ProcessState != nil {
		return !ts.process.ProcessState.Exited()
	}

	// Check if the process is still alive by sending a signal 0
	if ts.process.Process != nil {
		err := ts.process.Process.Signal(syscall.Signal(0))
		return err == nil
	}

	return false
}

// GetProcessID returns the process ID of the Thor node
func (ts *ThorServer) GetProcessID() int {
	if ts.process != nil && ts.process.Process != nil {
		return ts.process.Process.Pid
	}
	return -1
}
