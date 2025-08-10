package drone

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/spawn-mcp/coordinator/pkg/types"
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
	Subject   string             `json:"subject"`
	Summary   string             `json:"summary"`
	Citations []string           `json:"citations"`
	Entities  []types.Entity     `json:"entities"`
	Triples   []types.Triple     `json:"triples"`
	DurationS int                `json:"duration_s"`
	DroneID   string             `json:"drone_id"`
	Timestamp time.Time          `json:"timestamp"`
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
		if r.URL.Path != "/task" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
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
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// StartHTTPServer starts the HTTP server for the researcher drone.
func (d *ResearcherDrone) StartHTTPServer(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/health", d)
	mux.Handle("/task", d)
	log.Printf("Researcher Drone HTTP listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}