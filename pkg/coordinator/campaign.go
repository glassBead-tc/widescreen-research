package coordinator

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/spawn-mcp/coordinator/pkg/types"
)

// PlanCampaign validates a CampaignSpec, assigns a run ID, stores it, and returns a plan.
func (s *Server) PlanCampaign(ctx context.Context, spec types.CampaignSpec) (*types.CampaignPlan, error) {
	if spec.DatasetURI == "" {
		return nil, fmt.Errorf("dataset_uri is required")
	}
	if spec.DepthProfile == "" {
		spec.DepthProfile = "S1"
	}
	if spec.Parallelism <= 0 || spec.Parallelism > 100 {
		spec.Parallelism = 10
	}
	if spec.PerTaskTimeBudgetSec <= 0 {
		spec.PerTaskTimeBudgetSec = 180
	}
	if len(spec.Sources) == 0 {
		spec.Sources = []string{"exa", "wikipedia", "github"}
	}
	if spec.Mem0Space == "" {
		return nil, fmt.Errorf("mem0_space is required")
	}

	runID := uuid.New().String()
	spec.RunID = runID
	spec.CreatedAt = time.Now()

	// naive estimation: parallel tasks equals parallelism for now
	estimatedCost := s.estimateTaskCost(spec.Parallelism, spec.PerTaskTimeBudgetSec/60)
	plan := &types.CampaignPlan{
		RunID:            runID,
		Spec:             spec,
		TasksPlanned:     spec.Parallelism,
		EstimatedETA:     fmt.Sprintf("~%d min", spec.PerTaskTimeBudgetSec/60+2),
		EstimatedCostUSD: estimatedCost,
	}

	// Store spec and plan in Firestore
	if err := s.gcpClient.StoreDocument(ctx, "campaign_specs", runID, spec); err != nil {
		return nil, fmt.Errorf("store spec: %w", err)
	}
	if err := s.gcpClient.StoreDocument(ctx, "campaign_plans", runID, plan); err != nil {
		return nil, fmt.Errorf("store plan: %w", err)
	}
	return plan, nil
}

// LaunchFleet provisions workers and seeds queue for the given run.
func (s *Server) LaunchFleet(ctx context.Context, runID string, targetWorkers int) (string, error) {
	if targetWorkers <= 0 {
		targetWorkers = 10
	}
	// Placeholder: spawn research drones using existing SpawnDrone
	for i := 0; i < targetWorkers; i++ {
		_, _ = s.SpawnDrone(ctx, types.DroneConfig{Type: types.DroneTypeResearcher, Region: s.gcpClient.Region})
	}
	statusID := fmt.Sprintf("status-%s", runID)
	_ = s.gcpClient.StoreDocument(ctx, "campaign_status", runID, map[string]any{
		"run_id": runID,
		"workers": targetWorkers,
		"state": "launching",
		"updated_at": time.Now(),
	})
	return statusID, nil
}

// FleetStatus returns a minimal status payload.
func (s *Server) FleetStatus(ctx context.Context, runID string) (map[string]any, error) {
	return map[string]any{
		"run_id": runID,
		"active_drones": len(s.ListActiveDrones()),
		"state": "running",
		"updated_at": time.Now(),
	}, nil
}

// AbortRun scales down workers and marks run aborted.
func (s *Server) AbortRun(ctx context.Context, runID string) error {
	// Placeholder: no-op beyond status marker
	return s.gcpClient.StoreDocument(ctx, "campaign_status", runID, map[string]any{
		"run_id": runID,
		"state": "aborted",
		"updated_at": time.Now(),
	})
}

// ExportGraph placeholder; in MVP this would read mem0 and dump edges.
func (s *Server) ExportGraph(ctx context.Context, mem0Space, format string) (string, error) {
	// Return a GCS placeholder URL for now
	return fmt.Sprintf("gs://export-bucket/%s/graph.%s", mem0Space, format), nil
}