package coordinator

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/spawn-mcp/coordinator/pkg/gcp"
	"github.com/spawn-mcp/coordinator/pkg/types"
)

// Server represents the coordinator MCP server
type Server struct {
	gcpClient    *gcp.Client
	mcpClient    *MCPClient
	activeDrones map[string]*types.DroneInfo
	dronesMutex  sync.RWMutex
	taskResults  map[string][]*types.TaskResult
	resultsMutex sync.RWMutex
}

// NewServer creates a new coordinator MCP server
func NewServer(gcpClient *gcp.Client) *Server {
	server := &Server{
		gcpClient:    gcpClient,
		mcpClient:    NewMCPClient(gcpClient.ProjectID),
		activeDrones: make(map[string]*types.DroneInfo),
		taskResults:  make(map[string][]*types.TaskResult),
	}

	return server
}

// PlanDistributedTask creates an execution plan for a distributed task
func (s *Server) PlanDistributedTask(taskDescription string, parameters map[string]interface{}, timeConstraint int, droneType string) (*types.ExecutionPlan, error) {
	log.Printf("Planning distributed task: %s", taskDescription)

	// Analyze task requirements
	droneCount := s.calculateDroneRequirements(taskDescription, parameters)
	if droneCount > 100 {
		droneCount = 100 // Enforce maximum limit
	}

	// Create execution plan
	plan := &types.ExecutionPlan{
		ID: fmt.Sprintf("plan-%d", time.Now().Unix()),
		TaskDefinition: types.TaskDefinition{
			ID:             fmt.Sprintf("task-%d", time.Now().Unix()),
			Type:           droneType,
			Description:    taskDescription,
			Parameters:     parameters,
			RequiredDrones: droneCount,
			DroneType:      types.DroneType(droneType),
			TimeoutMinutes: timeConstraint,
			CheckpointConfig: types.CheckpointConfig{
				Enabled:         true,
				IntervalSeconds: 30,
				MaxRetries:      3,
			},
		},
		DroneCount:    droneCount,
		EstimatedCost: s.estimateTaskCost(droneCount, timeConstraint),
		EstimatedTime: time.Duration(timeConstraint) * time.Minute,
		Strategy:      "parallel-execution",
	}

	// Store plan in Firestore
	ctx := context.Background()
	err := s.gcpClient.StoreDocument(ctx, "execution_plans", plan.ID, plan)
	if err != nil {
		return nil, fmt.Errorf("failed to store execution plan: %w", err)
	}

	return plan, nil
}

// Helper methods
func (s *Server) calculateDroneRequirements(description string, params map[string]interface{}) int {
	// Simple heuristic - in practice, this would be more sophisticated
	baseCount := 5

	// Adjust based on description keywords
	if contains(description, []string{"research", "analyze", "investigate"}) {
		baseCount = 10
	}
	if contains(description, []string{"comprehensive", "detailed", "thorough"}) {
		baseCount *= 2
	}

	return baseCount
}

func (s *Server) getDroneImageURI(droneType types.DroneType) string {
	// TODO: Make these configurable via environment variables or config file
	baseRegistry := "gcr.io/" + s.gcpClient.ProjectID + "/spawn-mcp"

	switch droneType {
	case types.DroneTypeWorker:
		return baseRegistry + "/drone-worker:latest"
	case types.DroneTypeAnalyzer:
		return baseRegistry + "/drone-analyzer:latest"
	case types.DroneTypeProcessor:
		return baseRegistry + "/drone-processor:latest"
	case types.DroneTypeResearcher:
		return baseRegistry + "/drone-researcher:latest"
	case types.DroneTypeSynthesizer:
		return baseRegistry + "/drone-synthesizer:latest"
	default:
		// Default to worker type
		return baseRegistry + "/drone-worker:latest"
	}
}

func (s *Server) estimateTaskCost(droneCount, durationMinutes int) float64 {
	// Cloud Run pricing: $0.00002400/vCPU-second, $0.0000025/GiB-second
	cpuCostPerSecond := 0.00002400
	memoryCostPerSecond := 0.0000025 * 0.5 // 512Mi = 0.5 GiB

	durationSeconds := float64(durationMinutes * 60)
	totalCost := float64(droneCount) * durationSeconds * (cpuCostPerSecond + memoryCostPerSecond)

	// Add overhead for other services (Firestore, Pub/Sub, etc.)
	totalCost *= 1.25

	return totalCost
}

