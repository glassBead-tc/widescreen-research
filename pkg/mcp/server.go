// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/glassBead-tc/widescreen-research/pkg/coordinator"
	"github.com/glassBead-tc/widescreen-research/pkg/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPServerOfficial wraps the coordinator with MCP protocol support using official SDK
type MCPServerOfficial struct {
	coordinator *coordinator.Server
	mcpServer   *mcp.Server
}

// NewMCPServerOfficial creates a new MCP server using official SDK
func NewMCPServerOfficial(coord *coordinator.Server) *MCPServerOfficial {
	s := &MCPServerOfficial{
		coordinator: coord,
	}

	// Create MCP server with official SDK
	s.mcpServer = mcp.NewServer(&mcp.Implementation{
		Name:    "Spawn MCP Coordinator",
		Version: "1.0.0",
	}, nil)

	// Register all tools
	s.registerTools()

	return s
}

// Tool argument structs

type SpawnDroneArgs struct {
	DroneType string `json:"drone_type" jsonschema:"Type of drone to spawn"`
	Region    string `json:"region" jsonschema:"GCP region to deploy to"`
}

type ExecuteTaskArgs struct {
	TaskType    string  `json:"task_type" jsonschema:"Type of task to execute"`
	Description string  `json:"description" jsonschema:"Detailed description of the task"`
	MaxDrones   float64 `json:"max_drones" jsonschema:"Maximum number of drones to use"`
}

type DroneIDArgs struct {
	DroneID string `json:"drone_id" jsonschema:"ID of the drone"`
}

type PlanCampaignArgs struct {
	SpecJSON string `json:"spec_json" jsonschema:"JSON-encoded CampaignSpec"`
}

type LaunchFleetArgs struct {
	RunID         string  `json:"run_id" jsonschema:"Campaign run ID"`
	TargetWorkers float64 `json:"target_workers" jsonschema:"Target number of workers"`
}

type RunIDArgs struct {
	RunID string `json:"run_id" jsonschema:"Campaign run ID"`
}

type ExportGraphArgs struct {
	Mem0Space string `json:"mem0_space" jsonschema:"mem0 space identifier"`
	Format    string `json:"format" jsonschema:"Export format (jsonl or csv)"`
}

// registerTools registers all available MCP tools
func (s *MCPServerOfficial) registerTools() {
	// Tool: Spawn Drone Server
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "spawn_drone_server",
		Description: "Spawn a new drone MCP server on Cloud Run",
	}, s.handleSpawnDrone)

	// Tool: List Active Drones
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_active_drones",
		Description: "List all currently active drone servers",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
		drones := s.coordinator.ListActiveDrones()

		if len(drones) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "No active drones found"},
				},
			}, nil, nil
		}

		result := "Active Drones:\n"
		for _, drone := range drones {
			result += fmt.Sprintf("- ID: %s, Type: %s, Status: %s, Region: %s\n",
				drone.ID, drone.Type, drone.Status, drone.Region)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: result},
			},
		}, nil, nil
	})

	// Tool: Execute Distributed Task
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "execute_distributed_task",
		Description: "Execute a task across the drone fleet",
	}, s.handleExecuteTask)

	// Tool: Get Drone Status
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_drone_status",
		Description: "Get detailed status of a specific drone",
	}, s.handleGetDroneStatus)

	// Tool: Terminate Drone
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "terminate_drone",
		Description: "Terminate a specific drone server",
	}, s.handleTerminateDrone)

	// Campaign orchestration tools
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "plan_campaign",
		Description: "Validate a campaign spec and produce an execution plan",
	}, s.handlePlanCampaign)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "launch_fleet",
		Description: "Provision worker fleet and seed queue for a campaign run",
	}, s.handleLaunchFleet)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "fleet_status",
		Description: "Get current status and progress for a campaign run",
	}, s.handleFleetStatus)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "abort",
		Description: "Abort a campaign run and scale down workers",
	}, s.handleAbort)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "export_graph",
		Description: "Export collected graph for a mem0 space or run",
	}, s.handleExportGraph)
}

// Tool handlers

func (s *MCPServerOfficial) handleSpawnDrone(ctx context.Context, req *mcp.CallToolRequest, args SpawnDroneArgs) (*mcp.CallToolResult, any, error) {
	// Set default region
	if args.Region == "" {
		args.Region = "us-central1"
	}

	log.Printf("Spawning drone: type=%s, region=%s", args.DroneType, args.Region)

	// Create drone configuration
	droneConfig := types.DroneConfig{
		Type:   types.DroneType(args.DroneType),
		Region: args.Region,
		Capabilities: []string{
			"web_search",
			"data_analysis",
			"text_generation",
		},
	}

	// Spawn the drone using coordinator
	droneID, err := s.coordinator.SpawnDrone(ctx, droneConfig)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to spawn drone: %v", err)},
			},
		}, nil, nil
	}

	result := fmt.Sprintf("Successfully spawned drone %s of type %s in region %s", droneID, args.DroneType, args.Region)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}

