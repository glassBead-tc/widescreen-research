package drone

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/glassBead-tc/widescreen-research/pkg/types"
)

// researchRequest is the input payload for the drone HTTP endpoint.
type researchRequest struct {
	Subject   string            `json:"subject"`
	Policy    map[string]any    `json:"policy,omitempty"`
	BudgetSec int               `json:"budget_sec,omitempty"`
	Sources   []string          `json:"sources,omitempty"`
	RunID     string            `json:"run_id,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// researchResponse is the structured output including summary, citations, entities, triples.
type researchResponse struct {
	Subject   string         `json:"subject"`
	Summary   string         `json:"summary"`
	Citations []string       `json:"citations"`
	Entities  []types.Entity `json:"entities"`
	Triples   []types.Triple `json:"triples"`
	DurationS int            `json:"duration_s"`
	DroneID   string         `json:"drone_id"`
	Timestamp time.Time      `json:"timestamp"`
}

func (d *ResearcherDrone) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		return
	case http.MethodPost:
		if r.URL.Path == "/task" {
			d.handleTask(w, r)
			return
		} else if r.URL.Path == "/mcp" {
			d.handleMCP(w, r)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// handleTask handles task execution requests
func (d *ResearcherDrone) handleTask(w http.ResponseWriter, r *http.Request) {
	var req researchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// For MVP: call ConductResearch with basic mapping
	res, err := d.ConductResearch(req.Subject, "", req.Sources, 5)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Publish the result to Pub/Sub asynchronously
	go func() {
		ctx := context.Background()
		if err := d.publishResult(ctx, res); err != nil {
			log.Printf("ERROR: Failed to publish research result for subject '%s': %v", req.Subject, err)
		}
	}()

	// Respond immediately with 202 Accepted
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte("Task accepted for processing."))
}

// MCPRequest represents a JSON-RPC 2.0 request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC 2.0 response
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

// handleMCP handles MCP protocol requests (tools/list, tools/call)
func (d *ResearcherDrone) handleMCP(w http.ResponseWriter, r *http.Request) {
	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	var response MCPResponse
	response.JSONRPC = "2.0"
	response.ID = req.ID

	switch req.Method {
	case "tools/list":
		// Return available tools for this drone
		response.Result = map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "conduct_research",
					"description": "Conduct research on a given subject using web search and analysis",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"subject": map[string]interface{}{
								"type":        "string",
								"description": "The research subject or query",
							},
							"sources": map[string]interface{}{
								"type":        "array",
								"description": "Optional list of specific sources to search",
								"items": map[string]interface{}{
									"type": "string",
								},
							},
						},
						"required": []string{"subject"},
					},
				},
			},
		}

	case "tools/call":
		// Handle tool execution
		params, ok := req.Params.(map[string]interface{})
		if !ok {
			response.Error = &MCPError{
				Code:    -32602,
				Message: "Invalid params",
			}
		} else {
			toolName, ok := params["name"].(string)
			if !ok {
				response.Error = &MCPError{
					Code:    -32602,
					Message: "Missing tool name",
				}
			} else {
				switch toolName {
				case "conduct_research":
					args, _ := params["arguments"].(map[string]interface{})
					subject, _ := args["subject"].(string)
					sources, _ := args["sources"].([]string)

					if subject == "" {
						response.Error = &MCPError{
							Code:    -32602,
							Message: "Missing required parameter: subject",
						}
					} else {
						// Execute research
						res, err := d.ConductResearch(subject, "", sources, 5)
						if err != nil {
							response.Error = &MCPError{
								Code:    -32603,
								Message: fmt.Sprintf("Research failed: %v", err),
							}
						} else {
							response.Result = map[string]interface{}{
								"content": []map[string]interface{}{
									{
										"type": "text",
										"text": fmt.Sprintf("Research completed for: %s\n\nSummary: %s\n\nFindings: %v",
											res["topic"], res["summary"], res["findings"]),
									},
								},
							}
						}
					}
				default:
					response.Error = &MCPError{
						Code:    -32601,
						Message: fmt.Sprintf("Unknown tool: %s", toolName),
					}
				}
			}
		}

	default:
		response.Error = &MCPError{
			Code:    -32601,
			Message: fmt.Sprintf("Unknown method: %s", req.Method),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartHTTPServer starts the HTTP server for the researcher drone.
func (d *ResearcherDrone) StartHTTPServer(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/health", d)
	mux.Handle("/task", d)
	mux.Handle("/mcp", d)
	log.Printf("Researcher Drone HTTP listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}
