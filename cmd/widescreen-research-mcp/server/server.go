package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/operations"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/orchestrator"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/schemas"
)

// WidescreenResearchServer is the main MCP server that provides widescreen research capabilities
type WidescreenResearchServer struct {
	server       *mcp.Server
	orchestrator *orchestrator.Orchestrator
	operations   *operations.OperationRegistry
	elicitation  *ElicitationManager
}

// NewWidescreenResearchServer creates a new instance of the widescreen research server
func NewWidescreenResearchServer() (*WidescreenResearchServer, error) {
	// Create MCP server
	mcpServer := mcp.NewServer(
		"widescreen-research",
		"1.0.0",
		mcp.WithCapabilities([]string{
			"tools",
			"prompts",
			"resources",
			"experimental/elicitation",
		}),
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
	s.server.RegisterTool("widescreen-research", mcp.Tool{
		Description: "Perform comprehensive widescreen research using distributed research drones",
		InputSchema: schemas.WidescreenResearchInput{},
		Handler: func(ctx context.Context, request interface{}) (interface{}, error) {
			input, ok := request.(*schemas.WidescreenResearchInput)
			if !ok {
				return nil, fmt.Errorf("invalid input type")
			}

			// Check if we need elicitation
			if input.Operation == "" || input.Operation == "start" {
				// Start elicitation process
				return s.handleElicitation(ctx, input)
			}

			// Execute the requested operation
			return s.executeOperation(ctx, input)
		},
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
		return operation.Execute(ctx, input.Parameters)
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
		Handler:     s.handleOrchestrateResearch,
	})

	s.operations.Register("sequential-thinking", &operations.Operation{
		Name:        "sequential-thinking",
		Description: "Perform sequential thinking style reasoning",
		Handler:     s.handleSequentialThinking,
	})

	s.operations.Register("gcp-provision", &operations.Operation{
		Name:        "gcp-provision",
		Description: "Provision GCP resources for research",
		Handler:     s.handleGCPProvision,
	})

	s.operations.Register("analyze-findings", &operations.Operation{
		Name:        "analyze-findings",
		Description: "Analyze research findings from drones",
		Handler:     s.handleAnalyzeFindings,
	})
}

// registerResources registers available resources
func (s *WidescreenResearchServer) registerResources() {
	// Register research reports resource
	s.server.RegisterResource("research-reports", mcp.Resource{
		URI:         "research://reports",
		Name:        "Research Reports",
		Description: "Access completed research reports",
		MimeType:    "application/json",
		Handler: func(ctx context.Context, uri string) (interface{}, error) {
			// Return list of available reports
			reports := s.orchestrator.GetReports()
			return json.Marshal(reports)
		},
	})

	// Register research templates resource
	s.server.RegisterResource("research-templates", mcp.Resource{
		URI:         "research://templates",
		Name:        "Research Templates",
		Description: "Pre-orchestrated research workflows",
		MimeType:    "application/json",
		Handler: func(ctx context.Context, uri string) (interface{}, error) {
			// Return available templates
			templates := s.orchestrator.GetTemplates()
			return json.Marshal(templates)
		},
	})
}

// registerPrompts registers available prompts
func (s *WidescreenResearchServer) registerPrompts() {
	// Register research planning prompt
	s.server.RegisterPrompt("research-planning", mcp.Prompt{
		Name:        "Research Planning",
		Description: "Plan a comprehensive research strategy",
		Arguments: []mcp.PromptArgument{
			{
				Name:        "topic",
				Description: "Research topic",
				Required:    true,
			},
			{
				Name:        "scope",
				Description: "Research scope",
				Required:    false,
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			topic := args["topic"].(string)
			scope := ""
			if s, ok := args["scope"].(string); ok {
				scope = s
			}
			return fmt.Sprintf("Research Plan for: %s\nScope: %s\n\n[Planning template here]", topic, scope), nil
		},
	})
}

// Start starts the MCP server
func (s *WidescreenResearchServer) Start(ctx context.Context) error {
	// Initialize orchestrator
	if err := s.orchestrator.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize orchestrator: %w", err)
	}

	// Start the MCP server
	return s.server.Serve(ctx)
}

// Shutdown gracefully shuts down the server
func (s *WidescreenResearchServer) Shutdown() {
	log.Println("Shutting down widescreen research server...")
	s.orchestrator.Shutdown()
	s.server.Close()
}