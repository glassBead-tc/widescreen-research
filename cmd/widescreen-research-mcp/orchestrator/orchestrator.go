// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

package orchestrator

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	run "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research-mcp/schemas"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Orchestrator manages the research orchestration process
type Orchestrator struct {
	// GCP clients
	firestoreClient *firestore.Client
	pubsubClient    *pubsub.Client
	runClient       *run.ServicesClient

	// Research management
	activeSessions map[string]*ResearchSession
	reports        map[string]*schemas.ResearchReport
	templates      map[string]*ResearchTemplate
	mu             sync.RWMutex

	// Configuration
	projectID string
	region    string
}

// ResearchSession represents an active research session
type ResearchSession struct {
	Config    *schemas.ResearchConfig
	Drones    map[string]*DroneInfo
	Queue     *ResearchQueue
	StartTime time.Time
	Status    string
	Results   []schemas.DroneResult
	Report    *schemas.ResearchReport
}

// DroneInfo contains information about a deployed drone
type DroneInfo struct {
	ID          string
	ServiceURL  string
	Status      string
	StartTime   time.Time
	LastCheckin time.Time
}

// ResearchTemplate represents a pre-orchestrated workflow
type ResearchTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Workflow    map[string]interface{} `json:"workflow"`
}

// DataAnalysis contains analysis results from research data
type DataAnalysis struct {
	Patterns          []schemas.Pattern
	TopInsights       []string
	Statistics        map[string]interface{}
	Duration          time.Duration
	AverageConfidence float64
	Metrics           schemas.ResearchMetrics
}

// NewOrchestrator creates a new orchestrator instance
func NewOrchestrator() (*Orchestrator, error) {
	projectID := getEnvOrDefault("GOOGLE_CLOUD_PROJECT", "")
	region := getEnvOrDefault("GOOGLE_CLOUD_REGION", "us-central1")

	ctx := context.Background()

	// Initialize GCP clients if projectID is available
	var firestoreClient *firestore.Client
	var pubsubClient *pubsub.Client
	var runClient *run.ServicesClient

	if projectID != "" {
		log.Printf("Initializing GCP clients for project %s in region %s", projectID, region)
		
		// Initialize Firestore
		var err error
		firestoreClient, err = firestore.NewClient(ctx, projectID)
		if err != nil {
			log.Printf("Warning: Failed to initialize Firestore client: %v", err)
			firestoreClient = nil
		}

		// Initialize Pub/Sub
		pubsubClient, err = pubsub.NewClient(ctx, projectID)
		if err != nil {
			log.Printf("Warning: Failed to initialize Pub/Sub client: %v", err)
			pubsubClient = nil
		}

		// Initialize Cloud Run
		runClient, err = run.NewServicesClient(ctx)
		if err != nil {
			log.Printf("Warning: Failed to initialize Cloud Run client: %v", err)
			runClient = nil
		}

		log.Printf("GCP clients initialized (Firestore: %v, Pub/Sub: %v, Cloud Run: %v)",
			firestoreClient != nil, pubsubClient != nil, runClient != nil)
	} else {
		log.Println("No GOOGLE_CLOUD_PROJECT set - running in local mode without GCP")
	}

	orch := &Orchestrator{
		firestoreClient: firestoreClient,
		pubsubClient:    pubsubClient,
		runClient:       runClient,
		activeSessions:  make(map[string]*ResearchSession),
		reports:         make(map[string]*schemas.ResearchReport),
		templates:       make(map[string]*ResearchTemplate),
		projectID:       projectID,
		region:          region,
	}

	// Load templates
	orch.loadTemplates()

	return orch, nil
}

// Initialize initializes the orchestrator
func (o *Orchestrator) Initialize(ctx context.Context) error {
	// Create required Pub/Sub topics (only if GCP is configured)
	if o.pubsubClient != nil {
		if err := o.createPubSubTopics(ctx); err != nil {
			log.Printf("Warning: Pub/Sub topics creation failed (GCP features will be limited): %v", err)
		}
	} else {
		log.Println("Pub/Sub client not initialized - running in local mode without GCP")
	}

	return nil
}

