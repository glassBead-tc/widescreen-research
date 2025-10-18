// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research-mcp/schemas"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// OrchestratorClient is an MCP client that connects to widescreen-research-mcp
type OrchestratorClient struct {
	client          *mcp.Client
	session         *mcp.ClientSession
	orchestratorURL string
}

// NewOrchestratorClient creates a new orchestrator client
func NewOrchestratorClient(orchestratorURL string) (*OrchestratorClient, error) {
	return &OrchestratorClient{
		orchestratorURL: orchestratorURL,
	}, nil
}

// Connect establishes connection to the orchestrator MCP server
func (c *OrchestratorClient) Connect(ctx context.Context) error {
	// Determine transport type from URL
	var transport mcp.Transport
	var err error

	if strings.HasPrefix(c.orchestratorURL, "stdio://") {
		// Stdio transport - spawn the orchestrator process
		cmdPath := strings.TrimPrefix(c.orchestratorURL, "stdio://")
		log.Printf("Connecting to orchestrator via stdio: %s", cmdPath)
		transport = &mcp.StdioTransport{}
	} else if strings.HasPrefix(c.orchestratorURL, "http://") || strings.HasPrefix(c.orchestratorURL, "https://") {
		// HTTP transport - for now, not fully supported in client mode
		// The MCP SDK primarily supports stdio for client connections
		return fmt.Errorf("HTTP transport not yet implemented for client connections (use stdio://)")
	} else {
		return fmt.Errorf("unsupported orchestrator URL scheme: %s", c.orchestratorURL)
	}

	// Create MCP client
	c.client = mcp.NewClient(&mcp.Implementation{
		Name:    "widescreen-research-host",
		Version: "1.0.0",
	}, nil)

	// Connect to the server - returns ClientSession
	session, err := c.client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to orchestrator: %w", err)
	}

	c.session = session
	log.Println("Successfully connected to widescreen-research-mcp orchestrator")
	return nil
}

// StartGCPOrchestration calls the start-gcp-orchestration tool on the orchestrator server
// This delegates GCP resource provisioning and drone management to widescreen-research-mcp
// Returns the collected drone results (NOT a report - report generation happens in the host)
func (c *OrchestratorClient) StartGCPOrchestration(ctx context.Context, config *schemas.ResearchConfig) ([]schemas.DroneResult, error) {
	if c.session == nil {
		return nil, fmt.Errorf("client not connected")
	}

	// Build parameters as a map (will be JSON marshalled)
	params := map[string]interface{}{
		"session_id":       config.SessionID,
		"topic":            config.Topic,
		"researcher_count": config.ResearcherCount,
		"research_depth":   config.ResearchDepth,
		"timeout_minutes":  config.TimeoutMinutes,
		"priority_level":   config.PriorityLevel,
		"output_format":    config.OutputFormat,
	}

	log.Printf("Calling start-gcp-orchestration on orchestrator for session %s", config.SessionID)

	// Call the tool via ClientSession
	result, err := c.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "start-gcp-orchestration",
		Arguments: params,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to call start-gcp-orchestration: %w", err)
	}

	// Check for error response
	if result.IsError {
		errorMsg := "unknown error"
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
				errorMsg = textContent.Text
			}
		}
		return nil, fmt.Errorf("orchestration error: %s", errorMsg)
	}

	// Extract results from response
	if len(result.Content) == 0 {
		return nil, fmt.Errorf("empty response from orchestrator")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		return nil, fmt.Errorf("unexpected response type from orchestrator")
	}

	// Parse the response - expecting a JSON object with results field
	var response struct {
		SessionID string                `json:"session_id"`
		Status    string                `json:"status"`
		Results   []schemas.DroneResult `json:"results"`
	}

	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		return nil, fmt.Errorf("failed to parse orchestrator response: %w", err)
	}

	if response.Status != "completed" {
		return nil, fmt.Errorf("orchestration failed with status: %s", response.Status)
	}

	log.Printf("Received %d drone results from orchestrator for session %s", len(response.Results), config.SessionID)
	return response.Results, nil
}

// Close closes the connection to the orchestrator
func (c *OrchestratorClient) Close() error {
	if c.session != nil {
		return c.session.Close()
	}
	return nil
}
