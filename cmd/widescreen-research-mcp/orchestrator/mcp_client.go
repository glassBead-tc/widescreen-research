// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

package orchestrator

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/idtoken"
)

// MCPClientInterface defines the interface for MCP client operations
type MCPClientInterface interface {
	// Initialize sets up the client
	Initialize(ctx context.Context) error

	// ConnectToDrone establishes connection to a specific drone MCP server
	ConnectToDrone(ctx context.Context, droneURL string) error

	// CallTool invokes a tool on a specific drone
	CallTool(ctx context.Context, droneURL, toolName string, arguments map[string]interface{}) (*mcp.CallToolResult, error)

	// ListTools gets available tools from a drone
	ListTools(ctx context.Context, droneURL string) (*mcp.ListToolsResult, error)

	// DisconnectFromDrone closes connection to a specific drone
	DisconnectFromDrone(ctx context.Context, droneURL string) error

	// Shutdown closes all connections
	Shutdown() error
}

// MCPClient manages connections to other MCP servers (drones) using the official MCP Go SDK
type MCPClient struct {
	client    *mcp.Client
	sessions  map[string]*mcp.ClientSession // Map of drone URL to MCP session
	projectID string
	mu        sync.RWMutex
}

// NewMCPClient creates a new MCP client manager using the official SDK
func NewMCPClient() *MCPClient {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = "default-project"
	}

	// Create MCP client with implementation info
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "widescreen-research-orchestrator",
		Version: "v1.0.0",
	}, nil)

	return &MCPClient{
		client:    client,
		sessions:  make(map[string]*mcp.ClientSession),
		projectID: projectID,
	}
}

// Initialize initializes the MCP client connections
func (c *MCPClient) Initialize(ctx context.Context) error {
	log.Println("MCPClient initialized successfully")
	return nil
}

// ConnectToDrone establishes connection to a specific drone MCP server
func (c *MCPClient) ConnectToDrone(ctx context.Context, droneURL string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already connected
	if _, exists := c.sessions[droneURL]; exists {
		log.Printf("Already connected to drone: %s", droneURL)
		return nil
	}

	// Create authenticated HTTP client for service-to-service communication
	httpClient, err := c.createAuthenticatedClient(ctx, droneURL)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Create HTTP transport for MCP
	transport := &mcp.StreamableClientTransport{
		Endpoint:   droneURL,
		HTTPClient: httpClient,
	}

	// Connect and get session
	session, err := c.client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to drone %s: %w", droneURL, err)
	}

	// Store session
	c.sessions[droneURL] = session
	log.Printf("Successfully connected to drone: %s", droneURL)
	return nil
}

// createAuthenticatedClient creates an HTTP client with GCP Identity Token authentication
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

// CallTool invokes a tool on a specific drone using the official SDK
func (c *MCPClient) CallTool(ctx context.Context, droneURL, toolName string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	c.mu.RLock()
	session, exists := c.sessions[droneURL]
	c.mu.RUnlock()

	if !exists {
		// Try to connect if not already connected
		if err := c.ConnectToDrone(ctx, droneURL); err != nil {
			return nil, fmt.Errorf("failed to connect to drone %s: %w", droneURL, err)
		}
		c.mu.RLock()
		session = c.sessions[droneURL]
		c.mu.RUnlock()
	}

	// Call tool using session
	params := &mcp.CallToolParams{
		Name:      toolName,
		Arguments: arguments,
	}

	result, err := session.CallTool(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("tool call failed for %s on %s: %w", toolName, droneURL, err)
	}

	if result.IsError {
		return nil, fmt.Errorf("tool %s returned error on %s", toolName, droneURL)
	}

	return result, nil
}

// ListTools gets available tools from a drone using the official SDK
func (c *MCPClient) ListTools(ctx context.Context, droneURL string) (*mcp.ListToolsResult, error) {
	c.mu.RLock()
	session, exists := c.sessions[droneURL]
	c.mu.RUnlock()

	if !exists {
		// Try to connect if not already connected
		if err := c.ConnectToDrone(ctx, droneURL); err != nil {
			return nil, fmt.Errorf("failed to connect to drone %s: %w", droneURL, err)
		}
		c.mu.RLock()
		session = c.sessions[droneURL]
		c.mu.RUnlock()
	}

	// List tools using session
	result, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tools from %s: %w", droneURL, err)
	}

	return result, nil
}

// DisconnectFromDrone closes connection to a specific drone
func (c *MCPClient) DisconnectFromDrone(ctx context.Context, droneURL string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	session, exists := c.sessions[droneURL]
	if !exists {
		log.Printf("No active connection to drone: %s", droneURL)
		return nil
	}

	// Close the session
	if err := session.Close(); err != nil {
		log.Printf("Error closing session for drone %s: %v", droneURL, err)
	}

	// Remove from sessions map
	delete(c.sessions, droneURL)
	log.Printf("Disconnected from drone: %s", droneURL)
	return nil
}

// Shutdown closes all connections
func (c *MCPClient) Shutdown() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Println("Shutting down MCP client...")

	// Close all sessions
	for droneURL, session := range c.sessions {
		if err := session.Close(); err != nil {
			log.Printf("Error closing session for drone %s: %v", droneURL, err)
		}
	}

	// Clear sessions map
	c.sessions = make(map[string]*mcp.ClientSession)
	log.Println("MCP client shutdown complete")
	return nil
}

