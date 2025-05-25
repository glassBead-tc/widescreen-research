package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spawn-mcp/coordinator/pkg/coordinator"
	"github.com/spawn-mcp/coordinator/pkg/gcp"
	"github.com/spawn-mcp/coordinator/pkg/mcp"
)

func main() {
	log.Println("Starting Spawn MCP Coordinator...")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize GCP client (for now, we'll use a mock since the GCP client has issues)
	// In production, you would initialize this properly with your GCP project
	var gcpClient *gcp.Client

	// For now, we'll skip GCP initialization to focus on MCP functionality
	// gcpClient, err = gcp.NewClient(ctx, "your-project-id", "us-central1")
	// if err != nil {
	//     log.Fatalf("Failed to create GCP client: %v", err)
	// }
	// defer gcpClient.Close()

	// Create coordinator server
	coordinatorServer := coordinator.NewServer(gcpClient)

	// Create MCP server that wraps the coordinator
	mcpServer := mcp.NewMCPServer(coordinatorServer)

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start MCP server in a goroutine
	go func() {
		log.Println("Starting MCP server on stdio...")
		if err := mcpServer.Start(ctx); err != nil {
			log.Printf("MCP server error: %v", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	select {
	case <-sigChan:
		log.Println("Received shutdown signal")
	case <-ctx.Done():
		log.Println("Context cancelled")
	}

	// Graceful shutdown
	log.Println("Shutting down...")
	if err := mcpServer.Close(); err != nil {
		log.Printf("Error closing MCP server: %v", err)
	}

	log.Println("Shutdown complete")
}
