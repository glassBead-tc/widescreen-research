package types

import (
	"time"
)

// DroneType represents the different types of drones we can spawn
type DroneType string

const (
	DroneTypeWorker      DroneType = "worker"
	DroneTypeAnalyzer    DroneType = "analyzer"
	DroneTypeProcessor   DroneType = "processor"
	DroneTypeResearcher  DroneType = "researcher"
	DroneTypeSynthesizer DroneType = "synthesizer"
)

// DroneStatus represents the current state of a drone
type DroneStatus string

const (
	DroneStatusStarting    DroneStatus = "starting"
	DroneStatusReady       DroneStatus = "ready"
	DroneStatusBusy        DroneStatus = "busy"
	DroneStatusTerminating DroneStatus = "terminating"
	DroneStatusFailed      DroneStatus = "failed"
)

// DroneConfig holds the configuration for a drone instance
type DroneConfig struct {
	Type         DroneType            `json:"type"`
	Region       string               `json:"region"`
	Resources    ResourceRequirements `json:"resources"`
	Capabilities []string             `json:"capabilities"`
	Environment  map[string]string    `json:"environment"`
}

// ResourceRequirements specifies CPU and memory requirements
type ResourceRequirements struct {
	CPU    string `json:"cpu"`    // e.g., "1000m" for 1 CPU
	Memory string `json:"memory"` // e.g., "512Mi"
}

// DroneInfo represents a running drone instance
type DroneInfo struct {
	ID             string                 `json:"id"`
	ServiceName    string                 `json:"serviceName"`
	ServiceURL     string                 `json:"serviceUrl"`
	Type           string                 `json:"type"`
	Status         string                 `json:"status"`
	Region         string                 `json:"region"`
	CreatedAt      time.Time              `json:"createdAt"`
	LastPing       time.Time              `json:"lastPing"`
	LastSeen       time.Time              `json:"lastSeen"`
	TasksCompleted int                    `json:"tasksCompleted"`
	Capabilities   []string               `json:"capabilities"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// TaskDefinition defines a distributed task
type TaskDefinition struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"`
	Description      string                 `json:"description"`
	Parameters       map[string]interface{} `json:"parameters"`
	RequiredDrones   int                    `json:"requiredDrones"`
	DroneType        DroneType              `json:"droneType"`
	TimeoutMinutes   int                    `json:"timeoutMinutes"`
	CheckpointConfig CheckpointConfig       `json:"checkpointConfig"`
}

// CheckpointConfig defines checkpointing behavior
type CheckpointConfig struct {
	Enabled         bool `json:"enabled"`
	IntervalSeconds int  `json:"intervalSeconds"`
	MaxRetries      int  `json:"maxRetries"`
}

// TaskCheckpoint represents a saved state
type TaskCheckpoint struct {
	TaskID     string                 `json:"taskId"`
	DroneID    string                 `json:"droneId"`
	Progress   float64                `json:"progress"`
	State      map[string]interface{} `json:"state"`
	Timestamp  time.Time              `json:"timestamp"`
	RetryCount int                    `json:"retryCount"`
}

// TaskResult represents the output from a drone
type TaskResult struct {
	TaskID    string      `json:"taskId"`
	DroneID   string      `json:"droneId"`
	Status    string      `json:"status"`
	Data      interface{} `json:"data"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ExecutionPlan represents a plan for distributed execution
type ExecutionPlan struct {
	ID               string            `json:"id"`
	TaskDefinition   TaskDefinition    `json:"taskDefinition"`
	DroneCount       int               `json:"droneCount"`
	EstimatedCost    float64           `json:"estimatedCost"`
	EstimatedTime    time.Duration     `json:"estimatedTime"`
	Strategy         string            `json:"strategy"`
	DroneAllocations []DroneAllocation `json:"droneAllocations"`
}

// DroneAllocation represents how work is allocated to a drone
type DroneAllocation struct {
	DroneID    string                 `json:"droneId"`
	WorkItems  []interface{}          `json:"workItems"`
	Parameters map[string]interface{} `json:"parameters"`
}

// CostEstimate provides cost prediction details
type CostEstimate struct {
	TotalCost  float64            `json:"totalCost"`
	Breakdown  map[string]float64 `json:"breakdown"`
	Warnings   []string           `json:"warnings"`
	Confidence float64            `json:"confidence"`
}

// Task represents a task to be executed by drones
type Task struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	MaxDrones   int    `json:"maxDrones"`
}
