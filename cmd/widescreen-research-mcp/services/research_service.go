package services

import (
	"context"
	"fmt"

	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/operations"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/orchestrator"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/schemas"
)

// ResearchService handles research operations
type ResearchService struct {
	orchestrator *orchestrator.Orchestrator
	operations   *operations.OperationRegistry
}

// NewResearchService creates a new research service
func NewResearchService(orch *orchestrator.Orchestrator, ops *operations.OperationRegistry) *ResearchService {
	return &ResearchService{
		orchestrator: orch,
		operations:   ops,
	}
}

// ExecuteOperation executes a research operation
func (s *ResearchService) ExecuteOperation(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	// Look up operation handler
	op := s.operations.GetOperation(input.Operation)
	if op == nil {
		// Handle special websets operations
		if input.Operation == "websets-orchestrate" {
			return s.handleWebsetsOrchestrate(ctx, input)
		}
		if input.Operation == "websets-call" {
			return s.handleWebsetsCall(ctx, input)
		}
		return nil, fmt.Errorf("unknown operation: %s", input.Operation)
	}

	// Execute operation handler
	return op.Handler(ctx, input.Parameters)
}

// handleWebsetsOrchestrate handles websets orchestration
func (s *ResearchService) handleWebsetsOrchestrate(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	// For now, just return a placeholder
	// This would be implemented with proper websets integration
	return map[string]interface{}{
		"status": "websets-orchestrate not yet implemented",
		"parameters": input.Parameters,
	}, nil
}

// handleWebsetsCall handles direct websets API calls  
func (s *ResearchService) handleWebsetsCall(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	// For now, just return a placeholder
	// This would be implemented with proper websets integration
	return map[string]interface{}{
		"status": "websets-call not yet implemented",
		"parameters": input.Parameters,
	}, nil
}