package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/mark3labs/mcp-go/client"
)

// MCPClient manages connections to other MCP servers
type MCPClient struct {
	clients map[string]*client.Client
	mu      sync.RWMutex
}

// NewMCPClient creates a new MCP client manager
func NewMCPClient() *MCPClient {
	return &MCPClient{
		clients: make(map[string]*client.Client),
	}
}

// Initialize initializes the MCP client connections
func (c *MCPClient) Initialize(ctx context.Context) error {
	// Load MCP server configurations from environment or config
	servers := c.loadServerConfigs()

	for name, config := range servers {
		client, err := c.connectToServer(ctx, config)
		if err != nil {
			log.Printf("Failed to connect to MCP server %s: %v", name, err)
			continue
		}

		c.mu.Lock()
		c.clients[name] = client
		c.mu.Unlock()

		log.Printf("Connected to MCP server: %s", name)
	}

	return nil
}

// connectToServer connects to a single MCP server
func (c *MCPClient) connectToServer(ctx context.Context, config ServerConfig) (*client.Client, error) {
	// Create client with appropriate transport
	cl := client.NewClient(
		config.Name,
		config.Version,
		client.WithTransport(config.Transport),
	)

	// Connect to the server
	if err := cl.Connect(ctx, config.URL); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Initialize the client
	if err := cl.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize: %w", err)
	}

	return cl, nil
}

// CallTool calls a tool on a specific MCP server
func (c *MCPClient) CallTool(ctx context.Context, serverName string, toolName string, arguments interface{}) (interface{}, error) {
	c.mu.RLock()
	cl, exists := c.clients[serverName]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("MCP server %s not found", serverName)
	}

	// Call the tool
	result, err := cl.CallTool(ctx, toolName, arguments)
	if err != nil {
		return nil, fmt.Errorf("tool call failed: %w", err)
	}

	return result, nil
}

// GetResource gets a resource from a specific MCP server
func (c *MCPClient) GetResource(ctx context.Context, serverName string, uri string) (interface{}, error) {
	c.mu.RLock()
	cl, exists := c.clients[serverName]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("MCP server %s not found", serverName)
	}

	// Get the resource
	result, err := cl.GetResource(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("resource fetch failed: %w", err)
	}

	return result, nil
}

// ListTools lists all available tools from all connected servers
func (c *MCPClient) ListTools(ctx context.Context) map[string][]ToolInfo {
	tools := make(map[string][]ToolInfo)

	c.mu.RLock()
	defer c.mu.RUnlock()

	for serverName, cl := range c.clients {
		serverTools, err := cl.ListTools(ctx)
		if err != nil {
			log.Printf("Failed to list tools from %s: %v", serverName, err)
			continue
		}

		toolInfos := make([]ToolInfo, 0, len(serverTools))
		for _, tool := range serverTools {
			toolInfos = append(toolInfos, ToolInfo{
				Name:        tool.Name,
				Description: tool.Description,
				InputSchema: tool.InputSchema,
			})
		}

		tools[serverName] = toolInfos
	}

	return tools
}

// CallExternalResearch calls external research capabilities
func (c *MCPClient) CallExternalResearch(ctx context.Context, topic string, parameters map[string]interface{}) ([]map[string]interface{}, error) {
	// Try different research servers
	researchServers := []string{"exa-research", "web-research", "academic-research"}
	
	var allResults []map[string]interface{}
	
	for _, serverName := range researchServers {
		c.mu.RLock()
		_, exists := c.clients[serverName]
		c.mu.RUnlock()
		
		if !exists {
			continue
		}
		
		// Call research tool
		result, err := c.CallTool(ctx, serverName, "research", map[string]interface{}{
			"query":       topic,
			"parameters":  parameters,
			"numResults":  10,
		})
		
		if err != nil {
			log.Printf("Research call to %s failed: %v", serverName, err)
			continue
		}
		
		// Parse results
		if resultMap, ok := result.(map[string]interface{}); ok {
			allResults = append(allResults, resultMap)
		}
	}
	
	return allResults, nil
}

// Shutdown closes all MCP client connections
func (c *MCPClient) Shutdown() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for name, cl := range c.clients {
		if err := cl.Close(); err != nil {
			log.Printf("Error closing MCP client %s: %v", name, err)
		}
	}

	c.clients = make(map[string]*client.Client)
}

// ServerConfig represents MCP server configuration
type ServerConfig struct {
	Name      string
	Version   string
	URL       string
	Transport string
}

// ToolInfo represents information about a tool
type ToolInfo struct {
	Name        string
	Description string
	InputSchema interface{}
}

// loadServerConfigs loads MCP server configurations
func (c *MCPClient) loadServerConfigs() map[string]ServerConfig {
	// Load from environment or configuration file
	// For now, return a static configuration
	configs := make(map[string]ServerConfig)

	// Check for Exa research server
	if exaURL := getEnvOrDefault("EXA_MCP_URL", ""); exaURL != "" {
		configs["exa-research"] = ServerConfig{
			Name:      "exa-research",
			Version:   "1.0.0",
			URL:       exaURL,
			Transport: "stdio",
		}
	}

	// Add other research servers as needed
	if webURL := getEnvOrDefault("WEB_RESEARCH_MCP_URL", ""); webURL != "" {
		configs["web-research"] = ServerConfig{
			Name:      "web-research",
			Version:   "1.0.0",
			URL:       webURL,
			Transport: "stdio",
		}
	}

	return configs
}