// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	log.Println("Starting Simple Spawn MCP Server (Official SDK)...")

	// Create a new MCP server with the official SDK
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "Spawn MCP Coordinator",
		Version: "1.0.0",
	}, nil)

	// Add drone management tools
	addDroneTools(server)

	// Start the server using stdio transport
	log.Println("Starting MCP server on stdio...")
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func addDroneTools(server *mcp.Server) {
	// Tool: Spawn Drone Server
	type SpawnDroneArgs struct {
		DroneType string `json:"drone_type" jsonschema:"required,enum=researcher|analyst|writer|coder,description=Type of drone to spawn"`
		Region    string `json:"region" jsonschema:"description=GCP region to deploy to,default=us-central1"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "spawn_drone_server",
		Description: "Spawn a new drone MCP server on Cloud Run",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SpawnDroneArgs) (*mcp.CallToolResult, any, error) {
		// Set default region
		if args.Region == "" {
			args.Region = "us-central1"
		}

		droneCounter++
		droneID := fmt.Sprintf("drone-%s-%d", args.DroneType, droneCounter)

		// Store drone info
		activeDrones[droneID] = map[string]interface{}{
			"id":     droneID,
			"type":   args.DroneType,
			"region": args.Region,
			"status": "active",
		}

		log.Printf("Spawned drone: %s (type: %s, region: %s)", droneID, args.DroneType, args.Region)

		result := fmt.Sprintf("✅ Successfully spawned drone %s\n"+
			"Type: %s\n"+
			"Region: %s\n"+
			"Status: Active\n"+
			"\nThe drone is now ready to accept tasks!", droneID, args.DroneType, args.Region)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: result},
			},
		}, nil, nil
	})

	// Tool: List Active Drones
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_active_drones",
		Description: "List all currently active drone servers",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
		if len(activeDrones) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "No active drones found. Use spawn_drone_server to create some!"},
				},
			}, nil, nil
		}

		result := "🤖 Active Drones:\n\n"
		for _, drone := range activeDrones {
			result += fmt.Sprintf("• ID: %s\n  Type: %s\n  Region: %s\n  Status: %s\n\n",
				drone["id"], drone["type"], drone["region"], drone["status"])
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: result},
			},
		}, nil, nil
	})

	// Tool: Execute Distributed Task
	type ExecuteTaskArgs struct {
		TaskType    string  `json:"task_type" jsonschema:"required,enum=research|analysis|synthesis|coding,description=Type of task to execute"`
		Description string  `json:"description" jsonschema:"required,description=Detailed description of the task"`
		MaxDrones   float64 `json:"max_drones" jsonschema:"description=Maximum number of drones to use,default=3,minimum=1,maximum=10"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "execute_distributed_task",
		Description: "Execute a task across the drone fleet",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ExecuteTaskArgs) (*mcp.CallToolResult, any, error) {
		// Set default max drones
		if args.MaxDrones == 0 {
			args.MaxDrones = 3
		}

		if len(activeDrones) == 0 {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "No active drones available. Please spawn some drones first using spawn_drone_server."},
				},
			}, nil, nil
		}

		// Simulate task execution
		taskID := fmt.Sprintf("task-%s-%d", args.TaskType, droneCounter)
		droneCounter++

		availableDrones := len(activeDrones)
		dronesUsed := int(args.MaxDrones)
		if availableDrones < dronesUsed {
			dronesUsed = availableDrones
		}

		log.Printf("Executing task %s: %s (using %d drones)", taskID, args.Description, dronesUsed)

		result := fmt.Sprintf("🚀 Task Execution Started!\n\n"+
			"Task ID: %s\n"+
			"Type: %s\n"+
			"Description: %s\n"+
			"Drones Assigned: %d/%d\n"+
			"Status: In Progress\n\n"+
			"The task has been distributed across the drone fleet and is now executing...",
			taskID, args.TaskType, args.Description, dronesUsed, int(args.MaxDrones))

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: result},
			},
		}, nil, nil
	})

	// Tool: Get System Status
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_system_status",
		Description: "Get overall system status and metrics",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
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

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: result},
			},
		}, nil, nil
	})
}

// Simple in-memory drone tracking
var activeDrones = make(map[string]map[string]interface{})
var droneCounter = 0
