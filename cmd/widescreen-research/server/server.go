// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research-mcp/schemas"
	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research/aggregator"
	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research/client"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// WidescreenResearchHost is the bidirectional MCP entity (server to Claude Desktop, client to widescreen-research-mcp)
type WidescreenResearchHost struct {
	mcpServer          *mcp.Server
	orchestratorClient *client.OrchestratorClient
	aggregator         *aggregator.ReportAggregator

	// Session state management
	sessions map[string]*ResearchSession
	reports  map[string]*schemas.ResearchReport
	mu       sync.RWMutex
}

// ResearchSession tracks an active research session
type ResearchSession struct {
	ID          string
	Config      *schemas.ResearchConfig
	Status      string
	Results     []schemas.DroneResult
	Report      *schemas.ResearchReport
	StartTime   time.Time
	CompletedAt time.Time
}

// NewWidescreenResearchHost creates a new bidirectional host instance
func NewWidescreenResearchHost(orchestratorURL string) (*WidescreenResearchHost, error) {
	// Create orchestrator client (MCP client to widescreen-research-mcp)
	orchClient, err := client.NewOrchestratorClient(orchestratorURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create orchestrator client: %w", err)
	}

	// Create report aggregator
	agg := aggregator.NewReportAggregator()

	host := &WidescreenResearchHost{
		orchestratorClient: orchClient,
		aggregator:         agg,
		sessions:           make(map[string]*ResearchSession),
		reports:            make(map[string]*schemas.ResearchReport),
	}

	// Create MCP server (exposes tools to Claude Desktop)
	host.mcpServer = mcp.NewServer(&mcp.Implementation{
		Name:    "widescreen-research",
		Version: "1.0.0",
	}, nil)

	// Register tools
	host.registerTools()

	// TODO: Register resources for report access (optional enhancement)
	// host.registerResources()

	return host, nil
}

// OrchestrateResearchArgs defines arguments for orchestrate-research tool
type OrchestrateResearchArgs struct {
	Topic           string `json:"topic" jsonschema:"required,Research topic or query"`
	ResearcherCount int    `json:"researcher_count,omitempty" jsonschema:"Number of parallel research drones (default: 10)"`
	ResearchDepth   string `json:"research_depth,omitempty" jsonschema:"Research depth: quick, standard, or deep (default: standard)"`
	TimeoutMinutes  int    `json:"timeout_minutes,omitempty" jsonschema:"Timeout in minutes (default: 60)"`
	PriorityLevel   string `json:"priority_level,omitempty" jsonschema:"Priority level: low, normal, or high (default: normal)"`
}

// GetReportArgs defines arguments for get-report tool
type GetReportArgs struct {
	SessionID string `json:"session_id" jsonschema:"required,Session ID of the research report to retrieve"`
}

// registerTools registers all MCP tools exposed to Claude Desktop
func (h *WidescreenResearchHost) registerTools() {
	// Tool 1: orchestrate-research (exposed to Claude Desktop)
	mcp.AddTool(h.mcpServer, &mcp.Tool{
		Name:        "orchestrate-research",
		Description: "Orchestrate distributed research on a topic using parallel research drones. Returns when research is complete with report ready.",
	}, h.handleOrchestrateResearch)

	// Tool 2: get-report (exposed to Claude Desktop)
	mcp.AddTool(h.mcpServer, &mcp.Tool{
		Name:        "get-report",
		Description: "Retrieve a completed research report by session ID",
	}, h.handleGetReport)

	// Tool 3: list-sessions (exposed to Claude Desktop)
	mcp.AddTool(h.mcpServer, &mcp.Tool{
		Name:        "list-sessions",
		Description: "List all research sessions and their status",
	}, h.handleListSessions)
}

// registerResources registers MCP resources for report access (optional)
// func (h *WidescreenResearchHost) registerResources() {
// 	// Resource for accessing reports - TODO: implement if needed
// 	h.mcpServer.AddResource(&mcp.Resource{
// 		URI:         "widescreen://reports/{session_id}",
// 		Name:        "Research Report",
// 		Description: "Access research reports by session ID",
// 		MIMEType:    "application/json",
// 	}, h.handleReportResource)
// }

// handleOrchestrateResearch handles the orchestrate-research tool call from Claude Desktop
func (h *WidescreenResearchHost) handleOrchestrateResearch(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args OrchestrateResearchArgs,
) (*mcp.CallToolResult, any, error) {
	// Validate input
	if args.Topic == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "topic parameter is required"},
			},
		}, nil, nil
	}

	// Set defaults
	if args.ResearcherCount == 0 {
		args.ResearcherCount = 10
	}
	if args.ResearchDepth == "" {
		args.ResearchDepth = "standard"
	}
	if args.TimeoutMinutes == 0 {
		args.TimeoutMinutes = 60
	}
	if args.PriorityLevel == "" {
		args.PriorityLevel = "normal"
	}

	// Create research session
	sessionID := fmt.Sprintf("session-%s", uuid.New().String()[:8])
	config := &schemas.ResearchConfig{
		SessionID:       sessionID,
		Topic:           args.Topic,
		ResearcherCount: args.ResearcherCount,
		ResearchDepth:   args.ResearchDepth,
		TimeoutMinutes:  args.TimeoutMinutes,
		PriorityLevel:   args.PriorityLevel,
		OutputFormat:    "structured_json",
		CreatedAt:       time.Now(),
	}

	session := &ResearchSession{
		ID:        sessionID,
		Config:    config,
		Status:    "initializing",
		Results:   make([]schemas.DroneResult, 0),
		StartTime: time.Now(),
	}

	h.mu.Lock()
	h.sessions[sessionID] = session
	h.mu.Unlock()

	log.Printf("Created research session %s for topic: %s", sessionID, args.Topic)

	// Start orchestration asynchronously
	go h.runResearchSession(ctx, session)

	// Return immediately with session ID
	result := map[string]interface{}{
		"session_id": sessionID,
		"status":     "started",
		"message":    fmt.Sprintf("Research session started with %d drones. Use get-report or access resource widescreen://reports/%s when complete.", args.ResearcherCount, sessionID),
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, nil, nil
}

