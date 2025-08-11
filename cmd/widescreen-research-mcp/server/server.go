package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/operations"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/orchestrator"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/resources"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/schemas"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/tools"
)

// WidescreenResearchServer is the main MCP server that provides widescreen research capabilities
type WidescreenResearchServer struct {
	mcpServer    *mcpserver.MCPServer
	orchestrator *orchestrator.Orchestrator  // Lazy initialized
	operations   *operations.OperationRegistry
	elicitation  *ElicitationManager
	guides       *resources.GuideResource
	toolRegistry *tools.Registry
	mu           sync.Mutex  // Protects orchestrator initialization
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

	// Create operation registry
	opRegistry := operations.NewOperationRegistry()

	// Create elicitation manager
	elicitManager := NewElicitationManager()

	// Create guide resource
	guideResource := resources.NewGuideResource()

	// Create tool registry
	toolRegistry := tools.NewRegistry()

	srv := &WidescreenResearchServer{
		mcpServer:    mcpSrv,
		orchestrator: nil,  // Will be lazy initialized when needed
		operations:   opRegistry,
		elicitation:  elicitManager,
		guides:       guideResource,
		toolRegistry: toolRegistry,
	}

	// Register all tools
	srv.registerTools()

	// Register operations
	srv.registerOperations()

	// Resource and prompt registration are not supported with the current mcp-go server API used in this project.

	return srv, nil
}

// getOrCreateOrchestrator lazily initializes the orchestrator when first needed
func (s *WidescreenResearchServer) getOrCreateOrchestrator(ctx context.Context) (*orchestrator.Orchestrator, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.orchestrator != nil {
		return s.orchestrator, nil
	}
	
	// Create orchestrator now that environment variables should be available
	orch, err := orchestrator.NewOrchestrator()
	if err != nil {
		return nil, fmt.Errorf("failed to create orchestrator: %w", err)
	}
	
	// Initialize the orchestrator
	if err := orch.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize orchestrator: %w", err)
	}
	
	s.orchestrator = orch
	log.Println("Orchestrator initialized successfully")
	return s.orchestrator, nil
}


// registerTools registers all tools using the tool registry
func (s *WidescreenResearchServer) registerTools() {
	// For now, register tools directly until registry is fully integrated
	s.registerWidescreenResearchTool()
	s.registerGuideTool()
	
	// Tool registry will be used for advanced features in future
	// widescreenHandler := tools.NewWidescreenToolHandler(s.handleWidescreenRequest)
	// s.toolRegistry.Register(widescreenHandler.GetDefinition())
	// guideHandler := tools.NewGuideToolHandler(s.guides)
	// s.toolRegistry.Register(guideHandler.GetDefinition())
	// s.toolRegistry.RegisterWithServer(s.mcpServer)
}

// registerWidescreenResearchTool registers the main widescreen research tool
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

		result, err := s.handleWidescreenRequest(ctx, input)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return JSON-encoded result as text
		b, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(b)), nil
	})
}

// registerGuideTool registers the guide access tool
func (s *WidescreenResearchServer) registerGuideTool() {
	tool := mcp.NewTool(
		"get_guide",
		mcp.WithDescription("Get research system guides and documentation. Use 'list' as name to see all available guides."),
		mcp.WithString("name", mcp.Description("Guide name: 'main', 'websets', 'orchestration', 'quickstart', or 'list' to see all")),
	)

	s.mcpServer.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		guideName := req.GetString("name", "main")
		
		// Special case: list available guides
		if guideName == "list" {
			guides := s.guides.ListGuides()
			result := "Available guides:\n"
			for _, name := range guides {
				result += fmt.Sprintf("- %s\n", name)
			}
			result += "\nUse get_guide with the guide name to read it."
			return mcp.NewToolResultText(result), nil
		}
		
		guide, err := s.guides.GetGuide(guideName)
		if err != nil {
			availableGuides := s.guides.ListGuides()
			return mcp.NewToolResultError(fmt.Sprintf("Guide '%s' not found. Available guides: %v", guideName, availableGuides)), nil
		}
		
		return mcp.NewToolResultText(guide), nil
	})
}

