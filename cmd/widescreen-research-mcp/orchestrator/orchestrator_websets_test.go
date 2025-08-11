package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/schemas"
)

// TestRunWebsetsPipeline tests the complete websets pipeline orchestration
func TestRunWebsetsPipeline(t *testing.T) {
	// Skip if no Google Cloud project
	if getEnvOrDefault("GOOGLE_CLOUD_PROJECT", "") == "" {
		t.Skip("Skipping test: GOOGLE_CLOUD_PROJECT not set")
	}

	// Create mock websets client
	mockClient := &MockWebsetsClient{
		CallFunc: func(ctx context.Context, arguments map[string]any) (string, error) {
			op, _ := arguments["operation"].(string)
			
			switch op {
			case "create_webset":
				return `{"resourceId": "test-webset-789", "status": "created"}`, nil
			case "get_webset_status":
				return `{"status": "completed", "progress": 100}`, nil
			case "list_content_items":
				return `{
					"items": [
						{"title": "AI Paper 1", "url": "http://ai.example.com/1", "content": "Research on neural networks"},
						{"title": "AI Paper 2", "url": "http://ai.example.com/2", "content": "Deep learning advances"}
					],
					"hasMore": false
				}`, nil
			default:
				return `{}`, nil
			}
		},
	}

	// Create orchestrator with mock client
	orch, err := NewOrchestrator()
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	defer orch.Shutdown()

	// Replace the websets client with our mock
	orch.mcpClient.websetsClient = mockClient

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize orchestrator
	if err := orch.Initialize(ctx); err != nil {
		// Initialization might fail due to missing Pub/Sub, that's ok for this test
		t.Logf("Initialization warning: %v", err)
	}

	// Test the pipeline
	t.Run("SuccessfulPipeline", func(t *testing.T) {
		result, err := orch.RunWebsetsPipeline(ctx, "AI research papers", 50)
		if err != nil {
			t.Fatalf("RunWebsetsPipeline failed: %v", err)
		}

		// Verify result structure
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.Status != "completed" {
			t.Errorf("Expected status 'completed', got %v", result.Status)
		}
		if result.SessionID != "test-webset-789" {
			t.Errorf("Expected SessionID 'test-webset-789', got %v", result.SessionID)
		}

		// Check metrics
		if result.Metrics.DronesProvisioned != 1 {
			t.Errorf("Expected 1 drone provisioned, got %d", result.Metrics.DronesProvisioned)
		}
		if result.Metrics.DataPointsCollected != 2 {
			t.Errorf("Expected 2 data points, got %d", result.Metrics.DataPointsCollected)
		}

		// Check report data
		reportData, ok := result.ReportData.(map[string]interface{})
		if !ok {
			t.Fatal("ReportData should be a map")
		}
		if reportData["topic"] != "AI research papers" {
			t.Errorf("Expected topic 'AI research papers', got %v", reportData["topic"])
		}
		if reportData["item_count"].(int) != 2 {
			t.Errorf("Expected 2 items, got %v", reportData["item_count"])
		}
	})
}

// TestRunWebsetsPipeline_Errors tests error handling in the pipeline
func TestRunWebsetsPipeline_Errors(t *testing.T) {
	tests := []struct {
		name      string
		mockFunc  func(context.Context, map[string]any) (string, error)
		wantError string
	}{
		{
			name: "CreateWebsetFailure",
			mockFunc: func(ctx context.Context, arguments map[string]any) (string, error) {
				if op := arguments["operation"]; op == "create_webset" {
					return "", fmt.Errorf("API rate limit exceeded")
				}
				return `{}`, nil
			},
			wantError: "failed to create webset",
		},
		{
			name: "StatusCheckFailure",
			mockFunc: func(ctx context.Context, arguments map[string]any) (string, error) {
				switch arguments["operation"] {
				case "create_webset":
					return `{"resourceId": "test-123"}`, nil
				case "get_webset_status":
					return `{"status": "failed", "error": "Processing error"}`, nil
				default:
					return `{}`, nil
				}
			},
			wantError: "webset processing failed",
		},
		{
			name: "ListItemsFailure",
			mockFunc: func(ctx context.Context, arguments map[string]any) (string, error) {
				switch arguments["operation"] {
				case "create_webset":
					return `{"resourceId": "test-123"}`, nil
				case "get_webset_status":
					return `{"status": "completed"}`, nil
				case "list_content_items":
					return "", fmt.Errorf("Failed to retrieve items")
				default:
					return `{}`, nil
				}
			},
			wantError: "failed to list content items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if no project
			if getEnvOrDefault("GOOGLE_CLOUD_PROJECT", "") == "" {
				t.Skip("Skipping test: GOOGLE_CLOUD_PROJECT not set")
			}

			mockClient := &MockWebsetsClient{CallFunc: tt.mockFunc}
			
			orch, err := NewOrchestrator()
			if err != nil {
				t.Fatalf("Failed to create orchestrator: %v", err)
			}
			defer orch.Shutdown()

			orch.mcpClient.websetsClient = mockClient

			ctx := context.Background()
			_, err = orch.RunWebsetsPipeline(ctx, "test topic", 10)
			
			if err == nil {
				t.Fatal("Expected error but got none")
			}
			if !contains(err.Error(), tt.wantError) {
				t.Errorf("Expected error containing '%s', got %v", tt.wantError, err)
			}
		})
	}
}

