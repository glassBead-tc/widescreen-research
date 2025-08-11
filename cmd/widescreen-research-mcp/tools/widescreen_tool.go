package tools

import (
	"context"
	"encoding/json"

	mcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/schemas"
)

// WidescreenToolHandler handles the main widescreen research tool
type WidescreenToolHandler struct {
	executeFunc func(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error)
}

// NewWidescreenToolHandler creates a new widescreen tool handler
func NewWidescreenToolHandler(executeFunc func(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error)) *WidescreenToolHandler {
	return &WidescreenToolHandler{
		executeFunc: executeFunc,
	}
}

// GetDefinition returns the tool definition
func (h *WidescreenToolHandler) GetDefinition() *ToolDefinition {
	return &ToolDefinition{
		Name:        "widescreen_research",
		Description: "Perform comprehensive widescreen research using distributed research drones",
		Parameters: []mcp.ToolOption{
			mcp.WithString("operation", mcp.Description("Operation to execute")),
			mcp.WithString("session_id", mcp.Description("Session ID for elicitation and orchestration")),
			mcp.WithString("parameters_json", mcp.Description("JSON-encoded parameters for the operation")),
			mcp.WithString("elicitation_answers_json", mcp.Description("JSON-encoded elicitation answers")),
		},
		Handler: h.Handle,
	}
}

// Handle processes widescreen research requests
func (h *WidescreenToolHandler) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Build input from tool request
	op := req.GetString("operation", "")
	sessionID := req.GetString("session_id", "")

	params := map[string]interface{}{}
	if pstr := req.GetString("parameters_json", ""); pstr != "" {
		_ = json.Unmarshal([]byte(pstr), &params)
	}

	elicit := map[string]interface{}{}
	if estr := req.GetString("elicitation_answers_json", ""); estr != "" {
		_ = json.Unmarshal([]byte(estr), &elicit)
	}

	input := &schemas.WidescreenResearchInput{
		Operation:          op,
		SessionID:          sessionID,
		ElicitationAnswers: elicit,
		Parameters:         params,
	}

	result, err := h.executeFunc(ctx, input)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Return JSON-encoded result as text
	b, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(b)), nil
}