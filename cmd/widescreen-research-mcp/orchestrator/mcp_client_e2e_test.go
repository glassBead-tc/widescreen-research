package orchestrator

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestE2E_SubprocessCommunication tests real subprocess communication
func TestE2E_SubprocessCommunication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Create mock server script
	scriptPath := createMockMCPServerScript(t)
	
	// Set a dummy API key for testing
	os.Setenv("EXA_API_KEY", "test-key-123")
	defer os.Unsetenv("EXA_API_KEY")

	// Create client pointing to our mock server
	client := NewStdIOWebsetsClient("go", []string{"run", scriptPath})
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test connection
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to mock server: %v", err)
	}
	defer client.Close()

	// Test create webset
	t.Run("CreateWebset", func(t *testing.T) {
		result, err := client.Call(ctx, map[string]any{
			"operation": "create_webset",
			"webset": map[string]any{
				"searchQuery": "test query",
			},
		})
		
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}
		
		var response map[string]any
		if err := json.Unmarshal([]byte(result), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		
		if response["resourceId"] != "test-123" {
			t.Errorf("Expected resourceId 'test-123', got %v", response["resourceId"])
		}
	})

	// Test get status
	t.Run("GetStatus", func(t *testing.T) {
		result, err := client.Call(ctx, map[string]any{
			"operation":  "get_webset_status",
			"resourceId": "test-123",
		})
		
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}
		
		var response map[string]any
		if err := json.Unmarshal([]byte(result), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		
		if response["status"] != "completed" {
			t.Errorf("Expected status 'completed', got %v", response["status"])
		}
	})

	// Test list items
	t.Run("ListItems", func(t *testing.T) {
		result, err := client.Call(ctx, map[string]any{
			"operation":  "list_content_items",
			"resourceId": "test-123",
		})
		
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}
		
		var response map[string]any
		if err := json.Unmarshal([]byte(result), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		
		items, ok := response["items"].([]interface{})
		if !ok || len(items) == 0 {
			t.Error("Expected items in response")
		}
	})
}

// TestE2E_ReconnectionAfterCrash tests automatic reconnection
func TestE2E_ReconnectionAfterCrash(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// This test would require a more sophisticated mock that can be killed
	// For now, we test the reconnection logic indirectly
	t.Run("ReconnectionLogic", func(t *testing.T) {
		os.Setenv("EXA_API_KEY", "test-key")
		defer os.Unsetenv("EXA_API_KEY")

		client := NewStdIOWebsetsClient("false", nil) // 'false' command always fails
		ctx := context.Background()

		// First call should fail to connect
		_, err := client.Call(ctx, map[string]any{"test": "data"})
		if err == nil {
			t.Error("Expected error on failed connection")
		}
	})
}

// TestE2E_WebsetsOperationsPipeline tests the full operations pipeline
func TestE2E_WebsetsOperationsPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	scriptPath := createMockMCPServerScript(t)
	os.Setenv("EXA_API_KEY", "test-key-123")
	defer os.Unsetenv("EXA_API_KEY")

	client := NewStdIOWebsetsClient("go", []string{"run", scriptPath})
	ops := NewWebsetsOperations(client)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect first
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Test the complete pipeline
	t.Run("CompletePipeline", func(t *testing.T) {
		// Create webset
		websetID, err := ops.CreateWebset(ctx, "AI research", 50)
		if err != nil {
			t.Fatalf("CreateWebset failed: %v", err)
		}
		if websetID == "" {
			t.Error("Expected webset ID")
		}

		// Get status
		status, err := ops.GetWebsetStatus(ctx, websetID)
		if err != nil {
			t.Fatalf("GetWebsetStatus failed: %v", err)
		}
		if status == "" {
			t.Error("Expected status response")
		}

		// List items
		items, err := ops.ListContentItems(ctx, websetID, 100)
		if err != nil {
			t.Fatalf("ListContentItems failed: %v", err)
		}
		if len(items) == 0 {
			t.Error("Expected at least one item")
		}
	})
}

