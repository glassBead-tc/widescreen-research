package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// WebsetsClient defines the interface for interacting with the EXA Websets MCP server
type WebsetsClient interface {
	Connect(ctx context.Context) error
	Call(ctx context.Context, arguments map[string]any) (string, error)
	Close() error
}

// StdIOWebsetsClient implements WebsetsClient using stdio subprocess communication
type StdIOWebsetsClient struct {
	mu       sync.Mutex
	client   *sdk.Client
	session  *sdk.ClientSession
	cmd      *exec.Cmd
	started  bool
	binPath  string   // e.g., "exa-websets-mcp-server" or "node"
	binArgs  []string // e.g., ["./build/index.js"]
	env      []string // must include EXA_API_KEY
}

// NewStdIOWebsetsClient creates a new stdio-based websets client
func NewStdIOWebsetsClient(binPath string, binArgs []string) *StdIOWebsetsClient {
	if binPath == "" {
		binPath = "exa-websets-mcp-server" // default binary
	}
	return &StdIOWebsetsClient{
		binPath: binPath,
		binArgs: binArgs,
	}
}

// Connect establishes connection to the EXA Websets MCP server subprocess
func (c *StdIOWebsetsClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.started {
		return nil
	}

	// Ensure EXA_API_KEY is present
	if os.Getenv("EXA_API_KEY") == "" {
		return errors.New("EXA_API_KEY not set in environment")
	}

	// Prepare command
	cmd := exec.CommandContext(ctx, c.binPath, c.binArgs...)
	cmd.Env = os.Environ() // inherits EXA_API_KEY

	// Create MCP client with implementation details
	client := sdk.NewClient(
		&sdk.Implementation{
			Name:    "widescreen-client",
			Version: "v1.0.0",
		},
		&sdk.ClientOptions{},
	)

	// Create command transport for stdio communication
	transport := sdk.NewCommandTransport(cmd)

	// Connect to the subprocess MCP server
	session, err := client.Connect(ctx, transport)
	if err != nil {
		return fmt.Errorf("mcp connect failed: %w", err)
	}

	// Store connection details
	c.client = client
	c.session = session
	c.cmd = cmd
	c.started = true
	
	log.Printf("Connected to EXA Websets MCP server via %s", c.binPath)
	return nil
}

// Call invokes the unified websets_manager tool with the given arguments
func (c *StdIOWebsetsClient) Call(ctx context.Context, arguments map[string]any) (string, error) {
	c.mu.Lock()
	session := c.session
	c.mu.Unlock()

	// Auto-connect if not connected
	if session == nil {
		if err := c.Connect(ctx); err != nil {
			return "", fmt.Errorf("failed to connect: %w", err)
		}
		c.mu.Lock()
		session = c.session
		c.mu.Unlock()
	}

	// Prepare tool call parameters
	params := &sdk.CallToolParams{
		Name:      "websets_manager",
		Arguments: arguments,
	}

	// Execute tool call
	res, err := session.CallTool(ctx, params)
	if err != nil {
		// Check if it's a transport error, attempt one reconnect
		if isTransportError(err) {
			log.Println("Transport error detected, attempting reconnect...")
			c.mu.Lock()
			c.started = false
			if c.session != nil {
				_ = c.session.Close()
			}
			c.session = nil
			c.mu.Unlock()

			// Try reconnecting once
			if err := c.Connect(ctx); err != nil {
				return "", fmt.Errorf("reconnect failed: %w", err)
			}

			// Retry the call
			c.mu.Lock()
			session = c.session
			c.mu.Unlock()
			
			res, err = session.CallTool(ctx, params)
			if err != nil {
				return "", fmt.Errorf("tools/call failed after reconnect: %w", err)
			}
		} else {
			return "", fmt.Errorf("tools/call failed: %w", err)
		}
	}

	// Handle error response
	if res.IsError {
		// Extract first text content if present
		if len(res.Content) > 0 {
			if tc, ok := res.Content[0].(*sdk.TextContent); ok {
				return "", fmt.Errorf("tool error: %s", tc.Text)
			}
		}
		return "", errors.New("tool call returned isError=true")
	}

	// Extract successful response content
	if len(res.Content) == 0 {
		return "", nil
	}
	
	if tc, ok := res.Content[0].(*sdk.TextContent); ok {
		return tc.Text, nil
	}
	
	return "", nil
}

// Close shuts down the client and subprocess
func (c *StdIOWebsetsClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.session != nil {
		_ = c.session.Close()
	}
	
	c.client = nil
	c.session = nil
	c.started = false
	
	log.Println("Closed EXA Websets MCP client")
	return nil
}

// isTransportError checks if an error is related to transport/connection issues
func isTransportError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "transport") || 
		contains(errStr, "connection") || 
		contains(errStr, "pipe") ||
		contains(errStr, "EOF")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		len(s) > len(substr) && containsHelper(s[1:], substr)
}

func containsHelper(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if s[:len(substr)] == substr {
		return true
	}
	return containsHelper(s[1:], substr)
}

