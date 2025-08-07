package operations

import (
	"context"
	"fmt"

	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/orchestrator"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/schemas"
)

// SequentialThinking implements sequential thinking style reasoning
type SequentialThinking struct {
	claudeAgent *orchestrator.ClaudeAgent
}

// NewSequentialThinking creates a new sequential thinking operation
func NewSequentialThinking() *SequentialThinking {
	return &SequentialThinking{
		claudeAgent: orchestrator.NewClaudeAgent(),
	}
}

// Execute performs sequential thinking analysis
func (st *SequentialThinking) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	problem, ok := params["problem"].(string)
	if !ok {
		return nil, fmt.Errorf("problem parameter is required")
	}

	contextStr := ""
	if c, ok := params["context"].(string); ok {
		contextStr = c
	}

	// Parse additional parameters
	var steps []string
	if s, ok := params["steps"].([]interface{}); ok {
		for _, step := range s {
			if str, ok := step.(string); ok {
				steps = append(steps, str)
			}
		}
	}

	maxSteps := 10
	if ms, ok := params["max_steps"].(float64); ok {
		maxSteps = int(ms)
	}

	// Create request
	request := &schemas.SequentialThinkingRequest{
		Problem:  problem,
		Context:  contextStr,
		Steps:    steps,
		MaxSteps: maxSteps,
	}

	// Perform sequential thinking using Claude agent
	response, err := st.claudeAgent.AnalyzeSequentialThinking(ctx, request.Problem, request.Context)
	if err != nil {
		return nil, fmt.Errorf("sequential thinking failed: %w", err)
	}

	return response, nil
}

// GetDescription returns the operation description
func (st *SequentialThinking) GetDescription() string {
	return "Performs sequential thinking style reasoning to break down complex problems into logical steps"
}