// Advanced MCP Primitives for Research Enhancement
// Note: Some advanced primitives like Elicit and CreateMessage may not be available
// in the current SDK version and will be added when supported

// Error handling and resilience

// HealthCheck performs a health check on a drone
func (c *MCPClient) HealthCheck(ctx context.Context, droneURL string) error {
	c.mu.RLock()
	session, exists := c.sessions[droneURL]
	c.mu.RUnlock()

	if !exists {
		// Try to connect to check health
		if err := c.ConnectToDrone(ctx, droneURL); err != nil {
			return fmt.Errorf("health check failed - cannot connect to %s: %w", droneURL, err)
		}
		return nil
	}

	// Try a simple operation to verify the session is healthy
	_, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		// Session might be stale, try to reconnect
		log.Printf("Session to %s appears unhealthy, attempting reconnect", droneURL)
		c.mu.Lock()
		delete(c.sessions, droneURL)
		c.mu.Unlock()

		if reconnectErr := c.ConnectToDrone(ctx, droneURL); reconnectErr != nil {
			return fmt.Errorf("health check failed - reconnection to %s failed: %w", droneURL, reconnectErr)
		}
	}

	return nil
}

// CallToolWithRetry calls a tool with retry logic for resilience
func (c *MCPClient) CallToolWithRetry(ctx context.Context, droneURL, toolName string, arguments map[string]interface{}, maxRetries int) (*mcp.CallToolResult, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt*attempt) * time.Second
			log.Printf("Retrying tool call %s on %s (attempt %d/%d) after %v", toolName, droneURL, attempt+1, maxRetries+1, backoff)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		result, err := c.CallTool(ctx, droneURL, toolName, arguments)
		if err == nil {
			if attempt > 0 {
				log.Printf("Tool call %s on %s succeeded after %d retries", toolName, droneURL, attempt)
			}
			return result, nil
		}

		lastErr = err
		log.Printf("Tool call %s on %s failed (attempt %d/%d): %v", toolName, droneURL, attempt+1, maxRetries+1, err)

		// For certain errors, don't retry
		if isNonRetryableError(err) {
			log.Printf("Non-retryable error for tool call %s on %s: %v", toolName, droneURL, err)
			break
		}
	}

	return nil, fmt.Errorf("tool call %s on %s failed after %d retries: %w", toolName, droneURL, maxRetries+1, lastErr)
}

// isNonRetryableError determines if an error should not be retried
func isNonRetryableError(err error) bool {
	// Add logic to identify non-retryable errors
	// For now, assume all errors are potentially retryable
	return false
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitClosed CircuitBreakerState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreaker implements a simple circuit breaker pattern for drone connections
type CircuitBreaker struct {
	state        CircuitBreakerState
	failures     int
	maxFailures  int
	timeout      time.Duration
	lastFailTime time.Time
	mu           sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       CircuitClosed,
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if circuit should transition from open to half-open
	if cb.state == CircuitOpen && time.Since(cb.lastFailTime) > cb.timeout {
		cb.state = CircuitHalfOpen
		cb.failures = 0
	}

	// Reject calls if circuit is open
	if cb.state == CircuitOpen {
		return fmt.Errorf("circuit breaker is open")
	}

	// Execute the function
	err := fn()
	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()

		// Open circuit if max failures reached
		if cb.failures >= cb.maxFailures {
			cb.state = CircuitOpen
		}
		return err
	}

	// Reset on success
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
	}
	cb.failures = 0
	return nil
}

// ListResources gets available resources from a drone (for roots support)
func (c *MCPClient) ListResources(ctx context.Context, droneURL string) (*mcp.ListResourcesResult, error) {
	c.mu.RLock()
	session, exists := c.sessions[droneURL]
	c.mu.RUnlock()

	if !exists {
		if err := c.ConnectToDrone(ctx, droneURL); err != nil {
			return nil, fmt.Errorf("failed to connect to drone %s: %w", droneURL, err)
		}
		c.mu.RLock()
		session = c.sessions[droneURL]
		c.mu.RUnlock()
	}

	result, err := session.ListResources(ctx, &mcp.ListResourcesParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list resources from %s: %w", droneURL, err)
	}

	return result, nil
}

// ReadResource reads a specific resource from a drone (for roots support)
func (c *MCPClient) ReadResource(ctx context.Context, droneURL string, uri string) (*mcp.ReadResourceResult, error) {
	c.mu.RLock()
	session, exists := c.sessions[droneURL]
	c.mu.RUnlock()

	if !exists {
		if err := c.ConnectToDrone(ctx, droneURL); err != nil {
			return nil, fmt.Errorf("failed to connect to drone %s: %w", droneURL, err)
		}
		c.mu.RLock()
		session = c.sessions[droneURL]
		c.mu.RUnlock()
	}

	params := &mcp.ReadResourceParams{
		URI: uri,
	}

	result, err := session.ReadResource(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource %s from %s: %w", uri, droneURL, err)
	}

	return result, nil
}
