// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

package orchestrator

import (
	"context"
	"testing"
	"time"
)

func TestNewMCPClient(t *testing.T) {
	client := NewMCPClient()
	
	if client == nil {
		t.Fatal("NewMCPClient returned nil")
	}
	
	if client.client == nil {
		t.Error("MCP client not initialized")
	}
	
	if client.sessions == nil {
		t.Error("Sessions map not initialized")
	}
	
	if len(client.sessions) != 0 {
		t.Error("Sessions map should be empty initially")
	}
}

func TestMCPClientInitialize(t *testing.T) {
	client := NewMCPClient()
	ctx := context.Background()
	
	err := client.Initialize(ctx)
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}
}

func TestMCPClientShutdown(t *testing.T) {
	client := NewMCPClient()
	
	err := client.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
	
	if len(client.sessions) != 0 {
		t.Error("Sessions map should be empty after shutdown")
	}
}

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3, 5*time.Second)
	
	if cb.state != CircuitClosed {
		t.Error("Circuit breaker should start in closed state")
	}
	
	// Test successful calls
	err := cb.Call(func() error { return nil })
	if err != nil {
		t.Errorf("Successful call should not return error: %v", err)
	}
	
	if cb.state != CircuitClosed {
		t.Error("Circuit breaker should remain closed after successful call")
	}
}

func TestCircuitBreakerFailures(t *testing.T) {
	cb := NewCircuitBreaker(2, 5*time.Second)
	
	// First failure
	err := cb.Call(func() error { return &testError{"test error"} })
	if err == nil {
		t.Error("Expected error from failing call")
	}
	if cb.state != CircuitClosed {
		t.Error("Circuit should remain closed after first failure")
	}
	
	// Second failure - should open circuit
	err = cb.Call(func() error { return &testError{"test error"} })
	if err == nil {
		t.Error("Expected error from failing call")
	}
	if cb.state != CircuitOpen {
		t.Error("Circuit should be open after max failures")
	}
	
	// Third call should be rejected
	err = cb.Call(func() error { return nil })
	if err == nil {
		t.Error("Expected circuit breaker to reject call")
	}
	if err.Error() != "circuit breaker is open" {
		t.Errorf("Expected circuit breaker error, got: %v", err)
	}
}

func TestIsNonRetryableError(t *testing.T) {
	err := &testError{"test error"}
	
	// Currently all errors are considered retryable
	if isNonRetryableError(err) {
		t.Error("Test error should be retryable")
	}
}

// testError is a simple error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

// Integration test helpers (these would require actual MCP servers to test against)

func TestMCPClientInterface(t *testing.T) {
	// Verify that MCPClient implements MCPClientInterface
	var _ MCPClientInterface = &MCPClient{}
}

// Benchmark tests for performance validation

func BenchmarkNewMCPClient(b *testing.B) {
	for i := 0; i < b.N; i++ {
		client := NewMCPClient()
		_ = client
	}
}

func BenchmarkCircuitBreakerCall(b *testing.B) {
	cb := NewCircuitBreaker(10, 5*time.Second)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Call(func() error { return nil })
	}
}

// Mock implementations for testing (would be expanded for full integration tests)

type MockMCPServer struct {
	tools []string
	responses map[string]interface{}
}

func NewMockMCPServer() *MockMCPServer {
	return &MockMCPServer{
		tools: []string{"test_tool", "research_tool"},
		responses: map[string]interface{}{
			"test_tool": "test response",
		},
	}
}

// Example of how integration tests would work with real servers
func TestMCPClientIntegration(t *testing.T) {
	t.Skip("Integration test - requires running MCP server")
	
	client := NewMCPClient()
	ctx := context.Background()
	
	err := client.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}
	defer client.Shutdown()
	
	// This would test against a real MCP server
	// droneURL := "http://localhost:8080"
	// err = client.ConnectToDrone(ctx, droneURL)
	// if err != nil {
	//     t.Fatalf("Failed to connect to drone: %v", err)
	// }
	
	// tools, err := client.ListTools(ctx, droneURL)
	// if err != nil {
	//     t.Fatalf("Failed to list tools: %v", err)
	// }
	
	// if len(tools.Tools) == 0 {
	//     t.Error("Expected at least one tool")
	// }
}
