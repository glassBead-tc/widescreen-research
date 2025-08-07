package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	"github.com/google/uuid"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/schemas"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Orchestrator manages the research orchestration process
type Orchestrator struct {
	// GCP clients
	firestoreClient *firestore.Client
	pubsubClient    *pubsub.Client
	runClient       *run.ServicesClient

	// MCP client for connecting to other MCP servers
	mcpClient *MCPClient

	// Claude SDK agent
	claudeAgent *ClaudeAgent

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
	Config      *schemas.ResearchConfig
	Drones      map[string]*DroneInfo
	Queue       *ResearchQueue
	StartTime   time.Time
	Status      string
	Results     []schemas.DroneResult
	Report      *schemas.ResearchReport
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

// NewOrchestrator creates a new orchestrator instance
func NewOrchestrator() (*Orchestrator, error) {
	projectID := getEnvOrDefault("GOOGLE_CLOUD_PROJECT", "")
	if projectID == "" {
		return nil, fmt.Errorf("GOOGLE_CLOUD_PROJECT environment variable is required")
	}

	ctx := context.Background()

	// Initialize Firestore client
	firestoreClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %w", err)
	}

	// Initialize Pub/Sub client
	pubsubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Pub/Sub client: %w", err)
	}

	// Initialize Cloud Run client
	runClient, err := run.NewServicesClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run client: %w", err)
	}

	// Create MCP client
	mcpClient := NewMCPClient()

	// Create Claude agent
	claudeAgent := NewClaudeAgent()

	orch := &Orchestrator{
		firestoreClient: firestoreClient,
		pubsubClient:    pubsubClient,
		runClient:       runClient,
		mcpClient:       mcpClient,
		claudeAgent:     claudeAgent,
		activeSessions:  make(map[string]*ResearchSession),
		reports:         make(map[string]*schemas.ResearchReport),
		templates:       make(map[string]*ResearchTemplate),
		projectID:       projectID,
		region:          getEnvOrDefault("GOOGLE_CLOUD_REGION", "us-central1"),
	}

	// Load templates
	orch.loadTemplates()

	return orch, nil
}

// Initialize initializes the orchestrator
func (o *Orchestrator) Initialize(ctx context.Context) error {
	// Initialize MCP client connections
	if err := o.mcpClient.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	// Initialize Claude agent
	if err := o.claudeAgent.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize Claude agent: %w", err)
	}

	// Create required Pub/Sub topics
	if err := o.createPubSubTopics(ctx); err != nil {
		return fmt.Errorf("failed to create Pub/Sub topics: %w", err)
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

	// Start monitoring the session
	go o.monitorSession(ctx, session)

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
	result, err := o.waitForCompletion(ctx, session)
	if err != nil {
		session.Status = "failed"
		return nil, fmt.Errorf("research failed: %w", err)
	}

	// Generate report
	log.Printf("Generating report for session %s", config.SessionID)
	report, err := o.generateReport(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	session.Report = report
	session.Status = "completed"

	// Store report
	o.mu.Lock()
	o.reports[report.ID] = report
	o.mu.Unlock()

	// Clean up resources
	go o.cleanupSession(ctx, session)

	return &schemas.ResearchResult{
		SessionID:   config.SessionID,
		Status:      "completed",
		ReportURL:   fmt.Sprintf("/reports/%s", report.ID),
		ReportData:  report,
		Metrics:     o.calculateMetrics(session),
		CompletedAt: time.Now(),
	}, nil
}

// provisionDrones provisions the required number of research drones
func (o *Orchestrator) provisionDrones(ctx context.Context, session *ResearchSession) error {
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
						{Name: "DRONE_ID", Value: &runpb.EnvVar_Value{Value: droneID}},
						{Name: "SESSION_ID", Value: &runpb.EnvVar_Value{Value: config.SessionID}},
						{Name: "RESEARCH_TOPIC", Value: &runpb.EnvVar_Value{Value: config.Topic}},
						{Name: "RESEARCH_DEPTH", Value: &runpb.EnvVar_Value{Value: config.ResearchDepth}},
						{Name: "ORCHESTRATOR_URL", Value: &runpb.EnvVar_Value{Value: getOrchestratorURL()}},
						{Name: "PUBSUB_TOPIC", Value: &runpb.EnvVar_Value{Value: fmt.Sprintf("research-results-%s", config.SessionID)}},
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
			Timeout:                      &durationpb.Duration{Seconds: int64(config.TimeoutMinutes * 60)},
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
	// Create research instructions using Claude agent
	instructions, err := o.claudeAgent.GenerateResearchInstructions(ctx, session.Config)
	if err != nil {
		return fmt.Errorf("failed to generate research instructions: %w", err)
	}

	// Parse workflow templates if provided
	var workflow map[string]interface{}
	if session.Config.WorkflowTemplates != "" {
		if err := json.Unmarshal([]byte(session.Config.WorkflowTemplates), &workflow); err != nil {
			log.Printf("Failed to parse workflow templates: %v", err)
		}
	}

	// Send instructions to all drones
	o.mu.RLock()
	drones := make([]*DroneInfo, 0, len(session.Drones))
	for _, drone := range session.Drones {
		drones = append(drones, drone)
	}
	o.mu.RUnlock()

	// Distribute work across drones
	for i, drone := range drones {
		droneInstructions := o.assignDroneWork(instructions, workflow, i, len(drones))
		if err := o.sendInstructionsToDrone(ctx, drone, droneInstructions); err != nil {
			log.Printf("Failed to send instructions to drone %s: %v", drone.ID, err)
		}
	}

	// Start collecting results
	go o.collectResults(ctx, session)

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
			// Check completion status
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

// generateReport generates the final research report
func (o *Orchestrator) generateReport(ctx context.Context, session *ResearchSession) (*schemas.ResearchReport, error) {
	// Analyze collected data
	analysis, err := o.analyzeResults(ctx, session.Results)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze results: %w", err)
	}

	// Generate report using Claude agent
	report, err := o.claudeAgent.GenerateReport(ctx, session.Config, session.Results, analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	report.ID = uuid.New().String()
	report.SessionID = session.Config.SessionID
	report.CreatedAt = time.Now()

	// Store report in Firestore
	if err := o.storeReport(ctx, report); err != nil {
		log.Printf("Failed to store report: %v", err)
	}

	return report, nil
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
	
	// Shutdown MCP client
	o.mcpClient.Shutdown()
	
	// Shutdown Claude agent
	o.claudeAgent.Shutdown()
}