func contains(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(text) > 0 && len(keyword) > 0 {
			// Simple substring check - in practice, use proper text analysis
			for i := 0; i <= len(text)-len(keyword); i++ {
				if text[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}

// SpawnDrone spawns a new drone with the given configuration
func (s *Server) SpawnDrone(ctx context.Context, config types.DroneConfig) (string, error) {
	s.dronesMutex.Lock()
	defer s.dronesMutex.Unlock()

	droneID := fmt.Sprintf("drone-%s-%d", config.Type, time.Now().Unix())
	serviceName := fmt.Sprintf("drone-%s-%d", config.Type, time.Now().Unix())

	// Create drone info
	drone := &types.DroneInfo{
		ID:             droneID,
		ServiceName:    serviceName,
		Type:           string(config.Type),
		Status:         "spawning",
		Region:         config.Region,
		CreatedAt:      time.Now(),
		LastSeen:       time.Now(),
		TasksCompleted: 0,
		Capabilities:   config.Capabilities,
		Metadata:       make(map[string]interface{}),
	}

	// Store in active drones
	s.activeDrones[droneID] = drone

	// Prepare environment variables for the drone
	env := make(map[string]string)
	env["DRONE_ID"] = droneID
	env["DRONE_TYPE"] = string(config.Type)
	env["COORDINATOR_URL"] = "https://coordinator-service-url" // TODO: Make this configurable

	// Add any custom environment variables from config
	for key, value := range config.Environment {
		env[key] = value
	}

	// Determine the container image based on drone type
	imageURI := s.getDroneImageURI(config.Type)

	log.Printf("Creating Cloud Run service for drone %s (service: %s)", droneID, serviceName)

	// Create the Cloud Run service
	service, err := s.gcpClient.CreateCloudRunService(ctx, serviceName, imageURI, env)
	if err != nil {
		// Remove from active drones on failure
		delete(s.activeDrones, droneID)
		return "", fmt.Errorf("failed to create Cloud Run service for drone %s: %w", droneID, err)
	}

	// Wait for the service to be ready
	err = s.gcpClient.WaitForServiceReady(ctx, serviceName, 5*time.Minute)
	if err != nil {
		log.Printf("Warning: Service %s may not be fully ready: %v", serviceName, err)
		// Don't fail completely, just log the warning
	}

	// Get the service URL
	serviceURL, err := s.gcpClient.GetServiceURL(ctx, serviceName)
	if err != nil {
		log.Printf("Warning: Could not get service URL for %s: %v", serviceName, err)
		serviceURL = "" // Will be populated later when available
	}

	// Update drone info with service details
	drone.ServiceURL = serviceURL
	drone.Status = "active"
	drone.LastPing = time.Now()
	drone.Metadata["cloud_run_service"] = service.Name
	drone.Metadata["service_uri"] = service.Uri

	// Store drone info in Firestore for persistence
	err = s.gcpClient.StoreDocument(ctx, "drones", droneID, drone)
	if err != nil {
		log.Printf("Warning: Failed to store drone info in Firestore: %v", err)
		// Don't fail the spawn operation for this
	}

	log.Printf("Successfully spawned drone %s of type %s at %s", droneID, config.Type, serviceURL)

	return droneID, nil
}

// ListActiveDrones returns a list of all active drones
func (s *Server) ListActiveDrones() []*types.DroneInfo {
	s.dronesMutex.RLock()
	defer s.dronesMutex.RUnlock()

	drones := make([]*types.DroneInfo, 0, len(s.activeDrones))
	for _, drone := range s.activeDrones {
		drones = append(drones, drone)
	}

	return drones
}

// ExecuteTask executes a task across the drone fleet
func (s *Server) ExecuteTask(ctx context.Context, task types.Task) (string, error) {
	taskID := fmt.Sprintf("task-%s-%d", task.Type, time.Now().Unix())

	log.Printf("Executing task %s: %s", taskID, task.Description)

	// Find available drones of the required type
	s.dronesMutex.RLock()
	var availableDrones []*types.DroneInfo
	for _, drone := range s.activeDrones {
		if drone.Type == task.Type && drone.Status == "active" && drone.ServiceURL != "" {
			availableDrones = append(availableDrones, drone)
		}
	}
	s.dronesMutex.RUnlock()

	if len(availableDrones) == 0 {
		return "", fmt.Errorf("no available drones of type %s", task.Type)
	}

	// Limit to maxDrones if specified
	if task.MaxDrones > 0 && len(availableDrones) > task.MaxDrones {
		availableDrones = availableDrones[:task.MaxDrones]
	}

	log.Printf("Distributing task %s to %d drones", taskID, len(availableDrones))

	// Execute task on each drone (for now, just list their tools)
	var results []*types.TaskResult
	for _, drone := range availableDrones {
		result := &types.TaskResult{
			TaskID:    taskID,
			DroneID:   drone.ID,
			Status:    "executing",
			Timestamp: time.Now(),
		}

		// Call the drone to list its tools (as a test)
		response, err := s.mcpClient.ListTools(ctx, drone.ServiceURL)
		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			log.Printf("Failed to call drone %s: %v", drone.ID, err)
		} else {
			result.Status = "completed"
			result.Data = response.Result
			log.Printf("Successfully called drone %s", drone.ID)
		}

		results = append(results, result)
	}

	// Store results
	s.resultsMutex.Lock()
	s.taskResults[taskID] = results
	s.resultsMutex.Unlock()

	return taskID, nil
}

// ExecuteResearchTask executes a specific research task using Exa tools on research drones
func (s *Server) ExecuteResearchTask(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	taskID := fmt.Sprintf("research-task-%d", time.Now().Unix())

	log.Printf("Executing research task %s with tool %s", taskID, toolName)

	// Find available research drones
	s.dronesMutex.RLock()
	var researchDrones []*types.DroneInfo
	for _, drone := range s.activeDrones {
		if drone.Type == "research" && drone.Status == "active" && drone.ServiceURL != "" {
			researchDrones = append(researchDrones, drone)
		}
	}
	s.dronesMutex.RUnlock()

	if len(researchDrones) == 0 {
		return "", fmt.Errorf("no available research drones")
	}

	// Use the first available research drone
	drone := researchDrones[0]
	log.Printf("Using research drone %s for task %s", drone.ID, taskID)

	// Execute the research tool
	response, err := s.mcpClient.CallTool(ctx, drone.ServiceURL, toolName, arguments)
	if err != nil {
		return "", fmt.Errorf("failed to execute research tool %s on drone %s: %w", toolName, drone.ID, err)
	}

	// Store the result
	result := &types.TaskResult{
		TaskID:    taskID,
		DroneID:   drone.ID,
		Status:    "completed",
		Data:      response.Result,
		Timestamp: time.Now(),
	}

	if response.Error != nil {
		result.Status = "failed"
		result.Error = response.Error.Message
	}

	s.resultsMutex.Lock()
	s.taskResults[taskID] = []*types.TaskResult{result}
	s.resultsMutex.Unlock()

	log.Printf("Research task %s completed with status: %s", taskID, result.Status)

	return taskID, nil
}

// GetTaskResults returns the results for a specific task
func (s *Server) GetTaskResults(taskID string) ([]*types.TaskResult, error) {
	s.resultsMutex.RLock()
	defer s.resultsMutex.RUnlock()

	results, exists := s.taskResults[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	return results, nil
}

// GetDroneStatus returns the status of a specific drone
func (s *Server) GetDroneStatus(ctx context.Context, droneID string) (*types.DroneInfo, error) {
	s.dronesMutex.RLock()
	defer s.dronesMutex.RUnlock()

	drone, exists := s.activeDrones[droneID]
	if !exists {
		return nil, fmt.Errorf("drone %s not found", droneID)
	}

	return drone, nil
}

// CheckDroneHealth checks the health of a specific drone and updates its status
func (s *Server) CheckDroneHealth(ctx context.Context, droneID string) error {
	s.dronesMutex.Lock()
	defer s.dronesMutex.Unlock()

	drone, exists := s.activeDrones[droneID]
	if !exists {
		return fmt.Errorf("drone %s not found", droneID)
	}

	// If drone has a service URL, perform actual health check
	if drone.ServiceURL != "" {
		err := s.mcpClient.HealthCheck(ctx, drone.ServiceURL)
		if err != nil {
			log.Printf("Health check failed for drone %s: %v", droneID, err)
			drone.Status = "unhealthy"
		} else {
			drone.Status = "active"
			drone.LastPing = time.Now()
		}

		// Update in Firestore
		err = s.gcpClient.StoreDocument(ctx, "drones", droneID, drone)
		if err != nil {
			log.Printf("Warning: Failed to update drone health in Firestore: %v", err)
		}
	}

	return nil
}

// CheckAllDroneHealth checks the health of all active drones
func (s *Server) CheckAllDroneHealth(ctx context.Context) {
	s.dronesMutex.RLock()
	droneIDs := make([]string, 0, len(s.activeDrones))
	for droneID := range s.activeDrones {
		droneIDs = append(droneIDs, droneID)
	}
	s.dronesMutex.RUnlock()

	for _, droneID := range droneIDs {
		if err := s.CheckDroneHealth(ctx, droneID); err != nil {
			log.Printf("Health check failed for drone %s: %v", droneID, err)
		}
	}
}

// StartHealthCheckRoutine starts a background routine to periodically check drone health
func (s *Server) StartHealthCheckRoutine(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.CheckAllDroneHealth(ctx)
			}
		}
	}()
}

