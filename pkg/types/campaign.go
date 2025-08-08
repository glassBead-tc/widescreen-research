package types

import "time"

// CampaignSpec defines a distributed research campaign configuration
// aligned with the elicitation JSON shown in design docs.
type CampaignSpec struct {
	DatasetURI           string            `json:"dataset_uri"`
	DepthProfile         string            `json:"depth_profile"`
	Parallelism          int               `json:"parallelism"`
	PerTaskTimeBudgetSec int               `json:"per_task_time_budget_s"`
	Sources              []string          `json:"sources"`
	Mem0Space            string            `json:"mem0_space"`
	QualityBar           QualityBar        `json:"quality_bar"`
	RunID                string            `json:"run_id,omitempty"`
	CreatedAt            time.Time         `json:"created_at,omitempty"`
	Metadata             map[string]string `json:"metadata,omitempty"`
}

// QualityBar defines validation requirements for extracted facts
// and graph edges.
type QualityBar struct {
	MinSourcesPerFact int  `json:"min_sources_per_fact"`
	MustCite          bool `json:"must_cite"`
}

// CampaignPlan expands a CampaignSpec into an executable plan.
type CampaignPlan struct {
	RunID        string       `json:"run_id"`
	Spec         CampaignSpec `json:"spec"`
	TasksPlanned int          `json:"tasks_planned"`
	EstimatedETA string       `json:"estimated_eta"`
	EstimatedCostUSD float64  `json:"estimated_cost_usd"`
}