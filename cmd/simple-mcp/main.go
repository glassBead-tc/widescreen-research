package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	log.Println("Starting Simple Spawn MCP Server...")

	// Create a new MCP server
	s := server.NewMCPServer(
		"Spawn MCP Coordinator",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	// Add drone management tools
	addDroneTools(s)

	// Start the server using stdio transport
	log.Println("Starting MCP server on stdio...")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func addDroneTools(s *server.MCPServer) {
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

	s.AddTool(spawnDroneTool, handleSpawnDrone)

	// Tool: List Active Drones
	listDronesTool := mcp.NewTool("list_active_drones",
		mcp.WithDescription("List all currently active drone servers"),
	)

	s.AddTool(listDronesTool, handleListDrones)

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

	s.AddTool(executeTaskTool, handleExecuteTask)

	// Tool: Get System Status
	statusTool := mcp.NewTool("get_system_status",
		mcp.WithDescription("Get overall system status and metrics"),
	)

	s.AddTool(statusTool, handleGetStatus)
}

// Simple in-memory drone tracking
var activeDrones = make(map[string]map[string]interface{})
var droneCounter = 0

func handleSpawnDrone(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	droneType, err := request.RequireString("drone_type")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid drone_type: %v", err)), nil
	}

	region := request.GetString("region", "us-central1")

	droneCounter++
	droneID := fmt.Sprintf("drone-%s-%d", droneType, droneCounter)

	// Store drone info
	activeDrones[droneID] = map[string]interface{}{
		"id":     droneID,
		"type":   droneType,
		"region": region,
		"status": "active",
	}

	log.Printf("Spawned drone: %s (type: %s, region: %s)", droneID, droneType, region)

	result := fmt.Sprintf("✅ Successfully spawned drone %s\n"+
		"Type: %s\n"+
		"Region: %s\n"+
		"Status: Active\n"+
		"\nThe drone is now ready to accept tasks!", droneID, droneType, region)

	return mcp.NewToolResultText(result), nil
}

func handleListDrones(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if len(activeDrones) == 0 {
		return mcp.NewToolResultText("No active drones found. Use spawn_drone_server to create some!"), nil
	}

	result := "🤖 Active Drones:\n\n"
	for _, drone := range activeDrones {
		result += fmt.Sprintf("• ID: %s\n  Type: %s\n  Region: %s\n  Status: %s\n\n",
			drone["id"], drone["type"], drone["region"], drone["status"])
	}

	return mcp.NewToolResultText(result), nil
}

func handleExecuteTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskType, err := request.RequireString("task_type")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid task_type: %v", err)), nil
	}

	description, err := request.RequireString("description")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid description: %v", err)), nil
	}

	maxDrones := int(request.GetFloat("max_drones", 3))

	if len(activeDrones) == 0 {
		return mcp.NewToolResultError("No active drones available. Please spawn some drones first using spawn_drone_server."), nil
	}

	// Simulate task execution
	taskID := fmt.Sprintf("task-%s-%d", taskType, droneCounter)
	droneCounter++

	availableDrones := len(activeDrones)
	dronesUsed := maxDrones
	if availableDrones < maxDrones {
		dronesUsed = availableDrones
	}

	log.Printf("Executing task %s: %s (using %d drones)", taskID, description, dronesUsed)

	result := fmt.Sprintf("🚀 Task Execution Started!\n\n"+
		"Task ID: %s\n"+
		"Type: %s\n"+
		"Description: %s\n"+
		"Drones Assigned: %d/%d\n"+
		"Status: In Progress\n\n"+
		"The task has been distributed across the drone fleet and is now executing...",
		taskID, taskType, description, dronesUsed, maxDrones)

	return mcp.NewToolResultText(result), nil
}

func handleGetStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result := fmt.Sprintf("📊 Spawn MCP System Status\n\n"+
		"Active Drones: %d\n"+
		"Total Spawned: %d\n"+
		"System Status: Operational\n"+
		"MCP Protocol: v1.0.0\n\n"+
		"Available Commands:\n"+
		"• spawn_drone_server - Create new drones\n"+
		"• list_active_drones - View active drones\n"+
		"• execute_distributed_task - Run tasks across fleet\n"+
		"• get_system_status - View this status",
		len(activeDrones), droneCounter)

	return mcp.NewToolResultText(result), nil
}
