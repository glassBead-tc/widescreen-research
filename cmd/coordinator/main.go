package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spawn-mcp/coordinator/pkg/coordinator"
	"github.com/spawn-mcp/coordinator/pkg/gcp"
)

func main() {
	log.Println("Starting Spawn MCP Coordinator...")

	// Get configuration from environment variables
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT environment variable is required")
	}

	region := os.Getenv("GOOGLE_CLOUD_REGION")
	if region == "" {
		region = "us-central1" // Default region
	}

	// Create context
	ctx := context.Background()

	// Initialize GCP client
	gcpClient, err := gcp.NewClient(ctx, projectID, region)
	if err != nil {
		log.Fatalf("Failed to create GCP client: %v", err)
	}
	defer func() {
		if err := gcpClient.Close(); err != nil {
			log.Printf("Error closing GCP client: %v", err)
		}
	}()

	log.Printf("Initialized GCP client for project %s in region %s", projectID, region)

	// Create coordinator server
	server := coordinator.NewServer(gcpClient)

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Serve()
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

	log.Println("Coordinator MCP Server stopped")
}
