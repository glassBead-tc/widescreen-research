package mcp

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spawn-mcp/coordinator/pkg/coordinator"
	"github.com/spawn-mcp/coordinator/pkg/types"
)

// MCPServer wraps the coordinator with MCP protocol support
type MCPServer struct {
	coordinator *coordinator.Server
	mcpServer   *server.MCPServer
}

// NewMCPServer creates a new MCP server that exposes coordinator functionality
func NewMCPServer(coord *coordinator.Server) *MCPServer {
	mcpServer := server.NewMCPServer(
		"Spawn MCP Coordinator",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	s := &MCPServer{
		coordinator: coord,
		mcpServer:   mcpServer,
	}

	// Register MCP tools
	s.registerTools()

	return s
}

// registerTools registers all available MCP tools
func (s *MCPServer) registerTools() {
	// Tool: Spawn Drone Server
	spawnDroneTool := mcp.NewTool("spawn_drone_server",
		mcp.WithDescription("Spawn a new drone MCP server on Cloud Run"),
		mcp.WithString("drone_type",
			mcp.Required(),
			mcp.Description("Type of drone to spawn (researcher, analyst, etc.)"),
			mcp.Enum("researcher", "analyst", "writer", "coder"),
		),
		mcp.WithString("region",
			mcp.Description("GCP region to deploy to"),
			mcp.DefaultString("us-central1"),
		),
	)

	s.mcpServer.AddTool(spawnDroneTool, s.handleSpawnDrone)

	// Tool: List Active Drones
	listDronesTool := mcp.NewTool("list_active_drones",
		mcp.WithDescription("List all currently active drone servers"),
	)

	s.mcpServer.AddTool(listDronesTool, s.handleListDrones)

	// Tool: Execute Distributed Task
	executeTaskTool := mcp.NewTool("execute_distributed_task",
		mcp.WithDescription("Execute a task across the drone fleet"),
		mcp.WithString("task_type",
			mcp.Required(),
			mcp.Description("Type of task to execute"),
			mcp.Enum("research", "analysis", "synthesis", "coding"),
		),
		mcp.WithString("description",
			mcp.Required(),
			mcp.Description("Detailed description of the task"),
		),
		mcp.WithNumber("max_drones",
			mcp.Description("Maximum number of drones to use"),
			mcp.DefaultNumber(3),
			mcp.Min(1),
			mcp.Max(10),
		),
	)

	s.mcpServer.AddTool(executeTaskTool, s.handleExecuteTask)

	// Tool: Get Drone Status
	getDroneStatusTool := mcp.NewTool("get_drone_status",
		mcp.WithDescription("Get detailed status of a specific drone"),
		mcp.WithString("drone_id",
			mcp.Required(),
			mcp.Description("ID of the drone to check"),
		),
	)

	s.mcpServer.AddTool(getDroneStatusTool, s.handleGetDroneStatus)

	// Tool: Terminate Drone
	terminateDroneTool := mcp.NewTool("terminate_drone",
		mcp.WithDescription("Terminate a specific drone server"),
		mcp.WithString("drone_id",
			mcp.Required(),
			mcp.Description("ID of the drone to terminate"),
		),
	)

	s.mcpServer.AddTool(terminateDroneTool, s.handleTerminateDrone)

	// New tools for campaign orchestration
	planCampaign := mcp.NewTool("plan_campaign",
		mcp.WithDescription("Validate a campaign spec and produce an execution plan"),
		mcp.WithString("spec_json",
			mcp.Required(),
			mcp.Description("JSON-encoded CampaignSpec"),
		),
	)
	s.mcpServer.AddTool(planCampaign, s.handlePlanCampaign)

	launchFleet := mcp.NewTool("launch_fleet",
		mcp.WithDescription("Provision worker fleet and seed queue for a campaign run"),
		mcp.WithString("run_id", mcp.Required()),
		mcp.WithNumber("target_workers", mcp.DefaultNumber(10), mcp.Min(1), mcp.Max(100)),
	)
	s.mcpServer.AddTool(launchFleet, s.handleLaunchFleet)

	fleetStatus := mcp.NewTool("fleet_status",
		mcp.WithDescription("Get current status and progress for a campaign run"),
		mcp.WithString("run_id", mcp.Required()),
	)
	s.mcpServer.AddTool(fleetStatus, s.handleFleetStatus)

	abort := mcp.NewTool("abort",
		mcp.WithDescription("Abort a campaign run and scale down workers"),
		mcp.WithString("run_id", mcp.Required()),
	)
	s.mcpServer.AddTool(abort, s.handleAbort)

	exportGraph := mcp.NewTool("export_graph",
		mcp.WithDescription("Export collected graph for a mem0 space or run"),
		mcp.WithString("mem0_space", mcp.Required()),
		mcp.WithString("format", mcp.DefaultString("jsonl"), mcp.Enum("jsonl", "csv")),
	)
	s.mcpServer.AddTool(exportGraph, s.handleExportGraph)
}

// handleSpawnDrone handles the spawn_drone_server tool call
func (s *MCPServer) handleSpawnDrone(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	droneType, err := request.RequireString("drone_type")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid drone_type: %v", err)), nil
	}

	region := request.GetString("region", "us-central1")

	log.Printf("Spawning drone: type=%s, region=%s", droneType, region)

	// Create drone configuration
	droneConfig := types.DroneConfig{
		Type:   types.DroneType(droneType),
		Region: region,
		Capabilities: []string{
			"web_search",
			"data_analysis",
			"text_generation",
		},
	}

	// Spawn the drone using coordinator
	droneID, err := s.coordinator.SpawnDrone(ctx, droneConfig)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to spawn drone: %v", err)), nil
	}

	result := fmt.Sprintf("Successfully spawned drone %s of type %s in region %s", droneID, droneType, region)
	return mcp.NewToolResultText(result), nil
}

