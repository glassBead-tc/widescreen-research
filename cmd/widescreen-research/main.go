// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research/server"
)

func main() {
	// Parse flags
	httpMode := flag.Bool("http", false, "Run in HTTP mode instead of stdio")
	port := flag.String("port", "8080", "HTTP port (only used with -http)")
	orchestratorURL := flag.String("orchestrator", "", "URL of widescreen-research-mcp orchestrator (stdio:// or http://)")
	flag.Parse()

	// Check environment variable overrides
	if os.Getenv("MCP_TRANSPORT") == "http" {
		*httpMode = true
	}
	if envPort := os.Getenv("PORT"); envPort != "" {
		*port = envPort
	}
	if envOrchURL := os.Getenv("ORCHESTRATOR_URL"); envOrchURL != "" {
		*orchestratorURL = envOrchURL
	}

	// Default to stdio if not specified
	if *orchestratorURL == "" {
		*orchestratorURL = "stdio://widescreen-research-mcp"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create and start the bidirectional MCP host
	srv, err := server.NewWidescreenResearchHost(*orchestratorURL)
	if err != nil {
		log.Fatalf("Failed to create host server: %v", err)
	}

	// Start the server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if *httpMode {
			addr := ":" + *port
			errChan <- srv.RunHTTP(ctx, addr)
		} else {
			errChan <- srv.Run(ctx)
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