// handleWidescreenRequest processes the main widescreen research requests
func (s *WidescreenResearchServer) handleWidescreenRequest(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	if input.Operation == "" || input.Operation == "start" {
		return s.handleElicitation(ctx, input)
	}
	return s.executeOperation(ctx, input)
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
	case "websets-orchestrate":
		return s.handleWebsetsOrchestrate(ctx, input)
	case "websets-call":
		return s.handleWebsetsCall(ctx, input)
	default:
		return operation.Handler(ctx, input.Parameters)
	}
}

// handleOrchestrateResearch handles the main research orchestration
func (s *WidescreenResearchServer) handleOrchestrateResearch(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	// Get orchestrator (lazy initialization)
	orch, err := s.getOrCreateOrchestrator(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get orchestrator: %w", err)
	}
	
	// Get research configuration from elicitation
	config := s.elicitation.GetResearchConfig(input.SessionID)
	if config == nil {
		return nil, fmt.Errorf("no research configuration found for session")
	}

	// Start orchestration
	result, err := orch.OrchestrateResearch(ctx, config)
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

// handleWebsetsOrchestrate handles websets research orchestration
func (s *WidescreenResearchServer) handleWebsetsOrchestrate(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	// Get orchestrator (lazy initialization)
	orch, err := s.getOrCreateOrchestrator(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get orchestrator: %w", err)
	}
	
	// Extract parameters
	topic, ok := input.Parameters["topic"].(string)
	if !ok || topic == "" {
		return nil, fmt.Errorf("topic parameter is required")
	}
	
	resultCount := 50 // default
	if rc, ok := input.Parameters["result_count"].(float64); ok {
		resultCount = int(rc)
	}
	
	// Run websets pipeline
	result, err := orch.RunWebsetsPipeline(ctx, topic, resultCount)
	if err != nil {
		return nil, fmt.Errorf("websets orchestration failed: %w", err)
	}
	
	return result, nil
}

// handleWebsetsCall handles direct websets tool calls
func (s *WidescreenResearchServer) handleWebsetsCall(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	// Get orchestrator (lazy initialization)
	orch, err := s.getOrCreateOrchestrator(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get orchestrator: %w", err)
	}
	
	// Direct pass-through to websets manager
	mcpClient := orch.GetMCPClient()
	result, err := mcpClient.CallTool(ctx, "websets", "websets_manager", input.Parameters)
	if err != nil {
		return nil, fmt.Errorf("websets call failed: %w", err)
	}
	
	return result, nil
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
	
	s.operations.Register("websets-orchestrate", &operations.Operation{
		Name:        "websets-orchestrate",
		Description: "Orchestrate websets research pipeline (create → poll → list → publish)",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return s.handleWebsetsOrchestrate(ctx, &schemas.WidescreenResearchInput{Parameters: params})
		},
	})
	
	s.operations.Register("websets-call", &operations.Operation{
		Name:        "websets-call",
		Description: "Direct call to websets_manager tool for custom operations",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return s.handleWebsetsCall(ctx, &schemas.WidescreenResearchInput{Parameters: params})
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
	// Note: Orchestrator initialization is now lazy - happens on first use
	// This allows MCP protocol to establish and pass environment variables first
	
	// Start the MCP server
	return mcpserver.ServeStdio(s.mcpServer)
}

// Shutdown gracefully shuts down the server
func (s *WidescreenResearchServer) Shutdown() {
	log.Println("Shutting down widescreen research server...")
	
	// Only shutdown orchestrator if it was created
	s.mu.Lock()
	if s.orchestrator != nil {
		s.orchestrator.Shutdown()
	}
	s.mu.Unlock()
	
	// No explicit Close method in current mcp-go server API
}