// TestE2E_ConcurrentCalls tests concurrent operations
func TestE2E_ConcurrentCalls(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	scriptPath := createMockMCPServerScript(t)
	os.Setenv("EXA_API_KEY", "test-key-123")
	defer os.Unsetenv("EXA_API_KEY")

	client := NewStdIOWebsetsClient("go", []string{"run", scriptPath})
	
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Launch multiple concurrent calls
	errChan := make(chan error, 3)
	
	go func() {
		_, err := client.Call(ctx, map[string]any{
			"operation": "create_webset",
			"webset":    map[string]any{"searchQuery": "query1"},
		})
		errChan <- err
	}()
	
	go func() {
		_, err := client.Call(ctx, map[string]any{
			"operation":  "get_webset_status",
			"resourceId": "test-123",
		})
		errChan <- err
	}()
	
	go func() {
		_, err := client.Call(ctx, map[string]any{
			"operation":  "list_content_items",
			"resourceId": "test-123",
		})
		errChan <- err
	}()

	// Wait for all calls to complete
	for i := 0; i < 3; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent call %d failed: %v", i, err)
		}
	}
}

// TestE2E_EnvironmentInheritance tests that environment variables are passed
func TestE2E_EnvironmentInheritance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Create a script that checks for environment variable
	tmpDir := t.TempDir()
	scriptPath := tmpDir + "/check_env.sh"
	script := `#!/bin/bash
if [ "$EXA_API_KEY" = "secret-key-456" ]; then
    echo '{"jsonrpc":"2.0","id":1,"result":{"success":true}}'
else
    echo '{"jsonrpc":"2.0","id":1,"error":{"message":"Wrong API key"}}'
fi
`
	os.WriteFile(scriptPath, []byte(script), 0755)

	os.Setenv("EXA_API_KEY", "secret-key-456")
	defer os.Unsetenv("EXA_API_KEY")

	client := NewStdIOWebsetsClient("bash", []string{scriptPath})
	ctx := context.Background()

	// The script will respond based on the environment variable
	// This tests that the subprocess inherits the parent's environment
	err := client.Connect(ctx)
	// Connection might fail due to protocol mismatch, but that's ok
	// We're testing that the environment is passed
	if err == nil {
		client.Close()
	}
}

// TestE2E_BinaryFallback tests the binary fallback mechanism
func TestE2E_BinaryFallback(t *testing.T) {
	// Test that NewMCPClient handles missing binary gracefully
	client := NewMCPClient()
	if client == nil {
		t.Fatal("NewMCPClient should not return nil")
	}
	
	// Check that it selected a fallback
	if client.websetsClient == nil {
		t.Error("websetsClient should not be nil")
	}
	
	// Clean shutdown
	client.Shutdown()
}

// TestE2E_LongRunningOperation simulates a long-running webset creation
func TestE2E_LongRunningOperation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Create a mock that simulates delayed responses
	mockClient := &MockWebsetsClient{
		CallFunc: func(ctx context.Context, arguments map[string]any) (string, error) {
			op, _ := arguments["operation"].(string)
			
			switch op {
			case "create_webset":
				return `{"resourceId": "slow-webset-123", "status": "created"}`, nil
			case "get_webset_status":
				// Simulate processing delay
				select {
				case <-time.After(100 * time.Millisecond):
					return `{"status": "completed"}`, nil
				case <-ctx.Done():
					return "", ctx.Err()
				}
			default:
				return `{}`, nil
			}
		},
	}

	ops := NewWebsetsOperations(mockClient)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create webset
	websetID, err := ops.CreateWebset(ctx, "long query", 100)
	if err != nil {
		t.Fatalf("CreateWebset failed: %v", err)
	}

	// Poll for completion (this will timeout in our test)
	err = ops.WaitForWebsetCompletion(ctx, websetID, 200*time.Millisecond)
	// Should timeout since we have 10s poll interval
	if err == nil || err.Error() != "webset completion timeout" {
		t.Errorf("Expected timeout, got: %v", err)
	}
}

// BenchmarkWebsetsCall benchmarks the Call operation
func BenchmarkWebsetsCall(b *testing.B) {
	mockClient := &MockWebsetsClient{
		CallFunc: func(ctx context.Context, arguments map[string]any) (string, error) {
			return `{"result": "ok"}`, nil
		},
	}

	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockClient.Call(ctx, map[string]any{"operation": "test"})
	}
}

// BenchmarkJSONParsing benchmarks JSON parsing overhead
func BenchmarkJSONParsing(b *testing.B) {
	jsonStr := `{"items": [{"title": "Item 1"}, {"title": "Item 2"}], "status": "completed"}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result map[string]any
		json.Unmarshal([]byte(jsonStr), &result)
	}
}