func (s *MCPServerOfficial) handleExecuteTask(ctx context.Context, req *mcp.CallToolRequest, args ExecuteTaskArgs) (*mcp.CallToolResult, any, error) {
	// Set default max drones
	if args.MaxDrones == 0 {
		args.MaxDrones = 3
	}

	log.Printf("Executing distributed task: type=%s, maxDrones=%d", args.TaskType, int(args.MaxDrones))

	// Create task configuration
	task := types.Task{
		Type:        args.TaskType,
		Description: args.Description,
		MaxDrones:   int(args.MaxDrones),
	}

	// Execute the task using coordinator
	taskID, err := s.coordinator.ExecuteTask(ctx, task)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to execute task: %v", err)},
			},
		}, nil, nil
	}

	result := fmt.Sprintf("Successfully started task %s of type %s using up to %d drones", taskID, args.TaskType, int(args.MaxDrones))
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}

func (s *MCPServerOfficial) handleGetDroneStatus(ctx context.Context, req *mcp.CallToolRequest, args DroneIDArgs) (*mcp.CallToolResult, any, error) {
	drone, err := s.coordinator.GetDroneStatus(ctx, args.DroneID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get drone status: %v", err)},
			},
		}, nil, nil
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

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}

func (s *MCPServerOfficial) handleTerminateDrone(ctx context.Context, req *mcp.CallToolRequest, args DroneIDArgs) (*mcp.CallToolResult, any, error) {
	err := s.coordinator.TerminateDrone(ctx, args.DroneID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to terminate drone: %v", err)},
			},
		}, nil, nil
	}

	result := fmt.Sprintf("Successfully terminated drone %s", args.DroneID)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}

func (s *MCPServerOfficial) handlePlanCampaign(ctx context.Context, req *mcp.CallToolRequest, args PlanCampaignArgs) (*mcp.CallToolResult, any, error) {
	var spec types.CampaignSpec
	if err := json.Unmarshal([]byte(args.SpecJSON), &spec); err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("invalid spec_json: %v", err)},
			},
		}, nil, nil
	}

	plan, err := s.coordinator.PlanCampaign(ctx, spec)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: err.Error()},
			},
		}, nil, nil
	}

	resBytes, _ := json.Marshal(plan)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resBytes)},
		},
	}, nil, nil
}

func (s *MCPServerOfficial) handleLaunchFleet(ctx context.Context, req *mcp.CallToolRequest, args LaunchFleetArgs) (*mcp.CallToolResult, any, error) {
	// Set default target workers
	if args.TargetWorkers == 0 {
		args.TargetWorkers = 10
	}

	statusID, err := s.coordinator.LaunchFleet(ctx, args.RunID, int(args.TargetWorkers))
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: err.Error()},
			},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: statusID},
		},
	}, nil, nil
}

func (s *MCPServerOfficial) handleFleetStatus(ctx context.Context, req *mcp.CallToolRequest, args RunIDArgs) (*mcp.CallToolResult, any, error) {
	status, err := s.coordinator.FleetStatus(ctx, args.RunID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: err.Error()},
			},
		}, nil, nil
	}

	b, _ := json.Marshal(status)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(b)},
		},
	}, nil, nil
}

func (s *MCPServerOfficial) handleAbort(ctx context.Context, req *mcp.CallToolRequest, args RunIDArgs) (*mcp.CallToolResult, any, error) {
	if err := s.coordinator.AbortRun(ctx, args.RunID); err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: err.Error()},
			},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "aborted"},
		},
	}, nil, nil
}

func (s *MCPServerOfficial) handleExportGraph(ctx context.Context, req *mcp.CallToolRequest, args ExportGraphArgs) (*mcp.CallToolResult, any, error) {
	// Set default format
	if args.Format == "" {
		args.Format = "jsonl"
	}

	uri, err := s.coordinator.ExportGraph(ctx, args.Mem0Space, args.Format)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: err.Error()},
			},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: uri},
		},
	}, nil, nil
}

// Run starts the MCP server with stdio transport
func (s *MCPServerOfficial) Run(ctx context.Context) error {
	log.Println("Starting MCP coordinator server (Official SDK)...")
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}

// Close closes the MCP server
func (s *MCPServerOfficial) Close() error {
	log.Println("MCP coordinator server stopped")
	return nil
}