package coordinator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"google.golang.org/api/idtoken"
)

// MCPClient handles communication with remote MCP servers (drones)
type MCPClient struct {
	httpClient *http.Client
	projectID  string
}

// NewMCPClient creates a new MCP client for communicating with drones
func NewMCPClient(projectID string) *MCPClient {
	return &MCPClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		projectID: projectID,
	}
}

// MCPRequest represents a JSON-RPC 2.0 request to an MCP server
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC 2.0 response from an MCP server
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error response
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// CallTool calls a tool on a remote MCP server (drone)
func (c *MCPClient) CallTool(ctx context.Context, droneURL, toolName string, arguments map[string]interface{}) (*MCPResponse, error) {
	// Create authenticated HTTP client for service-to-service communication
	client, err := c.createAuthenticatedClient(ctx, droneURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Prepare MCP request
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}

	// Marshal request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", droneURL+"/mcp", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse MCP response
	var mcpResponse MCPResponse
	if err := json.Unmarshal(responseBody, &mcpResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &mcpResponse, nil
}

// ListTools lists available tools on a remote MCP server (drone)
func (c *MCPClient) ListTools(ctx context.Context, droneURL string) (*MCPResponse, error) {
	// Create authenticated HTTP client
	client, err := c.createAuthenticatedClient(ctx, droneURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Prepare MCP request
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	// Marshal request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", droneURL+"/mcp", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse MCP response
	var mcpResponse MCPResponse
	if err := json.Unmarshal(responseBody, &mcpResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &mcpResponse, nil
}

// createAuthenticatedClient creates an HTTP client with OIDC authentication for service-to-service communication
func (c *MCPClient) createAuthenticatedClient(ctx context.Context, targetURL string) (*http.Client, error) {
	// Create ID token source for the target audience (drone service URL)
	tokenSource, err := idtoken.NewTokenSource(ctx, targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create token source: %w", err)
	}

	// Get ID token
	token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get ID token: %w", err)
	}

	// Create HTTP client with authentication
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &authenticatedTransport{
			base:  http.DefaultTransport,
			token: token.AccessToken,
		},
	}

	return client, nil
}

// authenticatedTransport adds authentication headers to HTTP requests
type authenticatedTransport struct {
	base  http.RoundTripper
	token string
}

func (t *authenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	reqClone := req.Clone(req.Context())

	// Add authorization header
	reqClone.Header.Set("Authorization", "Bearer "+t.token)

	return t.base.RoundTrip(reqClone)
}

// HealthCheck performs a health check on a drone
func (c *MCPClient) HealthCheck(ctx context.Context, droneURL string) error {
	// Create authenticated HTTP client
	client, err := c.createAuthenticatedClient(ctx, droneURL)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Create health check request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", droneURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	// Send request
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send health check request: %w", err)
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}
