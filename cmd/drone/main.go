package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spawn-mcp/coordinator/pkg/drone"
)

func main() {
	log.Println("Starting Drone MCP Server...")

	// Create researcher drone
	researcherDrone, err := drone.NewResearcherDrone()
	if err != nil {
		log.Fatalf("Failed to create researcher drone: %v", err)
	}
	defer func() {
		if err := researcherDrone.Close(); err != nil {
			log.Printf("Error closing drone: %v", err)
		}
	}()

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- researcherDrone.StartHTTPServer(":8080")
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down gracefully...", sig)
	case err := <-serverErr:
		if err != nil {
			log.Printf("Server error: %v", err)
		}
	}

	log.Println("Drone MCP Server stopped")
}