// OrchestrateResearch orchestrates the research process
func (o *Orchestrator) OrchestrateResearch(ctx context.Context, config *schemas.ResearchConfig) (*schemas.ResearchResult, error) {
	o.mu.Lock()
	session := &ResearchSession{
		Config:    config,
		Drones:    make(map[string]*DroneInfo),
		Queue:     NewResearchQueue(config.SessionID),
		StartTime: time.Now(),
		Status:    "initializing",
		Results:   make([]schemas.DroneResult, 0),
	}
	o.activeSessions[config.SessionID] = session
	o.mu.Unlock()

	// Update progress file
	if err := o.updateProgressFile(session); err != nil {
		log.Printf("Warning: failed to update progress file for session %s: %v", session.Config.SessionID, err)
	}

	// Start monitoring the session
	go o.monitorSession(ctx, session)

	// Start collecting results from queue
	go o.collectResults(ctx, session)

	// Provision drones
	log.Printf("Provisioning %d research drones for session %s", config.ResearcherCount, config.SessionID)
	if err := o.provisionDrones(ctx, session); err != nil {
		session.Status = "failed"
		return nil, fmt.Errorf("failed to provision drones: %w", err)
	}

	// Start research coordination
	session.Status = "running"
	if err := o.coordinateResearch(ctx, session); err != nil {
		session.Status = "failed"
		return nil, fmt.Errorf("failed to coordinate research: %w", err)
	}

	// Wait for completion
	_, err := o.waitForCompletion(ctx, session)
	if err != nil {
		session.Status = "failed"
		o.updateProgressFile(session)
		return nil, fmt.Errorf("research failed: %w", err)
	}

	// ARCHITECTURAL CHANGE: Report generation removed - happens in host layer now
	session.Status = "completed"
	o.updateProgressFile(session)

	// Clean up resources
	go o.cleanupSession(ctx, session)

	// Return collected drone results (NOT a report - host will aggregate)
	return &schemas.ResearchResult{
		SessionID:   config.SessionID,
		Status:      "completed",
		Results:     session.Results, // Return raw drone results
		Metrics:     o.calculateMetrics(session),
		CompletedAt: time.Now(),
	}, nil
}

// provisionDrones provisions the required number of research drones
func (o *Orchestrator) provisionDrones(ctx context.Context, session *ResearchSession) error {
	// In local mode (no Cloud Run client), create mock drones for testing
	if o.runClient == nil {
		log.Printf("Running in local mode - creating mock drones (no actual deployment)")
		for i := 0; i < session.Config.ResearcherCount; i++ {
			droneID := fmt.Sprintf("mock-drone-%s-%d", session.Config.SessionID, i)
			o.mu.Lock()
			session.Drones[droneID] = &DroneInfo{
				ID:          droneID,
				ServiceURL:  fmt.Sprintf("http://localhost:808%d", i),
				Status:      "mock-deployed",
				StartTime:   time.Now(),
				LastCheckin: time.Now(),
			}
			o.mu.Unlock()
			log.Printf("Created mock drone %s", droneID)
		}
		return nil
	}

	var wg sync.WaitGroup
	errors := make(chan error, session.Config.ResearcherCount)

	for i := 0; i < session.Config.ResearcherCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			droneID := fmt.Sprintf("drone-%s-%d", session.Config.SessionID, index)
			serviceURL, err := o.deployDrone(ctx, droneID, session.Config)
			if err != nil {
				errors <- fmt.Errorf("failed to deploy drone %s: %w", droneID, err)
				return
			}

			o.mu.Lock()
			session.Drones[droneID] = &DroneInfo{
				ID:          droneID,
				ServiceURL:  serviceURL,
				Status:      "deployed",
				StartTime:   time.Now(),
				LastCheckin: time.Now(),
			}
			o.mu.Unlock()

			log.Printf("Successfully deployed drone %s at %s", droneID, serviceURL)
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	var provisionErrors []error
	for err := range errors {
		provisionErrors = append(provisionErrors, err)
	}

	if len(provisionErrors) > 0 {
		return fmt.Errorf("provisioning failed with %d errors: %v", len(provisionErrors), provisionErrors[0])
	}

	return nil
}

// deployDrone deploys a single research drone on Cloud Run
func (o *Orchestrator) deployDrone(ctx context.Context, droneID string, config *schemas.ResearchConfig) (string, error) {
	// Use the drone template image
	image := fmt.Sprintf("gcr.io/%s/research-drone:latest", o.projectID)

	// Create service configuration
	serviceConfig := &runpb.Service{
		Name: droneID,
		Template: &runpb.RevisionTemplate{
			Containers: []*runpb.Container{
				{
					Image: image,
					Env: []*runpb.EnvVar{
						{Name: "DRONE_ID", Values: &runpb.EnvVar_Value{Value: droneID}},
						{Name: "SESSION_ID", Values: &runpb.EnvVar_Value{Value: config.SessionID}},
						{Name: "GOOGLE_CLOUD_PROJECT", Values: &runpb.EnvVar_Value{Value: o.projectID}},
						// The drone will get its instructions via HTTP, but it needs to know which topic to publish results to.
						{Name: "PUBSUB_TOPIC", Values: &runpb.EnvVar_Value{Value: fmt.Sprintf("research-results-%s", config.SessionID)}},
					},
					Resources: &runpb.ResourceRequirements{
						Limits: map[string]string{
							"cpu":    o.getCPUForPriority(config.PriorityLevel),
							"memory": o.getMemoryForPriority(config.PriorityLevel),
						},
					},
				},
			},
			MaxInstanceRequestConcurrency: 1,
			Timeout:                       &durationpb.Duration{Seconds: int64(config.TimeoutMinutes * 60)},
		},
	}

	// Deploy the service
	operation, err := o.runClient.CreateService(ctx, &runpb.CreateServiceRequest{
		Parent:    fmt.Sprintf("projects/%s/locations/%s", o.projectID, o.region),
		ServiceId: droneID,
		Service:   serviceConfig,
	})
	if err != nil {
		return "", err
	}

	// Wait for deployment
	service, err := operation.Wait(ctx)
	if err != nil {
		return "", err
	}

	return service.Uri, nil
}

