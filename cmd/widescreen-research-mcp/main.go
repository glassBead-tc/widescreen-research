package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create and start the MCP server
	srv, err := server.NewWidescreenResearchServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start the server
	errChan := make(chan error, 1)
	go func() {
		if err := srv.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	// Wait for signal or error
	select {
	case sig := <-sigChan:
		fmt.Printf("Received signal %v, shutting down...\n", sig)
	case err := <-errChan:
		log.Fatalf("Server error: %v", err)
	}

	// Graceful shutdown
	srv.Shutdown()
}