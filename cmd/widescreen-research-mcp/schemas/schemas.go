// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

package schemas

import "time"

// ResearchConfig represents the configuration for a research session
type ResearchConfig struct {
	SessionID         string    `json:"session_id"`
	Topic             string    `json:"topic"`
	ResearcherCount   int       `json:"researcher_count"`
	ResearchDepth     string    `json:"research_depth"`
	OutputFormat      string    `json:"output_format"`
	TimeoutMinutes    int       `json:"timeout_minutes"`
	PriorityLevel     string    `json:"priority_level"`
	WorkflowTemplates string    `json:"workflow_templates,omitempty"`
	SpecificSources   string    `json:"specific_sources,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// ResearchResult represents the result of a research operation
type ResearchResult struct {
	SessionID   string          `json:"session_id"`
	Status      string          `json:"status"`
	ReportURL   string          `json:"report_url,omitempty"`
	ReportData  interface{}     `json:"report_data,omitempty"`
	Metrics     ResearchMetrics `json:"metrics"`
	CompletedAt time.Time       `json:"completed_at"`
}

// ResearchMetrics contains metrics about the research process
type ResearchMetrics struct {
	DronesProvisioned   int           `json:"drones_provisioned"`
	DronesCompleted     int           `json:"drones_completed"`
	DronesFailed        int           `json:"drones_failed"`
	TotalDuration       time.Duration `json:"total_duration"`
	DataPointsCollected int           `json:"data_points_collected"`
	CostEstimate        float64       `json:"cost_estimate"`
}

// DroneTask represents the input for a single research drone
type DroneTask struct {
	TaskID            string                 `json:"task_id"`
	Query             string                 `json:"query"`
	ResultPostbackURL string                 `json:"result_postback_url"`
	Parameters        map[string]interface{} `json:"parameters,omitempty"`
}

// DroneResult represents the result from a single research drone
type DroneResult struct {
	DroneID        string                 `json:"drone_id"`
	Status         string                 `json:"status"`
	Data           map[string]interface{} `json:"data"`
	Error          string                 `json:"error,omitempty"`
	CompletedAt    time.Time              `json:"completed_at"`
	ProcessingTime time.Duration          `json:"processing_time"`
}

// GCPProvisionRequest represents a request to provision GCP resources
type GCPProvisionRequest struct {
	ResourceType string                 `json:"resource_type"` // cloud_run, pubsub, firestore
	Count        int                    `json:"count"`
	Region       string                 `json:"region"`
	Config       map[string]interface{} `json:"config"`
}

// GCPProvisionResponse represents the response from GCP provisioning
type GCPProvisionResponse struct {
	Resources []GCPResource `json:"resources"`
	Status    string        `json:"status"`
	Message   string        `json:"message,omitempty"`
}

// GCPResource represents a provisioned GCP resource
type GCPResource struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	URL       string    `json:"url,omitempty"`
	Status    string    `json:"status"`
	Region    string    `json:"region"`
	CreatedAt time.Time `json:"created_at"`
}

// DataAnalysisRequest represents a request to analyze research data
type DataAnalysisRequest struct {
	Data         []DroneResult          `json:"data"`
	AnalysisType string                 `json:"analysis_type"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
}

// DataAnalysisResponse represents the response from data analysis
type DataAnalysisResponse struct {
	Summary        string                 `json:"summary"`
	Insights       []string               `json:"insights"`
	Patterns       []Pattern              `json:"patterns"`
	Statistics     map[string]interface{} `json:"statistics"`
	Visualizations []Visualization        `json:"visualizations,omitempty"`
}

// Pattern represents a discovered pattern in the data
type Pattern struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Frequency   int     `json:"frequency"`
	Confidence  float64 `json:"confidence"`
}

// Visualization represents a data visualization
type Visualization struct {
	Type   string                 `json:"type"`
	Title  string                 `json:"title"`
	Data   interface{}            `json:"data"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// ResearchReport represents a final research report
type ResearchReport struct {
	ID          string                 `json:"id"`
	SessionID   string                 `json:"session_id"`
	Title       string                 `json:"title"`
	Executive   string                 `json:"executive_summary"`
	Sections    []ReportSection        `json:"sections"`
	Methodology string                 `json:"methodology"`
	Data        map[string]interface{} `json:"data"`
	Metadata    ReportMetadata         `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ReportSection represents a section in the research report
type ReportSection struct {
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Insights []string               `json:"insights,omitempty"`
}

// ReportMetadata contains metadata about the research report
type ReportMetadata struct {
	ResearchTopic   string          `json:"research_topic"`
	ResearcherCount int             `json:"researcher_count"`
	Duration        time.Duration   `json:"duration"`
	DataPoints      int             `json:"data_points"`
	Sources         []string        `json:"sources"`
	Metrics         ResearchMetrics `json:"metrics"`
}
