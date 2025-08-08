package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spawn-mcp/coordinator/pkg/types"
)

func (s *MCPServer) handlePlanCampaign(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	specJSON, err := request.RequireString("spec_json")
	if err != nil {
		return mcp.NewToolResultError("spec_json required"), nil
	}
	var spec types.CampaignSpec
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid spec_json: %v", err)), nil
	}
	plan, err := s.coordinator.PlanCampaign(ctx, spec)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resBytes, _ := json.Marshal(plan)
	return mcp.NewToolResultText(string(resBytes)), nil
}

func (s *MCPServer) handleLaunchFleet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	runID, err := request.RequireString("run_id")
	if err != nil { return mcp.NewToolResultError("run_id required"), nil }
	tw := int(request.GetFloat("target_workers", 10))
	statusID, err := s.coordinator.LaunchFleet(ctx, runID, tw)
	if err != nil { return mcp.NewToolResultError(err.Error()), nil }
	return mcp.NewToolResultText(statusID), nil
}

func (s *MCPServer) handleFleetStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	runID, err := request.RequireString("run_id")
	if err != nil { return mcp.NewToolResultError("run_id required"), nil }
	status, err := s.coordinator.FleetStatus(ctx, runID)
	if err != nil { return mcp.NewToolResultError(err.Error()), nil }
	b, _ := json.Marshal(status)
	return mcp.NewToolResultText(string(b)), nil
}

func (s *MCPServer) handleAbort(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	runID, err := request.RequireString("run_id")
	if err != nil { return mcp.NewToolResultError("run_id required"), nil }
	if err := s.coordinator.AbortRun(ctx, runID); err != nil { return mcp.NewToolResultError(err.Error()), nil }
	return mcp.NewToolResultText("aborted"), nil
}

func (s *MCPServer) handleExportGraph(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	space, err := request.RequireString("mem0_space")
	if err != nil { return mcp.NewToolResultError("mem0_space required"), nil }
	format := request.GetString("format", "jsonl")
	uri, err := s.coordinator.ExportGraph(ctx, space, format)
	if err != nil { return mcp.NewToolResultError(err.Error()), nil }
	return mcp.NewToolResultText(uri), nil
}