// ScaleDrones scales the number of drones of a specific type
func (s *Server) ScaleDrones(ctx context.Context, droneType types.DroneType, targetCount int) error {
	s.dronesMutex.RLock()
	currentCount := 0
	for _, drone := range s.activeDrones {
		if drone.Type == string(droneType) && drone.Status == "active" {
			currentCount++
		}
	}
	s.dronesMutex.RUnlock()

	if currentCount == targetCount {
		log.Printf("Drone count for type %s already at target: %d", droneType, targetCount)
		return nil
	}

	if currentCount < targetCount {
		// Scale up
		needed := targetCount - currentCount
		log.Printf("Scaling up %s drones: need %d more", droneType, needed)

		for i := 0; i < needed; i++ {
			config := types.DroneConfig{
				Type:         droneType,
				Region:       s.gcpClient.Region,
				Capabilities: s.getDefaultCapabilities(droneType),
				Environment:  make(map[string]string),
			}

			_, err := s.SpawnDrone(ctx, config)
			if err != nil {
				log.Printf("Failed to spawn drone %d of %d: %v", i+1, needed, err)
				// Continue trying to spawn the rest
			}
		}
	} else {
		// Scale down
		excess := currentCount - targetCount
		log.Printf("Scaling down %s drones: need to remove %d", droneType, excess)

		// Find drones to terminate (prefer least recently used)
		s.dronesMutex.RLock()
		var dronesOfType []*types.DroneInfo
		for _, drone := range s.activeDrones {
			if drone.Type == string(droneType) && drone.Status == "active" {
				dronesOfType = append(dronesOfType, drone)
			}
		}
		s.dronesMutex.RUnlock()

		// Sort by last seen (oldest first)
		// TODO: Implement proper sorting

		// Terminate excess drones
		for i := 0; i < excess && i < len(dronesOfType); i++ {
			err := s.TerminateDrone(ctx, dronesOfType[i].ID)
			if err != nil {
				log.Printf("Failed to terminate drone %s: %v", dronesOfType[i].ID, err)
			}
		}
	}

	return nil
}