// coordinateResearch coordinates the research process across drones
func (o *Orchestrator) coordinateResearch(ctx context.Context, session *ResearchSession) error {
	// 1. Send research topic to each drone
	log.Printf("Distributing research topic to drones: %s", session.Config.Topic)

	// 2. Send research task to each drone
	o.mu.RLock()
	drones := make([]*DroneInfo, 0, len(session.Drones))
	for _, drone := range session.Drones {
		drones = append(drones, drone)
	}
	o.mu.RUnlock()

	for _, drone := range drones {
		// Each drone researches the topic - drones will use their own methods/sources
		task := map[string]interface{}{
			"subject": session.Config.Topic,
			"run_id":  session.Config.SessionID,
		}

		if err := o.sendInstructionsToDrone(ctx, drone, task); err != nil {
			log.Printf("Failed to send instructions to drone %s: %v", drone.ID, err)
			drone.Status = "failed_to_instruct"
		} else {
			log.Printf("Successfully sent research task to drone %s", drone.ID)
			drone.Status = "running"
		}
	}

	// Update progress file after dispatching all tasks
	if err := o.updateProgressFile(session); err != nil {
		log.Printf("Warning: failed to update progress file for session %s: %v", session.Config.SessionID, err)
	}

	return nil
}

// waitForCompletion waits for all drones to complete their research
func (o *Orchestrator) waitForCompletion(ctx context.Context, session *ResearchSession) (*schemas.ResearchResult, error) {
	timeout := time.Duration(session.Config.TimeoutMinutes) * time.Minute
	deadline := time.Now().Add(timeout)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// Check completion status (collectResults goroutine is updating session.Results in background)
			o.mu.RLock()
			completedCount := len(session.Results)
			totalCount := session.Config.ResearcherCount
			o.mu.RUnlock()

			if completedCount >= totalCount {
				log.Printf("All %d drones completed for session %s", totalCount, session.Config.SessionID)
				return &schemas.ResearchResult{
					SessionID: session.Config.SessionID,
					Status:    "completed",
				}, nil
			}

			if time.Now().After(deadline) {
				return nil, fmt.Errorf("research timeout after %v", timeout)
			}

			log.Printf("Research progress: %d/%d drones completed", completedCount, totalCount)
		}
	}
}

// Helper methods

func (o *Orchestrator) getCPUForPriority(priority string) string {
	switch priority {
	case "high":
		return "2000m"
	case "low":
		return "500m"
	default:
		return "1000m"
	}
}

func (o *Orchestrator) getMemoryForPriority(priority string) string {
	switch priority {
	case "high":
		return "2Gi"
	case "low":
		return "512Mi"
	default:
		return "1Gi"
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

//lint:ignore U1000 Reserved for future use
func getOrchestratorURL() string {
	return getEnvOrDefault("ORCHESTRATOR_URL", "http://localhost:8080")
}

// GetReports returns all available reports
func (o *Orchestrator) GetReports() []*schemas.ResearchReport {
	o.mu.RLock()
	defer o.mu.RUnlock()

	reports := make([]*schemas.ResearchReport, 0, len(o.reports))
	for _, report := range o.reports {
		reports = append(reports, report)
	}
	return reports
}

// GetTemplates returns all available templates
func (o *Orchestrator) GetTemplates() []*ResearchTemplate {
	o.mu.RLock()
	defer o.mu.RUnlock()

	templates := make([]*ResearchTemplate, 0, len(o.templates))
	for _, template := range o.templates {
		templates = append(templates, template)
	}
	return templates
}

// Shutdown gracefully shuts down the orchestrator
func (o *Orchestrator) Shutdown() {
	log.Println("Shutting down orchestrator...")

	// Close clients
	if o.firestoreClient != nil {
		o.firestoreClient.Close()
	}
	if o.pubsubClient != nil {
		o.pubsubClient.Close()
	}
	if o.runClient != nil {
		o.runClient.Close()
	}
}

