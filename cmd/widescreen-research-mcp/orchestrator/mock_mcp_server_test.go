package orchestrator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// MockMCPServer simulates an MCP server for testing
type MockMCPServer struct {
	responses map[string]interface{}
	stdin     io.Reader
	stdout    io.Writer
	stderr    io.Writer
}

// NewMockMCPServer creates a new mock MCP server
func NewMockMCPServer(stdin io.Reader, stdout io.Writer, stderr io.Writer) *MockMCPServer {
	return &MockMCPServer{
		responses: make(map[string]interface{}),
		stdin:     stdin,
		stdout:    stdout,
		stderr:    stderr,
	}
}

// Run starts the mock MCP server
func (m *MockMCPServer) Run() error {
	scanner := bufio.NewScanner(m.stdin)
	encoder := json.NewEncoder(m.stdout)

	// Send initial capability response
	initialResponse := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"result": map[string]interface{}{
			"protocolVersion": "1.0.0",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"websets_manager": map[string]interface{}{
						"description": "Manage websets",
					},
				},
			},
		},
	}
	if err := encoder.Encode(initialResponse); err != nil {
		return err
	}

	// Process incoming requests
	for scanner.Scan() {
		var request map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			continue // Skip malformed requests
		}

		// Handle different request types
		method, _ := request["method"].(string)
		id := request["id"]
		
		var response map[string]interface{}

		switch method {
		case "initialize":
			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"result": map[string]interface{}{
					"protocolVersion": "1.0.0",
					"capabilities":    map[string]interface{}{},
				},
			}

		case "tools/call":
			params, _ := request["params"].(map[string]interface{})
			toolName, _ := params["name"].(string)
			arguments, _ := params["arguments"].(map[string]interface{})
			
			if toolName == "websets_manager" {
				operation, _ := arguments["operation"].(string)
				response = m.handleWebsetsManager(id, operation, arguments)
			} else {
				response = m.errorResponse(id, "Unknown tool")
			}

		default:
			response = m.errorResponse(id, "Unknown method")
		}

		if err := encoder.Encode(response); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func (m *MockMCPServer) handleWebsetsManager(id interface{}, operation string, arguments map[string]interface{}) map[string]interface{} {
	var result interface{}
	
	switch operation {
	case "create_webset":
		result = map[string]interface{}{
			"resourceId": "mock-webset-123",
			"status":     "created",
		}
	case "get_webset_status":
		result = map[string]interface{}{
			"status":   "completed",
			"progress": 100,
		}
	case "list_content_items":
		result = map[string]interface{}{
			"items": []map[string]interface{}{
				{"title": "Mock Item 1", "url": "http://example.com/1"},
				{"title": "Mock Item 2", "url": "http://example.com/2"},
			},
			"hasMore": false,
		}
	default:
		return m.errorResponse(id, fmt.Sprintf("Unknown operation: %s", operation))
	}

	// Serialize result to JSON string (as the real server would)
	resultJSON, _ := json.Marshal(result)
	
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": string(resultJSON),
				},
			},
			"isError": false,
		},
	}
}

func (m *MockMCPServer) errorResponse(id interface{}, message string) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    -32603,
			"message": message,
		},
	}
}

// createMockMCPServerScript creates a Go script that runs a mock MCP server
func createMockMCPServerScript(t *testing.T) string {
	t.Helper()
	
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "mock_server.go")
	
	scriptContent := `package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)
	
	// Process requests
	for scanner.Scan() {
		var request map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			continue
		}
		
		method, _ := request["method"].(string)
		id := request["id"]
		
		var response map[string]interface{}
		
		switch method {
		case "initialize":
			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"result": map[string]interface{}{
					"protocolVersion": "1.0.0",
					"capabilities": map[string]interface{}{
						"tools": map[string]interface{}{},
					},
				},
			}
		case "tools/call":
			params, _ := request["params"].(map[string]interface{})
			arguments, _ := params["arguments"].(map[string]interface{})
			operation, _ := arguments["operation"].(string)
			
			var result interface{}
			switch operation {
			case "create_webset":
				result = map[string]interface{}{
					"resourceId": "test-123",
					"status": "created",
				}
			case "get_webset_status":
				result = map[string]interface{}{
					"status": "completed",
				}
			case "list_content_items":
				result = map[string]interface{}{
					"items": []interface{}{
						map[string]interface{}{"title": "Item 1"},
					},
				}
			default:
				result = map[string]interface{}{"error": "unknown op"}
			}
			
			resultJSON, _ := json.Marshal(result)
			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"result": map[string]interface{}{
					"content": []map[string]interface{}{
						{"type": "text", "text": string(resultJSON)},
					},
					"isError": false,
				},
			}
		default:
			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"error": map[string]interface{}{
					"code": -32601,
					"message": "Method not found",
				},
			}
		}
		
		encoder.Encode(response)
	}
}
`
	
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to create mock server script: %v", err)
	}
	
	return scriptPath
}