// getDefaultCapabilities returns default capabilities for a drone type
func (s *Server) getDefaultCapabilities(droneType types.DroneType) []string {
	switch droneType {
	case types.DroneTypeWorker:
		return []string{"basic-processing", "file-handling"}
	case types.DroneTypeAnalyzer:
		return []string{"data-analysis", "pattern-recognition", "statistical-processing"}
	case types.DroneTypeProcessor:
		return []string{"data-transformation", "batch-processing", "stream-processing"}
	case types.DroneTypeResearcher:
		return []string{"web-search", "document-analysis", "information-extraction"}
	case types.DroneTypeSynthesizer:
		return []string{"content-generation", "summarization", "synthesis"}
	default:
		return []string{"basic-processing"}
	}
}

// TerminateDrone terminates a specific drone
func (s *Server) TerminateDrone(ctx context.Context, droneID string) error {
	s.dronesMutex.Lock()
	defer s.dronesMutex.Unlock()

	drone, exists := s.activeDrones[droneID]
	if !exists {
		return fmt.Errorf("drone %s not found", droneID)
	}

	log.Printf("Terminating drone %s (service: %s)", droneID, drone.ServiceName)

	// Update status to terminating
	drone.Status = "terminating"

	// Delete the Cloud Run service
	if drone.ServiceName != "" {
		err := s.gcpClient.DeleteCloudRunService(ctx, drone.ServiceName)
		if err != nil {
			log.Printf("Warning: Failed to delete Cloud Run service %s: %v", drone.ServiceName, err)
			// Continue with cleanup even if service deletion fails
		}
	}

	// Remove from active drones
	delete(s.activeDrones, droneID)

	// Update status in Firestore (mark as terminated rather than delete)
	drone.Status = "terminated"
	drone.LastSeen = time.Now()
	err := s.gcpClient.StoreDocument(ctx, "drones_history", droneID, drone)
	if err != nil {
		log.Printf("Warning: Failed to store terminated drone info: %v", err)
	}

	log.Printf("Successfully terminated drone %s", droneID)

	return nil
}

// Serve starts the coordinator server
func (s *Server) Serve() error {
	log.Println("Starting Coordinator Server...")
	// For now, just keep running
	select {}
}
