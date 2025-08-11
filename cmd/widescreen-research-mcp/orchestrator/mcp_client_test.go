package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

// TestWebsetsClientInterface verifies the interface implementation
func TestWebsetsClientInterface(t *testing.T) {
	var _ WebsetsClient = (*StdIOWebsetsClient)(nil)
	var _ WebsetsClient = (*MockWebsetsClient)(nil)
}

// MockWebsetsClient implements WebsetsClient for testing
type MockWebsetsClient struct {
	ConnectFunc func(ctx context.Context) error
	CallFunc    func(ctx context.Context, arguments map[string]any) (string, error)
	CloseFunc   func() error
	CallCount   int
	Connected   bool
}

func (m *MockWebsetsClient) Connect(ctx context.Context) error {
	if m.ConnectFunc != nil {
		return m.ConnectFunc(ctx)
	}
	m.Connected = true
	return nil
}

func (m *MockWebsetsClient) Call(ctx context.Context, arguments map[string]any) (string, error) {
	m.CallCount++
	if m.CallFunc != nil {
		return m.CallFunc(ctx, arguments)
	}
	return `{"success": true}`, nil
}

func (m *MockWebsetsClient) Close() error {
	m.Connected = false
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// TestNewStdIOWebsetsClient tests client creation
func TestNewStdIOWebsetsClient(t *testing.T) {
	tests := []struct {
		name     string
		binPath  string
		binArgs  []string
		expected string
	}{
		{
			name:     "default binary",
			binPath:  "",
			binArgs:  nil,
			expected: "exa-websets-mcp-server",
		},
		{
			name:     "custom binary",
			binPath:  "node",
			binArgs:  []string{"./build/index.js"},
			expected: "node",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewStdIOWebsetsClient(tt.binPath, tt.binArgs)
			if client == nil {
				t.Fatal("NewStdIOWebsetsClient returned nil")
			}
			if client.binPath != tt.expected {
				t.Errorf("binPath = %v, want %v", client.binPath, tt.expected)
			}
		})
	}
}

// TestStdIOWebsetsClient_Connect tests connection establishment
func TestStdIOWebsetsClient_Connect(t *testing.T) {
	// Skip if EXA_API_KEY is not set
	if os.Getenv("EXA_API_KEY") == "" {
		t.Skip("Skipping test: EXA_API_KEY not set")
	}

	client := NewStdIOWebsetsClient("echo", []string{"connected"})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This will fail since echo doesn't implement MCP protocol
	// but we're testing the connection attempt
	err := client.Connect(ctx)
	if err == nil {
		t.Error("Expected error when connecting to non-MCP process")
		client.Close()
	}
}

