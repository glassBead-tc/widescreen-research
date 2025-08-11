package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/operations"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/orchestrator"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/schemas"
)

// WidescreenResearchServer is the main MCP server that provides widescreen research capabilities
type WidescreenResearchServer struct {
	mcpServer    *mcpserver.MCPServer
	orchestrator *orchestrator.Orchestrator
	operations   *operations.OperationRegistry
	elicitation  *ElicitationManager
}

// NewWidescreenResearchServer creates a new instance of the widescreen research server
func NewWidescreenResearchServer() (*WidescreenResearchServer, error) {
	// Create MCP server
	mcpSrv := mcpserver.NewMCPServer(
		"widescreen-research",
		"1.0.0",
		mcpserver.WithToolCapabilities(true),
		mcpserver.WithRecovery(),
	)

	// Create orchestrator
	orch, err := orchestrator.NewOrchestrator()
	if err != nil {
		return nil, fmt.Errorf("failed to create orchestrator: %w", err)
	}

	// Create operation registry
	opRegistry := operations.NewOperationRegistry()

	// Create elicitation manager
	elicitManager := NewElicitationManager()

	srv := &WidescreenResearchServer{
		mcpServer:    mcpSrv,
		orchestrator: orch,
		operations:   opRegistry,
		elicitation:  elicitManager,
	}

	// Register the main widescreen-research tool
	srv.registerWidescreenResearchTool()

	// Register operations
	srv.registerOperations()

	// Resource and prompt registration are not supported with the current mcp-go server API used in this project.

	return srv, nil
}

// registerWidescreenResearchTool registers the main tool that handles all operations
func (s *WidescreenResearchServer) registerWidescreenResearchTool() {
	tool := mcp.NewTool(
		"widescreen_research",
		mcp.WithDescription("Perform comprehensive widescreen research using distributed research drones"),
		mcp.WithString("operation", mcp.Description("Operation to execute")),
		mcp.WithString("session_id", mcp.Description("Session ID for elicitation and orchestration")),
		mcp.WithString("parameters_json", mcp.Description("JSON-encoded parameters for the operation")),
		mcp.WithString("elicitation_answers_json", mcp.Description("JSON-encoded elicitation answers")),
	)

	s.mcpServer.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Build input from tool request
		op := req.GetString("operation", "")
		sessionID := req.GetString("session_id", "")

		params := map[string]interface{}{}
		if pstr := req.GetString("parameters_json", ""); pstr != "" {
			_ = json.Unmarshal([]byte(pstr), &params)
		}

		elicit := map[string]interface{}{}
		if estr := req.GetString("elicitation_answers_json", ""); estr != "" {
			_ = json.Unmarshal([]byte(estr), &elicit)
		}

		input := &schemas.WidescreenResearchInput{
			Operation:          op,
			SessionID:          sessionID,
			ElicitationAnswers: elicit,
			Parameters:         params,
		}

		var result interface{}
		var err error

		if input.Operation == "" || input.Operation == "start" {
			result, err = s.handleElicitation(ctx, input)
		} else {
			result, err = s.executeOperation(ctx, input)
		}

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return JSON-encoded result as text
		b, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(b)), nil
	})
}

// handleElicitation manages the elicitation process
func (s *WidescreenResearchServer) handleElicitation(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
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
func (s *WidescreenResearchServer) executeOperation(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	operation := s.operations.GetOperation(input.Operation)
	if operation == nil {
		return nil, fmt.Errorf("unknown operation: %s", input.Operation)
	}

	// Execute operation based on type
	switch input.Operation {
	case "orchestrate-research":
		return s.handleOrchestrateResearch(ctx, input)
	case "sequential-thinking":
		return s.handleSequentialThinking(ctx, input)
	case "gcp-provision":
		return s.handleGCPProvision(ctx, input)
	case "analyze-findings":
		return s.handleAnalyzeFindings(ctx, input)
	default:
		return operation.Handler(ctx, input.Parameters)
	}
}

// handleOrchestrateResearch handles the main research orchestration
func (s *WidescreenResearchServer) handleOrchestrateResearch(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
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

// handleSequentialThinking handles sequential thinking operations
func (s *WidescreenResearchServer) handleSequentialThinking(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	thinking := operations.NewSequentialThinking()
	return thinking.Execute(ctx, input.Parameters)
}

// handleGCPProvision handles GCP resource provisioning
func (s *WidescreenResearchServer) handleGCPProvision(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	provisioner := operations.NewGCPProvisioner()
	return provisioner.Execute(ctx, input.Parameters)
}

// handleAnalyzeFindings handles data analysis of research findings
func (s *WidescreenResearchServer) handleAnalyzeFindings(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	analyzer := operations.NewDataAnalyzer()
	return analyzer.Execute(ctx, input.Parameters)
}

// registerOperations registers all available operations
func (s *WidescreenResearchServer) registerOperations() {
	// Register core operations
	s.operations.Register("orchestrate-research", &operations.Operation{
		Name:        "orchestrate-research",
		Description: "Orchestrate distributed research using multiple drones",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return s.handleOrchestrateResearch(ctx, &schemas.WidescreenResearchInput{Parameters: params})
		},
	})

	s.operations.Register("sequential-thinking", &operations.Operation{
		Name:        "sequential-thinking",
		Description: "Perform sequential thinking style reasoning",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return s.handleSequentialThinking(ctx, &schemas.WidescreenResearchInput{Parameters: params})
		},
	})

	s.operations.Register("gcp-provision", &operations.Operation{
		Name:        "gcp-provision",
		Description: "Provision GCP resources for research",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return s.handleGCPProvision(ctx, &schemas.WidescreenResearchInput{Parameters: params})
		},
	})

	s.operations.Register("analyze-findings", &operations.Operation{
		Name:        "analyze-findings",
		Description: "Analyze research findings from drones",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return s.handleAnalyzeFindings(ctx, &schemas.WidescreenResearchInput{Parameters: params})
		},
	})
}

// registerResources registers available resources
// registerResources is intentionally left unimplemented for the current mcp-go server API version
func (s *WidescreenResearchServer) registerResources() {}

// registerPrompts registers available prompts
// registerPrompts is intentionally left unimplemented for the current mcp-go server API version
func (s *WidescreenResearchServer) registerPrompts() {}

// Start starts the MCP server
func (s *WidescreenResearchServer) Start(ctx context.Context) error {
	// Initialize orchestrator
	if err := s.orchestrator.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize orchestrator: %w", err)
	}

	// Start the MCP server
	return mcpserver.ServeStdio(s.mcpServer)
}

// Shutdown gracefully shuts down the server
func (s *WidescreenResearchServer) Shutdown() {
	log.Println("Shutting down widescreen research server...")
	s.orchestrator.Shutdown()
	// No explicit Close method in current mcp-go server API
}
