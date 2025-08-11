package tools

import (
	"context"
	"fmt"

	mcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/resources"
)

// GuideToolHandler handles guide requests
type GuideToolHandler struct {
	guides *resources.GuideResource
}

// NewGuideToolHandler creates a new guide tool handler
func NewGuideToolHandler(guides *resources.GuideResource) *GuideToolHandler {
	return &GuideToolHandler{
		guides: guides,
	}
}

// GetDefinition returns the tool definition for the guide tool
func (h *GuideToolHandler) GetDefinition() *ToolDefinition {
	return &ToolDefinition{
		Name:        "get_guide",
		Description: "Get research system guides and documentation. Use 'list' as name to see all available guides.",
		Parameters: []mcp.ToolOption{
			mcp.WithString("name", mcp.Description("Guide name: 'main', 'websets', 'orchestration', 'quickstart', or 'list' to see all")),
		},
		Handler: h.Handle,
	}
}

// Handle processes guide requests
func (h *GuideToolHandler) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	guideName := req.GetString("name", "main")
	
	// Special case: list available guides
	if guideName == "list" {
		guides := h.guides.ListGuides()
		result := "Available guides:\n"
		for _, name := range guides {
			result += fmt.Sprintf("- %s\n", name)
		}
		result += "\nUse get_guide with the guide name to read it."
		return mcp.NewToolResultText(result), nil
	}
	
	guide, err := h.guides.GetGuide(guideName)
	if err != nil {
		availableGuides := h.guides.ListGuides()
		return mcp.NewToolResultError(fmt.Sprintf("Guide '%s' not found. Available guides: %v", guideName, availableGuides)), nil
	}
	
	return mcp.NewToolResultText(guide), nil
}