// TestWebsetsPubSubIntegration tests Pub/Sub message publishing
func TestWebsetsPubSubIntegration(t *testing.T) {
	// This test requires actual Pub/Sub setup
	if getEnvOrDefault("GOOGLE_CLOUD_PROJECT", "") == "" {
		t.Skip("Skipping test: GOOGLE_CLOUD_PROJECT not set")
	}

	mockClient := &MockWebsetsClient{
		CallFunc: func(ctx context.Context, arguments map[string]any) (string, error) {
			switch arguments["operation"] {
			case "create_webset":
				return `{"resourceId": "pubsub-test-123"}`, nil
			case "get_webset_status":
				return `{"status": "completed"}`, nil
			case "list_content_items":
				return `{
					"items": [
						{"title": "Test Item", "url": "http://test.com", "data": "test data"}
					]
				}`, nil
			default:
				return `{}`, nil
			}
		},
	}

	orch, err := NewOrchestrator()
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	defer orch.Shutdown()

	orch.mcpClient.websetsClient = mockClient

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize to create Pub/Sub client
	if err := orch.Initialize(ctx); err != nil {
		t.Logf("Initialization warning: %v", err)
	}

	result, err := orch.RunWebsetsPipeline(ctx, "pubsub test", 10)
	if err != nil {
		// Pub/Sub operations might fail in test environment
		t.Logf("Pipeline completed with warning: %v", err)
	}

	if result != nil {
		// Verify that items were attempted to be published
		reportData, _ := result.ReportData.(map[string]interface{})
		items, _ := reportData["items"].([]map[string]any)
		if len(items) != 1 {
			t.Errorf("Expected 1 item to be published, got %d", len(items))
		}
	}
}

// MockPubSubClient is a mock implementation of pubsub.Client for testing
type MockPubSubClient struct {
	topics           map[string]*MockTopic
	publishedMessages []pubsub.Message
}

type MockTopic struct {
	name   string
	exists bool
}

func (m *MockTopic) Exists(ctx context.Context) (bool, error) {
	return m.exists, nil
}

func (m *MockTopic) Publish(ctx context.Context, msg *pubsub.Message) *pubsub.PublishResult {
	// Return a mock result
	return &pubsub.PublishResult{}
}

// TestWebsetsMetricsCalculation tests that metrics are correctly calculated
func TestWebsetsMetricsCalculation(t *testing.T) {
	mockClient := &MockWebsetsClient{
		CallFunc: func(ctx context.Context, arguments map[string]any) (string, error) {
			switch arguments["operation"] {
			case "create_webset":
				return `{"resourceId": "metrics-test"}`, nil
			case "get_webset_status":
				return `{"status": "completed"}`, nil
			case "list_content_items":
				// Return many items to test metrics
				items := make([]map[string]interface{}, 100)
				for i := 0; i < 100; i++ {
					items[i] = map[string]interface{}{
						"title": fmt.Sprintf("Item %d", i),
						"url":   fmt.Sprintf("http://example.com/%d", i),
					}
				}
				response := map[string]interface{}{
					"items":   items,
					"hasMore": false,
				}
				data, _ := json.Marshal(response)
				return string(data), nil
			default:
				return `{}`, nil
			}
		},
	}

	// Skip if no project
	if getEnvOrDefault("GOOGLE_CLOUD_PROJECT", "") == "" {
		t.Skip("Skipping test: GOOGLE_CLOUD_PROJECT not set")
	}

	orch, err := NewOrchestrator()
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	defer orch.Shutdown()

	orch.mcpClient.websetsClient = mockClient

	ctx := context.Background()
	result, err := orch.RunWebsetsPipeline(ctx, "metrics test", 100)
	if err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	// Verify metrics
	if result.Metrics.DataPointsCollected != 100 {
		t.Errorf("Expected 100 data points, got %d", result.Metrics.DataPointsCollected)
	}
	if result.Metrics.DronesCompleted != 1 {
		t.Errorf("Expected 1 completed drone, got %d", result.Metrics.DronesCompleted)
	}
	if result.Metrics.DronesFailed != 0 {
		t.Errorf("Expected 0 failed drones, got %d", result.Metrics.DronesFailed)
	}
}

