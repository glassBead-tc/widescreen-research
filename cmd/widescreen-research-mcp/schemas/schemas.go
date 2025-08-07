package schemas

import "time"

// WidescreenResearchInput represents the input for the widescreen-research tool
type WidescreenResearchInput struct {
	Operation          string                 `json:"operation,omitempty"`
	SessionID          string                 `json:"session_id,omitempty"`
	ElicitationAnswers map[string]interface{} `json:"elicitation_answers,omitempty"`
	Parameters         map[string]interface{} `json:"parameters,omitempty"`
}

// ElicitationQuestion represents a question in the elicitation process
type ElicitationQuestion struct {
	ID       string                 `json:"id"`
	Question string                 `json:"question"`
	Type     string                 `json:"type"` // text, number, select, multiselect
	Required bool                   `json:"required"`
	Options  []ElicitationOption    `json:"options,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ElicitationOption represents an option for select/multiselect questions
type ElicitationOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ElicitationResponse represents the response from the elicitation process
type ElicitationResponse struct {
	Type      string                `json:"type"` // elicitation, ready, error
	Questions []ElicitationQuestion `json:"questions,omitempty"`
	SessionID string                `json:"session_id"`
	Message   string                `json:"message,omitempty"`
	Config    *ResearchConfig       `json:"config,omitempty"`
}

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
	SessionID    string                 `json:"session_id"`
	Status       string                 `json:"status"`
	ReportURL    string                 `json:"report_url,omitempty"`
	ReportData   interface{}            `json:"report_data,omitempty"`
	Metrics      ResearchMetrics        `json:"metrics"`
	CompletedAt  time.Time              `json:"completed_at"`
}

// ResearchMetrics contains metrics about the research process
type ResearchMetrics struct {
	DronesProvisioned int           `json:"drones_provisioned"`
	DronesCompleted   int           `json:"drones_completed"`
	DronesFailed      int           `json:"drones_failed"`
	TotalDuration     time.Duration `json:"total_duration"`
	DataPointsCollected int         `json:"data_points_collected"`
	CostEstimate      float64       `json:"cost_estimate"`
}

// DroneResult represents the result from a single research drone
type DroneResult struct {
	DroneID      string                 `json:"drone_id"`
	Status       string                 `json:"status"`
	Data         map[string]interface{} `json:"data"`
	Error        string                 `json:"error,omitempty"`
	CompletedAt  time.Time              `json:"completed_at"`
	ProcessingTime time.Duration        `json:"processing_time"`
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
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	URL          string    `json:"url,omitempty"`
	Status       string    `json:"status"`
	Region       string    `json:"region"`
	CreatedAt    time.Time `json:"created_at"`
}

// SequentialThinkingRequest represents a sequential thinking request
type SequentialThinkingRequest struct {
	Problem     string   `json:"problem"`
	Context     string   `json:"context,omitempty"`
	Steps       []string `json:"steps,omitempty"`
	MaxSteps    int      `json:"max_steps,omitempty"`
}

// SequentialThinkingResponse represents the response from sequential thinking
type SequentialThinkingResponse struct {
	Thoughts []ThoughtStep `json:"thoughts"`
	Solution string        `json:"solution"`
	Confidence float64     `json:"confidence"`
}

// ThoughtStep represents a single step in sequential thinking
type ThoughtStep struct {
	Step       int    `json:"step"`
	Thought    string `json:"thought"`
	Reasoning  string `json:"reasoning"`
	Confidence float64 `json:"confidence"`
}

// DataAnalysisRequest represents a request to analyze research data
type DataAnalysisRequest struct {
	Data       []DroneResult `json:"data"`
	AnalysisType string      `json:"analysis_type"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// DataAnalysisResponse represents the response from data analysis
type DataAnalysisResponse struct {
	Summary    string                 `json:"summary"`
	Insights   []string               `json:"insights"`
	Patterns   []Pattern              `json:"patterns"`
	Statistics map[string]interface{} `json:"statistics"`
	Visualizations []Visualization    `json:"visualizations,omitempty"`
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