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

	// Update progress file
	if err := o.updateProgressFile(session); err != nil {
		log.Printf("Warning: failed to update progress file for session %s: %v", session.Config.SessionID, err)
	}

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
	_, err := o.waitForCompletion(ctx, session)
	if err != nil {
		session.Status = "failed"
		o.updateProgressFile(session)
		return nil, fmt.Errorf("research failed: %w", err)
	}

	// Generate report
	log.Printf("Generating report for session %s", config.SessionID)
	report, err := o.generateReport(ctx, session)
	if err != nil {
		session.Status = "failed_report_generation"
		o.updateProgressFile(session)
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	session.Report = report
	session.Status = "completed"
	o.updateProgressFile(session)

	// Store report
	o.mu.Lock()
	o.reports[report.ID] = report
	o.mu.Unlock()

	// Clean up resources
	go o.cleanupSession(ctx, session)

	reportFilePath := fmt.Sprintf("reports/report_%s.md", session.Config.SessionID)

	return &schemas.ResearchResult{
		SessionID:   config.SessionID,
		Status:      "completed",
		ReportURL:   reportFilePath,
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
	// 1. Break down the high-level topic into specific sub-queries.
	log.Printf("Breaking down research topic: %s", session.Config.Topic)
	subQueries, err := o.claudeAgent.GenerateSubQueries(ctx, session.Config.Topic, session.Config.ResearcherCount)
	if err != nil {
		return fmt.Errorf("failed to generate sub-queries: %w", err)
	}
	log.Printf("Generated %d sub-queries for topic '%s'", len(subQueries), session.Config.Topic)

	// TODO: For now, we assume the number of drones matches the number of sub-queries.
	// A more robust implementation would use a queue to distribute subQueries to available drones.
	if len(subQueries) != len(session.Drones) {
		log.Printf("Warning: The number of sub-queries (%d) does not match the number of drones (%d). Adjusting drone count for this session.", len(subQueries), len(session.Drones))
		// This would be a place to dynamically adjust drone count if the architecture supported it.
		// For now, we'll just truncate the query list to match the drone count.
		if len(subQueries) > len(session.Drones) {
			subQueries = subQueries[:len(session.Drones)]
		}
	}

	// 2. Send a unique instruction to each drone.
	o.mu.RLock()
	drones := make([]*DroneInfo, 0, len(session.Drones))
	for _, drone := range session.Drones {
		drones = append(drones, drone)
	}
	o.mu.RUnlock()

	for i, drone := range drones {
		if i >= len(subQueries) {
			break // Don't send instructions if we have more drones than tasks.
		}

		// The drone needs to know its task ID (which can be the drone ID for simplicity)
		// and the query. The other info is passed via env vars.
		task := map[string]interface{}{
			"subject": subQueries[i],
			"run_id": session.Config.SessionID,
		}

		if err := o.sendInstructionsToDrone(ctx, drone, task); err != nil {
			log.Printf("Failed to send instructions to drone %s: %v", drone.ID, err)
			drone.Status = "failed_to_instruct"
		} else {
			log.Printf("Successfully sent task '%s' to drone %s", subQueries[i], drone.ID)
			drone.Status = "running"
		}
	}

	// Update progress file after dispatching all tasks
	if err := o.updateProgressFile(session); err != nil {
		log.Printf("Warning: failed to update progress file for session %s: %v", session.Config.SessionID, err)
	}

	// 3. Start collecting results from Pub/Sub.
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
	// 1. Save individual drone results
	resultFileDir := fmt.Sprintf("reports/results_%s", session.Config.SessionID)
	if err := os.MkdirAll(resultFileDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create results directory: %w", err)
	}

	var resultFilePaths []string
	for _, result := range session.Results {
		resultFilePath := fmt.Sprintf("%s/drone_%s.json", resultFileDir, result.DroneID)
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Printf("Warning: failed to marshal result for drone %s: %v", result.DroneID, err)
			continue
		}
		if err := os.WriteFile(resultFilePath, jsonData, 0644); err != nil {
			log.Printf("Warning: failed to save result for drone %s: %v", result.DroneID, err)
			continue
		}
		resultFilePaths = append(resultFilePaths, resultFilePath)
	}


	// 2. Analyze collected data
	analysis, err := o.analyzeResults(ctx, session.Results)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze results: %w", err)
	}

	// 3. Generate structured report using Claude agent
	report, err := o.claudeAgent.GenerateReport(ctx, session.Config, session.Results, analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	report.ID = uuid.New().String()
	report.SessionID = session.Config.SessionID
	report.CreatedAt = time.Now()

	// 4. Render the structured report to a user-facing Markdown file
	markdownContent, err := o.renderReportToMarkdown(report, resultFilePaths)
	if err != nil {
		return nil, fmt.Errorf("failed to render markdown report: %w", err)
	}
	reportFilePath := fmt.Sprintf("reports/report_%s.md", session.Config.SessionID)
	if err := os.WriteFile(reportFilePath, []byte(markdownContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to save markdown report: %w", err)
	}
	log.Printf("Final report saved to %s", reportFilePath)


	// 5. Store structured report in Firestore
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

// GetMCPClient returns the MCP client for direct access
func (o *Orchestrator) GetMCPClient() *MCPClient {
	return o.mcpClient
}

// RunWebsetsPipeline orchestrates the websets research pipeline
func (o *Orchestrator) RunWebsetsPipeline(ctx context.Context, topic string, resultCount int) (*schemas.ResearchResult, error) {
	log.Printf("Starting websets pipeline for topic: %s", topic)
	
	// Create websets operations handler
	websetsOps := NewWebsetsOperations(o.mcpClient.websetsClient)
	
	// Step 1: Create webset
	websetID, err := websetsOps.CreateWebset(ctx, topic, resultCount)
	if err != nil {
		return nil, fmt.Errorf("failed to create webset: %w", err)
	}
	log.Printf("Created webset with ID: %s", websetID)
	
	// Step 2: Wait for completion (15 minutes timeout)
	if err := websetsOps.WaitForWebsetCompletion(ctx, websetID, 15*time.Minute); err != nil {
		return nil, fmt.Errorf("webset processing failed: %w", err)
	}
	log.Printf("Webset %s completed successfully", websetID)
	
	// Step 3: Retrieve content items
	items, err := websetsOps.ListContentItems(ctx, websetID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to list content items: %w", err)
	}
	log.Printf("Retrieved %d content items from webset", len(items))
	
	// Step 4: Publish items to Pub/Sub for further processing
	if len(items) > 0 {
		topicName := fmt.Sprintf("websets-%s", websetID)
		topic := o.pubsubClient.Topic(topicName)
		
		// Ensure topic exists
		exists, err := topic.Exists(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to check topic existence: %w", err)
		}
		if !exists {
			topic, err = o.pubsubClient.CreateTopic(ctx, topicName)
			if err != nil {
				return nil, fmt.Errorf("failed to create topic: %w", err)
			}
		}
		
		// Publish items
		for _, item := range items {
			data, err := json.Marshal(item)
			if err != nil {
				log.Printf("Failed to marshal item: %v", err)
				continue
			}
			
			result := topic.Publish(ctx, &pubsub.Message{
				Data: data,
				Attributes: map[string]string{
					"webset_id": websetID,
					"type":      "content_item",
				},
			})
			
			// Wait for publish confirmation
			if _, err := result.Get(ctx); err != nil {
				log.Printf("Failed to publish item: %v", err)
			}
		}
		
		log.Printf("Published %d items to Pub/Sub topic %s", len(items), topicName)
	}
	
	// Step 5: Generate research result
	result := &schemas.ResearchResult{
		SessionID:   websetID,
		Status:      "completed",
		ReportURL:   fmt.Sprintf("/websets/%s", websetID),
		ReportData: map[string]interface{}{
			"webset_id":  websetID,
			"topic":      topic,
			"item_count": len(items),
			"items":      items,
			"summary":    fmt.Sprintf("Websets research completed for topic '%s'. Retrieved %d items.", topic, len(items)),
		},
		Metrics: schemas.ResearchMetrics{
			DronesProvisioned:   1, // websets acts as a single "drone"
			DronesCompleted:     1,
			DronesFailed:        0,
			TotalDuration:       15 * time.Minute, // approximate
			DataPointsCollected: len(items),
			CostEstimate:        0.0, // can be calculated based on EXA pricing
		},
		CompletedAt: time.Now(),
	}
	
	return result, nil
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