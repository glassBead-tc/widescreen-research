package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research-mcp/operations"
	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research-mcp/orchestrator"
	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research-mcp/schemas"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// WidescreenResearchServer is the main MCP server that provides widescreen research capabilities
type WidescreenResearchServer struct {
	server       *server.MCPServer
	orchestrator *orchestrator.Orchestrator
	operations   *operations.OperationRegistry
	elicitation  *ElicitationManager
}

// NewWidescreenResearchServer creates a new instance of the widescreen research server
func NewWidescreenResearchServer() (*WidescreenResearchServer, error) {
	// Create MCP server
	mcpServer := server.NewMCPServer(
		"widescreen-research",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
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
		server:       mcpServer,
		orchestrator: orch,
		operations:   opRegistry,
		elicitation:  elicitManager,
	}

	// Register the main widescreen-research tool
	srv.registerWidescreenResearchTool()

	// Register operations
	srv.registerOperations()

	// Register resources
	srv.registerResources()

	// Register prompts
	srv.registerPrompts()

	return srv, nil
}

// registerWidescreenResearchTool registers the main tool that handles all operations
func (s *WidescreenResearchServer) registerWidescreenResearchTool() {
	tool := mcp.NewTool("widescreen-research",
		mcp.WithDescription("Perform comprehensive widescreen research using distributed research drones"),
		mcp.WithString("operation", mcp.Required(), mcp.Description("Research operation to perform")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Research query or topic")),
	)

	s.server.AddTool(tool, s.handleWidescreenResearchTool)
}

// handleWidescreenResearchTool is the tool handler function
func (s *WidescreenResearchServer) handleWidescreenResearchTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters from arguments
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid arguments format")
	}

	operation, _ := args["operation"].(string)
	query, _ := args["query"].(string)

	// Create input struct
	input := &schemas.WidescreenResearchInput{
		Operation: operation,
		Parameters: map[string]interface{}{
			"query": query,
		},
	}

	// Check if we need elicitation
	if input.Operation == "" || input.Operation == "start" {
		// Start elicitation process
		result, err := s.handleElicitation(ctx, input)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("%v", result)), nil
	}

	// Execute the requested operation
	result, err := s.executeOperation(ctx, input)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("%v", result)), nil
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
			input := &schemas.WidescreenResearchInput{Parameters: params}
			return s.handleOrchestrateResearch(ctx, input)
		},
	})

	s.operations.Register("sequential-thinking", &operations.Operation{
		Name:        "sequential-thinking",
		Description: "Perform sequential thinking style reasoning",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			input := &schemas.WidescreenResearchInput{Parameters: params}
			return s.handleSequentialThinking(ctx, input)
		},
	})

	s.operations.Register("gcp-provision", &operations.Operation{
		Name:        "gcp-provision",
		Description: "Provision GCP resources for research",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			input := &schemas.WidescreenResearchInput{Parameters: params}
			return s.handleGCPProvision(ctx, input)
		},
	})

	s.operations.Register("analyze-findings", &operations.Operation{
		Name:        "analyze-findings",
		Description: "Analyze research findings from drones",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			input := &schemas.WidescreenResearchInput{Parameters: params}
			return s.handleAnalyzeFindings(ctx, input)
		},
	})
}

// registerResources registers available resources
func (s *WidescreenResearchServer) registerResources() {
	// Register research reports resource
	reportsResource := mcp.NewResource("research://reports", "Research Reports",
		mcp.WithResourceDescription("Access completed research reports"),
		mcp.WithMIMEType("application/json"),
	)
	s.server.AddResource(reportsResource, s.handleReportsResource)

	// Register research templates resource
	templatesResource := mcp.NewResource("research://templates", "Research Templates",
		mcp.WithResourceDescription("Pre-orchestrated research workflows"),
		mcp.WithMIMEType("application/json"),
	)
	s.server.AddResource(templatesResource, s.handleTemplatesResource)
}

// handleReportsResource handles requests for research reports
func (s *WidescreenResearchServer) handleReportsResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	reports := s.orchestrator.GetReports()
	data, err := json.Marshal(reports)
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{Text: string(data)},
	}, nil
}

// handleTemplatesResource handles requests for research templates
func (s *WidescreenResearchServer) handleTemplatesResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	templates := s.orchestrator.GetTemplates()
	data, err := json.Marshal(templates)
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{Text: string(data)},
	}, nil
}

// registerPrompts registers available prompts
func (s *WidescreenResearchServer) registerPrompts() {
	// Register research planning prompt
	prompt := mcp.NewPrompt("research-planning",
		mcp.WithPromptDescription("Plan a comprehensive research strategy"),
		mcp.WithArgument("topic", mcp.RequiredArgument(), mcp.ArgumentDescription("Research topic")),
		mcp.WithArgument("scope", mcp.ArgumentDescription("Research scope")),
	)

	s.server.AddPrompt(prompt, s.handleResearchPlanningPrompt)
}

// handleResearchPlanningPrompt handles the research planning prompt
func (s *WidescreenResearchServer) handleResearchPlanningPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	args := request.Params.Arguments
	topic := args["topic"]
	scope := args["scope"]

	// Generate research planning prompt
	promptText := fmt.Sprintf(`Research Plan for: %s
Scope: %s

## Suggested Research Strategy
1. Initial exploration and background research
2. Identify key stakeholders and sources
3. Conduct comprehensive analysis
4. Synthesize findings and insights

Please provide additional context or specific requirements for this research.`, topic, scope)

	messages := []mcp.PromptMessage{
		mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(promptText)),
	}

	return mcp.NewGetPromptResult("Research planning strategy", messages), nil
}

// Start starts the MCP server
func (s *WidescreenResearchServer) Start(ctx context.Context) error {
	// Initialize orchestrator
	if err := s.orchestrator.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize orchestrator: %w", err)
	}

	// MCP server is ready to handle requests
	log.Println("Widescreen research MCP server started")
	return nil
}

// Shutdown gracefully shuts down the server
func (s *WidescreenResearchServer) Shutdown() {
	log.Println("Shutting down widescreen research server...")
	s.orchestrator.Shutdown()
}