// handleListDrones handles the list_active_drones tool call
func (s *MCPServer) handleListDrones(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	drones := s.coordinator.ListActiveDrones()

	if len(drones) == 0 {
		return mcp.NewToolResultText("No active drones found"), nil
	}

	result := "Active Drones:\n"
	for _, drone := range drones {
		result += fmt.Sprintf("- ID: %s, Type: %s, Status: %s, Region: %s\n",
			drone.ID, drone.Type, drone.Status, drone.Region)
	}

	return mcp.NewToolResultText(result), nil
}

// handleExecuteTask handles the execute_distributed_task tool call
func (s *MCPServer) handleExecuteTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskType, err := request.RequireString("task_type")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid task_type: %v", err)), nil
	}

	description, err := request.RequireString("description")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid description: %v", err)), nil
	}

	maxDrones := int(request.GetFloat("max_drones", 3))

	log.Printf("Executing distributed task: type=%s, maxDrones=%d", taskType, maxDrones)

	// Create task configuration
	task := types.Task{
		Type:        taskType,
		Description: description,
		MaxDrones:   maxDrones,
	}

	// Execute the task using coordinator
	taskID, err := s.coordinator.ExecuteTask(ctx, task)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to execute task: %v", err)), nil
	}

	result := fmt.Sprintf("Successfully started task %s of type %s using up to %d drones", taskID, taskType, maxDrones)
	return mcp.NewToolResultText(result), nil
}

// handleGetDroneStatus handles the get_drone_status tool call
func (s *MCPServer) handleGetDroneStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	droneID, err := request.RequireString("drone_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid drone_id: %v", err)), nil
	}

	drone, err := s.coordinator.GetDroneStatus(ctx, droneID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get drone status: %v", err)), nil
	}

	result := fmt.Sprintf("Drone Status:\n"+
		"ID: %s\n"+
		"Type: %s\n"+
		"Status: %s\n"+
		"Region: %s\n"+
		"Created: %s\n"+
		"Last Seen: %s\n"+
		"Tasks Completed: %d",
		drone.ID, drone.Type, drone.Status, drone.Region,
		drone.CreatedAt.Format("2006-01-02 15:04:05"),
		drone.LastSeen.Format("2006-01-02 15:04:05"),
		drone.TasksCompleted)

	return mcp.NewToolResultText(result), nil
}

// handleTerminateDrone handles the terminate_drone tool call
func (s *MCPServer) handleTerminateDrone(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	droneID, err := request.RequireString("drone_id")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid drone_id: %v", err)), nil
	}

	err = s.coordinator.TerminateDrone(ctx, droneID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to terminate drone: %v", err)), nil
	}

	result := fmt.Sprintf("Successfully terminated drone %s", droneID)
	return mcp.NewToolResultText(result), nil
}

// Start starts the MCP server using stdio transport
func (s *MCPServer) Start(ctx context.Context) error {
	log.Println("Starting MCP server...")
	return server.ServeStdio(s.mcpServer)
}

// Close closes the MCP server
func (s *MCPServer) Close() error {
	// The mcp-go library doesn't expose a Close method, so we just log
	log.Println("MCP server stopped")
	return nil
}