// WebsetsOperations provides high-level operations for websets management
type WebsetsOperations struct {
	client WebsetsClient
}

// NewWebsetsOperations creates a new websets operations handler
func NewWebsetsOperations(client WebsetsClient) *WebsetsOperations {
	return &WebsetsOperations{client: client}
}

// CreateWebset creates a new webset with the given search query
func (o *WebsetsOperations) CreateWebset(ctx context.Context, searchQuery string, resultCount int) (string, error) {
	args := map[string]any{
		"operation": "create_webset",
		"webset": map[string]any{
			"searchQuery": searchQuery,
			"advanced": map[string]any{
				"resultCount": resultCount,
			},
		},
	}

	result, err := o.client.Call(ctx, args)
	if err != nil {
		return "", fmt.Errorf("create webset failed: %w", err)
	}

	// Parse response to extract webset ID
	var response map[string]any
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		return "", fmt.Errorf("failed to parse create response: %w", err)
	}

	if resourceID, ok := response["resourceId"].(string); ok {
		return resourceID, nil
	}

	return "", errors.New("no resourceId in create response")
}

// GetWebsetStatus polls the status of a webset
func (o *WebsetsOperations) GetWebsetStatus(ctx context.Context, websetID string) (string, error) {
	args := map[string]any{
		"operation":  "get_webset_status",
		"resourceId": websetID,
	}

	return o.client.Call(ctx, args)
}

// ListContentItems retrieves content items from a webset
func (o *WebsetsOperations) ListContentItems(ctx context.Context, websetID string, limit int) ([]map[string]any, error) {
	args := map[string]any{
		"operation":  "list_content_items",
		"resourceId": websetID,
		"query": map[string]any{
			"limit": limit,
		},
	}

	result, err := o.client.Call(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("list content items failed: %w", err)
	}

	// Parse response
	var response map[string]any
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		return nil, fmt.Errorf("failed to parse list response: %w", err)
	}

	if items, ok := response["items"].([]any); ok {
		var contentItems []map[string]any
		for _, item := range items {
			if contentItem, ok := item.(map[string]any); ok {
				contentItems = append(contentItems, contentItem)
			}
		}
		return contentItems, nil
	}

	return nil, nil
}

// WaitForWebsetCompletion polls webset status until completion or timeout
func (o *WebsetsOperations) WaitForWebsetCompletion(ctx context.Context, websetID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return errors.New("webset completion timeout")
			}

			status, err := o.GetWebsetStatus(ctx, websetID)
			if err != nil {
				log.Printf("Status check error: %v", err)
				continue
			}

			// Parse status response
			var statusResponse map[string]any
			if err := json.Unmarshal([]byte(status), &statusResponse); err != nil {
				log.Printf("Failed to parse status: %v", err)
				continue
			}

			if state, ok := statusResponse["status"].(string); ok {
				switch state {
				case "completed":
					return nil
				case "failed":
					return errors.New("webset processing failed")
				case "processing", "pending":
					// Continue polling
				default:
					log.Printf("Unknown webset status: %s", state)
				}
			}
		}
	}
}

// MCPClient manages connections to other MCP servers (backwards compatibility)
type MCPClient struct {
	mu            sync.RWMutex
	websetsClient WebsetsClient
}

// NewMCPClient creates a new MCP client manager
func NewMCPClient() *MCPClient {
	// Default to using exa-websets-mcp-server binary
	websetsClient := NewStdIOWebsetsClient("exa-websets-mcp-server", nil)
	
	// Fallback to node if binary not found
	if _, err := exec.LookPath("exa-websets-mcp-server"); err != nil {
		log.Println("exa-websets-mcp-server not found, using node fallback")
		websetsClient = NewStdIOWebsetsClient("node", []string{"./build/index.js"})
	}
	
	return &MCPClient{
		websetsClient: websetsClient,
	}
}

// Initialize initializes the MCP client connections
func (c *MCPClient) Initialize(ctx context.Context) error {
	log.Println("Initializing MCP client connections...")
	return c.websetsClient.Connect(ctx)
}

// CallTool calls a tool on a specific MCP server
func (c *MCPClient) CallTool(ctx context.Context, serverName string, toolName string, arguments interface{}) (interface{}, error) {
	// Route to appropriate client based on server name
	switch serverName {
	case "exa-websets", "websets":
		if toolName != "websets_manager" {
			return nil, fmt.Errorf("unknown tool %s for server %s", toolName, serverName)
		}
		
		args, ok := arguments.(map[string]any)
		if !ok {
			return nil, errors.New("arguments must be map[string]any for websets_manager")
		}
		
		return c.websetsClient.Call(ctx, args)
		
	default:
		return nil, fmt.Errorf("unknown MCP server: %s", serverName)
	}
}

// Shutdown closes all MCP client connections
func (c *MCPClient) Shutdown() {
	log.Println("Shutting down MCP client connections...")
	if c.websetsClient != nil {
		_ = c.websetsClient.Close()
	}
}