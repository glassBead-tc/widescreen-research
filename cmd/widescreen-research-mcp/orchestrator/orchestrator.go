// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

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
	run "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research-mcp/schemas"
	"github.com/google/uuid"
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
	// Note: projectID can be empty, GCP clients will be lazily initialized when needed

	ctx := context.Background()

	// GCP clients will be lazily initialized when first accessed
	var firestoreClient *firestore.Client
	var pubsubClient *pubsub.Client
	var runClient *run.ServicesClient
	var _ = ctx // Mark as used for future lazy init

	// Create MCP client
	mcpClient := NewMCPClient()

	orch := &Orchestrator{
		firestoreClient: firestoreClient,
		pubsubClient:    pubsubClient,
		runClient:       runClient,
		mcpClient:       mcpClient,
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
		log.Printf("Warning: MCP client initialization failed (will be unavailable): %v", err)
	}

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

	// Subscribe queue to Pub/Sub to collect results (if Pub/Sub is available)
	if o.pubsubClient != nil {
		if err := session.Queue.Subscribe(ctx, o.pubsubClient); err != nil {
			log.Printf("Warning: failed to subscribe queue (results won't be collected): %v", err)
		}
	}

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
			// Sync results from queue to session
			queueResults := session.Queue.GetResults()
			o.mu.Lock()
			session.Results = queueResults
			completedCount := len(session.Results)
			totalCount := session.Config.ResearcherCount
			o.mu.Unlock()

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

	// 3. Generate structured report from drone results
	report := &schemas.ResearchReport{
		ID:          uuid.New().String(),
		SessionID:   session.Config.SessionID,
		Title:       fmt.Sprintf("Research Report: %s", session.Config.Topic),
		Executive:   o.generateExecutiveSummary(session, analysis),
		Sections:    o.generateReportSections(session, analysis),
		Methodology: fmt.Sprintf("Distributed research using %d parallel drones with %s depth.", session.Config.ResearcherCount, session.Config.ResearchDepth),
		Data:        o.aggregateDroneData(session.Results),
		CreatedAt:   time.Now(),
		Metadata: schemas.ReportMetadata{
			ResearchTopic:   session.Config.Topic,
			ResearcherCount: session.Config.ResearcherCount,
			Duration:        analysis.Duration,
			DataPoints:      len(session.Results),
			Sources:         o.extractSources(session.Results),
			Metrics:         analysis.Metrics,
		},
	}

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

	// Shutdown MCP client
	o.mcpClient.Shutdown()
}

// generateExecutiveSummary creates an executive summary from research results
func (o *Orchestrator) generateExecutiveSummary(session *ResearchSession, analysis *DataAnalysis) string {
	successCount := 0
	for _, result := range session.Results {
		if result.Status == "completed" {
			successCount++
		}
	}

	summary := fmt.Sprintf("Research Summary: %s\n\n", session.Config.Topic)
	summary += fmt.Sprintf("Completed using %d research drones over %v.\n\n", session.Config.ResearcherCount, analysis.Duration)
	summary += fmt.Sprintf("Successfully collected data from %d out of %d drones.\n", successCount, len(session.Results))

	if len(analysis.TopInsights) > 0 {
		summary += "\nKey Findings:\n"
		for i, insight := range analysis.TopInsights {
			if i >= 5 {
				break
			}
			summary += fmt.Sprintf("- %s\n", insight)
		}
	}

	return summary
}

// generateReportSections creates report sections from drone results
func (o *Orchestrator) generateReportSections(session *ResearchSession, analysis *DataAnalysis) []schemas.ReportSection {
	sections := []schemas.ReportSection{
		{
			Title:    "Research Findings",
			Content:  fmt.Sprintf("Collected data from %d research drones. Identified %d patterns with average confidence of %.2f.", len(session.Results), len(analysis.Patterns), analysis.AverageConfidence),
			Insights: analysis.TopInsights,
			Data:     analysis.Statistics,
		},
	}

	return sections
}

// aggregateDroneData aggregates data from all drone results
func (o *Orchestrator) aggregateDroneData(results []schemas.DroneResult) map[string]interface{} {
	aggregated := make(map[string]interface{})

	var allData []map[string]interface{}
	for _, result := range results {
		if result.Status == "completed" && result.Data != nil {
			allData = append(allData, result.Data)
		}
	}

	aggregated["drone_results"] = allData
	aggregated["total_drones"] = len(results)
	aggregated["successful_drones"] = len(allData)

	return aggregated
}

// extractSources extracts unique sources from drone results
func (o *Orchestrator) extractSources(results []schemas.DroneResult) []string {
	sourceMap := make(map[string]bool)

	for _, result := range results {
		if sources, ok := result.Data["sources"].([]interface{}); ok {
			for _, source := range sources {
				if s, ok := source.(string); ok {
					sourceMap[s] = true
				}
			}
		}
	}

	sources := make([]string, 0, len(sourceMap))
	for source := range sourceMap {
		sources = append(sources, source)
	}

	return sources
}
