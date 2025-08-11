package tools

import (
	"context"
	"fmt"

	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// ToolHandler is the function signature for tool handlers
type ToolHandler = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error)

// ToolDefinition contains tool metadata and handler
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  []mcp.ToolOption
	Handler     ToolHandler
}

// Registry manages tool registration and discovery
type Registry struct {
	tools map[string]*ToolDefinition
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]*ToolDefinition),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(def *ToolDefinition) error {
	if _, exists := r.tools[def.Name]; exists {
		return fmt.Errorf("tool already registered: %s", def.Name)
	}
	r.tools[def.Name] = def
	return nil
}

// RegisterWithServer registers all tools with an MCP server
func (r *Registry) RegisterWithServer(server *mcpserver.MCPServer) {
	for _, def := range r.tools {
		// Create tool with options
		opts := append([]mcp.ToolOption{
			mcp.WithDescription(def.Description),
		}, def.Parameters...)
		
		tool := mcp.NewTool(def.Name, opts...)
		server.AddTool(tool, def.Handler)
	}
}

// Get retrieves a tool definition by name
func (r *Registry) Get(name string) (*ToolDefinition, bool) {
	def, exists := r.tools[name]
	return def, exists
}

// List returns all registered tool names
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}