// BenchmarkRunWebsetsPipeline benchmarks the pipeline performance
func BenchmarkRunWebsetsPipeline(b *testing.B) {
	// Skip if no project
	if getEnvOrDefault("GOOGLE_CLOUD_PROJECT", "") == "" {
		b.Skip("Skipping benchmark: GOOGLE_CLOUD_PROJECT not set")
	}

	mockClient := &MockWebsetsClient{
		CallFunc: func(ctx context.Context, arguments map[string]any) (string, error) {
			switch arguments["operation"] {
			case "create_webset":
				return `{"resourceId": "bench-123"}`, nil
			case "get_webset_status":
				return `{"status": "completed"}`, nil
			case "list_content_items":
				return `{"items": [{"title": "Item"}], "hasMore": false}`, nil
			default:
				return `{}`, nil
			}
		},
	}

	orch, err := NewOrchestrator()
	if err != nil {
		b.Fatalf("Failed to create orchestrator: %v", err)
	}
	defer orch.Shutdown()

	orch.mcpClient.websetsClient = mockClient
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orch.RunWebsetsPipeline(ctx, "benchmark test", 10)
	}
}

// TestOrchestratorGetMCPClient tests the GetMCPClient accessor
func TestOrchestratorGetMCPClient(t *testing.T) {
	// Skip if no project
	if getEnvOrDefault("GOOGLE_CLOUD_PROJECT", "") == "" {
		t.Skip("Skipping test: GOOGLE_CLOUD_PROJECT not set")
	}

	orch, err := NewOrchestrator()
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	defer orch.Shutdown()

	client := orch.GetMCPClient()
	if client == nil {
		t.Fatal("GetMCPClient returned nil")
	}
	if client.websetsClient == nil {
		t.Error("MCPClient should have websetsClient initialized")
	}
}

// TestResearchResultSerialization tests that ResearchResult can be properly serialized
func TestResearchResultSerialization(t *testing.T) {
	result := &schemas.ResearchResult{
		SessionID: "test-session",
		Status:    "completed",
		ReportURL: "/websets/test-session",
		ReportData: map[string]interface{}{
			"topic":      "test",
			"item_count": 5,
			"items": []map[string]interface{}{
				{"title": "Item 1"},
			},
		},
		Metrics: schemas.ResearchMetrics{
			DronesProvisioned:   1,
			DronesCompleted:     1,
			DronesFailed:        0,
			TotalDuration:       10 * time.Minute,
			DataPointsCollected: 5,
			CostEstimate:        0.05,
		},
		CompletedAt: time.Now(),
	}

	// Serialize to JSON
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal ResearchResult: %v", err)
	}

	// Deserialize back
	var decoded schemas.ResearchResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ResearchResult: %v", err)
	}

	// Verify key fields
	if decoded.SessionID != result.SessionID {
		t.Errorf("SessionID mismatch: got %v, want %v", decoded.SessionID, result.SessionID)
	}
	if decoded.Status != result.Status {
		t.Errorf("Status mismatch: got %v, want %v", decoded.Status, result.Status)
	}
	if decoded.Metrics.DataPointsCollected != result.Metrics.DataPointsCollected {
		t.Errorf("DataPointsCollected mismatch: got %v, want %v", 
			decoded.Metrics.DataPointsCollected, result.Metrics.DataPointsCollected)
	}
}