// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research-mcp/operations"
	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research-mcp/orchestrator"
	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research-mcp/schemas"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// WidescreenResearchServerOfficial is the main MCP server using official SDK
type WidescreenResearchServerOfficial struct {
	mcpServer    *mcp.Server
	orchestrator *orchestrator.Orchestrator
	operations   *operations.OperationRegistry
	elicitation  *ElicitationManager
}

// NewWidescreenResearchServerOfficial creates a new instance using official SDK
func NewWidescreenResearchServerOfficial() (*WidescreenResearchServerOfficial, error) {
	// Create orchestrator
	orch, err := orchestrator.NewOrchestrator()
	if err != nil {
		return nil, fmt.Errorf("failed to create orchestrator: %w", err)
	}

	// Create operation registry
	opRegistry := operations.NewOperationRegistry()

	// Create elicitation manager
	elicitManager := NewElicitationManager()

	srv := &WidescreenResearchServerOfficial{
		orchestrator: orch,
		operations:   opRegistry,
		elicitation:  elicitManager,
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

// WidescreenResearchArgs defines the arguments for the main tool
type WidescreenResearchArgs struct {
	Operation          string                 `json:"operation" jsonschema:"Research operation to perform"`
	Query              string                 `json:"query,omitempty" jsonschema:"Research query or topic"`
	SessionID          string                 `json:"session_id,omitempty" jsonschema:"Session ID for elicitation flow"`
	ElicitationAnswers map[string]interface{} `json:"elicitation_answers,omitempty" jsonschema:"Answers to elicitation questions"`
	Parameters         map[string]interface{} `json:"parameters,omitempty" jsonschema:"Additional operation parameters"`
}

// registerTools registers all MCP tools
func (s *WidescreenResearchServerOfficial) registerTools() {
	// Main widescreen-research tool
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "widescreen-research",
		Description: "Perform comprehensive widescreen research using distributed research drones",
	}, s.handleWidescreenResearchTool)
}

// handleWidescreenResearchTool is the main tool handler
func (s *WidescreenResearchServerOfficial) handleWidescreenResearchTool(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args WidescreenResearchArgs,
) (*mcp.CallToolResult, any, error) {
	// Create input struct
	input := &schemas.WidescreenResearchInput{
		Operation:          args.Operation,
		SessionID:          args.SessionID,
		ElicitationAnswers: args.ElicitationAnswers,
		Parameters: map[string]interface{}{
			"query": args.Query,
		},
	}

	// Merge additional parameters
	if args.Parameters != nil {
		for k, v := range args.Parameters {
			input.Parameters[k] = v
		}
	}

	// Check if we need elicitation
	if input.Operation == "" || input.Operation == "start" {
		result, err := s.handleElicitation(ctx, input)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Elicitation error: %v", err)},
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

	// Execute the requested operation
	result, err := s.executeOperation(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Operation error: %v", err)},
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

// handleElicitation manages the elicitation process
func (s *WidescreenResearchServerOfficial) handleElicitation(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	// Check current elicitation state
	state := s.elicitation.GetState(input.SessionID)

	if state == nil {
		// Start new elicitation
		questions := s.elicitation.GetInitialQuestions()
		return &schemas.ElicitationResponse{
			Type:      "elicitation",
			Questions: questions,
			SessionID: s.elicitation.CreateSession(),
		}, nil
	}

	// Process answers and get next questions
	nextQuestions, complete := s.elicitation.ProcessAnswers(input.SessionID, input.ElicitationAnswers)

	if !complete {
		return &schemas.ElicitationResponse{
			Type:      "elicitation",
			Questions: nextQuestions,
			SessionID: input.SessionID,
		}, nil
	}

	// Elicitation complete, prepare for research
	config := s.elicitation.GetResearchConfig(input.SessionID)
	return &schemas.ElicitationResponse{
		Type:      "ready",
		SessionID: input.SessionID,
		Message:   "Elicitation complete. Ready to start research.",
		Config:    config,
	}, nil
}

// executeOperation executes the requested operation
func (s *WidescreenResearchServerOfficial) executeOperation(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	operation := s.operations.GetOperation(input.Operation)
	if operation == nil {
		return nil, fmt.Errorf("unknown operation: %s", input.Operation)
	}

	// Execute operation based on type
	switch input.Operation {
	case "orchestrate-research":
		return s.handleOrchestrateResearch(ctx, input)
	case "gcp-provision":
		return s.handleGCPProvision(ctx, input)
	case "analyze-findings":
		return s.handleAnalyzeFindings(ctx, input)
	default:
		return operation.Handler(ctx, input.Parameters)
	}
}

// handleOrchestrateResearch handles the main research orchestration
func (s *WidescreenResearchServerOfficial) handleOrchestrateResearch(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	// Get research configuration from elicitation
	config := s.elicitation.GetResearchConfig(input.SessionID)
	if config == nil {
		return nil, fmt.Errorf("no research configuration found for session")
	}

	// Start orchestration
	result, err := s.orchestrator.OrchestrateResearch(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("orchestration failed: %w", err)
	}

	return result, nil
}

// handleGCPProvision handles GCP resource provisioning
func (s *WidescreenResearchServerOfficial) handleGCPProvision(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	provisioner := operations.NewGCPProvisioner()
	return provisioner.Execute(ctx, input.Parameters)
}

// handleAnalyzeFindings handles data analysis
func (s *WidescreenResearchServerOfficial) handleAnalyzeFindings(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	analyzer := operations.NewDataAnalyzer()
	return analyzer.Execute(ctx, input.Parameters)
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
