// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research-mcp/orchestrator"
	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research-mcp/schemas"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// WidescreenResearchServerOfficial is the main MCP server using official SDK
type WidescreenResearchServerOfficial struct {
	mcpServer    *mcp.Server
	orchestrator *orchestrator.Orchestrator
}

// NewWidescreenResearchServerOfficial creates a new instance using official SDK
func NewWidescreenResearchServerOfficial() (*WidescreenResearchServerOfficial, error) {
	// Create orchestrator
	orch, err := orchestrator.NewOrchestrator()
	if err != nil {
		return nil, fmt.Errorf("failed to create orchestrator: %w", err)
	}

	srv := &WidescreenResearchServerOfficial{
		orchestrator: orch,
	}

	// Create MCP server with official SDK
	srv.mcpServer = mcp.NewServer(&mcp.Implementation{
		Name:    "widescreen-research",
		Version: "1.0.0",
	}, nil)

	// Register tools
	srv.registerTools()

	// TODO: Register resources and prompts if needed
	// srv.registerResources()
	// srv.registerPrompts()

	return srv, nil
}

// registerTools registers all MCP tools
func (s *WidescreenResearchServerOfficial) registerTools() {
	// ARCHITECTURAL CHANGE: Renamed from "orchestrate-research" to "start-gcp-orchestration"
	// This server is now purely for GCP resource provisioning and drone management
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "start-gcp-orchestration",
		Description: "Start GCP orchestration: provision drones, manage Pub/Sub, collect results. Returns raw drone results (host handles report generation).",
	}, s.handleStartGCPOrchestration)
}

// handleStartGCPOrchestration is the main tool handler (renamed from handleWidescreenResearchTool)
func (s *WidescreenResearchServerOfficial) handleStartGCPOrchestration(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args map[string]interface{}, // Simplified args - just parameters
) (*mcp.CallToolResult, any, error) {
	// Build research configuration from parameters
	config := buildResearchConfig(args)

	// Start orchestration - returns drone results, not report
	result, err := s.orchestrator.OrchestrateResearch(ctx, config)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Orchestration error: %v", err)},
			},
		}, nil, nil
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, nil, nil
}

// buildResearchConfig builds a ResearchConfig from parameter map
func buildResearchConfig(params map[string]interface{}) *schemas.ResearchConfig {
	config := &schemas.ResearchConfig{
		SessionID:         getStringParam(params, "session_id", ""),
		Topic:             getStringParam(params, "topic", ""),
		ResearcherCount:   getIntParam(params, "researcher_count", 10),
		ResearchDepth:     getStringParam(params, "research_depth", "standard"),
		OutputFormat:      getStringParam(params, "output_format", "structured_json"),
		TimeoutMinutes:    getIntParam(params, "timeout_minutes", 60),
		PriorityLevel:     getStringParam(params, "priority_level", "normal"),
		WorkflowTemplates: getStringParam(params, "workflow_templates", ""),
		SpecificSources:   getStringParam(params, "specific_sources", ""),
		CreatedAt:         time.Now(),
	}

	// Generate session ID if not provided
	if config.SessionID == "" {
		config.SessionID = fmt.Sprintf("session-%d", time.Now().Unix())
	}

	return config
}

// Helper functions for parameter extraction
func getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if val, ok := params[key].(string); ok {
		return val
	}
	return defaultValue
}

func getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if val, ok := params[key].(float64); ok {
		return int(val)
	}
	if val, ok := params[key].(int); ok {
		return val
	}
	return defaultValue
}

// Run starts the MCP server with stdio transport
func (s *WidescreenResearchServerOfficial) Run(ctx context.Context) error {
	// Initialize orchestrator
	if err := s.orchestrator.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize orchestrator: %w", err)
	}

	log.Println("Widescreen research MCP server started (Official SDK) with stdio transport")

	// Run server with stdio transport
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}

// RunHTTP starts the MCP server with streamable HTTP transport
func (s *WidescreenResearchServerOfficial) RunHTTP(ctx context.Context, addr string) error {
	// Initialize orchestrator
	if err := s.orchestrator.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize orchestrator: %w", err)
	}

	// Create HTTP handler
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return s.mcpServer
	}, nil)

	log.Printf("Widescreen research MCP server started with HTTP transport on %s", addr)

	// Start HTTP server
	return http.ListenAndServe(addr, handler)
}

// Shutdown gracefully shuts down the server
func (s *WidescreenResearchServerOfficial) Shutdown() {
	log.Println("Shutting down widescreen research server...")
	s.orchestrator.Shutdown()
}