// runResearchSession executes the full research workflow
func (h *WidescreenResearchHost) runResearchSession(ctx context.Context, session *ResearchSession) {
	session.Status = "orchestrating"

	// Call widescreen-research-mcp's start-gcp-orchestration tool
	log.Printf("[Session %s] Delegating GCP orchestration to widescreen-research-mcp", session.ID)

	results, err := h.orchestratorClient.StartGCPOrchestration(ctx, session.Config)
	if err != nil {
		log.Printf("[Session %s] Orchestration failed: %v", session.ID, err)
		session.Status = "failed"
		return
	}

	// Store collected results
	h.mu.Lock()
	session.Results = results
	session.Status = "aggregating"
	h.mu.Unlock()

	log.Printf("[Session %s] Received %d drone results, generating report", session.ID, len(results))

	// Aggregate results into report (THIS IS THE KEY CHANGE - report generation happens HERE)
	report, err := h.aggregator.GenerateReport(ctx, session.Config, results)
	if err != nil {
		log.Printf("[Session %s] Report generation failed: %v", session.ID, err)
		session.Status = "failed_report_generation"
		return
	}

	// Store report
	h.mu.Lock()
	session.Report = report
	session.Status = "completed"
	session.CompletedAt = time.Now()
	h.reports[report.ID] = report
	h.mu.Unlock()

	log.Printf("[Session %s] Research complete! Report ID: %s", session.ID, report.ID)

	// TODO: Send notification to Claude Desktop when ready
	// This could be done via a callback mechanism or polling
}

// handleGetReport handles the get-report tool call
func (h *WidescreenResearchHost) handleGetReport(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args GetReportArgs,
) (*mcp.CallToolResult, any, error) {
	h.mu.RLock()
	session, exists := h.sessions[args.SessionID]
	h.mu.RUnlock()

	if !exists {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Session %s not found", args.SessionID)},
			},
		}, nil, nil
	}

	if session.Status != "completed" {
		result := map[string]interface{}{
			"session_id": args.SessionID,
			"status":     session.Status,
			"message":    fmt.Sprintf("Research is still %s. Please wait for completion.", session.Status),
		}
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(resultJSON)},
			},
		}, nil, nil
	}

	// Return the report
	reportJSON, _ := json.MarshalIndent(session.Report, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(reportJSON)},
		},
	}, nil, nil
}

// handleListSessions handles the list-sessions tool call
func (h *WidescreenResearchHost) handleListSessions(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args struct{},
) (*mcp.CallToolResult, any, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	sessionList := make([]map[string]interface{}, 0, len(h.sessions))
	for _, session := range h.sessions {
		sessionList = append(sessionList, map[string]interface{}{
			"session_id": session.ID,
			"topic":      session.Config.Topic,
			"status":     session.Status,
			"started_at": session.StartTime,
		})
	}

	result := map[string]interface{}{
		"sessions": sessionList,
		"total":    len(sessionList),
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, nil, nil
}

// handleReportResource handles resource requests for reports (optional feature)
// func (h *WidescreenResearchHost) handleReportResource(
// 	ctx context.Context,
// 	req *mcp.ReadResourceParams,
// ) (*mcp.ReadResourceResult, error) {
// 	// Extract session_id from URI
// 	// Implementation deferred - use get-report tool instead
// 	return nil, fmt.Errorf("resource access not yet implemented")
// }

// Run starts the MCP server with stdio transport
func (h *WidescreenResearchHost) Run(ctx context.Context) error {
	// Connect to orchestrator
	if err := h.orchestratorClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to orchestrator: %w", err)
	}

	log.Println("Widescreen research host started (bidirectional MCP) with stdio transport")
	return h.mcpServer.Run(ctx, &mcp.StdioTransport{})
}

// RunHTTP starts the MCP server with HTTP transport
func (h *WidescreenResearchHost) RunHTTP(ctx context.Context, addr string) error {
	// Connect to orchestrator
	if err := h.orchestratorClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to orchestrator: %w", err)
	}

	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return h.mcpServer
	}, nil)

	log.Printf("Widescreen research host started with HTTP transport on %s", addr)
	return http.ListenAndServe(addr, handler)
}

// Shutdown gracefully shuts down the host
func (h *WidescreenResearchHost) Shutdown() {
	log.Println("Shutting down widescreen research host...")
	if h.orchestratorClient != nil {
		h.orchestratorClient.Close()
	}
}