// TestStdIOWebsetsClient_ConnectMissingAPIKey tests connection without API key
func TestStdIOWebsetsClient_ConnectMissingAPIKey(t *testing.T) {
	// Store original value and restore after test
	originalKey := os.Getenv("EXA_API_KEY")
	os.Unsetenv("EXA_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("EXA_API_KEY", originalKey)
		}
	}()

	client := NewStdIOWebsetsClient("test-binary", nil)
	ctx := context.Background()

	err := client.Connect(ctx)
	if err == nil {
		t.Fatal("Expected error when EXA_API_KEY is not set")
	}
	if err.Error() != "EXA_API_KEY not set in environment" {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestWebsetsOperations_CreateWebset tests webset creation
func TestWebsetsOperations_CreateWebset(t *testing.T) {
	mockClient := &MockWebsetsClient{
		CallFunc: func(ctx context.Context, arguments map[string]any) (string, error) {
			// Verify the operation
			if op, ok := arguments["operation"].(string); !ok || op != "create_webset" {
				t.Errorf("Expected operation 'create_webset', got %v", arguments["operation"])
			}
			
			// Verify webset structure
			if webset, ok := arguments["webset"].(map[string]any); ok {
				if query, ok := webset["searchQuery"].(string); !ok || query != "test query" {
					t.Errorf("Expected searchQuery 'test query', got %v", query)
				}
			} else {
				t.Error("Missing webset in arguments")
			}
			
			return `{"resourceId": "test-webset-123", "status": "created"}`, nil
		},
	}

	ops := NewWebsetsOperations(mockClient)
	ctx := context.Background()

	websetID, err := ops.CreateWebset(ctx, "test query", 50)
	if err != nil {
		t.Fatalf("CreateWebset failed: %v", err)
	}
	if websetID != "test-webset-123" {
		t.Errorf("Expected websetID 'test-webset-123', got %v", websetID)
	}
	if mockClient.CallCount != 1 {
		t.Errorf("Expected 1 call, got %d", mockClient.CallCount)
	}
}

// TestWebsetsOperations_GetWebsetStatus tests status polling
func TestWebsetsOperations_GetWebsetStatus(t *testing.T) {
	mockClient := &MockWebsetsClient{
		CallFunc: func(ctx context.Context, arguments map[string]any) (string, error) {
			if resourceID, ok := arguments["resourceId"].(string); !ok || resourceID != "test-123" {
				t.Errorf("Expected resourceId 'test-123', got %v", arguments["resourceId"])
			}
			return `{"status": "processing", "progress": 75}`, nil
		},
	}

	ops := NewWebsetsOperations(mockClient)
	ctx := context.Background()

	status, err := ops.GetWebsetStatus(ctx, "test-123")
	if err != nil {
		t.Fatalf("GetWebsetStatus failed: %v", err)
	}
	
	var statusData map[string]any
	if err := json.Unmarshal([]byte(status), &statusData); err != nil {
		t.Fatalf("Failed to parse status: %v", err)
	}
	if statusData["status"] != "processing" {
		t.Errorf("Expected status 'processing', got %v", statusData["status"])
	}
}

// TestWebsetsOperations_ListContentItems tests content retrieval
func TestWebsetsOperations_ListContentItems(t *testing.T) {
	mockClient := &MockWebsetsClient{
		CallFunc: func(ctx context.Context, arguments map[string]any) (string, error) {
			return `{
				"items": [
					{"title": "Item 1", "url": "http://example.com/1"},
					{"title": "Item 2", "url": "http://example.com/2"}
				],
				"hasMore": false
			}`, nil
		},
	}

	ops := NewWebsetsOperations(mockClient)
	ctx := context.Background()

	items, err := ops.ListContentItems(ctx, "test-123", 100)
	if err != nil {
		t.Fatalf("ListContentItems failed: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
	if items[0]["title"] != "Item 1" {
		t.Errorf("Expected first item title 'Item 1', got %v", items[0]["title"])
	}
}

// TestWebsetsOperations_WaitForWebsetCompletion tests completion polling
func TestWebsetsOperations_WaitForWebsetCompletion(t *testing.T) {
	callCount := 0
	mockClient := &MockWebsetsClient{
		CallFunc: func(ctx context.Context, arguments map[string]any) (string, error) {
			callCount++
			// Simulate progression: pending -> processing -> completed
			switch callCount {
			case 1:
				return `{"status": "pending"}`, nil
			case 2:
				return `{"status": "processing"}`, nil
			case 3:
				return `{"status": "completed"}`, nil
			default:
				return `{"status": "completed"}`, nil
			}
		},
	}

	ops := NewWebsetsOperations(mockClient)
	ctx := context.Background()

	// Use short timeout for test
	err := ops.WaitForWebsetCompletion(ctx, "test-123", 1*time.Second)
	// This will timeout since we have 10s poll interval
	if err == nil || err.Error() != "webset completion timeout" {
		t.Errorf("Expected timeout error, got %v", err)
	}
}

// TestWebsetsOperations_WaitForWebsetCompletion_Failed tests failure handling
func TestWebsetsOperations_WaitForWebsetCompletion_Failed(t *testing.T) {
	mockClient := &MockWebsetsClient{
		CallFunc: func(ctx context.Context, arguments map[string]any) (string, error) {
			return `{"status": "failed", "error": "API error"}`, nil
		},
	}

	ops := NewWebsetsOperations(mockClient)
	ctx := context.Background()

	err := ops.WaitForWebsetCompletion(ctx, "test-123", 15*time.Minute)
	if err == nil {
		t.Fatal("Expected error for failed status")
	}
	if err.Error() != "webset processing failed" {
		t.Errorf("Expected 'webset processing failed', got %v", err)
	}
}

// TestMCPClient_BackwardCompatibility tests the backward-compatible MCPClient
func TestMCPClient_BackwardCompatibility(t *testing.T) {
	// Test that MCPClient can be created without error
	client := NewMCPClient()
	if client == nil {
		t.Fatal("NewMCPClient returned nil")
	}
	if client.websetsClient == nil {
		t.Fatal("websetsClient is nil")
	}
	
	// Test shutdown doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Shutdown panicked: %v", r)
		}
	}()
	client.Shutdown()
}

// TestMCPClient_CallTool tests routing to appropriate client
func TestMCPClient_CallTool(t *testing.T) {
	mockWebsets := &MockWebsetsClient{
		CallFunc: func(ctx context.Context, arguments map[string]any) (string, error) {
			return `{"result": "websets response"}`, nil
		},
	}

	client := &MCPClient{websetsClient: mockWebsets}
	ctx := context.Background()

	tests := []struct {
		name       string
		serverName string
		toolName   string
		arguments  interface{}
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid websets call",
			serverName: "websets",
			toolName:   "websets_manager",
			arguments:  map[string]any{"operation": "test"},
			wantErr:    false,
		},
		{
			name:       "valid exa-websets call",
			serverName: "exa-websets",
			toolName:   "websets_manager",
			arguments:  map[string]any{"operation": "test"},
			wantErr:    false,
		},
		{
			name:       "wrong tool for websets",
			serverName: "websets",
			toolName:   "wrong_tool",
			arguments:  map[string]any{},
			wantErr:    true,
			errMsg:     "unknown tool wrong_tool for server websets",
		},
		{
			name:       "unknown server",
			serverName: "unknown",
			toolName:   "some_tool",
			arguments:  map[string]any{},
			wantErr:    true,
			errMsg:     "unknown MCP server: unknown",
		},
		{
			name:       "wrong argument type",
			serverName: "websets",
			toolName:   "websets_manager",
			arguments:  "not a map",
			wantErr:    true,
			errMsg:     "arguments must be map[string]any for websets_manager",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.CallTool(ctx, tt.serverName, tt.toolName, tt.arguments)
			
			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected result but got nil")
				}
			}
		})
	}
}

// TestIsTransportError tests transport error detection
func TestIsTransportError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "transport error",
			err:  fmt.Errorf("transport connection failed"),
			want: true,
		},
		{
			name: "connection error",
			err:  fmt.Errorf("connection reset by peer"),
			want: true,
		},
		{
			name: "pipe error",
			err:  fmt.Errorf("broken pipe"),
			want: true,
		},
		{
			name: "EOF error",
			err:  fmt.Errorf("unexpected EOF"),
			want: true,
		},
		{
			name: "regular error",
			err:  fmt.Errorf("invalid argument"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTransportError(tt.err); got != tt.want {
				t.Errorf("isTransportError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestContains tests the string contains helper
func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "World", false},
		{"transport error", "transport", true},
		{"", "test", false},
		{"test", "", true},
		{"a", "ab", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s contains %s", tt.s, tt.substr), func(t *testing.T) {
			if got := contains(tt.s, tt.substr); got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}